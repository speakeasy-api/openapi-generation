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

variable "map_string" {
  type = map(string)
}

resource "testing_patch" "test" {
  map_string = var.map_string
}
