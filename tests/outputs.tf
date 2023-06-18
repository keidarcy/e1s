output "alb_dns" {
  description = "Public ALB DNS name"
  value       = aws_lb.main.dns_name
}

output "ecs_clusters" {
  description = "ECS Cluster ARNs"
  value       = aws_ecs_cluster.main[*].name
}

output "ecs_services" {
  description = "ECS Service ARNs"
  value       = aws_ecs_service.main[*].name
}
