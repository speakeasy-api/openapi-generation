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

// TestPatchResourcePatchSemantics tests that on update, only changed attributes are sent.
//
// This test validates the core patch semantics behavior:
// 1. Create a resource with string="initial", int64=100, bool=true
// 2. Update to string="updated", int64=200, bool=true (bool unchanged)
// 3. Validate captured requests to ensure update includes ONLY changed attributes
func TestPatchResourcePatchSemantics(t *testing.T) {
	t.Parallel()

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/patch",
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/patch/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_patch.test"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Create resource with multiple attributes
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
					"string":     config.StringVariable("initial"),
					"int64":      config.IntegerVariable(100),
					"bool":       config.BoolVariable(true),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("string"),
						knownvalue.StringExact("initial"),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("int64"),
						knownvalue.Int64Exact(100),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("bool"),
						knownvalue.Bool(true),
					),
				},
			},
			// Update only string and int64, bool stays the same
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
					"string":     config.StringVariable("updated"),
					"int64":      config.IntegerVariable(200),
					"bool":       config.BoolVariable(true),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("string"),
						knownvalue.StringExact("updated"),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("int64"),
						knownvalue.Int64Exact(200),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("bool"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})

	// ===== VALIDATE PATCH SEMANTICS =====
	// Access captured request bodies from the mock server
	bodies := endpoints.Create[0].CapturedBodies()

	if len(bodies) != 2 {
		t.Fatalf("Expected 2 POST requests (create + update), got %d", len(bodies))
	}

	// Validate CREATE request (first request)
	createReq := bodies[0]
	if val, ok := createReq["bool"]; !ok || val != true {
		t.Errorf("Create request: expected bool=true, got %v", createReq["bool"])
	}
	if val, ok := createReq["int64"]; !ok || val.(float64) != 100 {
		t.Errorf("Create request: expected int64=100, got %v", createReq["int64"])
	}
	if val, ok := createReq["string"]; !ok || val != "initial" {
		t.Errorf("Create request: expected string='initial', got %v", createReq["string"])
	}

	// Validate UPDATE request (second request) - THE KEY TEST
	updateReq := bodies[1]

	// Changed attributes MUST be present
	if val, ok := updateReq["string"]; !ok || val != "updated" {
		t.Errorf("Update request: expected string='updated', got %v", updateReq["string"])
	}
	if val, ok := updateReq["int64"]; !ok || val.(float64) != 200 {
		t.Errorf("Update request: expected int64=200, got %v", updateReq["int64"])
	}

	if val, hasBool := updateReq["bool"]; hasBool {
		t.Fatalf("❌ PATCH SEMANTICS VIOLATION: 'bool' should NOT be in update request (unchanged), but got: %v", val)
	}

	// Count non-empty attributes (excluding empty arrays/maps)
	nonEmptyCount := 0
	for _, val := range updateReq {
		if arr, isSlice := val.([]interface{}); isSlice && len(arr) == 0 {
			continue
		}
		if m, isMap := val.(map[string]interface{}); isMap && len(m) == 0 {
			continue
		}
		nonEmptyCount++
	}

	// Should have exactly 2 non-empty attributes
	if nonEmptyCount != 2 {
		t.Errorf("Expected exactly 2 non-empty attributes in update (string, int64), got %d", nonEmptyCount)
	}

	t.Logf("✓ Patch semantics validated: only changed attributes were sent in update request")
}

// TestPatchResourceUpdateString tests that a single string attribute update works correctly
func TestPatchResourceUpdateString(t *testing.T) {
	t.Parallel()

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/patch",
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/patch/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_patch.test"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Create resource with string = "initial"
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
					"string":     config.StringVariable("initial"),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("string"),
						knownvalue.StringExact("initial"),
					),
				},
			},
			// Update string to "updated"
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
					"string":     config.StringVariable("updated"),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("string"),
						knownvalue.StringExact("updated"),
					),
				},
			},
		},
	})
}

