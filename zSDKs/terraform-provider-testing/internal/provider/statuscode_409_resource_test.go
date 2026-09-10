package provider_test

import (
	"net/http"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-testing/internal/provider"
	"github.com/hashicorp/terraform-provider-testing/internal/tfmockserver"
)

// TestAccStatusCode409Resource verifies that when a Create operation
// returns a 409 Conflict status code, the generated provider shows a user-friendly
// error message instead of a generic "unexpected response code" error.
//
// This validates the TFGEN-223 implementation where 409 responses get special handling.
func TestAccStatusCode409Resource(t *testing.T) {
	t.Parallel()

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/statusCode/409",
				// Use ResponseFunc to always return a 409 Conflict response
				ResponseFunc: func(r *http.Request, w http.ResponseWriter) {
					// Return 409 Conflict with error message
					tfmockserver.JSONResponse(w, 409, map[string]any{
						"error": "Resource with this name already exists",
					})
				},
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.GetTestProviders(),
		Steps: []resource.TestStep{
			// Attempt to create a resource that will return 409 Conflict
			{
				Config: `
variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_status_code_409" "test" {
  name = "test-resource"
}
`,
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
				},
				// Expect the specific error message for 409 Conflict
				ExpectError: regexp.MustCompile(`Resource Already Exists`),
			},
		},
	})
}
