variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_import_matched_id" "my_importmatchedid" {
  parent_entity_id = "parent-456"
}

