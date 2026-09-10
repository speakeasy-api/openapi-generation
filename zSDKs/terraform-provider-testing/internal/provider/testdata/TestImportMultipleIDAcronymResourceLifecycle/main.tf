variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

variable "api_id" {
  type = string
}

variable "portal_id" {
  type = string
}

variable "request_body_property" {
  type = string
}

resource "testing_import_multiple_id_acronym" "my_importmultipleidacronym" {
  api_id                = var.api_id
  portal_id             = var.portal_id
  request_body_property = var.request_body_property
}

