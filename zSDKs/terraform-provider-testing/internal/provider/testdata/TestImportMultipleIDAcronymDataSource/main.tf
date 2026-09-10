variable "server_url" {
  type = string
}

variable "api_id" {
  type = string
}

variable "portal_id" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_import_multiple_id_acronym" "test" {
    api_id    = var.api_id
    portal_id = var.portal_id
}

data "testing_import_multiple_id_acronym" "test" {
    api_id    = testing_import_multiple_id_acronym.test.api_id
    portal_id = testing_import_multiple_id_acronym.test.portal_id
}

