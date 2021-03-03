# aws_rds_cluster
output "db" {
  value = module.db
  sensitive = true
}