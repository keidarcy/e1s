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
