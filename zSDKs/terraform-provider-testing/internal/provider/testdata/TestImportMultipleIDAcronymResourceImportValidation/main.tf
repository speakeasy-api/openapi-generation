variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_import_multiple_id_acronym" "my_importmultipleidacronym" {
  api_id                = "api-123"
  portal_id             = "portal-456"
  request_body_property = "test-value"
}