// TestPatchResourceUpdateInt64 tests that an int64 attribute update works correctly
func TestPatchResourceUpdateInt64(t *testing.T) {
	t.Parallel()

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/patch",
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/patch/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_patch.test"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Create resource with int64 = 100
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
					"int64":      config.IntegerVariable(100),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("int64"),
						knownvalue.Int64Exact(100),
					),
				},
			},
			// Update int64 to 200
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
					"int64":      config.IntegerVariable(200),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("int64"),
						knownvalue.Int64Exact(200),
					),
				},
			},
		},
	})
}

// TestPatchResourceUpdateBool tests that a bool attribute update works correctly
func TestPatchResourceUpdateBool(t *testing.T) {
	t.Parallel()

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/patch",
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/patch/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_patch.test"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Create resource with bool = false
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
					"bool":       config.BoolVariable(false),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("bool"),
						knownvalue.Bool(false),
					),
				},
			},
			// Update bool to true
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
					"bool":       config.BoolVariable(true),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("bool"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

// TestPatchResourceUpdateListString tests that a list attribute update works correctly
func TestPatchResourceUpdateListString(t *testing.T) {
	t.Parallel()

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/patch",
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/patch/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_patch.test"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Create resource with list_string = ["a", "b"]
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url":  config.StringVariable(mockServer.URL),
					"list_string": config.ListVariable(config.StringVariable("a"), config.StringVariable("b")),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_string"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("a"),
							knownvalue.StringExact("b"),
						}),
					),
				},
			},
			// Update list_string to ["c", "d", "e"]
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url":  config.StringVariable(mockServer.URL),
					"list_string": config.ListVariable(config.StringVariable("c"), config.StringVariable("d"), config.StringVariable("e")),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_string"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("c"),
							knownvalue.StringExact("d"),
							knownvalue.StringExact("e"),
						}),
					),
				},
			},
		},
	})
}

// TestPatchResourceUpdateMapString tests that a map attribute update works correctly
func TestPatchResourceUpdateMapString(t *testing.T) {
	t.Parallel()

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/patch",
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/patch/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_patch.test"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Create resource with map_string = {"key1": "value1"}
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
					"map_string": config.MapVariable(map[string]config.Variable{
						"key1": config.StringVariable("value1"),
					}),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_string"),
						knownvalue.MapExact(map[string]knownvalue.Check{
							"key1": knownvalue.StringExact("value1"),
						}),
					),
				},
			},
			// Update map_string to {"key2": "value2", "key3": "value3"}
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
					"map_string": config.MapVariable(map[string]config.Variable{
						"key2": config.StringVariable("value2"),
						"key3": config.StringVariable("value3"),
					}),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("map_string"),
						knownvalue.MapExact(map[string]knownvalue.Check{
							"key2": knownvalue.StringExact("value2"),
							"key3": knownvalue.StringExact("value3"),
						}),
					),
				},
			},
		},
	})
}

// TestPatchResourceUpdateObject tests that an object attribute update works correctly
func TestPatchResourceUpdateObject(t *testing.T) {
	t.Parallel()

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/patch",
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/patch/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_patch.test"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Create resource with object.object_string = "initial"
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url":    config.StringVariable(mockServer.URL),
					"object_string": config.StringVariable("initial"),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("object").AtMapKey("object_string"),
						knownvalue.StringExact("initial"),
					),
				},
			},
			// Update object.object_string to "updated"
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url":    config.StringVariable(mockServer.URL),
					"object_string": config.StringVariable("updated"),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("object").AtMapKey("object_string"),
						knownvalue.StringExact("updated"),
					),
				},
			},
		},
	})
}

