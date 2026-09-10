resource "testing_x_terraform_write_only" "my_xterraformwriteonly" {
  string                    = "...my_string..."
  string_with_oas_writeonly = "...my_string_with_oas_writeonly..."
  string_without_writeonly  = "...my_string_without_writeonly..."
}