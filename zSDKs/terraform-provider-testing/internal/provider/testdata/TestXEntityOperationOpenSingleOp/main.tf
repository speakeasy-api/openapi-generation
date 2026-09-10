variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

ephemeral "testing_x_entity_operation_open_single_op" "test" {}

# Ephemeral resource testing must use echo provider
provider "echo" {
  data = ephemeral.testing_x_entity_operation_open_single_op.test
}

resource "echo" "test" {}
