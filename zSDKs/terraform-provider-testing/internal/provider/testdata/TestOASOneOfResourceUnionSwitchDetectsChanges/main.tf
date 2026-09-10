terraform {
  required_providers {
    testing = {
      source = "hashicorp/testing"
    }
  }
}

variable "server_url" {
  type = string
}

variable "use_string" {
  type = bool
}

variable "use_number" {
  type = bool
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_oas_one_of" "test" {
  # Test union type switching between string and number primitives
  optional_inline_primitives_request_and_response = var.use_string ? {
    str = "test-string"
  } : (var.use_number ? {
    number = 42
  } : null)

  # Required fields - minimal values from example
  required_discriminator_request_and_response = {
    string = {
      nondiscriminator = "test"
      object_string = "test"
    }
  }

  required_discriminator_request_only = {
    string = {
      nondiscriminator = "test"
      object_string = "test"
    }
  }

  required_discriminator_request_optional_response = {
    allof = {
      nondiscriminator = "test"
      object_number = 1
      object_string = "test"
    }
  }

  required_inline_object_request_and_response = {
    oas_one_of_request_required_inline_object_request_and_response1 = {
      object_string = "test"
    }
  }

  required_inline_object_request_only = {
    required_inline_object_request_only2 = {
      object_number = 1
    }
  }

  required_inline_primitives_request_and_response = {
    number = 1
  }

  required_inline_primitives_request_only = {
    str = "test"
  }

  required_ref_object_request_and_response = {
    oas_one_of_object_string = {
      object_string = "test"
    }
  }

  required_ref_object_request_only = {
    oas_one_of_object_string = {
      object_string = "test"
    }
  }

  required_ref_primitives_request_and_response = {
    oas_one_of_number = 1
  }

  required_ref_primitives_request_only = {
    oas_one_of_number = 1
  }

  required_list_discriminator = [
    {
      string = {
        object_string = "test"
      }
    }
  ]

  required_set_discriminator = [
    {
      string = {
        object_string = "test"
      }
    }
  ]
}
