resource "testing_framework_type" "my_frameworktype" {
  bool    = true
  float32 = 0.51
  float64 = 3.05
  int32   = 4
  int64   = 1
  integer = 10
  list_any = [
    "{ \"see\": \"documentation\" }"
  ]
  list_bool = [
    false
  ]
  list_bool_null = [
    true
  ]
  list_enum_float32 = [
    4.84
  ]
  list_enum_float64 = [
    4.17
  ]
  list_enum_int32 = [
    2
  ]
  list_enum_int64 = [
    1
  ]
  list_enum_integer = [
    2
  ]
  list_enum_string = [
    "value1"
  ]
  list_float32 = [
    4.18
  ]
  list_float32_null = [
    6.74
  ]
  list_float64 = [
    2.98
  ]
  list_float64_null = [
    5.44
  ]
  list_int32 = [
    0
  ]
  list_int32_null = [
    0
  ]
  list_int64 = [
    5
  ]
  list_int64_null = [
    10
  ]
  list_integer = [
    2
  ]
  list_integer_null = [
    2
  ]
  list_list_float32 = [
    [
      # ...
    ]
  ]
  list_list_float64 = [
    [
      # ...
    ]
  ]
  list_list_int32 = [
    [
      # ...
    ]
  ]
  list_list_int64 = [
    [
      # ...
    ]
  ]
  list_list_integer = [
    [
      # ...
    ]
  ]
  list_list_number = [
    [
      # ...
    ]
  ]
  list_list_oneof_number_string = [
    [
      # ...
    ]
  ]
  list_list_oneof_number_string_null = [
    [
      # ...
    ]
  ]
  list_list_oneof_number_string_nullable = [
    [
      # ...
    ]
  ]
  list_list_string = [
    [
      # ...
    ]
  ]
  list_map_any = [
    {
      # ...
    }
  ]
  list_map_float32 = [
    {
      # ...
    }
  ]
  list_map_float64 = [
    {
      # ...
    }
  ]
  list_map_int32 = [
    {
      # ...
    }
  ]
  list_map_int64 = [
    {
      # ...
    }
  ]
  list_map_integer = [
    {
      # ...
    }
  ]
  list_map_number = [
    {
      # ...
    }
  ]
  list_map_object = [
    {
      # ...
    }
  ]
  list_map_object_empty = [
    {
      # ...
    }
  ]
  list_map_string = [
    {
      # ...
    }
  ]
  list_nested_empty = [
    {
      # ...
    }
  ]
  list_nested_string = [
    {
      list_nested_string = "...my_list_nested_string..."
    }
  ]
  list_null = [
    "..."
  ]
  list_null_enum_string = [
    "value2"
  ]
  list_null_list_string = [
    [
      # ...
    ]
  ]
  list_null_object = [
    {
      list_object_string = "...my_list_object_string..."
    }
  ]
  list_null_oneof_number_string = [
    {
      number = 2.77
    }
  ]
  list_nullable = [
    "..."
  ]
  list_nullable_enum_string = [
    "value1"
  ]
  list_nullable_list_string = [
    [
      # ...
    ]
  ]
  list_nullable_object = [
    {
      list_object_string = "...my_list_object_string..."
    }
  ]
  list_nullable_oneof_number_string = [
    {
      str = "...my_str..."
    }
  ]
  list_number = [
    8.12
  ]
  list_number_null = [
    5.45
  ]
  list_object = [
    {
      list_object_string = "...my_list_object_string..."
    }
  ]
  list_oneof_number_string = [
    {
      number = 5.69
    }
  ]
  list_oneof_number_string_null = [
    {
      number = 6.59
    }
  ]
  list_oneof_number_string_nullable = [
    {
      str = "...my_str..."
    }
  ]
  list_set_float32 = [
    [
      # ...
    ]
  ]
  list_set_float64 = [
    [
      # ...
    ]
  ]
  list_set_int32 = [
    [
      # ...
    ]
  ]
  list_set_int64 = [
    [
      # ...
    ]
  ]
  list_set_integer = [
    [
      # ...
    ]
  ]
  list_set_number = [
    [
      # ...
    ]
  ]
  list_set_string = [
    [
      # ...
    ]
  ]
  list_string = [
    "..."
  ]
  list_string_null = [
    "..."
  ]
  map_any = {
    key = jsonencode("value")
  }
  map_bool = {
    key = true
  }
  map_bool_null = {
    key = true
  }
  map_float32 = {
    key = 1.23
  }
  map_float64 = {
    key = 1.23
  }
  map_int32 = {
    key = 123
  }
  map_int64 = {
    key = 123
  }
  map_integer = {
    key = 123
  }
  map_list_bool = {
    key = [
      # ...
    ]
  }
  map_list_float32 = {
    key = [
      # ...
    ]
  }
  map_list_float64 = {
    key = [
      # ...
    ]
  }
  map_list_int32 = {
    key = [
      # ...
    ]
  }
  map_list_int64 = {
    key = [
      # ...
    ]
  }
  map_list_integer = {
    key = [
      # ...
    ]
  }
  map_list_number = {
    key = [
      # ...
    ]
  }
  map_list_string = {
    key = [
      # ...
    ]
  }
  map_map_bool = {
    key = {
      # ...
    }
  }
  map_map_float32 = {
    key = {
      # ...
    }
  }
  map_map_float64 = {
    key = {
      # ...
    }
  }
  map_map_int32 = {
    key = {
      # ...
    }
  }
  map_map_int64 = {
    key = {
      # ...
    }
  }
  map_map_integer = {
    key = {
      # ...
    }
  }
  map_map_string = {
    key = {
      # ...
    }
  }
  map_null_any = {
    key = jsonencode("value")
  }
  map_null_bool = {
    key = true
  }
  map_null_float32 = {
    key = 1.23
  }
  map_null_float64 = {
    key = 1.23
  }
  map_null_int32 = {
    key = 123
  }
  map_null_int64 = {
    key = 123
  }
  map_null_integer = {
    key = 123
  }
  map_null_list_string = {
    key = [
      # ...
    ]
  }
  map_null_number = {
    key = 1.23
  }
  map_null_set_string = {
    key = [
      # ...
    ]
  }
  map_null_string = {
    key = "value"
  }
  map_nullable_any = {
    key = jsonencode("value")
  }
  map_nullable_bool = {
    key = true
  }
  map_nullable_float32 = {
    key = 1.23
  }
  map_nullable_float64 = {
    key = 1.23
  }
  map_nullable_int32 = {
    key = 123
  }
  map_nullable_int64 = {
    key = 123
  }
  map_nullable_integer = {
    key = 123
  }
  map_nullable_list_string = {
    key = [
      # ...
    ]
  }
  map_nullable_number = {
    key = 1.23
  }
  map_nullable_set_string = {
    key = [
      # ...
    ]
  }
  map_nullable_string = {
    key = "value"
  }
  map_number = {
    key = 1.23
  }
  map_set_bool = {
    key = [
      # ...
    ]
  }
  map_set_float32 = {
    key = [
      # ...
    ]
  }
  map_set_float64 = {
    key = [
      # ...
    ]
  }
  map_set_int32 = {
    key = [
      # ...
    ]
  }
  map_set_int64 = {
    key = [
      # ...
    ]
  }
  map_set_integer = {
    key = [
      # ...
    ]
  }
  map_set_number = {
    key = [
      # ...
    ]
  }
  map_set_string = {
    key = [
      # ...
    ]
  }
  map_string = {
    key = "value"
  }
  map_string_null = {
    key = "value"
  }
  map_string_null_request_string_response = {
    key = "value"
  }
  map_string_request_string_nullable_response = {
    key = "value"
  }
  number = 4.16
  object = {
    object_string = "...my_object_string..."
  }
  oneof_number_string = {
    str = "...my_str..."
  }
  oneof_number_string_null = {
    str = "...my_str..."
  }
  oneof_number_string_nullable = {
    number = 6.12
  }
  set_any = [
    "{ \"see\": \"documentation\" }"
  ]
  set_bool = [
    false
  ]
  set_bool_null = [
    true
  ]
  set_enum_float32 = [
    3.87
  ]
  set_enum_float64 = [
    4.61
  ]
  set_enum_int32 = [
    2
  ]
  set_enum_int64 = [
    1
  ]
  set_enum_integer = [
    2
  ]
  set_enum_string = [
    "value2"
  ]
  set_float32 = [
    2.82
  ]
  set_float32_null = [
    8.85
  ]
  set_float64 = [
    0.56
  ]
  set_float64_null = [
    8.12
  ]
  set_int32 = [
    7
  ]
  set_int32_null = [
    9
  ]
  set_int64 = [
    1
  ]
  set_int64_null = [
    3
  ]
  set_integer = [
    3
  ]
  set_integer_null = [
    0
  ]
  set_list_float32 = [
    [
      # ...
    ]
  ]
  set_list_float64 = [
    [
      # ...
    ]
  ]
  set_list_int32 = [
    [
      # ...
    ]
  ]
  set_list_int64 = [
    [
      # ...
    ]
  ]
  set_list_integer = [
    [
      # ...
    ]
  ]
  set_list_number = [
    [
      # ...
    ]
  ]
  set_list_string = [
    [
      # ...
    ]
  ]
  set_map_any = [
    {
      # ...
    }
  ]
  set_map_float32 = [
    {
      # ...
    }
  ]
  set_map_float64 = [
    {
      # ...
    }
  ]
  set_map_int32 = [
    {
      # ...
    }
  ]
  set_map_int64 = [
    {
      # ...
    }
  ]
  set_map_integer = [
    {
      # ...
    }
  ]
  set_map_number = [
    {
      # ...
    }
  ]
  set_map_object = [
    {
      # ...
    }
  ]
  set_map_object_empty = [
    {
      # ...
    }
  ]
  set_map_string = [
    {
      # ...
    }
  ]
  set_nested_empty = [
    {
      # ...
    }
  ]
  set_nested_string = [
    {
      set_nested_string = "...my_set_nested_string..."
    }
  ]
  set_null = [
    "..."
  ]
  set_null_enum_string = [
    "value2"
  ]
  set_null_list_string = [
    [
      # ...
    ]
  ]
  set_null_object = [
    {
      set_object_string = "...my_set_object_string..."
    }
  ]
  set_null_oneof_number_string = [
    {
      number = 6.05
    }
  ]
  set_nullable = [
    "..."
  ]
  set_nullable_enum_string = [
    "value1"
  ]
  set_nullable_list_string = [
    [
      # ...
    ]
  ]
  set_nullable_object = [
    {
      set_object_string = "...my_set_object_string..."
    }
  ]
  set_nullable_oneof_number_string = [
    {
      str = "...my_str..."
    }
  ]
  set_number = [
    2.43
  ]
  set_number_null = [
    3.89
  ]
  set_object = [
    {
      set_object_string = "...my_set_object_string..."
    }
  ]
  set_oneof_number_string = [
    {
      number = 5.83
    }
  ]
  set_oneof_number_string_null = [
    {
      number = 0.22
    }
  ]
  set_oneof_number_string_nullable = [
    {
      str = "...my_str..."
    }
  ]
  set_set_float32 = [
    [
      # ...
    ]
  ]
  set_set_float64 = [
    [
      # ...
    ]
  ]
  set_set_int32 = [
    [
      # ...
    ]
  ]
  set_set_int64 = [
    [
      # ...
    ]
  ]
  set_set_integer = [
    [
      # ...
    ]
  ]
  set_set_number = [
    [
      # ...
    ]
  ]
  set_set_oneof_number_string = [
    [
      # ...
    ]
  ]
  set_set_oneof_number_string_null = [
    [
      # ...
    ]
  ]
  set_set_oneof_number_string_nullable = [
    [
      # ...
    ]
  ]
  set_set_string = [
    [
      # ...
    ]
  ]
  set_string = [
    "..."
  ]
  set_string_null = [
    "..."
  ]
  string           = "...my_string..."
  string_date      = "2022-10-21"
  string_date_time = "2022-12-30T22:07:49.394Z"
}