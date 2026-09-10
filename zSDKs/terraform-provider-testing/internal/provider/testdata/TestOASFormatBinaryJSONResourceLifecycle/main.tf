variable "server_url" {
  type = string
}

variable "content" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_oas_format_binary_json" "test" {
  name    = "test-oas-format-binary-json"
  content = var.content
}
