package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-testing/internal/provider"
	"github.com/hashicorp/terraform-provider-testing/internal/tfmockserver"
)

func TestOASServersPathDataSource(t *testing.T) {
	t.Parallel()

	providerServerEndpoints := tfmockserver.ResourceEndpoints{
		// Intentionally invalid.
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/oasservers/path/{id}",
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
				Endpoint: "POST /v0/oasservers/path",
				ResponseOverlay: map[string]any{
					"id": id,
				},
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/oasservers/path/{id}",
				ResponseOverlay: map[string]any{
					"id": id,
				},
			},
		},
		Update: tfmockserver.Endpoints{
			{
				Endpoint: "PATCH /v0/oasservers/path/{id}",
				ResponseOverlay: map[string]any{
					"id": id,
				},
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/oasservers/path/{id}",
			},
		},
	}

	providerMockServer := tfmockserver.StartServer(providerServerEndpoints, t)
	defer providerMockServer.Close()

	resourceMockServer := tfmockserver.StartServer(resourceServerEndpoints, t)
	defer resourceMockServer.Close()

	dataResourceAddress := "data.testing_oas_servers_path.test"
	managedResourceAddress := "testing_oas_servers_path.test"

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
				ConfigStateChecks: provider.AllModelFieldCompareChecks(
					&provider.OASServersPathDataSourceModel{},
					managedResourceAddress,
					dataResourceAddress,
					[]string{"id"},
				),
			},
			// Testing framework implicitly verifies resource delete.
		},
	})
}
