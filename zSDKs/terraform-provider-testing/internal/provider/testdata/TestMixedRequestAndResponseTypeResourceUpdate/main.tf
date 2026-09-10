variable "server_url" {
  type = string
}

variable "integer_with_update_request_int32" {
  type = number
}

variable "float_with_update_request_float32" {
  type = number
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_mixed_request_and_response_type" "my_resource" {
  integer_with_update_request_int32       = var.integer_with_update_request_int32
  float_with_update_request_float32 = var.float_with_update_request_float32
}
