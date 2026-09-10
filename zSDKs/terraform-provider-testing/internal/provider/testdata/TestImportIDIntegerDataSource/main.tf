variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_import_id_integer" "test" {}

data "testing_import_id_integer" "test" {
    id = testing_import_id_integer.test.id
}

