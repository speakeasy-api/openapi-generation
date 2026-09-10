variable "server_url" {
  type = string
}

variable "resource_name" {
  type = string
}

variable "resource_type" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_hoisted_readonly_default_minimal" "test" {
  name = var.resource_name
  type = var.resource_type
}
