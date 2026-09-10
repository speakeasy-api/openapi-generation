resource "terraform_data" "test" {
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.testing_custom.test]
    }
  }
}

action "testing_custom" "test" {
  config {}
}
