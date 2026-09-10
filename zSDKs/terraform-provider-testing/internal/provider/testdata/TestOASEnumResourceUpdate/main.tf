variable "server_url" {
  type = string
}

variable "string_request" {
  type = string
}

variable "int32_enum" {
  type = number
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_oas_enum" "my_oasenum" {
  string_request                              = var.string_request
  int32_enum_request_int32_enum_response     = var.int32_enum
}
