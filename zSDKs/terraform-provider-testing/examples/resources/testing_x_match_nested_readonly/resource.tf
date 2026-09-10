resource "testing_x_match_nested_readonly" "my_xmatchnestedreadonly" {
  name = "...my_name..."
  safety_settings = {
    enforce_approval = false
  }
}