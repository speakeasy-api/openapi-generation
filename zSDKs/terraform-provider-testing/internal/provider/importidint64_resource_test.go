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

func TestImportIDInt64ResourceLifecycle(t *testing.T) {
	t.Parallel()

	id := int64(123456789012345)
	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/import-id-int64",
				ResponseOverlay: map[string]any{
					"id": id,
				},
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/import-id-int64/{id}",
				ResponseOverlay: map[string]any{
					"id": id,
				},
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/import-id-int64/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_import_id_int64.my_importidint64"

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
			// Test import functionality with int64 ID
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

func TestImportIDInt64ResourceBoundaryValues(t *testing.T) {
	t.Parallel()

	id := int64(9223372036854775807) // max int64
	// Define endpoints for importidint64 resource with large int64 value
	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/import-id-int64",
				ResponseOverlay: map[string]any{
					"id": id,
				},
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/import-id-int64/{id}",
				ResponseOverlay: map[string]any{
					"id": id,
				},
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/import-id-int64/{id}",
			},
		},
	}

	// Start mock server
	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_import_id_int64.my_importidint64"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Test maximum int64 value
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
			// Test import with max int64 value
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
		},
	})
}

func TestImportIDInt64ResourceMinBoundaryValue(t *testing.T) {
	t.Parallel()

	id := int64(-9223372036854775808) // min int64
	// Define endpoints for importidint64 resource with minimum int64 value
	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/import-id-int64",
				ResponseOverlay: map[string]any{
					"id": id,
				},
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/import-id-int64/{id}",
				ResponseOverlay: map[string]any{
					"id": id,
				},
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/import-id-int64/{id}",
			},
		},
	}

	// Start mock server
	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_import_id_int64.my_importidint64"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Test minimum int64 value
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
			// Test import with min int64 value
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
		},
	})
}

func TestImportIDInt64ResourceImportValidation(t *testing.T) {
	t.Parallel()

	id := int64(123456789012345)
	// Define endpoints for importidint64 resource
	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/import-id-int64",
				ResponseOverlay: map[string]any{
					"id": id,
				},
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/import-id-int64/{id}",
				ResponseOverlay: map[string]any{
					"id": id,
				},
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/import-id-int64/{id}",
			},
		},
	}

	// Start mock server
	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_import_id_int64.my_importidint64"

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
