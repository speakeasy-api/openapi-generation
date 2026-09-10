package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-testing/internal/provider"
	"github.com/hashicorp/terraform-provider-testing/internal/tfmockserver"
)

func TestFrameworkTypeDataSource(t *testing.T) {
	t.Parallel()

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/frameworktypes",
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/frameworktype/{id}",
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/frameworktype/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	dataResourceAddress := "data.testing_framework_type.test"
	managedResourceAddress := "testing_framework_type.test"

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
				ConfigStateChecks: provider.AllModelFieldCompareChecks(
					&provider.FrameworkTypeDataSourceModel{},
					managedResourceAddress,
					dataResourceAddress,
					[]string{"id"},
				),
			},
		},
	})
}
