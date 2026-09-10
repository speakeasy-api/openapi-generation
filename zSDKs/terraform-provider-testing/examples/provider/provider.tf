terraform {
  required_providers {
    testing = {
      source  = "hashicorp/testing"
      version = "0.0.1"
    }
  }
}

provider "testing" {
  hostname    = "..." # Optional - can use TESTING_HOSTNAME environment variable
  server_port = "..." # Optional - can use TESTING_PORT environment variable
  server_url  = "..." # Optional
}