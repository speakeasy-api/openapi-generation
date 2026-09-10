variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_x_entity_missing_codes" "test" {
  name = "test-name"
}
