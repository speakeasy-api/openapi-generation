resource "testing_oas_write_only_nested" "my_oaswriteonlynested" {
  id   = "...my_id..."
  name = "...my_name..."
  ssh_credential = {
    keys = {
      passphrase  = "...my_passphrase..."
      private_key = "...my_private_key..."
    }
  }
}