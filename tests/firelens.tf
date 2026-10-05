#######################
# FireLens service for cloudwatch-log-overrides test
#######################
# Two containers ship their logs through a Fluent Bit sidecar(FireLens) to
# CloudWatch, so their task definition carries no awslogs options. The log
# router itself keeps the awslogs log driver. Use tests/e1s-config.yml to read
# all three containers' logs in e1s.
locals {
  firelens_name         = "${local.name}-firelens"
  firelens_group        = "/ecs/${local.name}/firelens"
  firelens_worker_group = "/ecs/${local.name}/firelens-worker"

  firelens_container_definitions = jsonencode([
    {
      name              = "log-router"
      image             = "public.ecr.aws/aws-observability/aws-for-fluent-bit:stable"
      essential         = true
      memoryReservation = 50
      firelensConfiguration = {
        type = "fluentbit"
      }
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          awslogs-group         = aws_cloudwatch_log_group.main.name
          awslogs-region        = "us-east-1"
          awslogs-stream-prefix = "firelens"
        }
      }
    },
    {
      name      = "app"
      image     = "public.ecr.aws/docker/library/busybox:stable"
      essential = true
      command   = ["sh", "-c", "while true; do echo \"app log line $(date)\"; sleep 5; done"]
      dependsOn = [{ containerName = "log-router", condition = "START" }]
      # stream name is ecs/app-firelens-<taskId>
      logConfiguration = {
        logDriver = "awsfirelens"
        options = {
          Name              = "cloudwatch_logs"
          region            = "us-east-1"
          log_group_name    = local.firelens_group
          log_stream_prefix = "ecs/"
          auto_create_group = "false"
        }
      }
    },
    {
      name      = "worker"
      image     = "public.ecr.aws/docker/library/busybox:stable"
      essential = true
      command   = ["sh", "-c", "while true; do echo \"worker log line $(date)\"; sleep 7; done"]
      dependsOn = [{ containerName = "log-router", condition = "START" }]
      # a group and a stream prefix of its own, stream name is worker/worker-firelens-<taskId>
      logConfiguration = {
        logDriver = "awsfirelens"
        options = {
          Name              = "cloudwatch_logs"
          region            = "us-east-1"
          log_group_name    = local.firelens_worker_group
          log_stream_prefix = "worker/"
          auto_create_group = "false"
        }
      }
    }
  ])
}

resource "aws_cloudwatch_log_group" "firelens" {
  count             = var.firelens ? 1 : 0
  name              = local.firelens_group
  retention_in_days = 1
}

resource "aws_cloudwatch_log_group" "firelens_worker" {
  count             = var.firelens ? 1 : 0
  name              = local.firelens_worker_group
  retention_in_days = 1
}

resource "aws_ecs_task_definition" "firelens" {
  count                    = var.firelens ? 1 : 0
  family                   = local.firelens_name
  requires_compatibilities = ["FARGATE"]
  task_role_arn            = aws_iam_role.task_role.arn
  execution_role_arn       = aws_iam_role.task_role.arn
  cpu                      = "256"
  memory                   = "512"
  network_mode             = "awsvpc"
  container_definitions    = local.firelens_container_definitions
}

resource "aws_ecs_service" "firelens" {
  count                  = var.firelens ? 1 : 0
  name                   = local.firelens_name
  cluster                = aws_ecs_cluster.main[0].name
  launch_type            = "FARGATE"
  desired_count          = 1
  enable_execute_command = true
  task_definition        = aws_ecs_task_definition.firelens[0].arn
  network_configuration {
    subnets         = aws_subnet.private[*].id
    security_groups = [aws_security_group.ecs.id]
  }
  depends_on = [aws_cloudwatch_log_group.firelens, aws_cloudwatch_log_group.firelens_worker]
}
