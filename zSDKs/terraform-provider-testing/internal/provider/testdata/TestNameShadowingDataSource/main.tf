variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_name_shadowing" "test" {}

data "testing_name_shadowing" "test" {
    id = testing_name_shadowing.test.id
}

