package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-testing/internal/provider"
	"github.com/hashicorp/terraform-provider-testing/internal/sdk/models/shared"
	"github.com/hashicorp/terraform-provider-testing/internal/tfmockserver"
)

func TestXPollingResourceLifecycle(t *testing.T) {
	t.Parallel()

	// var deleted bool

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/xpolling",
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/xpolling/{id}",
				// ResponseFunc: func(r *http.Request, w http.ResponseWriter) {
				// 	if deleted {
				// 		w.WriteHeader(404)
				// 		return
				// 	}

				// 	tfmockserver.JSONResponse(w, 200, map[string]any{
				// 		"status": shared.XPollingStatusCompleted,
				// 	})
				// },
				ResponseOverlay: map[string]any{
					"status":                          shared.XPollingStatusCompleted,
					"status_to_be_name_overridden": shared.XPollingStatusCompleted,
				},
			},
		},
		Update: tfmockserver.Endpoints{
			{
				Endpoint: "PATCH /v0/xpolling/{id}",
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/xpolling/{id}",
				// ResponseFunc: func(r *http.Request, w http.ResponseWriter) {
				// 	deleted = true
				// 	w.WriteHeader(200)
				// },
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_x_polling.test"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Verifies resource create and read.
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"name":       config.StringVariable("test-original"),
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
						tfjsonpath.New("status"),
						knownvalue.StringExact(string(shared.XPollingStatusCompleted)),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("status_name_overridden"),
						knownvalue.StringExact(string(shared.XPollingStatusCompleted)),
					),
				},
			},
			// Verifies import.
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"name":       config.StringVariable("test-original"),
					"server_url": config.StringVariable(mockServer.URL),
				},
				ResourceName:      resourceAddress,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Verifies resource update.
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"name":       config.StringVariable("test-updated"),
					"server_url": config.StringVariable(mockServer.URL),
				},
				// Check computed values.
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("status"),
						knownvalue.StringExact(string(shared.XPollingStatusCompleted)),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("status_name_overridden"),
						knownvalue.StringExact(string(shared.XPollingStatusCompleted)),
					),
				},
			},
			// Testing framework implicitly verifies resource delete.
		},
	})
}
