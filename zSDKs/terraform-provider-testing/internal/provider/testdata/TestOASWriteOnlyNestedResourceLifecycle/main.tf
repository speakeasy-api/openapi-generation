variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_oas_write_only_nested" "test" {
  id   = "test-id"
  name = "test-ssh-cred"

  ssh_credential = {
    keys = {
      private_key = "test-private-key"
      passphrase  = "test-passphrase"
    }
  }
}
