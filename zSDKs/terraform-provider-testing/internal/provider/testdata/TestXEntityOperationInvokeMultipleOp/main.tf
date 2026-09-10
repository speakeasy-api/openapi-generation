variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "terraform_data" "test" {
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.testing_x_entity_operation_invoke_multiple_op.test]
    }
  }
}

action "testing_x_entity_operation_invoke_multiple_op" "test" {
  config {
    op1_string_request_only         = "test-op1-request-only"
    op1_string_request_and_response = "test-op1-request-and-response"
    op2_string_request_only         = "test-op2-request-only"
    op2_string_request_and_response = "test-op2-request-and-response"
  }
}
