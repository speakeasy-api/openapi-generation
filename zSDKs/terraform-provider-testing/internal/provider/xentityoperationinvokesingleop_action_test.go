package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-testing/internal/provider"
	"github.com/hashicorp/terraform-provider-testing/internal/tfmockserver"
)

func TestXEntityOperationInvokeSingleOp(t *testing.T) {
	t.Parallel()

	invokeEndpoint := &tfmockserver.ResourceEndpoint{
		Endpoint: "POST /v0/xentityoperation/invoke-single-op",
	}

	endpoints := tfmockserver.ResourceEndpoints{
		Invoke: tfmockserver.Endpoints{
			invokeEndpoint,
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

	captured := invokeEndpoint.Captured()
	if len(captured) == 0 {
		t.Fatal("expected at least one invoke request to the mock server, got none")
	}

	if captured[0].Body["string_request_only"] != "test-request-only" {
		t.Errorf("expected string_request_only = 'test-request-only', got: %v", captured[0].Body["string_request_only"])
	}

	if captured[0].Body["string_request_and_response"] != "test-request-and-response" {
		t.Errorf("expected string_request_and_response = 'test-request-and-response', got: %v", captured[0].Body["string_request_and_response"])
	}
}
