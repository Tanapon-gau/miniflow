locals {
  service_names = toset(["api", "scheduler", "worker-go", "worker-py", "web"])
}

resource "aws_ecr_repository" "services" {
  for_each             = local.service_names
  name                 = "${var.project}/${each.key}"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  tags = { Name = "${var.project}-${each.key}" }
}
