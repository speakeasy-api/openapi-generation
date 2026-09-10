variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

data "testing_x_client_filter_array" "test" {}
