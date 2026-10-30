variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_const_container" "test" {
  name = "test-const-container"
}

data "testing_const_container" "test" {
  id = testing_const_container.test.id
}
