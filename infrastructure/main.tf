#Terraformer on Control Plane assumes the role arn:aws:iam::961214755549:role/ArmoryAdminRole which is the role for hosted-services AWS account
terraform {
  backend "s3" {
    bucket = "armory-dinghy-prod"
    key = "dinghy/terraform.tfstate"
    region = "us-west-2"
    profile = "hosted"
    role_arn = "arn:aws:iam::961214755549:role/ArmoryAdminRole"
  }
}

provider "aws" {
  region  = "us-west-2"
  profile = "hosted"
  assume_role {
    role_arn = "arn:aws:iam::961214755549:role/ArmoryAdminRole"
  }
}

module "db" {
  source = "./module/rds/aurora"
}

data "aws_sns_topic" "yeti_notification_topic" {
  name = "armory-cloud-services-prod-yeti-configuration-notifications"
}

module "cache_bust" {
  source = "./module/cachebust"
  notification_topic_arn = data.aws_sns_topic.yeti_notification_topic.arn
}
