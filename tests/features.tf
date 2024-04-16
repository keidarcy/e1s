#######################
# ElastiCache for port forwarding test
#######################
resource "aws_elasticache_subnet_group" "test_redis" {
  name       = "${local.name}-redis-subnet-group"
  subnet_ids = aws_subnet.private[*].id
}

# allow all traffic from private subnet
resource "aws_security_group" "test_redis" {
  name        = "${local.name}-redis-sg"
  description = "Security group for ${local.name} test environment"
  vpc_id      = aws_vpc.main.id
  ingress {
    from_port   = 6379
    to_port     = 6379
    protocol    = "tcp"
    cidr_blocks = [cidrsubnet(aws_vpc.main.cidr_block, 8, 1)]
  }
}

resource "aws_elasticache_cluster" "test_redis" {
  cluster_id           = "${local.name}-redis"
  engine               = "redis"
  engine_version       = "6.x"
  node_type            = "cache.t3.micro"
  num_cache_nodes      = 1
  parameter_group_name = "default.redis6.x"
  port                 = 6379
  subnet_group_name    = aws_elasticache_subnet_group.test_redis.name
  security_group_ids   = [aws_security_group.test_redis.id]
}


#######################
# S3 bucket for file transfer test
#######################
resource "aws_s3_bucket" "test_bucket" {
  bucket = "${local.name}-cp-test"
}
