package provider_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-testing/internal/provider"
	"github.com/hashicorp/terraform-provider-testing/internal/tfmockserver"
	"github.com/stretchr/testify/require"
)

// Create and update replace the whole group set through one operation. Delete
// is a separate operation, under another API prefix, whose body declares the
// same set with an empty const: the provider must send [] there, not the
// groups held in state. The response carries a readOnly const array that must
// never reach a request body.
func TestConstEmptySetDeleteResourceLifecycle(t *testing.T) {
	t.Parallel()

	supportedActions := map[string]any{"supported_actions": []any{"replace"}}
	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint:        "PUT /v0/constemptysetdelete",
				ResponseOverlay: supportedActions,
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint:        "GET /v0/constemptysetdelete/{id}",
				ResponseOverlay: supportedActions,
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "PUT /v1/constemptysetdelete/teardown",
				ResponseOverlay: map[string]any{
					"id":                tfmockserver.StoreKey,
					"scope_id":          "scope-1",
					"groups":            []any{},
					"supported_actions": []any{"replace"},
				},
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_const_empty_set_delete.test"
	variables := func(replicas int64) config.Variables {
		return config.Variables{
			"server_url": config.StringVariable(mockServer.URL),
			"replicas":   config.IntegerVariable(replicas),
		}
	}

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables:          variables(2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("id"),
						knownvalue.StringExact(tfmockserver.StoreKey),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("groups"),
						knownvalue.SetExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"provider": knownvalue.StringExact("aws"),
								"region":   knownvalue.StringExact("eu-west-1"),
								"replicas": knownvalue.Int64Exact(2),
							}),
						}),
					),
				},
			},
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables:          variables(2),
				PlanOnly:                 true,
			},
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables:          variables(3),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("groups").AtSliceIndex(0).AtMapKey("replicas"),
						knownvalue.Int64Exact(3),
					),
				},
			},
		},
	})

	group := func(replicas float64) map[string]any {
		return map[string]any{"provider": "aws", "region": "eu-west-1", "replicas": replicas}
	}
	upserts := endpoints.Create[0].Requests.Bodies()
	require.Len(t, upserts, 2, "create and update both go through the replace operation")
	for i, replicas := range []float64{2, 3} {
		tfmockserver.AssertRequestBodyField(t, upserts[i], "scope_id", "scope-1")
		tfmockserver.AssertRequestBodyFieldArray(t, upserts[i], "groups", []any{group(replicas)})
		tfmockserver.AssertRequestBodyFieldOmitted(t, upserts[i], "supported_actions")
	}

	deletes := endpoints.Delete[0].Requests.Bodies()
	require.Len(t, deletes, 1)
	tfmockserver.AssertRequestBodyField(t, deletes[0], "scope_id", "scope-1")
	tfmockserver.AssertRequestBodyFieldEmptyArray(t, deletes[0], "groups")
	require.Len(t, deletes[0], 2, "delete body carries only scope_id and the const empty set: %v", deletes[0])
}

func TestConstEmptySetDeleteResourceRejectsWrongResponseConst(t *testing.T) {
	t.Parallel()

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint:        "PUT /v0/constemptysetdelete",
				ResponseOverlay: map[string]any{"supported_actions": []any{"delete"}},
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory:          config.StaticDirectory("testdata/TestConstEmptySetDeleteResourceLifecycle"),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
					"replicas":   config.IntegerVariable(2),
				},
				ExpectError: regexp.MustCompile(`(?s)const field.*supported_actions.*does\s+not match`),
			},
		},
	})
}
