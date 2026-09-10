resource "testing_oas_format_binary_json" "my_oasformatbinaryjson" {
  content          = filebase64("${path.module}/example")
  name             = "...my_name..."
  optional_content = filebase64("${path.module}/example")
}