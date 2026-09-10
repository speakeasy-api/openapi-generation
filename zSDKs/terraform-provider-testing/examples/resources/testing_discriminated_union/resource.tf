resource "testing_discriminated_union" "my_discriminatedunion" {
  item = {
    item_a = {
      type_varying_field = {
        key = jsonencode("value")
      }
      value_a = "...my_value_a..."
    }
  }
}