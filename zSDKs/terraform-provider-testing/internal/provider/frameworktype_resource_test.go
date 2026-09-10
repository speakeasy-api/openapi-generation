package provider_test

import (
	"net/http"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-testing/internal/provider"
	"github.com/hashicorp/terraform-provider-testing/internal/tfmockserver"
)

func TestFrameworkTypeResourceLifecycle(t *testing.T) {
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

	resourceAddress := "testing_framework_type.my_frameworktype"

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
						knownvalue.StringExact(tfmockserver.StoreKey),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_any"),
						knownvalue.Null(), // to SDK is empty, from SDK has len > 0 check
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_bool"),
						knownvalue.Null(), // to SDK is empty, from SDK has len > 0 check
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_bool_null"),
						knownvalue.Null(), // to SDK is empty, from SDK has len > 0 check
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_float32"),
						knownvalue.Null(), // to SDK is empty, from SDK has len > 0 check
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_float64"),
						knownvalue.Null(), // to SDK is empty, from SDK has len > 0 check
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_int32"),
						knownvalue.Null(), // to SDK is empty, from SDK has len > 0 check
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_int64"),
						knownvalue.Null(), // to SDK is empty, from SDK has len > 0 check
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_integer"),
						knownvalue.Null(), // to SDK is empty, from SDK has len > 0 check
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_list_bool"),
						knownvalue.Null(), // to SDK is empty, from SDK has len > 0 check
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_list_float32"),
						knownvalue.Null(), // to SDK is empty, from SDK has len > 0 check
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_list_float64"),
						knownvalue.Null(), // to SDK is empty, from SDK has len > 0 check
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_list_int32"),
						knownvalue.Null(), // to SDK is empty, from SDK has len > 0 check
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_list_int64"),
						knownvalue.Null(), // to SDK is empty, from SDK has len > 0 check
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_list_integer"),
						knownvalue.Null(), // to SDK is empty, from SDK has len > 0 check
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_list_number"),
						knownvalue.Null(), // to SDK is empty, from SDK has len > 0 check
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_list_string"),
						knownvalue.Null(), // to SDK is empty, from SDK has len > 0 check
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_null_any"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_null_bool"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_null_float32"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_null_float64"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_null_int32"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_null_int64"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_null_integer"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_null_list_string"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_null_number"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_null_set_string"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_null_string"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_nullable_any"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_nullable_bool"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_nullable_float32"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_nullable_float64"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_nullable_int32"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_nullable_int64"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_nullable_integer"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_nullable_list_string"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_nullable_number"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_nullable_set_string"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_nullable_string"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_number"),
						knownvalue.Null(), // to SDK is empty, from SDK has len > 0 check
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_set_bool"),
						knownvalue.Null(), // to SDK is empty, from SDK has len > 0 check
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_set_float32"),
						knownvalue.Null(), // to SDK is empty, from SDK has len > 0 check
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_set_float64"),
						knownvalue.Null(), // to SDK is empty, from SDK has len > 0 check
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_set_int32"),
						knownvalue.Null(), // to SDK is empty, from SDK has len > 0 check
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_set_int64"),
						knownvalue.Null(), // to SDK is empty, from SDK has len > 0 check
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_set_integer"),
						knownvalue.Null(), // to SDK is empty, from SDK has len > 0 check
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_set_number"),
						knownvalue.Null(), // to SDK is empty, from SDK has len > 0 check
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_set_string"),
						knownvalue.Null(), // to SDK is empty, from SDK has len > 0 check
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_string"),
						knownvalue.Null(), // to SDK is empty, from SDK has len > 0 check
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_string_null"),
						knownvalue.Null(), // to SDK is empty, from SDK has len > 0 check
					),
				},
			},
			// Testing framework implicitly verifies resource delete.
		},
	})

	actualGetCount := len(endpoints.Get[0].Requests.All())
	expectedGetCount := 1

	if actualGetCount != expectedGetCount {
		t.Errorf("TFGEN-287 regression: expected exactly %d requests (e.g. no additional read after create), got %d requests", expectedGetCount, actualGetCount)
	}
}

func TestFrameworkTypeResourceValidateListObjectValidator(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Verifies resource create and read.
			{
				Config: `resource "testing_framework_type" "my_frameworktype" {
					list_object = [
					null,
					{
						list_object_string = "test"
					}
					]
				}`,
				// Setup provider for this test
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ExpectError:              regexp.MustCompile("Missing Attribute Value"),
			},
			// Testing framework implicitly verifies resource delete.
		},
	})
}

// TestFrameworkTypeResource_NullableListDriftDetection tests TFGEN-182: drift detection
// for nullable list fields when the API returns null/absent for a field that was previously set.
//
// This test verifies that when:
// 1. A resource is created with list_nullable = ["value1", "value2"]
// 2. The API returns a response where list_nullable is absent (simulating external deletion)
// 3. Terraform refresh correctly detects the drift and plans to restore the value
func TestFrameworkTypeResource_NullableListDriftDetection(t *testing.T) {
	t.Parallel()

	// Track read call count to simulate drift after Step 1 completes.
	// Step 1 does: POST, GET (post-create) = 1 GET
	// Step 2 does: GET (refresh) where we simulate drift by omitting list_nullable, then plan/apply.
	// We simulate drift on the 2nd+ GET to ensure Step 2's refresh sees the missing field.
	readCallCount := 0

	// Keep a reference to the update endpoint to check if an update has been applied.
	// Once an update is applied, we stop simulating drift so the post-apply verification passes.
	updateEndpoint := &tfmockserver.ResourceEndpoint{
		Endpoint: "PATCH /v0/frameworktype/{id}",
	}

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/frameworktypes",
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/frameworktype/{id}",
				ResponseFuncWithStore: func(r *http.Request, w http.ResponseWriter, storedData map[string]any) {
					readCallCount++
					updateApplied := len(updateEndpoint.Requests.All()) > 0

					// Make a copy to avoid modifying the underlying store
					response := make(map[string]any)
					for k, v := range storedData {
						response[k] = v
					}

					// On 2nd+ reads (Step 2 refresh), simulate drift by removing list_nullable.
					// Stop simulating drift after the update has been applied.
					if readCallCount >= 2 && !updateApplied {
						delete(response, "list_nullable")
					}

					tfmockserver.JSONResponse(w, 200, response)
				},
			},
		},
		Update: tfmockserver.Endpoints{
			updateEndpoint,
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/frameworktype/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_framework_type.test"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Step 1: Create resource with list_nullable set
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_nullable"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("value1"),
							knownvalue.StringExact("value2"),
						}),
					),
				},
			},
			// Step 2: Re-apply the same config. Terraform will refresh state first (triggering
			// a Read where the mock returns list_nullable as absent), detect drift, and plan
			// an update to restore the values.
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						// TFGEN-182: Verify drift detection - plan shows resource needs update
						// because list_nullable drifted from config values to null
						plancheck.ExpectResourceAction(resourceAddress, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_nullable"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("value1"),
							knownvalue.StringExact("value2"),
						}),
					),
				},
			},
		},
	})
}
