package provider_test

import (
	"net/http"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-testing/internal/provider"
	"github.com/hashicorp/terraform-provider-testing/internal/tfmockserver"
)

// TestXAdditionalPropertiesNameResourceLifecycle tests the full lifecycle of a resource
// with additionalProperties. This validates that:
// 1. The generated code compiles correctly (interface{} vs map[string]any fix)
// 2. JSON additionalProperties values are correctly sent to the API
// 3. JSON additionalProperties values are correctly read back from the API
//
// This test covers the fix for a type mismatch bug where terraform provider generation
// was using `interface{}` for additionalProperties fields instead of `map[string]any`,
// causing a compile error: "cannot use additional (variable of type interface{}) as
// map[string]any value in struct literal: need type assertion"
func TestXAdditionalPropertiesNameResourceLifecycle(t *testing.T) {
	t.Parallel()

	// Store to track the resource state across requests
	var mu sync.Mutex
	var storedData map[string]any

	// Create endpoint references so we can access captured requests in ResponseFunc
	createEndpoint := &tfmockserver.ResourceEndpoint{
		Endpoint: "POST /v0/xadditionalpropertiesname/{slug}",
	}
	getEndpoint := &tfmockserver.ResourceEndpoint{
		Endpoint: "GET /v0/xadditionalpropertiesname/{slug}/{id}",
	}
	deleteEndpoint := &tfmockserver.ResourceEndpoint{
		Endpoint: "DELETE /v0/xadditionalpropertiesname/{slug}/{id}",
	}

	// Set up ResponseFuncs that use captured request data.
	// We need custom ResponseFuncs because the default mock server behavior
	// adds path parameters to the response body, which would be interpreted
	// as additionalProperties and cause state drift.
	createEndpoint.ResponseFunc = func(r *http.Request, w http.ResponseWriter) {
		// Get the most recently captured request body
		bodies := createEndpoint.CapturedBodies()
		if len(bodies) == 0 {
			http.Error(w, "no captured request", http.StatusInternalServerError)
			return
		}
		body := bodies[len(bodies)-1]

		mu.Lock()
		storedData = make(map[string]any)
		for k, v := range body {
			storedData[k] = v
		}
		storedData["id"] = tfmockserver.StoreKey
		dataCopy := make(map[string]any)
		for k, v := range storedData {
			dataCopy[k] = v
		}
		mu.Unlock()

		tfmockserver.JSONResponse(w, 200, dataCopy)
	}

	getEndpoint.ResponseFunc = func(r *http.Request, w http.ResponseWriter) {
		mu.Lock()
		if storedData == nil {
			mu.Unlock()
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		dataCopy := make(map[string]any)
		for k, v := range storedData {
			dataCopy[k] = v
		}
		mu.Unlock()

		tfmockserver.JSONResponse(w, 200, dataCopy)
	}

	deleteEndpoint.ResponseFunc = func(r *http.Request, w http.ResponseWriter) {
		mu.Lock()
		storedData = nil
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{createEndpoint},
		Get:    tfmockserver.Endpoints{getEndpoint},
		Delete: tfmockserver.Endpoints{deleteEndpoint},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_x_additional_properties_name.test"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Verifies resource create and read with additionalProperties.
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
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("slug"),
						knownvalue.StringExact("test-slug"),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("string"),
						knownvalue.StringExact("test-string-value"),
					),
					// The x_additional_properties_name should be stored as normalized JSON
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("x_additional_properties_name"),
						knownvalue.NotNull(),
					),
				},
			},
			// Testing framework implicitly verifies resource delete.
		},
	})
}
