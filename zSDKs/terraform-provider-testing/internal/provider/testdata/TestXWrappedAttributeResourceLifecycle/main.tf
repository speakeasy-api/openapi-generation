variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_x_wrapped_attribute" "test" {
  wrapped_list_string = ["test-value-1"]
}
