package provider_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-testing/internal/provider"
	"github.com/hashicorp/terraform-provider-testing/internal/tfmockserver"
)

func TestXWrappedAttributeResourceLifecycle(t *testing.T) {
	t.Parallel()

	wrappedResponse := map[string]any{
		"wrapped_list_object": []map[string]any{
			{
				"wrapped_list_object_string": "wrapped-value-1",
			},
		},
	}

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/xwrappedattribute",
			},
			{
				// The wrapped endpoint is called as create#2 after the main
				// create. Use ResponseFunc to return a fixed response without
				// interfering with the shared mock store.
				Endpoint: "PUT /v0/xwrappedattribute/{id}/wrapped",
				ResponseFunc: func(r *http.Request, w http.ResponseWriter) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(200)
					json.NewEncoder(w).Encode(wrappedResponse)
				},
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/xwrappedattribute/{id}",
			},
			{
				Endpoint: "GET /v0/xwrappedattribute/{id}/wrapped",
				Response: wrappedResponse,
			},
		},
		Update: tfmockserver.Endpoints{
			{
				Endpoint: "PUT /v0/xwrappedattribute/{id}",
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/xwrappedattribute/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_x_wrapped_attribute.test"

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
						tfjsonpath.New("wrapped_list_string"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("test-value-1"),
						}),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("wrapper"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"wrapped_list_object": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"wrapped_list_object_string": knownvalue.StringExact("wrapped-value-1"),
								}),
							}),
						}),
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
				// This is similar to a real world scenario where the request
				// accepts a list of strings but the response returns a list of
				// objects instead. Without additional configuration, such as
				// x-speakeasy-transform-from-api on the response, the import
				// (read) data would not be aligned. The test endpoints do not
				// need the additional setup for perfection here.
				ImportStateVerifyIgnore: []string{"wrapped_list_string"},
			},
		},
	})
}
