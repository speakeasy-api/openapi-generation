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

func TestOASReadOnlyResourceLifecycle(t *testing.T) {
	t.Parallel()

	id := "test-id"
	inlineString := "inline-string-value"
	responseOverlay := map[string]any{
		"id": id,
		"inline_object_request_only": map[string]any{
			"inline_string": inlineString,
		},
		"inline_object_request_response": map[string]any{
			"inline_string": inlineString,
		},
		"inline_object_response_only": map[string]any{
			"inline_string": inlineString,
		},
		"ref_object_request_only": map[string]any{
			"inline_string": inlineString,
		},
		"ref_object_request_response": map[string]any{
			"inline_string": inlineString,
		},
		"ref_object_response_only": map[string]any{
			"inline_string": inlineString,
		},
	}

	serverEndpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint:        "POST /v0/oasreadonly",
				ResponseOverlay: responseOverlay,
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint:        "GET /v0/oasreadonly/{id}",
				ResponseOverlay: responseOverlay,
			},
		},
		Update: tfmockserver.Endpoints{
			{
				Endpoint:        "PATCH /v0/oasreadonly/{id}",
				ResponseOverlay: responseOverlay,
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/oasreadonly/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(serverEndpoints, t)
	defer mockServer.Close()
	resourceAddress := "testing_oas_read_only.test"

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
						resourceAddress,
						tfjsonpath.New("id"),
						knownvalue.StringExact(id),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("inline_object_request_only"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"inline_string": knownvalue.StringExact(inlineString),
						}),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("inline_object_request_response"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"inline_string": knownvalue.StringExact(inlineString),
						}),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("inline_object_response_only"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"inline_string": knownvalue.StringExact(inlineString),
						}),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("ref_object_request_only"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"inline_string": knownvalue.StringExact(inlineString),
						}),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("ref_object_request_response"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"inline_string": knownvalue.StringExact(inlineString),
						}),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("ref_object_response_only"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"inline_string": knownvalue.StringExact(inlineString),
						}),
					),
				},
			},
		},
	})
}
