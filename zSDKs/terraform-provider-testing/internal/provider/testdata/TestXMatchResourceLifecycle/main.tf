variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_x_match" "my_xmatch" {
  object = {
    id = "object-123"
  }
}
