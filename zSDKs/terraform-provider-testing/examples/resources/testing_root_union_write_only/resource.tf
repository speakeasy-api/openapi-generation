resource "testing_root_union_write_only" "my_rootunionwriteonly" {
  alpha = {
    credentials = {
      client_id     = "...my_client_id..."
      client_secret = "...my_client_secret..."
    }
    name = "...my_name..."
  }
}