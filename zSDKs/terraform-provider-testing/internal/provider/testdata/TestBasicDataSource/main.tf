variable "server_url" {
  type = string
}

variable "name"{
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_basic" "test" {
    name = var.name
}

data "testing_basic" "test" {
    name = testing_basic.test.name
}
