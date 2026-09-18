variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_import_defaulted_id" "my_importdefaultedid" {
  request_body_property = "test-request-body"
}
