package provider_test

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-testing/internal/provider"
	"github.com/hashicorp/terraform-provider-testing/internal/tfmockserver"
)

func TestImportMultipleIDResourceLifecycle(t *testing.T) {
	t.Parallel()

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/import-multiple-id/{param1}/{param2}",
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/import-multiple-id/{param1}/{param2}",
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/import-multiple-id/{param1}/{param2}",
			},
		},
	}
	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_import_multiple_id.my_importmultipleid"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Verifies resource create and read.
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
				},
				// Check computed values.
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("param1"),
						knownvalue.StringExact("test-param1"),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("param2"),
						knownvalue.StringExact("test-param2"),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("request_body_property"),
						knownvalue.StringExact("test-request-body"),
					),
				},
			},

			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
				},
				ResourceName: resourceAddress,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					importIDBytes, err := json.Marshal(struct {
						Param1 string `json:"param1"`
						Param2 string `json:"param2"`
					}{
						Param1: s.RootModule().Resources[resourceAddress].Primary.Attributes["param1"],
						Param2: s.RootModule().Resources[resourceAddress].Primary.Attributes["param2"],
					})

					return string(importIDBytes), err
				},
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "param1",
			},
			// Testing framework implicitly verifies resource delete.
		},
	})
}
