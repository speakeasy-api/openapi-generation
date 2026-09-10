resource "testing_oas_default" "my_oasdefault" {
  bool_request                 = false
  bool_request_and_response    = true
  bool_response                = true
  float32_request              = 1.1
  float32_request_and_response = 12.12
  float32_response             = 4.65
  float64_request              = 1.1
  float64_request_and_response = 12.12
  float64_response             = 5.44
  int32_request                = 1
  int32_request_and_response   = 12
  int32_response               = 8
  int64_request                = 1
  int64_request_and_response   = 12
  int64_response               = 3
  list_bool_request_and_response_empty = [
    true
  ]
  list_bool_request_and_response_value = [
    true
  ]
  list_bool_request_empty = [
    true
  ]
  list_bool_request_value = [
    false
  ]
  list_bool_response_empty = [
    false
  ]
  list_bool_response_value = [
    true
  ]
  list_float32_request_and_response_empty = [
    2.38
  ]
  list_float32_request_and_response_value = [
    3.13
  ]
  list_float32_request_empty = [
    0.1
  ]
  list_float32_request_value = [
    7.52
  ]
  list_float32_response_empty = [
    3.39
  ]
  list_float32_response_value = [
    1.27
  ]
  list_float64_request_and_response_empty = [
    1.57
  ]
  list_float64_request_and_response_value = [
    0.65
  ]
  list_float64_request_empty = [
    2.21
  ]
  list_float64_request_value = [
    6.53
  ]
  list_float64_response_empty = [
    9.67
  ]
  list_float64_response_value = [
    4.96
  ]
  list_int32_request_and_response_empty = [
    8.18
  ]
  list_int32_request_and_response_value = [
    3.23
  ]
  list_int32_request_empty = [
    1.76
  ]
  list_int32_request_value = [
    4.36
  ]
  list_int32_response_empty = [
    7.49
  ]
  list_int32_response_value = [
    0.46
  ]
  list_int64_request_and_response_empty = [
    9.63
  ]
  list_int64_request_and_response_value = [
    1.22
  ]
  list_int64_request_empty = [
    5.34
  ]
  list_int64_request_value = [
    4.22
  ]
  list_int64_response_empty = [
    2.63
  ]
  list_int64_response_value = [
    6.51
  ]
  list_integer_request_and_response_empty = [
    6
  ]
  list_integer_request_and_response_value = [
    1
  ]
  list_integer_request_empty = [
    4
  ]
  list_integer_request_value = [
    1
  ]
  list_integer_response_empty = [
    3
  ]
  list_integer_response_value = [
    0
  ]
  list_nested_request = [
    {
      list_nested_integer = 1
      list_nested_string  = "one"
    }
  ]
  list_nested_request_and_response = [
    {
      list_nested_integer = 12
      list_nested_string  = "three"
    }
  ]
  list_nested_response = [
    {
      list_nested_integer = 8
      list_nested_string  = "...my_list_nested_string..."
    }
  ]
  list_string_request_and_response_empty = [
    "..."
  ]
  list_string_request_and_response_value = [
    "..."
  ]
  list_string_request_empty = [
    "..."
  ]
  list_string_request_value = [
    "..."
  ]
  list_string_response_empty = [
    "..."
  ]
  list_string_response_value = [
    "..."
  ]
  map_bool_request_and_response_empty = {
    key = true
  }
  map_bool_request_and_response_value = {
    key = true
  }
  map_bool_request_empty = {
    key = true
  }
  map_bool_request_value = {
    key = true
  }
  map_bool_response_empty = {
    key = true
  }
  map_bool_response_value = {
    key = true
  }
  map_float32_request_and_response_empty = {
    key = 1.23
  }
  map_float32_request_and_response_value = {
    key = 1.23
  }
  map_float32_request_empty = {
    key = 1.23
  }
  map_float32_request_value = {
    key = 1.23
  }
  map_float32_response_empty = {
    key = 1.23
  }
  map_float32_response_value = {
    key = 1.23
  }
  map_float64_request_and_response_empty = {
    key = 1.23
  }
  map_float64_request_and_response_value = {
    key = 1.23
  }
  map_float64_request_empty = {
    key = 1.23
  }
  map_float64_request_value = {
    key = 1.23
  }
  map_float64_response_empty = {
    key = 1.23
  }
  map_float64_response_value = {
    key = 1.23
  }
  map_int32_request_and_response_empty = {
    key = 1.23
  }
  map_int32_request_and_response_value = {
    key = 1.23
  }
  map_int32_request_empty = {
    key = 1.23
  }
  map_int32_request_value = {
    key = 1.23
  }
  map_int32_response_empty = {
    key = 1.23
  }
  map_int32_response_value = {
    key = 1.23
  }
  map_int64_request_and_response_empty = {
    key = 1.23
  }
  map_int64_request_and_response_value = {
    key = 1.23
  }
  map_int64_request_empty = {
    key = 1.23
  }
  map_int64_request_value = {
    key = 1.23
  }
  map_int64_response_empty = {
    key = 1.23
  }
  map_int64_response_value = {
    key = 1.23
  }
  map_integer_request_and_response_empty = {
    key = 123
  }
  map_integer_request_and_response_value = {
    key = 123
  }
  map_integer_request_empty = {
    key = 123
  }
  map_integer_request_value = {
    key = 123
  }
  map_integer_response_empty = {
    key = 123
  }
  map_integer_response_value = {
    key = 123
  }
  map_string_request_and_response_empty = {
    key = "value"
  }
  map_string_request_and_response_value = {
    key = "value"
  }
  map_string_request_empty = {
    key = "value"
  }
  map_string_request_value = {
    key = "value"
  }
  map_string_response_empty = {
    key = "value"
  }
  map_string_response_value = {
    key = "value"
  }
  number_request              = 1.1
  number_request_and_response = 12.12
  number_response             = 6.23
  object_request_and_response_null = {
    object_integer = 7
    object_list_string = [
      "..."
    ]
    object_map_string = {
      key = "value"
    }
    object_set_string = [
      "..."
    ]
    object_string = "...my_object_string..."
  }
  object_request_and_response_null_and_property = {
    object_integer = 12
    object_string  = "three"
  }
  object_request_and_response_object = {
    object_integer = 12
    object_string  = "three"
  }
  object_request_and_response_property = {
    object_integer = 12
    object_string  = "three"
  }
  object_request_null = {
    object_integer = 1
    object_list_string = [
      "..."
    ]
    object_map_string = {
      key = "value"
    }
    object_set_string = [
      "..."
    ]
    object_string = "...my_object_string..."
  }
  object_request_null_and_property = {
    object_integer = 1
    object_string  = "one"
  }
  object_request_object = {
    object_integer = 0
    object_string  = "...my_object_string..."
  }
  object_request_property = {
    object_integer = 1
    object_string  = "one"
  }
  object_response_null = {
    object_integer = 10
    object_list_string = [
      "..."
    ]
    object_map_string = {
      key = "value"
    }
    object_set_string = [
      "..."
    ]
    object_string = "...my_object_string..."
  }
  object_response_null_and_property = {
    object_integer = 9
    object_string  = "...my_object_string..."
  }
  object_response_object = {
    object_integer = 5
    object_string  = "...my_object_string..."
  }
  object_response_property = {
    object_integer = 5
    object_string  = "...my_object_string..."
  }
  set_bool_request_and_response_empty = [
    true
  ]
  set_bool_request_and_response_value = [
    false
  ]
  set_bool_request_empty = [
    false
  ]
  set_bool_request_value = [
    false
  ]
  set_bool_response_empty = [
    true
  ]
  set_bool_response_value = [
    false
  ]
  set_float32_request_and_response_empty = [
    2.49
  ]
  set_float32_request_and_response_value = [
    8.07
  ]
  set_float32_request_empty = [
    5.6
  ]
  set_float32_request_value = [
    9.96
  ]
  set_float32_response_empty = [
    0.89
  ]
  set_float32_response_value = [
    8.04
  ]
  set_float64_request_and_response_empty = [
    8.37
  ]
  set_float64_request_and_response_value = [
    0.21
  ]
  set_float64_request_empty = [
    2.64
  ]
  set_float64_request_value = [
    7.07
  ]
  set_float64_response_empty = [
    3.64
  ]
  set_float64_response_value = [
    2.84
  ]
  set_int32_request_and_response_empty = [
    3.4
  ]
  set_int32_request_and_response_value = [
    3.88
  ]
  set_int32_request_empty = [
    9.99
  ]
  set_int32_request_value = [
    4.1
  ]
  set_int32_response_empty = [
    7.85
  ]
  set_int32_response_value = [
    4.63
  ]
  set_int64_request_and_response_empty = [
    3.05
  ]
  set_int64_request_and_response_value = [
    6.95
  ]
  set_int64_request_empty = [
    3.18
  ]
  set_int64_request_value = [
    1.04
  ]
  set_int64_response_empty = [
    4.66
  ]
  set_int64_response_value = [
    5.09
  ]
  set_integer_request_and_response_empty = [
    3
  ]
  set_integer_request_and_response_value = [
    9
  ]
  set_integer_request_empty = [
    3
  ]
  set_integer_request_value = [
    8
  ]
  set_integer_response_empty = [
    9
  ]
  set_integer_response_value = [
    1
  ]
  set_nested_request = [
    {
      set_nested_integer = 1
      set_nested_string  = "one"
    }
  ]
  set_nested_request_and_response = [
    {
      set_nested_integer = 12
      set_nested_string  = "three"
    }
  ]
  set_nested_response = [
    {
      set_nested_integer = 3
      set_nested_string  = "...my_set_nested_string..."
    }
  ]
  set_string_request_and_response_empty = [
    "..."
  ]
  set_string_request_and_response_value = [
    "..."
  ]
  set_string_request_empty = [
    "..."
  ]
  set_string_request_value = [
    "..."
  ]
  set_string_response_empty = [
    "..."
  ]
  set_string_response_value = [
    "..."
  ]
  string_request              = "one"
  string_request_and_response = "three"
  string_response             = "...my_string_response..."
  string_sanitization         = "Example value with {{ templating }} character sequences"
}