resource "testing_oas_content_media_type" "my_oascontentmediatype" {
  application_json = jsonencode({})
  application_json_with_example = jsonencode({
    key = "value"
  })
}