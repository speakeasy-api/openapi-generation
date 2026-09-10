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

func TestImportIDStringAcronymResourceLifecycle(t *testing.T) {
	t.Parallel()

	id := "test-api-id-123"
	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/import-id-string-acronym",
				ResponseOverlay: map[string]any{
					"apiId": id,
				},
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/import-id-string-acronym/{apiId}",
				ResponseOverlay: map[string]any{
					"apiId": id,
				},
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/import-id-string-acronym/{apiId}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_import_id_string_acronym.my_importidstringacronym"

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
				// Check computed values.
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("api_id"),
						knownvalue.StringExact(id),
					),
				},
			},
			// Test import functionality with string ID
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
				},
				ResourceName:                         resourceAddress,
				ImportState:                          true,
				ImportStateId:                        id,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "api_id",
			},
			// Testing framework implicitly verifies resource delete.
		},
	})
}
