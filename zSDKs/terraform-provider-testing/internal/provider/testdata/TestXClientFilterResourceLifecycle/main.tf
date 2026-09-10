variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_x_client_filter" "test" {
  name     = "test-filter"
  category = "system"
}
