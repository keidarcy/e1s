# #######################
# # ALB Blue/Green Deployment Configuration
# #######################
#
# This configuration follows AWS documentation for blue/green deployments:
# https://docs.aws.amazon.com/AmazonECS/latest/developerguide/alb-resources-for-blue-green.html
#
# Key components:
# - Two target groups: blue (primary) and green (alternate)
# - Production listener on port 80
# - Test listener on port 8080 for validation
# - Listener rules for traffic routing

# Primary (Blue) Target Group
resource "aws_lb_target_group" "blue" {
  name        = "${local.name}-blue-tg"
  vpc_id      = aws_vpc.main.id
  port        = 80
  protocol    = "HTTP"
  target_type = "ip"

  health_check {
    enabled             = true
    healthy_threshold   = 2
    interval            = 30
    matcher             = "200"
    path                = "/"
    port                = "traffic-port"
    protocol            = "HTTP"
    timeout             = 5
    unhealthy_threshold = 2
  }

  tags = {
    Name = "${local.name}-blue-target-group"
  }
}

# Alternate (Green) Target Group
resource "aws_lb_target_group" "green" {
  name        = "${local.name}-green-tg"
  vpc_id      = aws_vpc.main.id
  port        = 80
  protocol    = "HTTP"
  target_type = "ip"

  health_check {
    enabled             = true
    healthy_threshold   = 2
    interval            = 30
    matcher             = "200"
    path                = "/"
    port                = "traffic-port"
    protocol            = "HTTP"
    timeout             = 5
    unhealthy_threshold = 2
  }

  tags = {
    Name = "${local.name}-green-target-group"
  }
}

# Production Listener (Port 80)
# Initially forwards traffic to blue target group
resource "aws_lb_listener" "production" {
  port              = "80"
  protocol          = "HTTP"
  load_balancer_arn = aws_lb.main.arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.blue.arn
  }
}

# Test Listener (Port 8080)
# Forwards traffic to green target group for validation
resource "aws_lb_listener" "test" {
  port              = "8080"
  protocol          = "HTTP"
  load_balancer_arn = aws_lb.main.arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.green.arn
  }
}

# Production Listener Rule for blue/green routing
resource "aws_lb_listener_rule" "production" {
  listener_arn = aws_lb_listener.production.arn
  priority     = 100

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.blue.arn
  }

  condition {
    path_pattern {
      values = ["*"]
    }
  }
}

# Test Listener Rule for path-based testing
resource "aws_lb_listener_rule" "test_path" {
  listener_arn = aws_lb_listener.production.arn
  priority     = 200

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.green.arn
  }

  condition {
    path_pattern {
      values = ["/test/*"]
    }
  }
}

# Test Listener Rule for header-based testing
resource "aws_lb_listener_rule" "test_header" {
  listener_arn = aws_lb_listener.production.arn
  priority     = 300

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.green.arn
  }

  condition {
    http_header {
      http_header_name = "X-Environment"
      values           = ["test"]
    }
  }
}

# ALB
# https://www.terraform.io/docs/providers/aws/d/lb.html
resource "aws_lb" "main" {
  load_balancer_type = "application"
  name               = "${local.name}-alb"

  security_groups = ["${aws_security_group.alb.id}"]
  subnets         = aws_subnet.public[*].id
}

# SecurityGroup Rule for HTTP (Port 80)
resource "aws_security_group_rule" "alb_http" {
  security_group_id = aws_security_group.alb.id
  type              = "ingress"
  from_port         = 80
  to_port           = 80
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
}

# SecurityGroup Rule for Test Port (Port 8080)
resource "aws_security_group_rule" "alb_test" {
  security_group_id = aws_security_group.alb.id
  type              = "ingress"
  from_port         = 8080
  to_port           = 8080
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group" "alb" {
  name   = "${local.name}-alb-sg"
  vpc_id = aws_vpc.main.id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# Amazon ECS infrastructure IAM role for load balancers
# This role allows Amazon ECS to manage load balancer resources for blue/green deployments
resource "aws_iam_role" "ecs_infrastructure_role" {
  name = "${local.name}-ecs-infrastructure-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AllowAccessToECSForInfrastructureManagement"
        Effect = "Allow"
        Principal = {
          Service = "ecs.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })

  tags = {
    Name        = "${local.name}-ecs-infrastructure-role"
    Description = "ECS infrastructure role for load balancer management"
  }
}

# Attach the AWS managed policy for ECS infrastructure role
resource "aws_iam_role_policy_attachment" "ecs_infrastructure_policy" {
  role       = aws_iam_role.ecs_infrastructure_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonECSInfrastructureRolePolicyForLoadBalancers"
}

# IAM policy to allow passing the infrastructure role to ECS
# This should be attached to the user/role that creates ECS services
resource "aws_iam_policy" "pass_ecs_infrastructure_role" {
  name        = "${local.name}-pass-ecs-infrastructure-role"
  description = "Allow passing ECS infrastructure role to ECS service"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action   = "iam:PassRole"
        Effect   = "Allow"
        Resource = aws_iam_role.ecs_infrastructure_role.arn
        Condition = {
          StringEquals = {
            "iam:PassedToService" = "ecs.amazonaws.com"
          }
        }
      }
    ]
  })
}
