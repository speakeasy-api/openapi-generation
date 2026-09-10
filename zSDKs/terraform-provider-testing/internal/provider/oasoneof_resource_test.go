package provider_test

import (
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

// TestOASOneOfResourceUnionSwitchDetectsChanges verifies the fix for the union type bug where:
// - Post-update, both union submembers incorrectly ended up in state (only one should be present)
// - Terraform failed to detect infrastructure changes without a refresh
//
// The bug was caused by union member fields being marked as both Computed: true and Optional: true,
// which prevented proper plan diff detection.
//
// The fix:
// 1. Union subtype fields now have Computed: false when Optional: true
// 2. Union subtypes marked with x-speakeasy-terraform-plan-only explicitly set x-speakeasy-param-suppress-computed-diff to false
// 3. Plan-only mutable fields never have Computed: true
func TestOASOneOfResourceUnionSwitchDetectsChanges(t *testing.T) {
	t.Parallel()

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/oasoneof",
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/oasoneof/{id}",
			},
		},
		Update: tfmockserver.Endpoints{
			{
				Endpoint: "PATCH /v0/oasoneof/{id}",
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/oasoneof/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_oas_one_of.test"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Step 1: Create resource with inline primitives union set to string
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
					"use_string": config.BoolVariable(true),
					"use_number": config.BoolVariable(false),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify only string member is present, not number
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("optional_inline_primitives_request_and_response").AtMapKey("str"),
						knownvalue.StringExact("test-string"),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("optional_inline_primitives_request_and_response").AtMapKey("number"),
						knownvalue.Null(),
					),
				},
			},
			// Step 2: Switch to number member - verify plan detects change WITHOUT refresh
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
					"use_string": config.BoolVariable(false),
					"use_number": config.BoolVariable(true),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						// CRITICAL: This proves the bug is fixed - plan must detect change immediately
						plancheck.ExpectNonEmptyPlan(),
						// Verify the specific attribute change is detected
						plancheck.ExpectResourceAction(resourceAddress, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					// After apply, verify only number member is present
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("optional_inline_primitives_request_and_response").AtMapKey("number"),
						knownvalue.Float64Exact(42.0),
					),
					// Verify string member is NOT in state (the original bug)
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("optional_inline_primitives_request_and_response").AtMapKey("str"),
						knownvalue.Null(),
					),
				},
			},
			// Step 3: Run plan again with same config - should be empty (no drift)
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
					"use_string": config.BoolVariable(false),
					"use_number": config.BoolVariable(true),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						// Subsequent plan should be empty - no drift
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestOASOneOfResourceUnionMemberFieldUpdateDetected verifies that updating a field
// within the active union member is immediately detected in plan without requiring a refresh.
// This ensures union member fields are properly marked as Optional (not Computed).
func TestOASOneOfResourceUnionMemberFieldUpdateDetected(t *testing.T) {
	t.Parallel()

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/oasoneof",
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/oasoneof/{id}",
			},
		},
		Update: tfmockserver.Endpoints{
			{
				Endpoint: "PATCH /v0/oasoneof/{id}",
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/oasoneof/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_oas_one_of.test"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Step 1: Create with string value "v1"
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url":   config.StringVariable(mockServer.URL),
					"use_string":   config.BoolVariable(true),
					"use_number":   config.BoolVariable(false),
					"string_value": config.StringVariable("v1"),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("optional_inline_primitives_request_and_response").AtMapKey("str"),
						knownvalue.StringExact("v1"),
					),
				},
			},
			// Step 2: Update string value to "v2" - verify plan detects change
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url":   config.StringVariable(mockServer.URL),
					"use_string":   config.BoolVariable(true),
					"use_number":   config.BoolVariable(false),
					"string_value": config.StringVariable("v2"),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						// Plan must detect change immediately (proves fields are Optional, not Computed)
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction(resourceAddress, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("optional_inline_primitives_request_and_response").AtMapKey("str"),
						knownvalue.StringExact("v2"),
					),
					// Number member remains null
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("optional_inline_primitives_request_and_response").AtMapKey("number"),
						knownvalue.Null(),
					),
				},
			},
			// Step 3: Verify no drift after update
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url":   config.StringVariable(mockServer.URL),
					"use_string":   config.BoolVariable(true),
					"use_number":   config.BoolVariable(false),
					"string_value": config.StringVariable("v2"),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestOASOneOfResourceDiscriminatedUnionSwitchExclusivity tests switching between
