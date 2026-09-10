variable "server_url" {
  type = string
}

variable "name" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_x_match_prior_state" "test" {
  name        = var.name
  description = "Test resource for usePriorState functionality"
}
