variable "server_url" {
  type = string
}

variable "name" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_basic" "my_basic" {
  name = var.name
}
