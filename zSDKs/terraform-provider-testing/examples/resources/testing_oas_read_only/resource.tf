resource "testing_oas_read_only" "my_oasreadonly" {
  inline_bool    = true
  inline_float32 = 9.39
  inline_float64 = 4.99
  inline_int32   = 9.3
  inline_int64   = 3.85
  inline_integer = 2.62
  inline_list_nested = [
    {
      inline_string = "...my_inline_string..."
    }
  ]
  inline_list_string = [
    "..."
  ]
  inline_map_nested = {
    key = {
      inline_string = "...my_inline_string..."
    }
  }
  inline_map_string = {
    key = "value"
  }
  inline_number = 1.89
  inline_object_response_only = {
    inline_string = "...my_inline_string..."
  }
  inline_set_nested = [
    {
      inline_string = "...my_inline_string..."
    }
  ]
  inline_set_string = [
    "..."
  ]
  inline_string             = "...my_inline_string..."
  ref_bool_response_only    = false
  ref_float32_response_only = 7.65
  ref_float64_response_only = 0.17
  ref_int32_response_only   = 2.49
  ref_int64_response_only   = 6.81
  ref_integer_response_only = 3.39
  ref_list_nested_inline_request_only = [
    {
      # ...
    }
  ]
  ref_list_nested_inline_request_response = [
    {
      # ...
    }
  ]
  ref_list_nested_inline_response_only = [
    {
      inline_string = "...my_inline_string..."
    }
  ]
  ref_list_nested_ref_request_only = [
    {
      # ...
    }
  ]
  ref_list_nested_ref_request_response = [
    {
      # ...
    }
  ]
  ref_list_nested_ref_response_only = [
    {
      inline_string = "...my_inline_string..."
    }
  ]
  ref_list_string_response_only = [
    "..."
  ]
  ref_map_nested_inline_request_only = {
    key = {
      inline_string = "...my_inline_string..."
    }
  }
  ref_map_nested_inline_request_response = {
    inline_string = "...my_inline_string..."
  }
  ref_map_nested_inline_response_only = {
    inline_string = "...my_inline_string..."
  }
  ref_map_nested_ref_response_only = {
    key = {
      inline_string = "...my_inline_string..."
    }
  }
  ref_map_string_response_only = {
    key = "value"
  }
  ref_number_response_only = 2.78
  ref_object_request_only = {
    # ...
  }
  ref_object_request_response = {
    # ...
  }
  ref_object_response_only = {
    inline_string = "...my_inline_string..."
  }
  ref_set_nested_inline_request_only = [
    {
      # ...
    }
  ]
  ref_set_nested_inline_request_response = [
    {
      # ...
    }
  ]
  ref_set_nested_inline_response_only = [
    {
      inline_string = "...my_inline_string..."
    }
  ]
  ref_set_nested_ref_request_only = [
    {
      # ...
    }
  ]
  ref_set_nested_ref_request_response = [
    {
      # ...
    }
  ]
  ref_set_nested_ref_response_only = [
    {
      inline_string = "...my_inline_string..."
    }
  ]
  ref_set_string_response_only = [
    "..."
  ]
  ref_string_response_only = "...my_ref_string_response_only..."
}