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

func TestAPICreateAndUpdateDataSource(t *testing.T) {
	t.Parallel()

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/apicreateandupdate",
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/apicreateandupdate",
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/apicreateandupdate",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	dataResourceAddress := "data.testing_api_create_and_update.test"
	managedResourceAddress := "testing_api_create_and_update.test"

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
