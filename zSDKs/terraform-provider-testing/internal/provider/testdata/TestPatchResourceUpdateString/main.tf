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

variable "string" {
  type = string
}

resource "testing_patch" "test" {
  string = var.string
}
