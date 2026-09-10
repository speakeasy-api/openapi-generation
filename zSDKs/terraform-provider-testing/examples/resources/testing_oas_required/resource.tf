resource "testing_oas_required" "my_oasrequired" {
  inline_bool    = true
  inline_float32 = 9.3
  inline_float64 = 0.68
  inline_int32   = 0.41
  inline_int64   = 7.86
  inline_integer = 4.61
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
  inline_number = 3.01
  inline_object = {
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
  inline_string = "...my_inline_string..."
  ref_bool      = false
  ref_float32   = 2.33
  ref_float64   = 4.81
  ref_int32     = 9.35
  ref_int64     = 4.62
  ref_integer   = 3.38
  ref_list_nested_inline = [
    {
      inline_string = "...my_inline_string..."
    }
  ]
  ref_list_nested_ref = [
    {
      inline_string = "...my_inline_string..."
    }
  ]
  ref_list_string = [
    "..."
  ]
  ref_map_nested_inline = {
    key = {
      inline_string = "...my_inline_string..."
    }
  }
  ref_map_nested_ref = {
    key = {
      inline_string = "...my_inline_string..."
    }
  }
  ref_map_string = {
    key = "value"
  }
  ref_number = 8.59
  ref_object = {
    inline_string = "...my_inline_string..."
  }
  ref_set_nested_inline = [
    {
      inline_string = "...my_inline_string..."
    }
  ]
  ref_set_nested_ref = [
    {
      inline_string = "...my_inline_string..."
    }
  ]
  ref_set_string = [
    "..."
  ]
  ref_string = "...my_ref_string..."
}