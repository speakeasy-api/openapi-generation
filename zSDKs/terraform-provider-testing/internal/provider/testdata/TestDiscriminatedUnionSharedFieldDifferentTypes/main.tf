variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_discriminated_union" "test" {
  item = {
    item_b = {
      value_b = "test-value"
      type_varying_field = {
        setting = "my-setting"
        enabled = true
      }
    }
  }
}