// discriminated union members to ensure mutual exclusivity in state.
// Uses optional_discriminator_request_and_response which doesn't have the request/response split issue.
func TestOASOneOfResourceDiscriminatedUnionSwitchExclusivity(t *testing.T) {
	t.Parallel()

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/oasoneof",
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/oasoneof/{id}",
			},
		},
		Update: tfmockserver.Endpoints{
			{
				Endpoint: "PATCH /v0/oasoneof/{id}",
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/oasoneof/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_oas_one_of.test"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Step 1: Create with discriminated union - string variant
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url":          config.StringVariable(mockServer.URL),
					"use_disc_string":     config.BoolVariable(true),
					"use_disc_number":     config.BoolVariable(false),
					"discriminator_value": config.StringVariable("v1"),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify string member is present (discriminator field removed — implicit in block structure)
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("optional_discriminator_request_and_response").
							AtMapKey("string").
							AtMapKey("object_string"),
						knownvalue.StringExact("v1"),
					),
					// Verify number member is NOT present
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("optional_discriminator_request_and_response").
							AtMapKey("number"),
						knownvalue.Null(),
					),
					// Verify allof member is NOT present
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("optional_discriminator_request_and_response").
							AtMapKey("allof"),
						knownvalue.Null(),
					),
				},
			},
			// Step 2: Switch to number variant
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url":          config.StringVariable(mockServer.URL),
					"use_disc_string":     config.BoolVariable(false),
					"use_disc_number":     config.BoolVariable(true),
					"discriminator_value": config.StringVariable("v1"),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						// Must detect the switch immediately
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction(resourceAddress, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					// After apply, verify number member is present (discriminator field removed — implicit in block structure)
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("optional_discriminator_request_and_response").
							AtMapKey("number").
							AtMapKey("object_number"),
						knownvalue.Float64Exact(42.0),
					),
					// Verify string member is NOT in state (exclusivity check)
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("optional_discriminator_request_and_response").
							AtMapKey("string"),
						knownvalue.Null(),
					),
					// Verify allof member is NOT present
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("optional_discriminator_request_and_response").
							AtMapKey("allof"),
						knownvalue.Null(),
					),
				},
			},
			// Step 3: Verify no drift
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url":          config.StringVariable(mockServer.URL),
					"use_disc_string":     config.BoolVariable(false),
					"use_disc_number":     config.BoolVariable(true),
					"discriminator_value": config.StringVariable("v1"),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})

	// Verify that the discriminator value is still sent in the HTTP request body,
	// even though it's been removed from the Terraform schema (preApplyUnionDiscriminators).
	// The Go SDK's const tag marshaling should inject the discriminator automatically.
	capturedCreate := endpoints.Create[0].Captured()
	if len(capturedCreate) != 1 {
		t.Fatalf("Expected 1 create request, got %d", len(capturedCreate))
	}

	createDiscUnion, ok := capturedCreate[0].Body["optional_discriminator_request_and_response"].(map[string]any)
	if !ok {
		t.Fatal("Expected optional_discriminator_request_and_response to be a nested object in create request body")
	}

	tfmockserver.AssertRequestBodyField(t, createDiscUnion, "discriminator", "string")
	tfmockserver.AssertRequestBodyField(t, createDiscUnion, "object_string", "v1")

	capturedUpdate := endpoints.Update[0].Captured()
	if len(capturedUpdate) != 1 {
		t.Fatalf("Expected 1 update request, got %d", len(capturedUpdate))
	}

	updateDiscUnion, ok := capturedUpdate[0].Body["optional_discriminator_request_and_response"].(map[string]any)
	if !ok {
		t.Fatal("Expected optional_discriminator_request_and_response to be a nested object in update request body")
	}

	tfmockserver.AssertRequestBodyField(t, updateDiscUnion, "discriminator", "number")
	tfmockserver.AssertRequestBodyField(t, updateDiscUnion, "object_number", float64(42))
}
