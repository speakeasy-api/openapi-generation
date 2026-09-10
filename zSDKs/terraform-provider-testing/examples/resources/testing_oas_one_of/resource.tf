resource "testing_oas_one_of" "my_oasoneof" {
  optional_discriminator_request_and_response = {
    allof = {
      nondiscriminator = "...my_nondiscriminator..."
      object_number    = 8.57
      object_string    = "...my_object_string..."
    }
  }
  optional_discriminator_request_only = {
    all_of = {
      nondiscriminator = "...my_nondiscriminator..."
      object_number    = 1.54
      object_string    = "...my_object_string..."
    }
  }
  optional_discriminator_request_required_response = {
    number = {
      nondiscriminator = "...my_nondiscriminator..."
      object_number    = 8.87
    }
  }
  optional_inline_object_request_and_response = {
    # ...
  }
  optional_inline_object_request_only = {
    optional_inline_object_request_only1 = {
      object_string              = "...my_object_string..."
      object_string_with_default = "default-value"
    }
  }
  optional_inline_primitives_request_and_response = {
    str = "optional_inline_primitives_request_and_response oneOf level example"
  }
  optional_inline_primitives_request_only = {
    number = 9.71
  }
  optional_inline_strings_request_and_response = "second"
  optional_inline_strings_request_only         = "first"
  optional_list_discriminator = [
    {
      number = {
        object_number = 2.31
      }
    }
  ]
  optional_ref_object_request_and_response = {
    oas_one_of_object_number = {
      object_number = 7.95
    }
  }
  optional_ref_object_request_only = {
    oas_one_of_object_number = {
      object_number = 7.89
    }
  }
  optional_ref_primitives_request_and_response = {
    oas_one_of_string = "OASOneOfString example"
  }
  optional_ref_primitives_request_only = {
    oas_one_of_string = "OASOneOfString example"
  }
  optional_set_discriminator = [
    {
      all_of = {
        object_number = 8.84
        object_string = "...my_object_string..."
      }
    }
  ]
  required_discriminator_request_and_response = {
    string = {
      nondiscriminator = "...my_nondiscriminator..."
      object_string    = "...my_object_string..."
    }
  }
  required_discriminator_request_only = {
    string = {
      nondiscriminator = "...my_nondiscriminator..."
      object_string    = "...my_object_string..."
    }
  }
  required_discriminator_request_optional_response = {
    allof = {
      nondiscriminator = "...my_nondiscriminator..."
      object_number    = 3.68
      object_string    = "...my_object_string..."
    }
  }
  required_inline_object_request_and_response = {
    oas_one_of_request_required_inline_object_request_and_response1 = {
      object_string              = "...my_object_string..."
      object_string_with_default = "default-value"
    }
  }
  required_inline_object_request_only = {
    required_inline_object_request_only2 = {
      object_number = 4.59
    }
  }
  required_inline_primitives_request_and_response = {
    number = 9.55
  }
  required_inline_primitives_request_only = {
    str = "required_inline_primitives_request_only oneOf level example"
  }
  required_list_discriminator = [
    {
      all_of = {
        object_number = 9.56
        object_string = "...my_object_string..."
      }
    }
  ]
  required_ref_object_request_and_response = {
    oas_one_of_object_string = {
      object_string = "...my_object_string..."
    }
  }
  required_ref_object_request_only = {
    oas_one_of_object_string = {
      object_string = "...my_object_string..."
    }
  }
  required_ref_primitives_request_and_response = {
    oas_one_of_number = 123
  }
  required_ref_primitives_request_only = {
    oas_one_of_number = 123
  }
  required_set_discriminator = [
    {
      string = {
        object_string = "...my_object_string..."
      }
    }
  ]
}