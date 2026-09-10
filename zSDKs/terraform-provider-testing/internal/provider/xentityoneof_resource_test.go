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

// TestXEntityOneOfHoistedFieldNoDrift verifies that hoisted readonly fields retain API values
// and don't cause drift. Tests the UseHoistedValue plan modifier for TFGEN-276.
func TestXEntityOneOfHoistedFieldNoDrift(t *testing.T) {
	t.Parallel()

	serverSharedNameValue := "server-generated-value"
	// json.Marshal sorts map keys, so the normalized state value is in key order.
	serverConfigValue := `{"enabled":true,"mode":"test-mode"}`
	responseOverlay := map[string]any{
		"discriminator_property": "branch-one",
		"branch_one_config":      "test-config",
		"shared_name":            serverSharedNameValue,
		"config": map[string]any{
			"mode":    "test-mode",
			"enabled": true,
		},
	}

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{{Endpoint: "POST /v0/xentity/oneOf", ResponseOverlay: responseOverlay}},
		Get:    tfmockserver.Endpoints{{Endpoint: "GET /v0/xentity/oneOf/{id}", ResponseOverlay: responseOverlay}},
		Update: tfmockserver.Endpoints{{Endpoint: "PATCH /v0/xentity/oneOf/{id}", ResponseOverlay: responseOverlay}},
		Delete: tfmockserver.Endpoints{{Endpoint: "DELETE /v0/xentity/oneOf/{id}"}},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_x_entity_one_of.test"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables:          config.Variables{"server_url": config.StringVariable(mockServer.URL)},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceAddress, tfjsonpath.New("shared_name"), knownvalue.StringExact(serverSharedNameValue)),
					statecheck.ExpectKnownValue(resourceAddress, tfjsonpath.New("config"), knownvalue.StringExact(serverConfigValue)),
					statecheck.ExpectKnownValue(resourceAddress, tfjsonpath.New("branch_one").AtMapKey("config"), knownvalue.StringExact(serverConfigValue)),
				},
			},
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables:          config.Variables{"server_url": config.StringVariable(mockServer.URL)},
			},
		},
	})
}
