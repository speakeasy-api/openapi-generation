variable "name" {
  type = string
}

variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_x_polling" "test" {
  name = var.name
}
