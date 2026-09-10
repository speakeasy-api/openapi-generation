variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_x_entity_object_nested_optional" "test" {}

data "testing_x_entity_object_nested_optional" "test" {
  id = testing_x_entity_object_nested_optional.test.id
}
