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

func TestXPaginationSingletonListDataSource(t *testing.T) {
	t.Parallel()

	// Define test variables
	filterName := "test-filter"

	// Mock server with empty data response
	endpoints := tfmockserver.ResourceEndpoints{
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/xpagination/singletonAndList",
				Response: map[string]any{
					"data": []map[string]any{}, // Empty data array
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

	dataResourceAddress := "data.testing_x_pagination_singleton_list.test"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url":  config.StringVariable(mockServer.URL),
					"filter_name": config.StringVariable(filterName),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify filter_name is set correctly
					statecheck.ExpectKnownValue(
						dataResourceAddress,
						tfjsonpath.New("filter_name"),
						knownvalue.StringExact(filterName),
					),
					// Verify the data array is empty
					statecheck.ExpectKnownValue(
						dataResourceAddress,
						tfjsonpath.New("data"),
						knownvalue.ListExact([]knownvalue.Check{}),
					),
				},
			},
		},
	})
}
