package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-testing/internal/provider"
	"github.com/hashicorp/terraform-provider-testing/internal/tfmockserver"
)

// The data source schema hides the const properties and reads a response
// that carries them.
func TestConstContainerDataSource(t *testing.T) {
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

	dataResourceAddress := "data.testing_const_container.test"
	managedResourceAddress := "testing_const_container.test"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						managedResourceAddress,
						tfjsonpath.New("name"),
						dataResourceAddress,
						tfjsonpath.New("name"),
						compare.ValuesSame(),
					),
				},
			},
		},
	})
}
