package provider_test

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-testing/internal/provider"
	"github.com/hashicorp/terraform-provider-testing/internal/tfmockserver"
)

func TestXMatchResourceLifecycle(t *testing.T) {
	t.Parallel()

	id := "resource-123"
	endpointsBuilder := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/xmatch",
				ResponseOverlay: map[string]any{
					"id": id,
				},
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/xmatch/{objectNestedString}/{string}",
				ResponseOverlay: map[string]any{
					"id": id,
				},
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/xmatch/{objectNestedString}/{string}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpointsBuilder, t)
	defer mockServer.Close()

	t.Logf("🚀 [TEST] Mock server started at: %s", mockServer.URL)

	resourceAddress := "testing_x_match.my_xmatch"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Verifies resource create and read.
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
				},
				// Check computed values.
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("id"),
						knownvalue.StringExact(id),
					),
				},
			},
			// Test import functionality with JSON format
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
				},
				ResourceName: resourceAddress,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources[resourceAddress]
					jsonBytes, err := json.Marshal(map[string]string{
						"id":        rs.Primary.Attributes["id"],
						"object_id": rs.Primary.Attributes["object.id"],
					})
					return string(jsonBytes), err
				},
				ImportStateVerify: true,
			},
			// Testing framework implicitly verifies resource delete.
		},
	})
}
