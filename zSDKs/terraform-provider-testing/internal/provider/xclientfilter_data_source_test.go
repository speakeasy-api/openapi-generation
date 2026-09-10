package provider_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-testing/internal/provider"
	"github.com/hashicorp/terraform-provider-testing/internal/tfmockserver"
)

func TestXClientFilterDataSource(t *testing.T) {
	t.Parallel()

	endpoints := tfmockserver.ResourceEndpoints{
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/xclientfilter",
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
							"name":     "gamma",
							"category": "system",
						},
					},
				},
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "data.testing_x_client_filter.test"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Filters the response array by name and returns the matching item.
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
					"name":       config.StringVariable("beta"),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("id"),
						knownvalue.StringExact("cf-002"),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("name"),
						knownvalue.StringExact("beta"),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("category"),
						knownvalue.StringExact("user"),
					),
				},
			},
		},
	})
}

func TestXClientFilterDataSourceNoMatch(t *testing.T) {
	t.Parallel()

	endpoints := tfmockserver.ResourceEndpoints{
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/xclientfilter",
				Response: map[string]any{
					"data": []map[string]any{
						{
							"id":       "cf-001",
							"name":     "alpha",
							"category": "system",
						},
					},
				},
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Filter value does not match any items — expects a diagnostic error.
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
					"name":       config.StringVariable("nonexistent"),
				},
				ExpectError: regexp.MustCompile("No matching items found"),
			},
		},
	})
}

func TestXClientFilterDataSourceEmptyResponse(t *testing.T) {
	t.Parallel()

	endpoints := tfmockserver.ResourceEndpoints{
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/xclientfilter",
				Response: map[string]any{
					"data": []map[string]any{},
				},
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// API returns empty array — error should indicate missing data, not filter mismatch.
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
					"name":       config.StringVariable("anything"),
				},
				ExpectError: regexp.MustCompile("Missing response body array data"),
			},
		},
	})
}

func TestXClientFilterDataSourceNoFilter(t *testing.T) {
	t.Parallel()

	endpoints := tfmockserver.ResourceEndpoints{
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/xclientfilter",
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

	resourceAddress := "data.testing_x_client_filter.test"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Without a filter, returns the first item in the response array.
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
						knownvalue.StringExact("cf-001"),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("name"),
						knownvalue.StringExact("alpha"),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("category"),
						knownvalue.StringExact("system"),
					),
				},
			},
		},
	})
}
