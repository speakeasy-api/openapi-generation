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

provider "testing" {
  server_url = var.server_url
}

# TFGEN-276: Test that hoisted shared_name field (a field with default that exists in all variants)
# does not cause false drift when the API returns a non-default value.
#
# XEntityOneOf has a root-level discriminated oneOf where shared_name gets hoisted to root.
# The shared_name field has default: '' in the spec.
# Without the UseHoistedValue fix, this would cause drift on every plan.
resource "testing_x_entity_one_of" "test" {
  # Use branch_one variant
  branch_one = {
    discriminator_property = "branch-one"
    branch_one_config      = "test-config"
    # shared_name intentionally omitted - API will return "server-generated-value"
  }
}
