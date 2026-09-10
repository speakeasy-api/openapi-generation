variable "server_url" {
  type = string
}

variable "filter_name" {
  type = string
  default = ""
}

provider "testing" {
  server_url = var.server_url
}

data "testing_x_pagination_singleton" "test" {
  filter_name = var.filter_name
}

