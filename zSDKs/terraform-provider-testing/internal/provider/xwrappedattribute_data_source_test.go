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

func TestXWrappedAttributeDataSource(t *testing.T) {
	t.Parallel()

	endpoints := tfmockserver.ResourceEndpoints{
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/xwrappedattribute/{id}",
				Response: map[string]any{
					"id":                 "test-id",
					"non_wrapped_string": "test-non-wrapped",
				},
			},
			{
				Endpoint: "GET /v0/xwrappedattribute/{id}/wrapped",
				Response: map[string]any{
					"wrapped_list_object": []map[string]any{
						{
							"wrapped_list_object_string": "wrapped-value-1",
						},
					},
				},
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	dataSourceAddress := "data.testing_x_wrapped_attribute.test"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Verifies data source read, exercising both
			// RefreshFromSharedXWrappedAttributeResponse and
			// RefreshFromOperationsGetXSpeakeasyWrappedAttributeWrappedResponseBody.
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						dataSourceAddress,
						tfjsonpath.New("id"),
						knownvalue.StringExact("test-id"),
					),
					statecheck.ExpectKnownValue(
						dataSourceAddress,
						tfjsonpath.New("non_wrapped_string"),
						knownvalue.StringExact("test-non-wrapped"),
					),
					statecheck.ExpectKnownValue(
						dataSourceAddress,
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
		},
	})
}
