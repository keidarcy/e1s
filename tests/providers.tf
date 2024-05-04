terraform {
  required_version = ">= 1.5"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.55"
    }
  }
}

provider "aws" {
  region  = "us-east-1"
  profile = "e1s" // match your ~/.aws/config profile
}

# provider "aws" {
#   region     = "us-east-1"
#   access_key = "mock_access_key" // Use mock credentials for LocalStack
#   secret_key = "mock_secret_key" // Use mock credentials for LocalStack
#   profile    = "localstack"      // match your ~/.aws/config profile

#   // LocalStack uses the same endpoint for all services
#   endpoints {
#     elasticache = "http://localhost:4566"
#     ec2         = "http://localhost:4566"
#     ecs         = "http://localhost:4566"
#     s3          = "http://localhost:4566"
#   }
#   skip_credentials_validation = true
#   skip_requesting_account_id  = true
#   skip_metadata_api_check     = true
# }
