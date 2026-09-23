resource "testing_import_defaulted_id" "my_importdefaultedid" {
  region                = "oas-region"
  request_body_property = "...my_request_body_property..."
  tier                  = "basic"
  workspace             = "default-workspace"
}