// TestPatchResourceUnchangedListNotSent validates that unchanged lists are NOT included in update requests.
//
// This test verifies patch semantics for list attributes:
// 1. Create a resource with string="initial", list_string=["a","b"]
// 2. Update only string to "updated", leaving list_string unchanged
// 3. Validate that list_string is NOT present in the update request
func TestPatchResourceUnchangedListNotSent(t *testing.T) {
	t.Parallel()

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/patch",
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/patch/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_patch.test"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Create resource with string and list
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url":  config.StringVariable(mockServer.URL),
					"string":      config.StringVariable("initial"),
					"list_string": config.ListVariable(config.StringVariable("a"), config.StringVariable("b")),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("string"),
						knownvalue.StringExact("initial"),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_string"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("a"),
							knownvalue.StringExact("b"),
						}),
					),
				},
			},
			// Update ONLY string, leave list_string unchanged
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url":  config.StringVariable(mockServer.URL),
					"string":      config.StringVariable("updated"),
					"list_string": config.ListVariable(config.StringVariable("a"), config.StringVariable("b")),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("string"),
						knownvalue.StringExact("updated"),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_string"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("a"),
							knownvalue.StringExact("b"),
						}),
					),
				},
			},
		},
	})

	// ===== VALIDATE PATCH SEMANTICS FOR LISTS =====
	bodies := endpoints.Create[0].CapturedBodies()

	if len(bodies) != 2 {
		t.Fatalf("Expected 2 POST requests (create + update), got %d", len(bodies))
	}

	createReq := bodies[0]
	updateReq := bodies[1]

	// CREATE should have both string and list_string
	if val, ok := createReq["string"]; !ok || val != "initial" {
		t.Errorf("Create request: expected string='initial', got %v", createReq["string"])
	}
	if _, ok := createReq["list_string"]; !ok {
		t.Error("Create request: expected list_string to be present")
	}

	// UPDATE should have ONLY string (not list_string, since it's unchanged)
	if val, ok := updateReq["string"]; !ok || val != "updated" {
		t.Errorf("Update request: expected string='updated', got %v", updateReq["string"])
	}

	// The critical check: unchanged list_string should NOT be present
	if val, hasListString := updateReq["list_string"]; hasListString {
		t.Fatalf("❌ PATCH SEMANTICS VIOLATION: 'list_string' should NOT be in update request (unchanged), but got: %v", val)
	}

	t.Logf("✓ List patch semantics validated: unchanged list was NOT sent in update request")
}

// TestPatchResourceChangedListIsSent validates that changed lists ARE included in update requests.
//
// This test verifies patch semantics for list attributes:
// 1. Create a resource with list_string=["a","b"], string="initial"
// 2. Update both to list_string=["c","d"], string="updated"
// 3. Validate that list_string IS present in the update request
func TestPatchResourceChangedListIsSent(t *testing.T) {
	t.Parallel()

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/patch",
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/patch/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Create resource with string and list
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url":  config.StringVariable(mockServer.URL),
					"string":      config.StringVariable("initial"),
					"list_string": config.ListVariable(config.StringVariable("a"), config.StringVariable("b")),
				},
			},
			// Update BOTH string and list_string
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url":  config.StringVariable(mockServer.URL),
					"string":      config.StringVariable("updated"),
					"list_string": config.ListVariable(config.StringVariable("c"), config.StringVariable("d")), // CHANGED
				},
			},
		},
	})

	// ===== VALIDATE PATCH SEMANTICS FOR LISTS =====
	bodies := endpoints.Create[0].CapturedBodies()

	if len(bodies) != 2 {
		t.Fatalf("Expected 2 POST requests (create + update), got %d", len(bodies))
	}

	updateReq := bodies[1]

	// UPDATE should have both string and list_string (both changed)
	if val, ok := updateReq["string"]; !ok || val != "updated" {
		t.Errorf("Update request: expected string='updated', got %v", updateReq["string"])
	}

	// The critical check: changed list_string SHOULD be present
	if listVal, ok := updateReq["list_string"]; !ok {
		t.Fatal("❌ PATCH SEMANTICS VIOLATION: 'list_string' SHOULD be in update request (changed)")
	} else {
		// Verify the list contains the new values
		listArray, isArray := listVal.([]interface{})
		if !isArray {
			t.Fatalf("Expected list_string to be an array, got %T", listVal)
		}
		if len(listArray) != 2 || listArray[0] != "c" || listArray[1] != "d" {
			t.Errorf("Expected list_string=['c','d'], got %v", listArray)
		}
	}

	t.Logf("✓ List patch semantics validated: changed list WAS sent in update request")
}

