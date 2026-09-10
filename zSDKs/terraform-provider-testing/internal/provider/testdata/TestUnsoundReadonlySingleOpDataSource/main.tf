variable "server_url" {
  type = string
}

variable "name" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_unsound_readonly_single_op" "test" {
  name = var.name
}

data "testing_unsound_readonly_single_op" "test" {
  id = testing_unsound_readonly_single_op.test.id
}
