terraform {
  required_providers {
    testing = {
      source  = "hashicorp/testing"
      version = "0.0.1"
    }
  }
}

provider "testing" {
  host_name      = "..." # Optional - can use TESTING_HOST_NAME environment variable
  port           = "..." # Optional - can use TESTING_PORT environment variable
  server_url     = "..." # Optional - can use TESTING_SERVER_URL environment variable
  server_version = "..." # Optional - can use TESTING_SERVER_VERSION environment variable
  subdomain      = "..." # Optional - can use TESTING_SUBDOMAIN environment variable
}