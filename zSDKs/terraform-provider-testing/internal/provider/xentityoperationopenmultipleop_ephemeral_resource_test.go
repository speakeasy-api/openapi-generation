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

func TestXEntityOperationOpenMultipleOp(t *testing.T) {
	t.Parallel()

	endpoints := tfmockserver.ResourceEndpoints{
		Open: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/xentityoperation/open-multiple-op",
				ResponseOverlay: map[string]any{
					"op1_string_for_op2":       "op1-string-for-op2-value",
					"op1_string_response_only": "computed-value-1",
				},
			},
			{
				Endpoint: "POST /v0/xentityoperation/open-multiple-op/op1-string-for-op2-value",
				ResponseOverlay: map[string]any{
					"op2_string_response_only": "computed-value-2",
				},
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	// Ephemeral resource testing must use echo provider.
	echoResourceAddress := "echo.test"

	resource.UnitTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						echoResourceAddress,
						tfjsonpath.New("data").AtMapKey("op1_string_response_only"),
						knownvalue.StringExact("computed-value-1"),
					),
					statecheck.ExpectKnownValue(
						echoResourceAddress,
						tfjsonpath.New("data").AtMapKey("op2_string_response_only"),
						knownvalue.StringExact("computed-value-2"),
					),
				},
			},
		},
	})
}
