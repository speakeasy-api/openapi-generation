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

variable "int64" {
  type = number
}

variable "bool" {
  type = bool
}

resource "testing_patch" "test" {
  string = var.string
  int64  = var.int64
  bool   = var.bool
}
