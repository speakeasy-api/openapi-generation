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

func TestXEntityOperationCloseSingleOp(t *testing.T) {
	t.Parallel()

	endpoints := tfmockserver.ResourceEndpoints{
		Open: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/xentityoperation/close-single-op",
				ResponseOverlay: map[string]any{
					"string_response_only": "computed-value",
				},
			},
		},
		Close: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/xentityoperation/close-single-op",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	// Ephemeral resource testing must use echo provider.
	echoResourceAddress := "echo.test"

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
						echoResourceAddress,
						tfjsonpath.New("data").AtMapKey("string_response_only"),
						knownvalue.StringExact("computed-value"),
					),
				},
			},
		},
	})

	// Verify the close endpoint was called.
	closeRequests := endpoints.Close[0].Requests.All()

	if len(closeRequests) == 0 {
		t.Fatal("expected close endpoint to be called, but it was not")
	}

	// Verify the close request body contains the expected data from private state.
	closeBody := closeRequests[0].Body
	if closeBody["string_request_and_response"] == nil {
		t.Fatal("expected close request body to contain string_request_and_response")
	}
}
