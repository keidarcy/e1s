# Terraform ECS Clusters for `e1s` testing purposes

- ECS cluster using Fargate (on-demand and spot) capacity providers
- Example ECS service that utilizes
  - Load balancer target group attachment
  - Security group for access to the example service
  - Task role for exec shell access to the containers
  - Task definition using FluentBit sidecar container definition

## Usage

To run this example you need to execute.

```bash
$ terraform init
$ terraform plan
$ terraform apply
```

__THIS WILL BE CHARGED TO YOUR AWS ACCOUNT__

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