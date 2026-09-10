resource "testing_oas_enum" "my_oasenum" {
  float32_request                            = 6.67
  float32_request_and_response               = 6.03
  float32_response                           = 5.34
  float64_request                            = 8.26
  float64_request_and_response               = 0.73
  float64_response                           = 7.87
  int32_enum_request_int32_enum_response     = 1
  int32_enum_request_int32_noenum_response   = 2
  int32_enum_request_int64_enum_response     = 2
  int32_enum_request_int64_noenum_response   = 1
  int32_enum_request_integer_enum_response   = 2
  int32_enum_request_integer_noenum_response = 2
  int32_long_enum                            = 1000000001
  int32_noenum_request_int32_enum_response   = 2
  int32_noenum_request_int64_enum_response   = 2
  int32_noenum_request_integer_enum_response = 2
  int64_enum_request_int32_enum_response     = 1
  int64_enum_request_int32_noenum_response   = 1
  int64_enum_request_int64_enum_response     = 2
  int64_enum_request_int64_noenum_response   = 1
  int64_enum_request_integer_enum_response   = 1
  int64_enum_request_integer_noenum_response = 2
  int64_noenum_request_int32_enum_response   = 1
  int64_noenum_request_int64_enum_response   = 2
  int64_noenum_request_integer_enum_response = 1
  list_int32_enum_request_int32_enum_response = [
    2
  ]
  list_int32_enum_request_int32_enum_response_nullable = [
    1
  ]
  list_int32_enum_request_int32_noenum_response = [
    2
  ]
  list_int32_enum_request_int32_noenum_response_nullable = [
    2
  ]
  list_int32_enum_request_int64_enum_response = [
    2
  ]
  list_int32_enum_request_int64_enum_response_nullable = [
    2
  ]
  list_int32_enum_request_int64_noenum_response = [
    2
  ]
  list_int32_enum_request_int64_noenum_response_nullable = [
    2
  ]
  list_int32_enum_request_integer_enum_response = [
    1
  ]
  list_int32_enum_request_integer_enum_response_nullable = [
    1
  ]
  list_int32_enum_request_integer_noenum_response = [
    1
  ]
  list_int32_enum_request_integer_noenum_response_nullable = [
    1
  ]
  list_int32_noenum_request_int32_enum_response = [
    2
  ]
  list_int32_noenum_request_int32_enum_response_nullable = [
    1
  ]
  list_int32_noenum_request_int64_enum_response = [
    2
  ]
  list_int32_noenum_request_int64_enum_response_nullable = [
    1
  ]
  list_int32_noenum_request_integer_enum_response = [
    1
  ]
  list_int32_noenum_request_integer_enum_response_nullable = [
    2
  ]
  list_int64_enum_request_int32_enum_response = [
    2
  ]
  list_int64_enum_request_int32_enum_response_nullable = [
    1
  ]
  list_int64_enum_request_int32_noenum_response = [
    1
  ]
  list_int64_enum_request_int32_noenum_response_nullable = [
    2
  ]
  list_int64_enum_request_int64_enum_response = [
    1
  ]
  list_int64_enum_request_int64_enum_response_nullable = [
    1
  ]
  list_int64_enum_request_int64_noenum_response = [
    1
  ]
  list_int64_enum_request_int64_noenum_response_nullable = [
    2
  ]
  list_int64_enum_request_integer_enum_response = [
    2
  ]
  list_int64_enum_request_integer_enum_response_nullable = [
    2
  ]
  list_int64_enum_request_integer_noenum_response = [
    2
  ]
  list_int64_enum_request_integer_noenum_response_nullable = [
    2
  ]
  list_int64_noenum_request_int32_enum_response = [
    1
  ]
  list_int64_noenum_request_int32_enum_response_nullable = [
    1
  ]
  list_int64_noenum_request_int64_enum_response = [
    2
  ]
  list_int64_noenum_request_int64_enum_response_nullable = [
    2
  ]
  list_int64_noenum_request_integer_enum_response = [
    1
  ]
  list_int64_noenum_request_integer_enum_response_nullable = [
    1
  ]
  list_integer_request = [
    2
  ]
  list_integer_request_and_response = [
    2
  ]
  list_integer_response = [
    2
  ]
  list_nested_request = [
    {
      list_nested_integer = 2
      list_nested_string  = "one"
    }
  ]
  list_nested_request_and_response = [
    {
      list_nested_integer = 2
      list_nested_string  = "two"
    }
  ]
  list_nested_response = [
    {
      list_nested_integer = 2
      list_nested_string  = "two"
    }
  ]
  list_string_request = [
    "one"
  ]
  list_string_request_and_response = [
    "one"
  ]
  list_string_response = [
    "two"
  ]
  number_request              = 4.36
  number_request_and_response = 4.32
  number_response             = 2.62
  object_request = {
    object_integer = 2
    object_string  = "one"
  }
  object_request_and_response = {
    object_integer = 1
    object_string  = "two"
  }
  object_response = {
    object_integer = 1
    object_string  = "two"
  }
  set_int32_enum_request_int32_enum_response = [
    1
  ]
  set_int32_enum_request_int32_enum_response_nullable = [
    1
  ]
  set_int32_enum_request_int32_noenum_response = [
    1
  ]
  set_int32_enum_request_int32_noenum_response_nullable = [
    2
  ]
  set_int32_enum_request_int64_enum_response = [
    1
  ]
  set_int32_enum_request_int64_enum_response_nullable = [
    1
  ]
  set_int32_enum_request_int64_noenum_response = [
    1
  ]
  set_int32_enum_request_int64_noenum_response_nullable = [
    2
  ]
  set_int32_enum_request_integer_enum_response = [
    2
  ]
  set_int32_enum_request_integer_enum_response_nullable = [
    1
  ]
  set_int32_enum_request_integer_noenum_response = [
    2
  ]
  set_int32_enum_request_integer_noenum_response_nullable = [
    2
  ]
  set_int32_noenum_request_int32_enum_response = [
    2
  ]
  set_int32_noenum_request_int32_enum_response_nullable = [
    1
  ]
  set_int32_noenum_request_int64_enum_response = [
    1
  ]
  set_int32_noenum_request_int64_enum_response_nullable = [
    1
  ]
  set_int32_noenum_request_integer_enum_response = [
    2
  ]
  set_int32_noenum_request_integer_enum_response_nullable = [
    2
  ]
  set_int64_enum_request_int32_enum_response = [
    1
  ]
  set_int64_enum_request_int32_enum_response_nullable = [
    2
  ]
  set_int64_enum_request_int32_noenum_response = [
    1
  ]
  set_int64_enum_request_int32_noenum_response_nullable = [
    1
  ]
  set_int64_enum_request_int64_enum_response = [
    2
  ]
  set_int64_enum_request_int64_enum_response_nullable = [
    1
  ]
  set_int64_enum_request_int64_noenum_response = [
    1
  ]
  set_int64_enum_request_int64_noenum_response_nullable = [
    2
  ]
  set_int64_enum_request_integer_enum_response = [
    2
  ]
  set_int64_enum_request_integer_enum_response_nullable = [
    2
  ]
  set_int64_enum_request_integer_noenum_response = [
    1
  ]
  set_int64_enum_request_integer_noenum_response_nullable = [
    2
  ]
  set_int64_noenum_request_int32_enum_response = [
    1
  ]
  set_int64_noenum_request_int32_enum_response_nullable = [
    1
  ]
  set_int64_noenum_request_int64_enum_response = [
    1
  ]
  set_int64_noenum_request_int64_enum_response_nullable = [
    1
  ]
  set_int64_noenum_request_integer_enum_response = [
    2
  ]
  set_int64_noenum_request_integer_enum_response_nullable = [
    1
  ]
  set_integer_request = [
    2
  ]
  set_integer_request_and_response = [
    2
  ]
  set_integer_response = [
    1
  ]
  set_nested_request = [
    {
      set_nested_integer = 2
      set_nested_string  = "one"
    }
  ]
  set_nested_request_and_response = [
    {
      set_nested_integer = 1
      set_nested_string  = "two"
    }
  ]
  set_nested_response = [
    {
      set_nested_integer = 1
      set_nested_string  = "one"
    }
  ]
  set_string_request = [
    "one"
  ]
  set_string_request_and_response = [
    "two"
  ]
  set_string_response = [
    "two"
  ]
  string_request              = "two"
  string_request_and_response = "two"
  string_response             = "one"
}