package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-testing/internal/provider"
	"github.com/hashicorp/terraform-provider-testing/internal/tfmockserver"
)

func TestOASServersOperationResourceLifecycle(t *testing.T) {
	t.Parallel()

	providerServerEndpoints := tfmockserver.ResourceEndpoints{
		// Intentionally empty/invalid.
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/oasservers/operation",
				ResponseOverlay: map[string]any{
					"id": "invalid-id",
				},
			},
		},
	}
	id := "test-id"
	resourceServerEndpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/oasservers/operation",
				ResponseOverlay: map[string]any{
					"id": id,
				},
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/oasservers/operation/{id}",
				ResponseOverlay: map[string]any{
					"id": id,
				},
			},
		},
		Update: tfmockserver.Endpoints{
			{
				Endpoint: "PATCH /v0/oasservers/operation/{id}",
				ResponseOverlay: map[string]any{
					"id": id,
				},
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/oasservers/operation/{id}",
			},
		},
	}

	providerMockServer := tfmockserver.StartServer(providerServerEndpoints, t)
	defer providerMockServer.Close()

	resourceMockServer := tfmockserver.StartServer(resourceServerEndpoints, t)
	defer resourceMockServer.Close()

	resourceAddress := "testing_oas_servers_operation.test"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Verifies resource create and read.
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"provider_server_url": config.StringVariable(providerMockServer.URL),
					"resource_server_url": config.StringVariable(resourceMockServer.URL),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("id"),
						knownvalue.StringExact(id),
					),
				},
			},
			// Testing framework implicitly verifies resource delete.
		},
	})
}
