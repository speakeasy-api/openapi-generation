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

// The request schema declares const values on array and map properties. The
// provider must hide them from the resource schema and send them verbatim,
// and a read must accept the const values however the server formats them.
func TestConstContainerResourceLifecycle(t *testing.T) {
	t.Parallel()

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/constcontainers",
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/constcontainer/{id}",
				// Return the const values with whitespace, map key order and
				// number formatting the client did not send.
				ResponseFuncWithStore: func(r *http.Request, w http.ResponseWriter, storedData map[string]any) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{
						"id": "` + tfmockserver.StoreKey + `",
						"name": "` + storedData["name"].(string) + `",
						"empty_list": [ ],
						"fixed_list": [ "read" , "write" ],
						"empty_map": { },
						"fixed_map": { "max" : 1e1 }
					}`))
				},
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/constcontainer/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_const_container.test"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
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
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-const-container"),
					),
				},
			},
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

	creates := endpoints.Create[0].Requests.Bodies()
	require.Len(t, creates, 1)
	body := creates[0]
	tfmockserver.AssertRequestBodyField(t, body, "name", "test-const-container")
	tfmockserver.AssertRequestBodyFieldEmptyArray(t, body, "empty_list")
	tfmockserver.AssertRequestBodyFieldArray(t, body, "fixed_list", []any{"read", "write"})
	tfmockserver.AssertRequestBodyField(t, body, "empty_map", map[string]any{})
	tfmockserver.AssertRequestBodyField(t, body, "fixed_map", map[string]any{"max": float64(10)})
	require.Len(t, body, 5, "request body carries only name and the four const fields: %v", body)
}
