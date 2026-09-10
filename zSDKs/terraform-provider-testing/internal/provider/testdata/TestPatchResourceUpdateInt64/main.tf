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

variable "int64" {
  type = number
}

resource "testing_patch" "test" {
  int64 = var.int64
}
