resource "testing_x_plan_modifiers" "my_xplanmodifiers" {
  any     = "{ \"see\": \"documentation\" }"
  bool    = true
  float32 = 1.21
  float64 = 4.15
  int32   = 7
  int64   = 2
  list = [
    "..."
  ]
  list_nested = [
    {
      list_nested_string = "...my_list_nested_string..."
    }
  ]
  map = {
    key = jsonencode("value")
  }
  object = {
    object_string = "...my_object_string..."
  }
  set = [
    "..."
  ]
  set_nested = [
    {
      set_nested_string = "...my_set_nested_string..."
    }
  ]
  string = "...my_string..."
}