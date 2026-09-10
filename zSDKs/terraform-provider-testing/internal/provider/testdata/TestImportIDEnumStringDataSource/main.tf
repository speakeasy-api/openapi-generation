variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_import_id_enum_string" "test" {}

data "testing_import_id_enum_string" "test" {
    id = testing_import_id_enum_string.test.id
}
