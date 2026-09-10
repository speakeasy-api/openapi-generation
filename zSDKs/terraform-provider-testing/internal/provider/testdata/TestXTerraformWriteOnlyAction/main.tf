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
      actions = [action.testing_x_terraform_write_only_action.test]
    }
  }
}

action "testing_x_terraform_write_only_action" "test" {
  config {
    string                    = "test-write-only"
    string_with_oas_writeonly = "test-write-only"
    string_without_writeonly  = "test-regular"
  }
}
