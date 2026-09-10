package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-testing/internal/provider"
	"github.com/hashicorp/terraform-provider-testing/internal/tfmockserver"
)

func TestImportMultipleIDDataSource(t *testing.T) {
	t.Parallel()

	param1 := "test-param1"
	param2 := "test-param2"
	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/import-multiple-id/{param1}/{param2}",
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/import-multiple-id/{param1}/{param2}",
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/import-multiple-id/{param1}/{param2}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	dataResourceAddress := "data.testing_import_multiple_id.test"
	managedResourceAddress := "testing_import_multiple_id.test"

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
					"param1":     config.StringVariable(param1),
					"param2":     config.StringVariable(param2),
				},
				// Verify that the data source and managed resource have the same field values
				ConfigStateChecks: provider.AllModelFieldCompareChecks(provider.ImportMultipleIDDataSourceModel{}, managedResourceAddress, dataResourceAddress, []string{}),
			},
		},
	})
}
