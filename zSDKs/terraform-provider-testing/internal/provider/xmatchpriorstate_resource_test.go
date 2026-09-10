package provider_test

import (
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-testing/internal/provider"
	"github.com/hashicorp/terraform-provider-testing/internal/tfmockserver"
	"github.com/stretchr/testify/require"
)

// TestXMatchPriorStateResourceUpdate tests the critical usePriorState functionality.
// This verifies that when updating a resource with a renameable identifier:
// 1. The path parameter uses the PRIOR state value (old name)
// 2. The request body contains the NEW value (new name)
func TestXMatchPriorStateResourceUpdate(t *testing.T) {
	t.Parallel()

	initialName := "initial-name"
	updatedName := "updated-name"

	// We'll use these endpoints to capture and track the UPDATE request
	var updateEndpoint *tfmockserver.ResourceEndpoint

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/xmatchpriorstate",
				// Use default behavior - framework will capture the request
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/xmatchpriorstate/{name}",
				// We need to use ResponseFunc to handle the name correctly.
				// After UPDATE, the store has both 'name: initial-name' and 'new_name: updated-name'.
				// But the GET request comes with the path /v0/xmatchpriorstate/{updated-name},
				// so we need to return 'name: updated-name' in the response.
				ResponseFunc: func(r *http.Request, w http.ResponseWriter) {
					// Return the resource with the name from the path parameter
					// This simulates a real API where GET returns the current state of the resource
					tfmockserver.JSONResponse(w, 200, map[string]any{
						"name":        r.PathValue("name"),
						"description": "Test resource for usePriorState functionality",
					})
				},
			},
		},
		Update: tfmockserver.Endpoints{
			{
				Endpoint: "PATCH /v0/xmatchpriorstate/{name}",
				// The ResponseOverlay will be applied after the default merge logic.
				// The framework:
				// 1. Decodes the body (new_name + description)
				// 2. Captures the request (so we can verify path params and body later)
				// 3. Merges body into existing data
				// 4. Applies our ResponseOverlay to set name=updated-name
				//
				// Note: In a real API, when you PATCH with new_name, the response would
				// return the resource with name set to the new value (not new_name).
				ResponseOverlay: map[string]any{
					"name": updatedName, // Override to use the new name
				},
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/xmatchpriorstate/{name}",
			},
		},
	}

	// Save reference to the update endpoint so we can access captured requests
	updateEndpoint = endpoints.Update[0]

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_x_match_prior_state.test"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Step 1: Create resource with initial name
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
					"name":       config.StringVariable(initialName),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("name"),
						knownvalue.StringExact(initialName),
					),
				},
			},
			// Step 2: Update resource - CHANGE THE NAME
			// This is the CRITICAL test for usePriorState functionality
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
					"name":       config.StringVariable(updatedName), // New name!
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("name"),
						knownvalue.StringExact(updatedName),
					),
				},
			},
		},
	})

	// Verify the UPDATE request was captured by the framework
	capturedRequests := updateEndpoint.Captured()
	require.NotEmpty(t, capturedRequests, "UPDATE request should have been captured")

	// Get the first (and should be only) UPDATE request
	updateReq := capturedRequests[0]

	// CRITICAL ASSERTION: Verify prior state was used for path parameter
	require.Equal(t, initialName, updateReq.PathParams["name"],
		"Path parameter should use PRIOR state value (initial-name), not updated value")

	// The body should contain 'new_name' field with the NEW value
	// (The OpenAPI spec uses 'new_name' which gets mapped to 'name' in Terraform via x-speakeasy-name-override)
	bodyNewName, hasNewName := updateReq.Body["new_name"]
	require.True(t, hasNewName, "Request body should contain 'new_name' field")
	require.Equal(t, updatedName, bodyNewName,
		"Request body 'new_name' should contain the NEW name (updated-name)")
}
