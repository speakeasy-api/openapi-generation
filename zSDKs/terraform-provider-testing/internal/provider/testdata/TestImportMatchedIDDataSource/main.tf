variable "server_url" {
  type = string
}

variable "parent_entity_id" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_import_matched_id" "test" {
    parent_entity_id = var.parent_entity_id
}

data "testing_import_matched_id" "test" {
    parent_entity_id = testing_import_matched_id.test.parent_entity_id
}

