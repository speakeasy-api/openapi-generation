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
      actions = [action.testing_x_entity_operation_invoke_single_op.test]
    }
  }
}

action "testing_x_entity_operation_invoke_single_op" "test" {
  config {
    string_request_only         = "test-request-only"
    string_request_and_response = "test-request-and-response"
  }
}
