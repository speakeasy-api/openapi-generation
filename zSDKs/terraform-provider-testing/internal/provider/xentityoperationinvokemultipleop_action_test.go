package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-testing/internal/provider"
	"github.com/hashicorp/terraform-provider-testing/internal/tfmockserver"
)

func TestXEntityOperationInvokeMultipleOp(t *testing.T) {
	t.Parallel()

	invokeEndpoint1 := &tfmockserver.ResourceEndpoint{
		Endpoint: "POST /v0/xentityoperation/invoke-multiple-op",
		ResponseOverlay: map[string]any{
			"op1_string_for_op2": "op1-string-for-op2-value",
		},
	}
	invokeEndpoint2 := &tfmockserver.ResourceEndpoint{
		Endpoint: "POST /v0/xentityoperation/invoke-multiple-op/op1-string-for-op2-value",
	}

	endpoints := tfmockserver.ResourceEndpoints{
		Invoke: tfmockserver.Endpoints{
			invokeEndpoint1,
			invokeEndpoint2,
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		Steps: []resource.TestStep{
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
				},
			},
		},
	})

	captured1 := invokeEndpoint1.Captured()
	if len(captured1) == 0 {
		t.Fatal("expected at least one invoke request to endpoint 1, got none")
	}

	// Verify op1 request body contains the expected fields
	tfmockserver.AssertRequestBodyField(t, captured1[0].Body, "op1_string_request_only", "test-op1-request-only")
	tfmockserver.AssertRequestBodyField(t, captured1[0].Body, "op1_string_request_and_response", "test-op1-request-and-response")

	captured2 := invokeEndpoint2.Captured()
	if len(captured2) == 0 {
		t.Fatal("expected at least one invoke request to endpoint 2, got none")
	}

	// Verify op2 path includes the op1 response value (op1_string_for_op2),
	// confirming the intermediate response flow between operations works.
	if captured2[0].Path != "/v0/xentityoperation/invoke-multiple-op/op1-string-for-op2-value" {
		t.Errorf("expected endpoint 2 path to include op1 response value, got: %v", captured2[0].Path)
	}

	// Verify op2 request body contains the expected fields
	tfmockserver.AssertRequestBodyField(t, captured2[0].Body, "op2_string_request_only", "test-op2-request-only")
	tfmockserver.AssertRequestBodyField(t, captured2[0].Body, "op2_string_request_and_response", "test-op2-request-and-response")
}
