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

variable "set_string" {
  type = set(string)
}

resource "testing_patch" "test" {
  string     = var.string
  set_string = var.set_string
}
