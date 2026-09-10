variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_api_create_and_update" "test" {
    name = "example-name"
}

data "testing_api_create_and_update" "test" {
  depends_on = [testing_api_create_and_update.test]
}
