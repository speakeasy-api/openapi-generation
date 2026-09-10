variable "server_url" {
  type = string
}

variable "long_enum" {
  type = number
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_oas_enum" "my_oasenum" {
  int32_long_enum = var.long_enum
}
