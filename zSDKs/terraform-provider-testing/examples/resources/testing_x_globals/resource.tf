resource "testing_x_globals" "my_xglobals" {
  global_boolean              = true
  global_boolean_with_default = true
  global_enum_float32         = 1.2
  global_enum_float64         = 1.2
  global_enum_int32           = 1
  global_enum_int64           = 3
  global_enum_integer         = 5
  global_enum_number          = 1.2
  global_enum_string          = "value1"
  global_float32              = 1.2
  global_float32_with_default = 1.2
  global_float64              = 3.4
  global_float64_with_default = 3.4
  global_int32                = 12
  global_int32_with_default   = 12
  global_int64                = 34
  global_int64_with_default   = 34
  global_integer              = 56
  global_integer_with_default = 56
  global_number               = 5.6
  global_number_with_default  = 5.6
  global_string               = "global_string_example"
  global_string_with_default  = "DEFAULT"
  local_string                = "...my_local_string..."
}