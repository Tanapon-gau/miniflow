variable "region" {
  type    = string
  default = "ap-southeast-1"
}

variable "project" {
  type    = string
  default = "miniflow"
}

variable "db_password" {
  type      = string
  sensitive = true
}
