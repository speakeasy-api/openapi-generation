variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_discriminated_union" "test" {
  item = {
    item_a = {
      value_a = "my-secret-value"
    }
  }
}
