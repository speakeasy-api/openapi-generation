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

// TestOASWriteOnlyNestedResourceLifecycle tests TFGEN-226: Generation for OPS schema with writeOnly properties
// This test verifies that nested objects with all writeOnly properties generate non-empty types
// and that writeOnly fields are properly preserved during create/read cycles.
func TestOASWriteOnlyNestedResourceLifecycle(t *testing.T) {
	t.Parallel()

	responseOverlay := map[string]any{
		"id":     "test-id",
		"name":   "test-ssh-cred",
		"status": "active",
		"ssh_credential": map[string]any{
			"keys": map[string]any{
				"key_type": "ssh",
			},
		},
	}

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint:        "POST /v0/oaswriteonly-nested",
				ResponseOverlay: responseOverlay,
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint:        "GET /v0/oaswriteonly-nested",
				ResponseOverlay: responseOverlay,
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/oaswriteonly-nested",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_oas_write_only_nested.test"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Verifies resource create and read with writeOnly nested properties.
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
						knownvalue.StringExact("test-id"),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-ssh-cred"),
					),
					// Verify writeOnly fields are preserved in state
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("ssh_credential").AtMapKey("keys").AtMapKey("private_key"),
						knownvalue.StringExact("test-private-key"),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("ssh_credential").AtMapKey("keys").AtMapKey("passphrase"),
						knownvalue.StringExact("test-passphrase"),
					),
					// Verify readOnly field from response is present
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("status"),
						knownvalue.StringExact("active"),
					),
				},
			},
			// Testing framework implicitly verifies resource delete.
		},
	})
}
