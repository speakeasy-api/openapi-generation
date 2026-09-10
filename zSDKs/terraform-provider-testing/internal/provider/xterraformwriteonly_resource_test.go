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

func TestXTerraformWriteOnlyResourceLifecycle(t *testing.T) {
	t.Parallel()

	ephemeralStringOriginal := "test"
	ephemeralStringUpdated := "test-updated"
	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/xterraformwriteonly",
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/xterraformwriteonly/{id}",
			},
		},
		Update: tfmockserver.Endpoints{
			{
				Endpoint: "PATCH /v0/xterraformwriteonly/{id}",
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/xterraformwriteonly/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_x_terraform_write_only.test"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"ephemeral_string": config.StringVariable(ephemeralStringOriginal),
					"server_url":       config.StringVariable(mockServer.URL),
					"string":           config.StringVariable(ephemeralStringOriginal),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("id"),
						knownvalue.StringExact(tfmockserver.StoreKey),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("string"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("string_with_oas_writeonly"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("string_without_writeonly"),
						knownvalue.StringExact(ephemeralStringOriginal),
					),
				},
			},
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"ephemeral_string": config.StringVariable(ephemeralStringOriginal),
					"server_url":       config.StringVariable(mockServer.URL),
					"string":           config.StringVariable(ephemeralStringOriginal),
				},
				ResourceName:      resourceAddress,
				ImportState:       true,
				ImportStateVerify: true,
				// Terraform intentionally cannot import write only attributes.
				ImportStateVerifyIgnore: []string{
					"string",
					"string_with_oas_writeonly",
				},
			},
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"ephemeral_string": config.StringVariable(ephemeralStringUpdated),
					"server_url":       config.StringVariable(mockServer.URL),
					"string":           config.StringVariable(ephemeralStringUpdated),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("string"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("string_with_oas_writeonly"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("string_without_writeonly"),
						knownvalue.StringExact(ephemeralStringUpdated),
					),
				},
			},
		},
	})

	capturedCreateRequest := endpoints.Create[0].Captured()

	if len(capturedCreateRequest) != 1 {
		t.Fatalf("Expected 1 create request, got %d", len(capturedCreateRequest))
	}

	capturedCreateRequestBody := capturedCreateRequest[0].Body

	tfmockserver.AssertRequestBodyField(t, capturedCreateRequestBody, "string", ephemeralStringOriginal)
	tfmockserver.AssertRequestBodyField(t, capturedCreateRequestBody, "string_with_oas_writeonly", ephemeralStringOriginal)

	capturedUpdateRequest := endpoints.Update[0].Captured()

	if len(capturedUpdateRequest) != 1 {
		t.Fatalf("Expected 1 update request, got %d", len(capturedUpdateRequest))
	}

	capturedUpdateRequestBody := capturedUpdateRequest[0].Body

	tfmockserver.AssertRequestBodyField(t, capturedUpdateRequestBody, "string", ephemeralStringUpdated)
	tfmockserver.AssertRequestBodyField(t, capturedUpdateRequestBody, "string_with_oas_writeonly", ephemeralStringUpdated)
}
