package provider_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-testing/internal/provider"
	"github.com/hashicorp/terraform-provider-testing/internal/tfmockserver"
)

func TestXPaginationSingletonDataSource(t *testing.T) {
	t.Parallel()

	// Define test variables
	filterName := "test-filter"

	// Mock server with empty data response
	endpoints := tfmockserver.ResourceEndpoints{
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/xpagination/singletonAndList",
				Response: map[string]any{
					"data": []map[string]any{}, // Empty data array - should cause error for singleton
					"meta": map[string]any{
						"page": map[string]any{
							"number": 1,
							"size":   10,
							"total":  0,
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
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url":  config.StringVariable(mockServer.URL),
					"filter_name": config.StringVariable(filterName),
				},
				// Expect an error because the singleton data source requires at least one item
				ExpectError: regexp.MustCompile("Missing response body array data"),
			},
		},
	})
}

// TestXPaginationSingletonDataSourceNoPagination tests TFGEN-209: ensures that
// pagination logic is NOT applied to singleton extraction data sources, even when
// the underlying operation has x-speakeasy-pagination configured.
// Before the fix, this would panic with "index out of range" when pagination tried
// to call the Next() method and extract singleton from an empty subsequent page.
func TestXPaginationSingletonDataSourceNoPagination(t *testing.T) {
	t.Parallel()

	// Define test variables
	filterName := "test-item"

	// Mock server that returns data on first call
	// If pagination was incorrectly applied, it would try to call Next() which would
	// make additional HTTP requests beyond the expected 3 (one per terraform operation: plan, refresh, apply)
	getEndpoint := &tfmockserver.ResourceEndpoint{
		Endpoint: "GET /v0/xpagination/singletonAndList",
		Response: map[string]any{
			"data": []map[string]any{
				{
					"id":   "singleton-1",
					"name": "Test Singleton Item",
				},
			},
			"meta": map[string]any{
				"page": map[string]any{
					"number": 1,
					"size":   10,
					"total":  1,
				},
			},
		},
	}

	endpoints := tfmockserver.ResourceEndpoints{
		Get: tfmockserver.Endpoints{getEndpoint},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url":  config.StringVariable(mockServer.URL),
					"filter_name": config.StringVariable(filterName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.testing_x_pagination_singleton.test", "id", "singleton-1"),
					resource.TestCheckResourceAttr("data.testing_x_pagination_singleton.test", "name", "Test Singleton Item"),
				),
			},
		},
	})

	// Assert that exactly 3 requests were made (no pagination)
	// Terraform data sources make 3 requests during plan/refresh/apply phases
	// If pagination was incorrectly applied, there would be many more requests
	requests := getEndpoint.Requests.All()
	expectedCount := 3
	actualCount := len(requests)
	if actualCount != expectedCount {
		t.Errorf("TFGEN-209 regression: expected exactly %d requests (no pagination), got %d requests", expectedCount, actualCount)
	}
}
