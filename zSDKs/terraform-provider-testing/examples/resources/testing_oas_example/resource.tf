resource "testing_oas_example" "my_oasexample" {
  float32_request              = 1.1
  float32_request_and_response = 12.12
  float32_response             = 2.2
  float64_request              = 1.1
  float64_request_and_response = 12.12
  float64_response             = 2.2
  int32_request                = 1
  int32_request_and_response   = 12
  int32_response               = 2
  int64_request                = 1
  int64_request_and_response   = 12
  int64_response               = 2
  list_integer_request = [
    1,
  ]
  list_integer_request_and_response = [
    12,
  ]
  list_integer_request_and_response_item = [
    12
  ]
  list_integer_request_item = [
    1
  ]
  list_integer_response = [
    2,
  ]
  list_integer_response_item = [
    2
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
      list_nested_integer = 2
      list_nested_string  = "two"
    }
  ]
  list_string_request = [
    "one",
  ]
  list_string_request_and_response = [
    "three",
  ]
  list_string_request_and_response_item = [
    "three"
  ]
  list_string_request_item = [
    "one"
  ]
  list_string_response = [
    "two",
  ]
  list_string_response_item = [
    "two"
  ]
  number_request              = 1.1
  number_request_and_response = 12.12
  number_response             = 2.2
  object_request = {
    object_integer = 1
    object_string  = "one"
  }
  object_request_and_response = {
    object_integer = 12
    object_string  = "three"
  }
  object_response = {
    object_integer = 2
    object_string  = "two"
  }
  set_integer_request = [
    1,
  ]
  set_integer_request_and_response = [
    12,
  ]
  set_integer_request_and_response_item = [
    12
  ]
  set_integer_request_item = [
    1
  ]
  set_integer_response = [
    2,
  ]
  set_integer_response_item = [
    2
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
      set_nested_integer = 2
      set_nested_string  = "two"
    }
  ]
  set_string_request = [
    "one",
  ]
  set_string_request_and_response = [
    "three",
  ]
  set_string_request_and_response_item = [
    "three"
  ]
  set_string_request_item = [
    "one"
  ]
  set_string_response = [
    "two",
  ]
  set_string_response_item = [
    "two"
  ]
  string_interpolation_sequence      = "`.$${env[\"WORKER_ID\"]}.$${__format}`"
  string_request                     = "one"
  string_request_and_response        = "three"
  string_response                    = "two"
  string_template_directive_sequence = "%%{if condition}true%%{endif}"
}