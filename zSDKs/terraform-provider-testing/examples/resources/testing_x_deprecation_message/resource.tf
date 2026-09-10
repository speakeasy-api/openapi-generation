resource "testing_x_deprecation_message" "my_xdeprecationmessage" {
  float32_request              = 1.29
  float32_request_and_response = 0.82
  float32_response             = 3.07
  float64_request              = 9.16
  float64_request_and_response = 4.34
  float64_response             = 5.14
  int32_request                = 6
  int32_request_and_response   = 3
  int32_response               = 6
  int64_request                = 4
  int64_request_and_response   = 0
  int64_response               = 10
  list_nested_request = [
    {
      list_nested_integer = 1
      list_nested_string  = "...my_list_nested_string..."
    }
  ]
  list_nested_request_and_response = [
    {
      list_nested_integer = 7
      list_nested_string  = "...my_list_nested_string..."
    }
  ]
  list_nested_response = [
    {
      list_nested_integer = 7
      list_nested_string  = "...my_list_nested_string..."
    }
  ]
  list_request = [
    "..."
  ]
  list_request_and_response = [
    "..."
  ]
  list_response = [
    "..."
  ]
  number_request              = 1.46
  number_request_and_response = 9.95
  number_response             = 6.17
  object_request = {
    object_integer = 1
    object_string  = "...my_object_string..."
  }
  object_request_and_response = {
    object_integer = 5
    object_string  = "...my_object_string..."
  }
  object_response = {
    object_integer = 0
    object_string  = "...my_object_string..."
  }
  set_nested_request = [
    {
      set_nested_integer = 6
      set_nested_string  = "...my_set_nested_string..."
    }
  ]
  set_nested_request_and_response = [
    {
      set_nested_integer = 6
      set_nested_string  = "...my_set_nested_string..."
    }
  ]
  set_nested_response = [
    {
      set_nested_integer = 8
      set_nested_string  = "...my_set_nested_string..."
    }
  ]
  set_request = [
    "..."
  ]
  set_request_and_response = [
    "..."
  ]
  set_response = [
    "..."
  ]
  string_request              = "...my_string_request..."
  string_request_and_response = "...my_string_request_and_response..."
  string_response             = "...my_string_response..."
}