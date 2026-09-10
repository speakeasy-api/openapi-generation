resource "testing_x_terraform_custom_default" "my_xterraformcustomdefault" {
  bool    = true
  float32 = 7.89
  float64 = 5.31
  int32   = 1
  int64   = 1
  integer = 3
  list_string = [
    "..."
  ]
  map_string = {
    key = "value"
  }
  number = 5.12
  object = {
    object_string = "...my_object_string..."
  }
  object_with_underlying_custom_defaults = {
    bool    = false
    float32 = 3.48
    float64 = 4.18
    int32   = 6
    int64   = 8
    integer = 9
    list_string = [
      "..."
    ]
    map_string = {
      key = "value"
    }
    number = 5.75
    object = {
      object_string = "...my_object_string..."
    }
    set_string = [
      "..."
    ]
    string = "...my_string..."
  }
  set_string = [
    "..."
  ]
  string = "...my_string..."
}