resource "testing_x_param_suppress_computed_diff" "my_xparamsuppresscomputeddiff" {
  bool    = true
  float32 = 7.43
  float64 = 0.92
  int32   = 4
  int64   = 9
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
  number = 3.56
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
  string_map = {
    key = "value"
  }
}