variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_import_multiple_id" "my_importmultipleid" {
  param1                = "test-param1"
  param2                = "test-param2"
  request_body_property = "test-request-body"
}
