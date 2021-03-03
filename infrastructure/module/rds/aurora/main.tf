module "db" {
  source = "terraform-aws-modules/rds-aurora/aws"
  name   = "dinghy-${terraform.workspace}"

  engine = "aurora-mysql"

  engine_mode           = "serverless"
  engine_version        = null
  replica_count         = 0

  subnets                 = var.subnets
  vpc_id                  = var.vpc_id
  allowed_security_groups = var.allowed_security_groups

  monitoring_interval             = 10
  performance_insights_enabled    = true

  skip_final_snapshot = true
  apply_immediately   = true
  storage_encrypted   = true
  backup_retention_period = 15
  deletion_protection     = false

  enable_http_endpoint    = true
  database_name = "dinghy"
  username      = "dinghy"

  scaling_configuration = {
    auto_pause               = true
    min_capacity             = 2
    max_capacity             = 16
    seconds_until_auto_pause = 300
    timeout_action           = "ForceApplyCapacityChange"
  }

  tags = {
    Environment = terraform.workspace
    Application = "dinghy"
  }
}