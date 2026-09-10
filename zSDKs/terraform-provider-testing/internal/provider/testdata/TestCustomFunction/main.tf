# Since provider-defined functions are loaded earlier in Terraform, the
# required_providers block must be defined to avoid an unknown provider function
# error.
terraform {
  required_providers {
    testing = {
      source = "hashicorp/testing" 
    }
  }
}

output "test" {
  value = provider::testing::custom("test-value")
}
