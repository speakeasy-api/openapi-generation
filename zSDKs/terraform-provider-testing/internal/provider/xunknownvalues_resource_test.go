package provider_test

import (
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

// TestXUnknownValuesResourceLifecycle verifies that open enums (marked with
// x-speakeasy-unknown-values: allow) correctly handle values outside the
// declared enum set. The mock server responds with unknown enum values on
// response-only and optional+computed fields to simulate an API that has
// added new enum values not yet known to the provider.
func TestXUnknownValuesResourceLifecycle(t *testing.T) {
	t.Parallel()

	// Response overlay injects unknown enum values — values that are NOT in
	// the declared enum sets. This simulates an API returning a newly added
	// enum value that the provider schema doesn't know about yet.
	unknownValuesOverlay := map[string]any{
		// "extra_large" is not in [small, medium, large]
		"optional_computed_open_enum_string": "extra_large",
		// 99 is not in [1, 2, 3]
		"optional_computed_open_enum_int64": 99,
		// "archived" is not in [active, inactive, pending]
		"response_only_open_enum_string": "archived",
		// 999 is not in [100, 200, 300]
		"response_only_open_enum_int64": 999,
	}

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint:        "POST /v0/xunknownvalues",
				ResponseOverlay: unknownValuesOverlay,
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint:        "GET /v0/xunknownvalues/{id}",
				ResponseOverlay: unknownValuesOverlay,
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/xunknownvalues/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_x_unknown_values.my_xunknownvalues"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Step 1: Create + Read — verify unknown enum values propagate
			// to state without validation errors.
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
						knownvalue.StringExact(tfmockserver.StoreKey),
					),
					// Required field uses config value.
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("required_open_enum_string"),
						knownvalue.StringExact("alpha"),
					),
					// Optional+Computed string with unknown value from API.
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("optional_computed_open_enum_string"),
						knownvalue.StringExact("extra_large"),
					),
					// Optional+Computed int64 with unknown value from API.
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("optional_computed_open_enum_int64"),
						knownvalue.Int64Exact(99),
					),
					// Response-only (Computed) string with unknown value.
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("response_only_open_enum_string"),
						knownvalue.StringExact("archived"),
					),
					// Response-only (Computed) int64 with unknown value.
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("response_only_open_enum_int64"),
						knownvalue.Int64Exact(999),
					),
				},
			},
			// Step 2: Import — verify imported state also handles unknown values.
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
				},
				ResourceName:      resourceAddress,
				ImportState:       true,
				ImportStateId:     tfmockserver.StoreKey,
				ImportStateVerify: true,
			},
		},
	})
}

// TestXUnknownValuesClosedEnumRejectsUnknownValue verifies that closed enums
// (marked with x-speakeasy-unknown-values: disallow) still enforce OneOf
// validation. An invalid value on the closed_enum_string field should trigger
// a plan-time validation error, proving the open enum behavior is granular.
func TestXUnknownValuesClosedEnumRejectsUnknownValue(t *testing.T) {
	t.Parallel()

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/xunknownvalues",
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/xunknownvalues/{id}",
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/xunknownvalues/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
				},
				ExpectError: regexp.MustCompile(`value must be one of`),
			},
		},
	})
}
