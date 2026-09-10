variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_oas_enum" "my_oasenum" {
  string_request                              = "one"
  int32_enum_request_int32_enum_response     = 1
  int64_enum_request_int64_enum_response     = 2
}
