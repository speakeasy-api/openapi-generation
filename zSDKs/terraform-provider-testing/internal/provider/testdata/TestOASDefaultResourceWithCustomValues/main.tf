variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_oas_default" "my_oasdefault" {
  # Test with custom values to ensure they override defaults
  string_request  = "custom_string"
  int32_request   = 42
  float32_request = 3.14
  number_request  = 2.718

  # Test nested object with custom values
  object_request = {
    object_integer = 100
    object_string  = "custom_object_string"
  }

  # Test list with custom values
  list_string_request  = ["item1", "item2"]
  list_integer_request = [10, 20, 30]

  # Test set with custom values
  set_string_request  = ["set_item1", "set_item2"]
  set_integer_request = [40, 50, 60]

  # Test nested list with custom values
  list_nested_request = [
    {
      list_nested_integer = 200
      list_nested_string  = "nested_item1"
    },
    {
      list_nested_integer = 300
      list_nested_string  = "nested_item2"
    }
  ]

  # Test nested set with custom values
  set_nested_request = [
    {
      set_nested_integer = 400
      set_nested_string  = "nested_set_item1"
    },
    {
      set_nested_integer = 500
      set_nested_string  = "nested_set_item2"
    }
  ]
}
