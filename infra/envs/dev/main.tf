module "network" {
  source = "../../modules/network"

  project              = var.project
  vpc_cidr             = "10.0.0.0/16"
  availability_zones   = ["${var.region}a", "${var.region}b"]
  public_subnet_cidrs  = ["10.0.1.0/24", "10.0.2.0/24"]
  private_subnet_cidrs = ["10.0.11.0/24", "10.0.12.0/24"]
}

module "db" {
  source = "../../modules/db"

  project           = var.project
  subnet_ids        = module.network.private_subnet_ids
  security_group_id = module.network.db_sg_id
  db_password       = var.db_password
}

module "redis" {
  source = "../../modules/redis"

  project           = var.project
  subnet_ids        = module.network.private_subnet_ids
  security_group_id = module.network.redis_sg_id
}

module "services" {
  source = "../../modules/services"

  project            = var.project
  region             = var.region
  vpc_id             = module.network.vpc_id
  public_subnet_ids  = module.network.public_subnet_ids
  private_subnet_ids = module.network.private_subnet_ids
  alb_sg_id          = module.network.alb_sg_id
  app_sg_id          = module.network.app_sg_id

  db_host     = module.db.host
  db_port     = module.db.port
  db_name     = module.db.db_name
  db_username = module.db.db_username
  db_password = var.db_password

  redis_host = module.redis.host
  redis_port = module.redis.port
}
