variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_x_additional_properties_name" "test" {
  slug   = "test-slug"
  string = "test-string-value"
  x_additional_properties_name = jsonencode({
    "key1" = "value1"
    "key2" = 42
    "nested" = {
      "foo" = "bar"
    }
  })
}
