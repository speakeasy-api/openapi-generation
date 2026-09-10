resource "testing_name_shadowing" "my_nameshadowing" {
  big = "...my_big..."
  container_object_for_json_array = {
    json = [
      {
        key   = "...my_key..."
        value = "...my_value..."
      }
    ]
  }
  json = "...my_json..."
  operations = {
    name = "...my_name..."
    operations = {
      id    = "...my_id..."
      label = "...my_label..."
    }
    status = "...my_status..."
  }
  time                        = "...my_time..."
  types                       = "...my_types..."
  z_requires_big_import       = 0.11
  z_requires_json_import      = "{ \"see\": \"documentation\" }"
  z_requires_sdk_types_import = "2020-08-24"
  z_requires_time_import      = "2022-03-22T19:30:46.099Z"
}