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

// TestDiscriminatedUnionSharedFieldDifferentTypes verifies that discriminated
// union variants with a shared field name ("type_varying_field") but structurally
// different types (map vs class) generate correctly and function at runtime. This
// regression test covers TFGEN-316 where empty OriginalName in discriminator
// mapping caused incorrect oneOf variant matching during shard merge.
//
// ItemA has type_varying_field as map[string]any (additionalProperties: true).
// ItemB has type_varying_field as DiscriminatedUnionClassTypeVaryingField (defined properties).
// Without the fix, the shard merge would fail with a type mismatch error.
func TestDiscriminatedUnionSharedFieldDifferentTypes(t *testing.T) {
	t.Parallel()

	id := "test-shared-field"

	overlay := map[string]any{
		"id": id,
	}

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint:        "POST /v0/discriminated-union",
				ResponseOverlay: overlay,
			},
		},
		Update: tfmockserver.Endpoints{
			{
				Endpoint:        "PATCH /v0/discriminated-union",
				ResponseOverlay: overlay,
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/discriminated-union/{id}",
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/discriminated-union/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_discriminated_union.test"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigDirectory:          config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("id"),
						knownvalue.StringExact(id),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("item").AtMapKey("item_b").AtMapKey("value_b"),
						knownvalue.StringExact("test-value"),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("item").AtMapKey("item_b").AtMapKey("type_varying_field").AtMapKey("setting"),
						knownvalue.StringExact("my-setting"),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("item").AtMapKey("item_b").AtMapKey("type_varying_field").AtMapKey("enabled"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

// TestDiscriminatedUnionWriteOnlyPreservation verifies that write-only fields
// in discriminated union variants are preserved through the RefreshFromShared
// lifecycle, and that terraform import works without panicking.
//
// The discriminated union has ItemA with:
//   - value_a: write-only (set by user, not returned by API)
//   - value_a_sensitive: computed (returned by API as value_a_sensitive)
//
// Step 1 (Create): Verify write-only value_a preserved and computed value_a_sensitive mapped.
// Step 2 (Empty plan): Re-apply same config, ExpectEmptyPlan proves no state drift.
// Step 3 (Import): ImportState triggers Read → RefreshFromShared with nil prior data.
func TestDiscriminatedUnionWriteOnlyPreservation(t *testing.T) {
	t.Parallel()

	id := "test-id"

	// Static response for GET — simulates what the API returns for a read.
	// The API never returns the write-only value_a; only computed fields come back.
	apiResponse := map[string]any{
		"id": id,
		"item": map[string]any{
			"type":              "item_a",
			"value_a_sensitive": "sensitive-from-api",
		},
	}

	// ResponseOverlay for POST/PATCH merges computed fields on top of stored request data.
	apiOverlay := map[string]any{
		"id": id,
		"item": map[string]any{
			"type":              "item_a",
			"value_a_sensitive": "sensitive-from-api",
		},
	}

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint:        "POST /v0/discriminated-union",
				ResponseOverlay: apiOverlay,
			},
		},
		Update: tfmockserver.Endpoints{
			{
				Endpoint:        "PATCH /v0/discriminated-union",
				ResponseOverlay: apiOverlay,
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/discriminated-union/{id}",
				Response: apiResponse,
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/discriminated-union/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_discriminated_union.test"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Step 1: Create — verify write-only field preserved AND computed field mapped
			{
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigDirectory:          config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("id"),
						knownvalue.StringExact(id),
					),
					// Write-only field must be preserved from config/prior state
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("item").AtMapKey("item_a").AtMapKey("value_a"),
						knownvalue.StringExact("my-secret-value"),
					),
					// Computed field must be mapped from API response
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("item").AtMapKey("item_a").AtMapKey("value_a_sensitive"),
						knownvalue.StringExact("sensitive-from-api"),
					),
				},
			},
			// Step 2: Re-plan with same config — must be empty (no drift)
			// If write-only value_a was lost during RefreshFromShared, Terraform
			// would detect a diff and plan an update.
			{
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigDirectory:          config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// Step 3: Import — triggers Read with nil prior data (Bug 1 scenario)
			// Before the fix, RefreshFromShared would panic accessing nil prior data.
			// After import, write-only value_a cannot survive (API doesn't return it),
			// so we ignore it in verification.
			{
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigDirectory:          config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
				},
				ResourceName:            resourceAddress,
				ImportState:             true,
				ImportStateId:           id,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"item.item_a.value_a"},
			},
			// Testing framework implicitly verifies resource delete.
		},
	})
}
