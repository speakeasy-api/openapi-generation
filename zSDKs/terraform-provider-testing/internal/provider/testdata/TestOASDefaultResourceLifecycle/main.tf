variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_oas_default" "my_oasdefault" {
  # This test relies on default values being applied
  # No custom values are specified to test defaults
}
