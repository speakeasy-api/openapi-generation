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

func TestXClientFilterArrayDataSource(t *testing.T) {
	t.Parallel()

	endpoints := tfmockserver.ResourceEndpoints{
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/xclientfilterarray",
				Response: map[string]any{
					"data": []map[string]any{
						{
							"id":       "cf-001",
							"name":     "alpha",
							"category": "system",
						},
						{
							"id":       "cf-002",
							"name":     "beta",
							"category": "user",
						},
						{
							"id":       "cf-003",
							"name":     "alpha",
							"category": "admin",
						},
					},
				},
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "data.testing_x_client_filter_array.test"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Filters the response array by name and returns all matching items.
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
					"name":       config.StringVariable("alpha"),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					// Should return 2 items matching name=alpha.
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("data"),
						knownvalue.ListSizeExact(2),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("data").AtSliceIndex(0).AtMapKey("id"),
						knownvalue.StringExact("cf-001"),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("data").AtSliceIndex(0).AtMapKey("name"),
						knownvalue.StringExact("alpha"),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("data").AtSliceIndex(1).AtMapKey("id"),
						knownvalue.StringExact("cf-003"),
					),
				},
			},
		},
	})
}

func TestXClientFilterArrayDataSourceNoFilter(t *testing.T) {
	t.Parallel()

	endpoints := tfmockserver.ResourceEndpoints{
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/xclientfilterarray",
				Response: map[string]any{
					"data": []map[string]any{
						{
							"id":       "cf-001",
							"name":     "alpha",
							"category": "system",
						},
						{
							"id":       "cf-002",
							"name":     "beta",
							"category": "user",
						},
					},
				},
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "data.testing_x_client_filter_array.test"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Without a filter, returns all items in the response array.
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("data"),
						knownvalue.ListSizeExact(2),
					),
				},
			},
		},
	})
}
