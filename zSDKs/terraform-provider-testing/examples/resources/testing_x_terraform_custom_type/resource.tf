resource "testing_x_terraform_custom_type" "my_xterraformcustomtype" {
  bool    = false
  float32 = 9.7
  float64 = 3.59
  int32   = 1
  int64   = 4
  integer = 9
  list_string = [
    "..."
  ]
  map_string = {
    key = "value"
  }
  number = 6.88
  object = {
    object_string = "...my_object_string..."
  }
  object_with_underlying_custom_types = {
    bool    = false
    float32 = 7.21
    float64 = 9.04
    int32   = 5
    int64   = 0
    integer = 10
    list_string = [
      "..."
    ]
    map_string = {
      key = "value"
    }
    number = 8.23
    object = {
      object_string = "...my_object_string..."
    }
    set_string = [
      "..."
    ]
    string                   = "...my_string..."
    string_timetypes_rfc3339 = "2022-09-01T14:28:53.490Z"
  }
  set_string = [
    "..."
  ]
  string                   = "...my_string..."
  string_timetypes_rfc3339 = "2022-03-11T06:39:08.704Z"
}