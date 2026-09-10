resource "testing_oas_write_only" "my_oaswriteonly" {
  array_of_objects_with_mixed_writeonly = [
    {
      api_key  = "...my_api_key..."
      password = "...my_password..."
      username = "...my_username..."
    }
  ]
  deep_nested_with_writeonly_leaf = {
    level1_name = "...my_level1_name..."
    level2 = {
      level2_name = "...my_level2_name..."
      level3 = {
        level3_name = "...my_level3_name..."
        level4 = {
          encryption_key = "...my_encryption_key..."
          level4_name    = "...my_level4_name..."
          secret_key     = "...my_secret_key..."
        }
      }
    }
  }
  map_of_objects_with_writeonly = {
    key = {
      api_token        = "...my_api_token..."
      db_password      = "...my_db_password..."
      environment_name = "...my_environment_name..."
    }
  }
  object_no_writeonly = {
    object_level_1_no_writeonly = {
      string_level_2_no_writeonly = "...my_string_level_2_no_writeonly..."
      string_level_2_writeonly    = "...my_string_level_2_writeonly..."
    }
    object_level_1_writeonly = {
      string_level_2_no_writeonly = "...my_string_level_2_no_writeonly..."
      string_level_2_writeonly    = "...my_string_level_2_writeonly..."
    }
    string_level_1_no_writeonly = "...my_string_level_1_no_writeonly..."
    string_level_1_writeonly    = "...my_string_level_1_writeonly..."
  }
  object_writeonly = {
    object_level_1_no_writeonly = {
      string_level_2_no_writeonly = "...my_string_level_2_no_writeonly..."
      string_level_2_writeonly    = "...my_string_level_2_writeonly..."
    }
    object_level_1_writeonly = {
      string_level_2_no_writeonly = "...my_string_level_2_no_writeonly..."
      string_level_2_writeonly    = "...my_string_level_2_writeonly..."
    }
    string_level_1_no_writeonly = "...my_string_level_1_no_writeonly..."
    string_level_1_writeonly    = "...my_string_level_1_writeonly..."
  }
  string_no_writeonly = "...my_string_no_writeonly..."
  string_writeonly    = "...my_string_writeonly..."
  union_with_partial_writeonly = {
    token = {
      created_at = "2021-08-06T05:39:21.170Z"
      token_id   = "...my_token_id..."
      token_name = "...my_token_name..."
    }
  }
}