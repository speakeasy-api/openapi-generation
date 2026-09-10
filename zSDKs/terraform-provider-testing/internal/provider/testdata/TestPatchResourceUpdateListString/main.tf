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

variable "list_string" {
  type = list(string)
}

resource "testing_patch" "test" {
  list_string = var.list_string
}
