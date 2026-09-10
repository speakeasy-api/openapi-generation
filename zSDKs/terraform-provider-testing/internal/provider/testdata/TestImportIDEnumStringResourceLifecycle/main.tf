variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_import_id_enum_string" "my_importidenumstring" {
  # No attributes needed - this resource only has an ID
}
