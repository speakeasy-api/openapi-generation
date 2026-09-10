variable "provider_server_url" {
  type = string
}

variable "resource_server_url" {
  type = string
}

provider "testing" {
  server_url = var.provider_server_url
}

resource "testing_oas_servers_operation" "test" {
  server_url = var.resource_server_url
}

data "testing_oas_servers_operation" "test" {
  id         = testing_oas_servers_operation.test.id
  server_url = var.resource_server_url
}
