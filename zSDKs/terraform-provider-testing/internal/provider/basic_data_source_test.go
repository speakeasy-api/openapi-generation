package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-testing/internal/provider"
	"github.com/hashicorp/terraform-provider-testing/internal/tfmockserver"
)

func TestBasicDataSource(t *testing.T) {
	t.Parallel()

	name := "example-name"
	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/basics",
			},
		},
		Get: tfmockserver.Endpoints{

			{
				Endpoint: "GET /v0/basic/{id}",
			},
			// Adding a response overlay here because the response is nested under `data` field. If not specified
			// the mock server responds with id and name hoisted out of `data` field
			{
				Endpoint: "GET /v0/basic",
				ResponseOverlay: map[string]any{
					"data": []map[string]any{
						{
							"id":   tfmockserver.StoreKey,
							"name": name,
						},
					},
				},
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/basic/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	dataResourceAddress := "data.testing_basic.test"
	managedResourceAddress := "testing_basic.test"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Verifies resource create and read.
			{
				// Specifies the directory to look for the terraform configuration files.
				// By default it will look in the folder `testdata/<name of the test>`
				ConfigDirectory: config.TestNameDirectory(),
				// Setup provider for this test
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
					"name":       config.StringVariable(name),
				},
				// Verify that the data source and managed resource have the same field values
				// No need to check "id" as Terraform will error
				// if applied state value differs from configuration.
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						managedResourceAddress,
						tfjsonpath.New("name"),
						dataResourceAddress,
						tfjsonpath.New("name"),
						compare.ValuesSame(),
					),
				},
			},
		},
	})
}
