variable "server_url" {
  type = string
}


provider "testing" {
  server_url = var.server_url
}

resource "testing_framework_type" "my_frameworktype" {
  bool = true
}
