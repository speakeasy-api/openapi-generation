package provider_test

import (
	"encoding/base64"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-testing/internal/provider"
	"github.com/hashicorp/terraform-provider-testing/internal/tfmockserver"
)

func TestOASFormatBinaryJSONResourceLifecycle(t *testing.T) {
	t.Parallel()

	// "Hello, World!" base64-encoded
	contentBytes := []byte("Hello, World!")
	contentBase64 := base64.StdEncoding.EncodeToString(contentBytes)

	// Server-side computed content (returned in response overlay)
	computedContentBytes := []byte("computed-binary-data")
	computedContentBase64 := base64.StdEncoding.EncodeToString(computedContentBytes)

	responseOverlay := map[string]any{
		"computed_content": computedContentBase64,
		"size":            len(contentBytes),
	}

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint:        "POST /v0/oasformat/binary/json",
				ResponseOverlay: responseOverlay,
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint:        "GET /v0/oasformat/binary/json/{id}",
				ResponseOverlay: responseOverlay,
			},
		},
		Update: tfmockserver.Endpoints{
			{
				Endpoint:        "PATCH /v0/oasformat/binary/json/{id}",
				ResponseOverlay: responseOverlay,
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/oasformat/binary/json/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_oas_format_binary_json.test"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Verifies resource create and read with binary content.
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
					"content":    config.StringVariable(contentBase64),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("id"),
						knownvalue.StringExact(tfmockserver.StoreKey),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-oas-format-binary-json"),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("content"),
						knownvalue.StringExact(contentBase64),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("computed_content"),
						knownvalue.StringExact(computedContentBase64),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("size"),
						knownvalue.Int64Exact(int64(len(contentBytes))),
					),
				},
			},
			// Import resource
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
					"content":    config.StringVariable(contentBase64),
				},
				ResourceName:      resourceAddress,
				ImportState:       true,
				ImportStateId:     tfmockserver.StoreKey,
				ImportStateVerify: true,
			},
			// Testing framework implicitly verifies resource delete.
		},
	})
}
