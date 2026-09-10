variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_basic" "my_basic" {
  name = "test-basic-resource"
}
