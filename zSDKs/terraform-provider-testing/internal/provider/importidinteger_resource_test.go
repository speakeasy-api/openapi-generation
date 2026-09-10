package provider_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-testing/internal/provider"
	"github.com/hashicorp/terraform-provider-testing/internal/tfmockserver"
)

func TestImportIDIntegerResourceLifecycle(t *testing.T) {
	t.Parallel()

	id := int64(987654321)
	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/import-id-integer",
				ResponseOverlay: map[string]any{
					"id": id,
				},
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/import-id-integer/{id}",
				ResponseOverlay: map[string]any{
					"id": id,
				},
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/import-id-integer/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_import_id_integer.my_importidinteger"

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
						tfjsonpath.New("id"),
						knownvalue.Int64Exact(id),
					),
				},
			},
			// Test import functionality with integer ID
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
				},
				ResourceName:      resourceAddress,
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%d", id),
				ImportStateVerify: true,
			},
			// Testing framework implicitly verifies resource delete.
		},
	})
}

func TestImportIDIntegerResourceImportValidation(t *testing.T) {
	t.Parallel()

	id := int64(987654321)
	// Define endpoints for importidinteger resource
	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/import-id-integer",
				ResponseOverlay: map[string]any{
					"id": id,
				},
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/import-id-integer/{id}",
				ResponseOverlay: map[string]any{
					"id": id,
				},
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/import-id-integer/{id}",
			},
		},
	}

	// Start mock server
	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_import_id_integer.my_importidinteger"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Create resource first
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("id"),
						knownvalue.Int64Exact(id),
					),
				},
			},
			// Test import with non-numeric string
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
				},
				ResourceName:  resourceAddress,
				ImportState:   true,
				ImportStateId: "not-a-number",
				ExpectError:   regexp.MustCompile("ID must be an integer"),
			},
			// Test import with floating point number
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
				},
				ResourceName:  resourceAddress,
				ImportState:   true,
				ImportStateId: "123.456",
				ExpectError:   regexp.MustCompile("ID must be an integer"),
			},
		},
	})
}
