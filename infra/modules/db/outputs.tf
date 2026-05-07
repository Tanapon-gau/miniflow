output "host" {
  value = aws_db_instance.main.address
}

output "port" {
  value = aws_db_instance.main.port
}

output "db_name" {
  value = var.db_name
}

output "db_username" {
  value = var.db_username
}
