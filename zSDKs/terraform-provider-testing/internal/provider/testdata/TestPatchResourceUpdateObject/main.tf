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

variable "object_string" {
  type = string
}

resource "testing_patch" "test" {
  object = {
    object_string = var.object_string
  }
}
