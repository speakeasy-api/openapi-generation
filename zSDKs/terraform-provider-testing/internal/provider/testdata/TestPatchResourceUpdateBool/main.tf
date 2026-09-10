terraform {
  required_providers {
    testing = {
      source = "hashicorp/testing"
    }
  }
}

provider "testing" {
  server_url = var.server_url
}

variable "server_url" {
  type = string
}

variable "bool" {
  type = bool
}

resource "testing_patch" "test" {
  bool = var.bool
}
