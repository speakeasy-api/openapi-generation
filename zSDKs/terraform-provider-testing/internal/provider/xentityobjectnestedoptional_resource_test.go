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

func TestXEntityObjectNestedOptionalResource(t *testing.T) {
	t.Parallel()

	entityObjectString := "entity-object-string-value"
	id := "id-value"
	piblingObjectString := "pibling-object-string-value"
	piblingObjectReadonlyStringReadonly := "pibling-object-readonly-string-readonly-value"
	piblingString := "pibling-string-value"
	responseOverlay := map[string]any{
		"data": map[string]any{
			"entity_object_string": entityObjectString,
			"id":                   id,
		},
		"pibling_object": map[string]any{
			"pibling_object_string": piblingObjectString,
		},
		"pibling_object_readonly": map[string]any{
			"pibling_object_readonly_string_readonly": piblingObjectReadonlyStringReadonly,
		},
		"pibling_string": piblingString,
	}

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint:        "POST /v0/xentity/object/nested/optional",
				ResponseOverlay: responseOverlay,
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint:        "GET /v0/xentity/object/nested/optional/{id}",
				ResponseOverlay: responseOverlay,
			},
		},
		Update: tfmockserver.Endpoints{
			{
				Endpoint:        "PUT /v0/xentity/object/nested/optional/{id}",
				ResponseOverlay: responseOverlay,
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/xentity/object/nested/optional/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_x_entity_object_nested_optional.test"

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
						tfjsonpath.New("entity_object_string"),
						knownvalue.StringExact(entityObjectString),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("pibling_object_string"),
						knownvalue.StringExact(piblingObjectString),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("pibling_object_readonly_string_readonly"),
						knownvalue.StringExact(piblingObjectReadonlyStringReadonly),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("pibling_string"),
						knownvalue.StringExact(piblingString),
					),
				},
			},
		},
	})
}
