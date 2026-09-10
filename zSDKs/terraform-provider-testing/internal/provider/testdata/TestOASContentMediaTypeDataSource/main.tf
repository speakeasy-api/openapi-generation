variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_oas_content_media_type" "test" {
  application_json = jsonencode({ key = "value" })
}

data "testing_oas_content_media_type" "test" {
  id = testing_oas_content_media_type.test.id
}
