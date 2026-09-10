variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_framework_type" "test" {}

data "testing_framework_type" "test" {
    id = testing_framework_type.test.id
}
