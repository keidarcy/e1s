#######################
# ECS Managed Daemons
#######################

locals {
  daemon_name = "e1s-daemon"
}

resource "aws_ecs_cluster" "daemon" {
  name = "${local.daemon_name}-cluster"
}

resource "aws_iam_role" "ecs_infrastructure" {
  name = "${local.daemon_name}-infra-role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ecs.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "ecs_infrastructure" {
  role       = aws_iam_role.ecs_infrastructure.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonECSInfrastructureRolePolicyForManagedInstances"
}

resource "aws_iam_role_policy" "ecs_infrastructure_pass_role" {
  name = "pass-role"
  role = aws_iam_role.ecs_infrastructure.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = "iam:PassRole"
      Resource = aws_iam_role.daemon_instance.arn
    }]
  })
}

resource "aws_iam_role" "daemon_instance" {
  name = "${local.daemon_name}-instance-role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "daemon_instance_ecs" {
  role       = aws_iam_role.daemon_instance.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonECSInstanceRolePolicyForManagedInstances"
}

resource "aws_iam_role_policy_attachment" "daemon_instance_ssm" {
  role       = aws_iam_role.daemon_instance.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

resource "aws_iam_instance_profile" "daemon" {
  name = "${local.daemon_name}-instance-profile"
  role = aws_iam_role.daemon_instance.name
}

resource "aws_ecs_capacity_provider" "daemon" {
  name    = "${local.daemon_name}-cp"
  cluster = aws_ecs_cluster.daemon.name

  managed_instances_provider {
    infrastructure_role_arn = aws_iam_role.ecs_infrastructure.arn

    instance_launch_template {
      ec2_instance_profile_arn = aws_iam_instance_profile.daemon.arn
      monitoring               = "BASIC"

      network_configuration {
        subnets         = aws_subnet.private[*].id
        security_groups = [aws_security_group.ecs.id]
      }
    }
  }

  # Managed Instances need outbound internet during deprovisioning to allow the
  # ECS agent to drain tasks and deregister. On destroy, Terraform must keep the
  # full network path (IGW -> public route -> NAT -> private route -> association)
  # and IAM roles alive until the capacity provider is fully deleted.
  depends_on = [
    aws_iam_role_policy_attachment.ecs_infrastructure,
    aws_iam_role_policy.ecs_infrastructure_pass_role,
    aws_iam_role_policy_attachment.daemon_instance_ecs,
    aws_iam_role_policy_attachment.daemon_instance_ssm,
    aws_internet_gateway.main,
    aws_route.public,
    aws_route_table_association.public_1a,
    aws_nat_gateway.nat,
    aws_route.private_1a,
    aws_route_table_association.private,
  ]
}

resource "aws_ecs_cluster_capacity_providers" "daemon" {
  cluster_name       = aws_ecs_cluster.daemon.name
  capacity_providers = [aws_ecs_capacity_provider.daemon.name]
}

resource "aws_ecs_daemon_task_definition" "main" {
  family             = "${local.daemon_name}-monitoring"
  cpu                = "256"
  memory             = "512"
  task_role_arn      = aws_iam_role.task_role.arn
  execution_role_arn = aws_iam_role.task_role.arn

  container_definition {
    name               = "monitoring-agent"
    image              = "public.ecr.aws/docker/library/busybox:latest"
    essential          = true
    command            = ["sh", "-c", "while true; do echo 'Daemon running'; sleep 30; done"]
    memory_reservation = 128

    log_configuration {
      log_driver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.daemon.name
        "awslogs-region"        = "us-east-1"
        "awslogs-stream-prefix" = "daemon"
      }
    }
  }
}

resource "aws_ecs_daemon" "main" {
  name                       = local.daemon_name
  cluster_arn                = aws_ecs_cluster.daemon.arn
  daemon_task_definition_arn = aws_ecs_daemon_task_definition.main.arn
  enable_execute_command     = true

  capacity_provider_arns = [aws_ecs_capacity_provider.daemon.arn]

  depends_on = [
    aws_ecs_cluster_capacity_providers.daemon,
    aws_internet_gateway.main,
    aws_route.public,
    aws_route_table_association.public_1a,
    aws_nat_gateway.nat,
    aws_route.private_1a,
    aws_route_table_association.private,
  ]
}

#######################
# Nginx service on Managed Instances
#######################
resource "aws_ecs_task_definition" "daemon_nginx" {
  family                   = "${local.daemon_name}-nginx"
  task_role_arn            = aws_iam_role.task_role.arn
  execution_role_arn       = aws_iam_role.task_role.arn
  requires_compatibilities = ["EC2"]
  network_mode             = "awsvpc"
  cpu                      = "256"
  memory                   = "512"

  container_definitions = jsonencode([{
    name      = "nginx"
    image     = "nginx:alpine"
    essential = true
    portMappings = [{ containerPort = 80 }]
    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.daemon.name
        "awslogs-region"        = "us-east-1"
        "awslogs-stream-prefix" = "nginx"
      }
    }
  }])
}

resource "aws_ecs_service" "daemon_nginx" {
  name                   = "${local.daemon_name}-nginx"
  cluster                = aws_ecs_cluster.daemon.id
  task_definition        = aws_ecs_task_definition.daemon_nginx.arn
  desired_count          = 1
  enable_execute_command = true

  capacity_provider_strategy {
    capacity_provider = aws_ecs_capacity_provider.daemon.name
    weight            = 1
    base              = 1
  }

  network_configuration {
    subnets         = aws_subnet.private[*].id
    security_groups = [aws_security_group.ecs.id]
  }

  depends_on = [
    aws_ecs_cluster_capacity_providers.daemon,
    aws_internet_gateway.main,
    aws_route.public,
    aws_route_table_association.public_1a,
    aws_nat_gateway.nat,
    aws_route.private_1a,
    aws_route_table_association.private,
  ]
}
resource "aws_cloudwatch_log_group" "daemon" {
  name = "/aws/ecs/${local.daemon_name}"
}
