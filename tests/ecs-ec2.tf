###############
# EC2 Cluster
###############

resource "aws_ecs_cluster" "ec2" {
  name = "${local.name}-ec2-cluster"
  setting {
    name  = "containerInsights"
    value = "disabled"
  }
}

# EC2 iam role
resource "aws_iam_role" "asg-ec2" {
  name = "${local.name}-ec2-iam-role"
  assume_role_policy = jsonencode(
    {
      Version = "2012-10-17"
      Statement = [
        {
          Effect = "Allow"
          Action = ["sts:AssumeRole"]
          Principal = {
            Service = ["ec2.amazonaws.com"]
          }
        }
      ]
    }
  )
  managed_policy_arns = ["arn:aws:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"]
}


resource "aws_iam_instance_profile" "asg-ec2" {
  name = "ecs-instance-profile"
  role = aws_iam_role.asg-ec2.name
}

resource "aws_launch_configuration" "ec2" {
  name_prefix          = "${local.name}-config"
  image_id             = "ami-0759f51a90924c166"
  instance_type        = "t2.micro"
  iam_instance_profile = aws_iam_instance_profile.asg-ec2.name

  user_data = <<-EOT
        #!/bin/bash

        cat <<'EOF' >> /etc/ecs/ecs.config
        ECS_CLUSTER="${local.name}-ec2-cluster"
        ECS_LOGLEVEL=debug
        ECS_CONTAINER_INSTANCE_TAGS=${jsonencode({ Name = "${local.name}-ec2-cluster" })}
        ECS_ENABLE_TASK_IAM_ROLE=true
        ECS_ENABLE_SPOT_INSTANCE_DRAINING=true
        EOF
      EOT

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_autoscaling_group" "ec2" {
  desired_capacity     = 1
  max_size             = 2
  min_size             = 0
  launch_configuration = aws_launch_configuration.ec2.id
  vpc_zone_identifier  = aws_subnet.private[*].id

  lifecycle {
    create_before_destroy = true
  }
  tag {
    key                 = "Name"
    value               = "my-asg-instance"
    propagate_at_launch = true
  }
}

resource "aws_ecs_capacity_provider" "ec2" {
  name = "${local.name}-ec2-capacity"
  auto_scaling_group_provider {
    auto_scaling_group_arn         = aws_autoscaling_group.ec2.arn
    managed_termination_protection = "DISABLED"
    managed_scaling {
      maximum_scaling_step_size = 3
      minimum_scaling_step_size = 1
      status                    = "ENABLED"
      target_capacity           = 1
    }
  }
}

resource "aws_ecs_cluster_capacity_providers" "ec2" {
  cluster_name       = aws_ecs_cluster.ec2.name
  capacity_providers = [aws_ecs_capacity_provider.ec2.name]
  depends_on         = [aws_ecs_capacity_provider.ec2]
}

# ECS on EC2 task role
resource "aws_iam_role" "ec2" {
  name = "${local.name}-ec2-task-role"
  assume_role_policy = jsonencode(
    {
      Version = "2012-10-17"
      Statement = [
        {
          Effect = "Allow"
          Action = ["sts:AssumeRole"]
          Principal = {
            Service = ["ecs-tasks.amazonaws.com"]
          }
        }
      ]
    }
  )
  inline_policy {
    name = "${local.name}-custom-policy"
    policy = jsonencode(
      {
        Statement = concat([
          {
            Action = [
              "ec2:AuthorizeSecurityGroupIngress",
              "ec2:Describe*",
              "elasticloadbalancing:DeregisterInstancesFromLoadBalancer",
              "elasticloadbalancing:DeregisterTargets",
              "elasticloadbalancing:Describe*",
              "elasticloadbalancing:RegisterInstancesWithLoadBalancer",
              "elasticloadbalancing:RegisterTargets",
              "ec2:DescribeTags",
              "logs:CreateLogGroup",
              "logs:CreateLogStream",
              "logs:DescribeLogStreams",
              "logs:PutSubscriptionFilter",
              "logs:PutLogEvents"
            ]
            Effect   = "Allow"
            Resource = "*"
            Sid      = "EC2"
          },
          {
            Action = [
              "ssmmessages:OpenDataChannel",
              "ssmmessages:OpenControlChannel",
              "ssmmessages:CreateDataChannel",
              "ssmmessages:CreateControlChannel"
            ]
            Effect   = "Allow"
            Resource = "*"
            Sid      = "UseECSExec"
          },
        ])
        Version = "2012-10-17"
      }
    )
  }
}

# ECS on EC2 task definition
resource "aws_ecs_task_definition" "ec2" {
  family                = "${local.name}-ec2-task-definition"
  task_role_arn         = aws_iam_role.ec2.arn
  cpu                   = "256"
  memory                = "512"
  network_mode          = "awsvpc"
  container_definitions = local.container_definitions_without_fluent_bit
}

# ECS on EC2 service
resource "aws_ecs_service" "service" {
  name            = "${local.name}-ec2-service"
  cluster         = aws_ecs_cluster.ec2.id
  task_definition = aws_ecs_task_definition.ec2.arn
  desired_count   = local.task_count

  load_balancer {
    target_group_arn = aws_lb_target_group.main.arn
    container_name   = local.container_name
    container_port   = local.container_port
  }

  network_configuration {
    subnets         = aws_subnet.private[*].id
    security_groups = ["${aws_security_group.ecs.id}"]
  }

  # ## Spread tasks evenly accross all Availability Zones for High Availability
  # ordered_placement_strategy {
  #   type  = "spread"
  #   field = "attribute:ecs.availability-zone"
  # }

  # ## Make use of all available space on the Container Instances
  # ordered_placement_strategy {
  #   type  = "binpack"
  #   field = "memory"
  # }
}
