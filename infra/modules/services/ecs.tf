locals {
  db_url_asyncpg = "postgresql+asyncpg://${var.db_username}:${var.db_password}@${var.db_host}:${var.db_port}/${var.db_name}"
  db_url_pg      = "postgres://${var.db_username}:${var.db_password}@${var.db_host}:${var.db_port}/${var.db_name}"
  db_url_std     = "postgresql://${var.db_username}:${var.db_password}@${var.db_host}:${var.db_port}/${var.db_name}"
  redis_addr     = "${var.redis_host}:${var.redis_port}"
  redis_url      = "redis://${var.redis_host}:${var.redis_port}"
}

resource "aws_ecs_cluster" "main" {
  name = var.project

  setting {
    name  = "containerInsights"
    value = "disabled"
  }

  tags = { Name = var.project }
}

resource "aws_cloudwatch_log_group" "services" {
  for_each          = local.service_names
  name              = "/ecs/${var.project}/${each.key}"
  retention_in_days = 7
}

# --- API ---

resource "aws_ecs_task_definition" "api" {
  family                   = "${var.project}-api"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = 256
  memory                   = 512
  execution_role_arn       = aws_iam_role.task_execution.arn
  task_role_arn            = aws_iam_role.task.arn

  container_definitions = jsonencode([{
    name      = "api"
    image     = "${aws_ecr_repository.services["api"].repository_url}:${var.api_image_tag}"
    essential = true
    portMappings = [{ containerPort = 8000 }]
    environment = [
      { name = "DATABASE_URL", value = local.db_url_asyncpg }
    ]
    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.services["api"].name
        "awslogs-region"        = var.region
        "awslogs-stream-prefix" = "ecs"
      }
    }
  }])
}

resource "aws_ecs_service" "api" {
  name            = "${var.project}-api"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.api.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets         = var.private_subnet_ids
    security_groups = [var.app_sg_id]
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.api.arn
    container_name   = "api"
    container_port   = 8000
  }

  depends_on = [aws_lb_listener.api]
}

# --- Scheduler ---

resource "aws_ecs_task_definition" "scheduler" {
  family                   = "${var.project}-scheduler"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = 256
  memory                   = 512
  execution_role_arn       = aws_iam_role.task_execution.arn
  task_role_arn            = aws_iam_role.task.arn

  container_definitions = jsonencode([{
    name      = "scheduler"
    image     = "${aws_ecr_repository.services["scheduler"].repository_url}:${var.scheduler_image_tag}"
    essential = true
    environment = [
      { name = "DATABASE_URL", value = local.db_url_pg },
      { name = "REDIS_ADDR", value = local.redis_addr }
    ]
    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.services["scheduler"].name
        "awslogs-region"        = var.region
        "awslogs-stream-prefix" = "ecs"
      }
    }
  }])
}

resource "aws_ecs_service" "scheduler" {
  name            = "${var.project}-scheduler"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.scheduler.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets         = var.private_subnet_ids
    security_groups = [var.app_sg_id]
  }
}

# --- Worker Go ---

resource "aws_ecs_task_definition" "worker_go" {
  family                   = "${var.project}-worker-go"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = 256
  memory                   = 512
  execution_role_arn       = aws_iam_role.task_execution.arn
  task_role_arn            = aws_iam_role.task.arn

  container_definitions = jsonencode([{
    name      = "worker-go"
    image     = "${aws_ecr_repository.services["worker-go"].repository_url}:${var.worker_go_image_tag}"
    essential = true
    environment = [
      { name = "DATABASE_URL", value = local.db_url_pg },
      { name = "REDIS_ADDR", value = local.redis_addr }
    ]
    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.services["worker-go"].name
        "awslogs-region"        = var.region
        "awslogs-stream-prefix" = "ecs"
      }
    }
  }])
}

resource "aws_ecs_service" "worker_go" {
  name            = "${var.project}-worker-go"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.worker_go.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets         = var.private_subnet_ids
    security_groups = [var.app_sg_id]
  }
}

# --- Worker Python ---

resource "aws_ecs_task_definition" "worker_py" {
  family                   = "${var.project}-worker-py"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = 512
  memory                   = 1024
  execution_role_arn       = aws_iam_role.task_execution.arn
  task_role_arn            = aws_iam_role.task.arn

  container_definitions = jsonencode([{
    name      = "worker-py"
    image     = "${aws_ecr_repository.services["worker-py"].repository_url}:${var.worker_py_image_tag}"
    essential = true
    environment = [
      { name = "DATABASE_URL", value = local.db_url_std },
      { name = "REDIS_URL", value = local.redis_url }
    ]
    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.services["worker-py"].name
        "awslogs-region"        = var.region
        "awslogs-stream-prefix" = "ecs"
      }
    }
  }])
}

resource "aws_ecs_service" "worker_py" {
  name            = "${var.project}-worker-py"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.worker_py.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets         = var.private_subnet_ids
    security_groups = [var.app_sg_id]
  }
}

# --- Web ---

resource "aws_ecs_task_definition" "web" {
  family                   = "${var.project}-web"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = 256
  memory                   = 512
  execution_role_arn       = aws_iam_role.task_execution.arn
  task_role_arn            = aws_iam_role.task.arn

  container_definitions = jsonencode([{
    name      = "web"
    image     = "${aws_ecr_repository.services["web"].repository_url}:${var.web_image_tag}"
    essential = true
    portMappings = [{ containerPort = 80 }]
    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.services["web"].name
        "awslogs-region"        = var.region
        "awslogs-stream-prefix" = "ecs"
      }
    }
  }])
}

resource "aws_ecs_service" "web" {
  name            = "${var.project}-web"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.web.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets         = var.private_subnet_ids
    security_groups = [var.app_sg_id]
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.web.arn
    container_name   = "web"
    container_port   = 80
  }

  depends_on = [aws_lb_listener.web]
}
