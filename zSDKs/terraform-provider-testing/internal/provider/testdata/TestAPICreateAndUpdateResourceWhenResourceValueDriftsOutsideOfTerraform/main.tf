variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_api_create_and_update" "my_apicreateandupdate" {
  name = "initial-name"
}