// TestPatchResourceUnchangedSetNotSent validates that unchanged sets are NOT included in update requests.
//
// This test verifies patch semantics for set attributes:
// 1. Create a resource with string="initial", set_string=["x","y"]
// 2. Update only string to "updated", leaving set_string unchanged
// 3. Validate that set_string is NOT present in the update request
func TestPatchResourceUnchangedSetNotSent(t *testing.T) {
	t.Parallel()

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/patch",
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/patch/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_patch.test"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Create resource with string and set
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
					"string":     config.StringVariable("initial"),
					"set_string": config.SetVariable(config.StringVariable("x"), config.StringVariable("y")),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("string"),
						knownvalue.StringExact("initial"),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("set_string"),
						knownvalue.SetSizeExact(2),
					),
				},
			},
			// Update ONLY string, leave set_string unchanged
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
					"string":     config.StringVariable("updated"),
					"set_string": config.SetVariable(config.StringVariable("x"), config.StringVariable("y")),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("string"),
						knownvalue.StringExact("updated"),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("set_string"),
						knownvalue.SetSizeExact(2),
					),
				},
			},
		},
	})

	// ===== VALIDATE PATCH SEMANTICS FOR SETS =====
	bodies := endpoints.Create[0].CapturedBodies()

	if len(bodies) != 2 {
		t.Fatalf("Expected 2 POST requests (create + update), got %d", len(bodies))
	}

	createReq := bodies[0]
	updateReq := bodies[1]

	// CREATE should have both string and set_string
	if val, ok := createReq["string"]; !ok || val != "initial" {
		t.Errorf("Create request: expected string='initial', got %v", createReq["string"])
	}
	if _, ok := createReq["set_string"]; !ok {
		t.Error("Create request: expected set_string to be present")
	}

	// UPDATE should have ONLY string (not set_string, since it's unchanged)
	if val, ok := updateReq["string"]; !ok || val != "updated" {
		t.Errorf("Update request: expected string='updated', got %v", updateReq["string"])
	}

	// The critical check: unchanged set_string should NOT be present
	if val, hasSetString := updateReq["set_string"]; hasSetString {
		t.Fatalf("❌ PATCH SEMANTICS VIOLATION: 'set_string' should NOT be in update request (unchanged), but got: %v", val)
	}

	t.Logf("✓ Set patch semantics validated: unchanged set was NOT sent in update request")
}

// TestPatchResourceChangedSetIsSent validates that changed sets ARE included in update requests.
//
// This test verifies patch semantics for set attributes:
// 1. Create a resource with set_string=["x","y"], string="initial"
// 2. Update both to set_string=["z"], string="updated"
// 3. Validate that set_string IS present in the update request
func TestPatchResourceChangedSetIsSent(t *testing.T) {
	t.Parallel()

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/patch",
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/patch/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Create resource with string and set
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
					"string":     config.StringVariable("initial"),
					"set_string": config.SetVariable(config.StringVariable("x"), config.StringVariable("y")),
				},
			},
			// Update BOTH string and set_string
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
					"string":     config.StringVariable("updated"),
					"set_string": config.SetVariable(config.StringVariable("z")), // CHANGED
				},
			},
		},
	})

	// ===== VALIDATE PATCH SEMANTICS FOR SETS =====
	bodies := endpoints.Create[0].CapturedBodies()

	if len(bodies) != 2 {
		t.Fatalf("Expected 2 POST requests (create + update), got %d", len(bodies))
	}

	updateReq := bodies[1]

	// UPDATE should have both string and set_string (both changed)
	if val, ok := updateReq["string"]; !ok || val != "updated" {
		t.Errorf("Update request: expected string='updated', got %v", updateReq["string"])
	}

	// The critical check: changed set_string SHOULD be present
	if setVal, ok := updateReq["set_string"]; !ok {
		t.Fatal("❌ PATCH SEMANTICS VIOLATION: 'set_string' SHOULD be in update request (changed)")
	} else {
		// Verify the set contains the new value
		setArray, isArray := setVal.([]interface{})
		if !isArray {
			t.Fatalf("Expected set_string to be an array, got %T", setVal)
		}
		if len(setArray) != 1 || setArray[0] != "z" {
			t.Errorf("Expected set_string=['z'], got %v", setArray)
		}
	}

	t.Logf("✓ Set patch semantics validated: changed set WAS sent in update request")
}
