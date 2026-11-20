variable "server_url" {
  type = string
}

variable "replicas" {
  type = number
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_const_empty_set_delete" "test" {
  scope_id = "scope-1"
  groups = [{
    provider = "aws"
    region   = "eu-west-1"
    replicas = var.replicas
  }]
}
