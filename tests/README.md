# Terraform ECS Clusters for `e1s` testing purposes

- ECS cluster using Fargate (on-demand and spot) capacity providers
- Example ECS service that utilizes
  - Load balancer target group attachment
  - Security group for access to the example service
  - Task role for exec shell access to the containers
  - Task definition using FluentBit sidecar container definition
- Example ECS service `e1s-firelens` shipping its logs through FireLens to CloudWatch(`firelens = false` to skip it)

## Usage

To run this example you need to execute.

```bash
$ terraform init
$ terraform plan
$ terraform apply
```

__THIS WILL BE CHARGED TO YOUR AWS ACCOUNT__

### FireLens logs

`e1s-firelens` runs a Fluent Bit log router and two containers whose logs FireLens ships to CloudWatch, `app` to `/ecs/e1s/firelens` and `worker` to `/ecs/e1s/firelens-worker`. Their task definition carries no awslogs options, so `e1s` reads them through the `cloudwatch-log-overrides` entries of [e1s-config.yml](./e1s-config.yml):

```bash
$ e1s --config-file tests/e1s-config.yml
```

Logs of `app`, `worker` and `log-router` are readable on the task and the container, and the nginx container of `e1s-service-0` still reads its own awslogs group.

### Cleanup

```bash
$ terraform destroy
```

### Resources

Following resources are created in this example:

- [ECS Cluster](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/ecs_cluster)
- [ECS Service](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/ecs_service)
- [ECS Task Definition](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/ecs_task_definition)
- [ALB](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/lb)
- [ALB Target Group](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/lb_target_group)
- [VPC](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/vpc)
- [Subnet](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/subnet)
- S3
- ElastiCache