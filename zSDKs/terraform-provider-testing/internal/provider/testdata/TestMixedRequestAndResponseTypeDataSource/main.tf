variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_mixed_request_and_response_type" "test" {
}

data "testing_mixed_request_and_response_type" "test" {
  id = testing_mixed_request_and_response_type.test.id
}
