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

func TestXTerraformIgnoreResourceLifecycle(t *testing.T) {
	t.Parallel()

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/xterraformignore",
			},
			{
				Endpoint: "POST /v0/xterraformignore/{id}/task",
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/xterraformignore/{id}",
			},
			{
				Endpoint: "GET /v0/xterraformignore/{id}/task/{taskId}",
				ResponseOverlay: map[string]any{
					"status": shared.XPollingStatusCompleted,
				},
			},
		},
		Update: tfmockserver.Endpoints{
			{
				Endpoint: "PATCH /v0/xterraformignore/{id}",
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/xterraformignore/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_x_terraform_ignore.test"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Verifies resource create and read.
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"configurable": config.StringVariable("test-original"),
					"server_url":   config.StringVariable(mockServer.URL),
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
				},
			},
			// Verifies import.
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"configurable": config.StringVariable("test-original"),
					"server_url":   config.StringVariable(mockServer.URL),
				},
				ResourceName:      resourceAddress,
				ImportState:       true,
				ImportStateVerify: true,
				// Secondary operation is not implemented with read support and
				// therefore import support, which is fine for this particular
				// resource and its testing of x-speakeasy-terraform-ignore.
				ImportStateVerifyIgnore: []string{
					"secondary_operation_configurable",
					"status",
				},
			},
			// Verifies resource update.
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"configurable": config.StringVariable("test-updated"),
					"server_url":   config.StringVariable(mockServer.URL),
				},
				// Check computed values.
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("status"),
						knownvalue.StringExact(string(shared.XPollingStatusCompleted)),
					),
				},
			},
			// Testing framework implicitly verifies resource delete.
		},
	})
}
