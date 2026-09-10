variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_x_match_nested_readonly" "test" {
  name = "test-resource"

  safety_settings = {
    enforce_approval = true
  }
}
