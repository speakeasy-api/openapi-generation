variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_status_code_2xx" "test" {
  name = "test"
}
