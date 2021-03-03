// default variables defined down below are based on a previously provisioned infrastructure on hosted-services AWS account.

// default vpc_id is the base-vpc, not default
variable "vpc_id" {
  default = "vpc-03eb15cc673d291e0"
}

// default subnets are the database subnets
variable "subnets" {
  default = ["subnet-09d504ad4e7febc4b", "subnet-0e941a39df45ddfe8", "subnet-0f2c8d8c487b633b1"]
}

// default security groups are the worker node security groups
variable "allowed_security_groups" {
  default = ["sg-0a954dd7745b8a986"]
}