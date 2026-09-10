variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_import_id_integer" "my_importidinteger" {
}

