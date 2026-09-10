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

func TestXEntityOperationCloseMultipleOp(t *testing.T) {
	t.Parallel()

	endpoints := tfmockserver.ResourceEndpoints{
		Open: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/xentityoperation/close-multiple-op",
				ResponseOverlay: map[string]any{
					"string_response_only": "computed-value",
				},
			},
		},
		Close: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/xentityoperation/close-multiple-op",
			},
			{
				Endpoint: "DELETE /v0/xentityoperation/close-multiple-op/{string_request_and_response}",
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

	// Verify the first close endpoint was called.
	close1Requests := endpoints.Close[0].Requests.All()

	if len(close1Requests) == 0 {
		t.Fatal("expected first close endpoint to be called, but it was not")
	}

	// Verify the second close endpoint was called.
	close2Requests := endpoints.Close[1].Requests.All()

	if len(close2Requests) == 0 {
		t.Fatal("expected second close endpoint to be called, but it was not")
	}
}
