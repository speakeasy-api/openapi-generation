variable "server_url" {
  type = string
}

variable "param1" {
  type = string
}

variable "param2" {
  type = string
}


provider "testing" {
  server_url = var.server_url
}

resource "testing_import_multiple_id" "test" {
    param1                = var.param1
    param2                = var.param2
    # request_body_property = var.request_body_property
}

data "testing_import_multiple_id" "test" {
    param1 = testing_import_multiple_id.test.param1
    param2 = testing_import_multiple_id.test.param2
}

