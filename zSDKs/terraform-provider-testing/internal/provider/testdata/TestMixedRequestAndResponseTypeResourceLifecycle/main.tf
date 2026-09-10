variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_mixed_request_and_response_type" "my_resource" {
}
