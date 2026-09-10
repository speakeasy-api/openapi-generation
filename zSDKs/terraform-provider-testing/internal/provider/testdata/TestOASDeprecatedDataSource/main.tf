variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_oas_deprecated" "test" {}

data "testing_oas_deprecated" "test" {
    id = testing_oas_deprecated.test.id
}

