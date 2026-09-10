resource "testing_x_plan_validators" "my_xplanvalidators" {
  bool    = true
  float32 = 2.41
  float64 = 8
  int32   = 5
  int64   = 0
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