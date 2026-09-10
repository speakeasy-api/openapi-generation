resource "testing_oas_pattern" "my_oaspattern" {
  complex_pattern            = "abc"
  incompatible_pattern       = "...my_incompatible_pattern..."
  invalid_pattern            = "...my_invalid_pattern..."
  oneof_request              = "request1"
  oneof_request_and_response = "request_and_response1"
  oneof_response             = "response1"
  request                    = "request"
  request_and_response       = "request_and_response"
  response                   = "response"
}