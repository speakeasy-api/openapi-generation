resource "testing_x_unknown_values" "my_xunknownvalues" {
  closed_enum_string                 = "off"
  optional_computed_open_enum_int64  = 1
  optional_computed_open_enum_string = "small"
  optional_open_enum_string          = "red"
  required_open_enum_string          = "gamma"
}