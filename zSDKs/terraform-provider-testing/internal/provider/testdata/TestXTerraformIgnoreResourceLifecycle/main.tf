variable "configurable" {
  type = string
}

variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_x_terraform_ignore" "test" {
  configurable                     = var.configurable
  secondary_operation_configurable = var.configurable
}
