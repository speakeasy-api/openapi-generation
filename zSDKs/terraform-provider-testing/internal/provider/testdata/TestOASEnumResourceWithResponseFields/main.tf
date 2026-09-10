variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_oas_enum" "my_oasenum" {
  # Minimal configuration to test response-only fields
}
