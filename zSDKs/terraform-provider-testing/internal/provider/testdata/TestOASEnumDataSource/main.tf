variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_oas_enum" "test" {}

data "testing_oas_enum" "test" {
    id = testing_oas_enum.test.id
}

