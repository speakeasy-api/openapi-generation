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

variable "float64" {
  type = number
}

resource "testing_patch" "test" {
  float64 = var.float64
}
