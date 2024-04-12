locals {
  name                                  = "e1s"
  cidr_block                            = "10.0.0.0/16"
  availability_zones                    = ["us-east-1a", "us-east-1c", "us-east-1d"]
  cluster_count                         = var.cluster_count
  service_count                         = var.service_count
  task_count                            = var.task_count
  container_name                        = "nginx"
  container_port                        = "80"
  container_definitions_with_fluent_bit = <<EOL
[
  {
    "name": "nginx",
    "image": "nginx:1.14",
    "essential": true,
    "portMappings": [
      {
        "containerPort": 80,
        "hostPort": 80
      }
    ]
  },
  {
    "name": "fluent-bit",
    "image": "public.ecr.aws/aws-observability/aws-for-fluent-bit:stable",
    "memory": 200,
    "cpu": 10,
    "memoryReservation": 50
  }
]
EOL

  container_definitions_without_fluent_bit = <<EOL
[
  {
    "name": "nginx",
    "image": "nginx:1.14",
    "essential": true,
    "portMappings": [
      {
        "containerPort": 80,
        "hostPort": 80
      }
    ]
  }
]
EOL
  ecs_service_capacity_provider_strategies = [
    {
      capacity_provider = "FARGATE_SPOT"
      base              = 0
      weight            = 1
    },
    {
      capacity_provider = "FARGATE"
      base              = 1
      weight            = 1
    }
  ]
}


#######################
# ECS
#######################
resource "aws_ecs_cluster" "main" {
  count = local.cluster_count
  name  = "${local.name}-cluster-${count.index}"
  setting {
    name  = "containerInsights"
    value = "disabled"
  }
  lifecycle {
    ignore_changes = [setting]
  }
}

resource "aws_ecs_cluster_capacity_providers" "main" {
  cluster_name = aws_ecs_cluster.main[0].name

  capacity_providers = ["FARGATE", "FARGATE_SPOT"]

  dynamic "default_capacity_provider_strategy" {
    for_each = local.ecs_service_capacity_provider_strategies
    content {
      capacity_provider = default_capacity_provider_strategy.value.capacity_provider
      base              = default_capacity_provider_strategy.value.base
      weight            = default_capacity_provider_strategy.value.weight
    }
  }
}

# ECS Service
# https://www.terraform.io/docs/providers/aws/r/ecs_service.html
resource "aws_ecs_service" "main" {
  count                  = local.service_count
  name                   = "${local.name}-service-${count.index}"
  depends_on             = [aws_lb_listener_rule.main]
  cluster                = aws_ecs_cluster.main[0].name
  launch_type            = "FARGATE"
  desired_count          = count.index == 0 ? local.task_count : 0
  enable_execute_command = true
  task_definition        = aws_ecs_task_definition.main[count.index].arn
  network_configuration {
    subnets         = aws_subnet.private[*].id
    security_groups = ["${aws_security_group.ecs.id}"]
  }
  load_balancer {
    target_group_arn = aws_lb_target_group.main.arn
    container_name   = local.container_name
    container_port   = local.container_port
  }
  deployment_circuit_breaker {
    enable   = true
    rollback = true
  }
}

resource "aws_ecs_task_definition" "main" {
  count                    = local.service_count
  family                   = "${local.name}-task-definition-${count.index}"
  requires_compatibilities = ["FARGATE"]
  task_role_arn            = aws_iam_role.task_role.arn
  cpu                      = "256"
  memory                   = "512"
  network_mode             = "awsvpc"
  container_definitions    = count.index == 0 ? local.container_definitions_with_fluent_bit : local.container_definitions_without_fluent_bit
}

#######################
# Supporting Resources
#######################
resource "aws_service_discovery_http_namespace" "main" {
  name = local.name
}

resource "aws_iam_role" "task_role" {
  name = "${local.name}-task-role"
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
              "ssmmessages:OpenDataChannel",
              "ssmmessages:OpenControlChannel",
              "ssmmessages:CreateDataChannel",
              "ssmmessages:CreateControlChannel"
            ]
            Effect   = "Allow"
            Resource = "*"
            Sid      = "UseECSExec"
          },
          {
            Action = [
              "ssm:GetParameters",
              "ssm:GetParameter",
              "logs:PutLogEvents",
              "logs:CreateLogStream",
              "logs:CreateLogGroup",
              "kms:Decrypt",
            ],
            Effect = "Allow",
            Resource : "*",
          }
        ])
        Version = "2012-10-17"
      }
    )
  }
}


resource "aws_lb_target_group" "main" {
  name        = "${local.name}-tg"
  vpc_id      = aws_vpc.main.id
  port        = 80
  protocol    = "HTTP"
  target_type = "ip"
  health_check {
    healthy_threshold   = 2
    unhealthy_threshold = 2
    interval            = 5
    timeout             = 3
    path                = "/"
    protocol            = "HTTP"
    matcher             = "200"
  }
}

resource "aws_lb_listener_rule" "main" {
  listener_arn = aws_lb_listener.main.arn

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.main.id
  }

  condition {
    path_pattern {
      values = ["*"]
    }
  }
}


resource "aws_security_group" "ecs" {
  name = "${local.name}-ecs-sg"

  vpc_id = aws_vpc.main.id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}


# SecurityGroup Rule
# https://www.terraform.io/docs/providers/aws/r/security_group.html
resource "aws_security_group_rule" "ecs" {
  security_group_id = aws_security_group.ecs.id
  type              = "ingress"
  from_port         = 80
  to_port           = 80
  protocol          = "tcp"
  cidr_blocks       = ["10.0.0.0/16"]
}
