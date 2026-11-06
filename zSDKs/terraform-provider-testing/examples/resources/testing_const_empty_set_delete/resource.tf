resource "testing_const_empty_set_delete" "my_constemptysetdelete" {
  groups = [
    {
      provider = "...my_provider..."
      region   = "...my_region..."
      replicas = 10
    }
  ]
  scope_id = "...my_scope_id..."
}