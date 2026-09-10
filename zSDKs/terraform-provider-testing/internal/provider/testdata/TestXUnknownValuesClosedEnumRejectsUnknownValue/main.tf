variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_x_unknown_values" "my_xunknownvalues" {
  required_open_enum_string = "alpha"
  closed_enum_string        = "invalid_value"
}
