locals {
  ec2_name = "e1s-ec2"
}

data "aws_ami" "ecs_optimized" {
  most_recent = true

  filter {
    name   = "name"
    values = ["amzn2-ami-ecs-hvm-*-arm64-ebs"]
  }

  owners = ["amazon"]
}

# EC2 Instance IAM resources
data "aws_iam_policy_document" "ec2_instance" {
  statement {
    actions = [
      "cloudwatch:PutMetricData",
    ]

    resources = [
      "*",
    ]
  }

  statement {
    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
      "logs:DescribeLogStreams",
    ]

    resources = [
      "*",
    ]
  }

  statement {
    actions = [
      "ec2:*",
      "ecs:*",
      "kms:*",
    ]

    resources = [
      "*",
    ]
  }
}

#######################
# EC2 IAM
#######################
resource "aws_iam_role" "ec2_instance" {
  name = "${local.ec2_name}-instance-role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      },
    ]
  })

  inline_policy {
    name   = "instance-policy"
    policy = data.aws_iam_policy_document.ec2_instance.json
  }
}


resource "aws_iam_role_policy_attachment" "ec2_instance_ssm" {
  role       = aws_iam_role.ec2_instance.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

resource "aws_iam_instance_profile" "ec2_instance" {
  name = "${local.ec2_name}-instance-profile"
  role = aws_iam_role.ec2_instance.name
}


resource "aws_security_group" "ec2_instance" {
  name        = "${local.ec2_name}-instance-sg"
  description = "Security group for ${local.ec2_name} test environment"
  vpc_id      = aws_vpc.main.id

  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }
}

resource "aws_launch_template" "ec2_instance" {
  name                                 = "${local.ec2_name}-instance-lt"
  image_id                             = data.aws_ami.ecs_optimized.image_id
  instance_type                        = "t4g.medium"
  ebs_optimized                        = true
  instance_initiated_shutdown_behavior = "terminate"

  tag_specifications {
    resource_type = "instance"

    tags = {
      Name = "${local.ec2_name}-instance"
    }
  }

  iam_instance_profile {
    name = aws_iam_instance_profile.ec2_instance.id
  }

  monitoring {
    enabled = true
  }

  network_interfaces {
    associate_public_ip_address = true
    security_groups             = [aws_security_group.ec2_instance.id]
  }

  user_data = base64encode(<<-EOF
    #!/usr/bin/env bash

    cat <<EOF >> /etc/ecs/ecs.config
    ECS_CLUSTER=${aws_ecs_cluster.ec2_instance[0].name}
    ECS_ENABLE_CONTAINER_METADATA=true
    ECS_CONTAINER_INSTANCE_PROPAGATE_TAGS_FROM=ec2_instance
    EOF
  )

  instance_market_options {
    market_type = "spot"
  }
}

resource "aws_autoscaling_group" "ec2_instance" {
  name                = "${local.ec2_name}-instance-asg"
  vpc_zone_identifier = aws_subnet.public[*].id
  max_size            = 1
  min_size            = 0
  desired_capacity    = 1

  launch_template {
    id      = aws_launch_template.ec2_instance.id
    version = "$Latest"
  }

  health_check_grace_period = 300
  health_check_type         = "EC2"

  lifecycle {
    create_before_destroy = true
  }
}

#######################
# ECS
#######################
resource "aws_ecs_cluster" "ec2_instance" {
  count = local.cluster_count
  name  = "${local.ec2_name}-cluster-${count.index}"
}

resource "aws_ecs_service" "ec2_instance" {
  count                  = local.service_count
  name                   = "${local.ec2_name}-service-${count.index}"
  cluster                = aws_ecs_cluster.ec2_instance[0].name
  task_definition        = aws_ecs_task_definition.ec2_launch.arn
  desired_count          = count.index == 0 ? local.task_count : 0
  enable_execute_command = true
  launch_type            = "EC2"
}


resource "aws_cloudwatch_log_group" "ec2_instance" {
  name = "/aws/ecs/${local.ec2_name}"
}

resource "aws_ecs_task_definition" "ec2_launch" {
  family                   = "${local.ec2_name}-ec2-launch"
  task_role_arn            = aws_iam_role.task_role.arn
  execution_role_arn       = aws_iam_role.task_role.arn
  requires_compatibilities = ["EC2"]

  runtime_platform {
    operating_system_family = "LINUX"
  }

  container_definitions = jsonencode([
    {
      name      = "nginx"
      image     = "nginx:alpine"
      cpu       = 256
      memory    = 512
      essential = true
      portMappings = [
        {
          containerPort = 80
        }
      ]
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group" : aws_cloudwatch_log_group.ec2_instance.id
          "awslogs-region" : "us-east-1"
          "awslogs-stream-prefix" : "ecs"
        }
      }
    },
    {
      name      = "redis"
      image     = "redis:alpine"
      cpu       = 256
      memory    = 512
      essential = false
      portMappings = [
        {
          containerPort = 6379
      }]
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group" : aws_cloudwatch_log_group.ec2_instance.id
          "awslogs-region" : "us-east-1"
          "awslogs-stream-prefix" : "ecs"
        }
      }
    },
  ])
}
