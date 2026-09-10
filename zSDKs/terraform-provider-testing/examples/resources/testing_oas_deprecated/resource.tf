resource "testing_oas_deprecated" "my_oasdeprecated" {
  float32_request              = 9.32
  float32_request_and_response = 2.45
  float32_response             = 1.41
  float64_request              = 6.18
  float64_request_and_response = 8.19
  float64_response             = 9.89
  int32_request                = 4
  int32_request_and_response   = 9
  int32_response               = 2
  int64_request                = 4
  int64_request_and_response   = 9
  int64_response               = 8
  list_nested_request = [
    {
      list_nested_integer = 4
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
      list_nested_integer = 8
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
  number_request              = 0.24
  number_request_and_response = 4.46
  number_response             = 1.2
  object_request = {
    object_integer = 0
    object_string  = "...my_object_string..."
  }
  object_request_and_response = {
    object_integer = 3
    object_string  = "...my_object_string..."
  }
  object_response = {
    object_integer = 3
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
      set_nested_integer = 8
      set_nested_string  = "...my_set_nested_string..."
    }
  ]
  set_nested_response = [
    {
      set_nested_integer = 9
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