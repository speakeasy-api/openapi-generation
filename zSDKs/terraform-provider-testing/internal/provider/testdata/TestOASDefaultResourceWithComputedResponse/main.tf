variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_oas_default" "my_oasdefault" {
  # string_response is response-only computed field
}
