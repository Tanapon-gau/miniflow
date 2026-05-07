output "web_url" {
  description = "URL for the web UI"
  value       = "http://${module.services.alb_dns_name}"
}

output "api_url" {
  description = "URL for the API"
  value       = "http://${module.services.alb_dns_name}:8000"
}

output "ecr_repository_urls" {
  description = "ECR repo URLs — use these to tag and push Docker images"
  value       = module.services.ecr_repository_urls
}
