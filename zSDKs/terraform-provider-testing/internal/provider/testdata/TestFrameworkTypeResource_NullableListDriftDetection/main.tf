variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_framework_type" "test" {
  bool          = true
  list_nullable = ["value1", "value2"]
}
