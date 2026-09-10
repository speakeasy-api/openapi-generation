variable "ephemeral_string" {
  ephemeral = true
  type      = string
}

variable "server_url" {
  type = string
}

variable "string" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_x_terraform_write_only" "test" {
  string                    = var.ephemeral_string
  string_with_oas_writeonly = var.ephemeral_string
  string_without_writeonly  = var.string
}
