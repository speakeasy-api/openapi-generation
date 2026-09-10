package provider_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-testing/internal/provider"
	"github.com/hashicorp/terraform-provider-testing/internal/sdk"
	"github.com/hashicorp/terraform-provider-testing/internal/sdk/models/operations"
	"github.com/hashicorp/terraform-provider-testing/internal/tfmockserver"
)

func TestXEntityMissingCodesResourceDelete(t *testing.T) {
	t.Parallel()

	// TODO: Update to t.Context() instead of ctx when Go module uses go 1.24
	ctx := context.Background()
	id := "id-value"
	responseOverlay := map[string]any{
		"id": id,
	}

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint:        "POST /v0/xentitymissingcodes",
				ResponseOverlay: responseOverlay,
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint:        "GET /v0/xentitymissingcodes/{id}",
				ResponseOverlay: responseOverlay,
			},
		},
		Update: tfmockserver.Endpoints{
			{
				Endpoint:        "PUT /v0/xentitymissingcodes/{id}",
				ResponseOverlay: responseOverlay,
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint:           "DELETE /v0/xentitymissingcodes/{id}",
				NotFoundStatusCode: http.StatusGone,
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_x_entity_missing_codes.test"

	resource.Test(t, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{
			Plan: resource.PlanOptions{
				// Required to destroy without first detecting resource missing.
				NoRefresh: true,
			},
		},
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigDirectory:          config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("id"),
						knownvalue.StringExact(id),
					),
				},
			},
			{
				PreConfig: func() {
					client := sdk.New(sdk.WithServerURL(mockServer.URL))
					request := operations.DeleteXEntityMissingCodesRequest{
						ID: id,
					}
					_, err := client.DeleteXEntityMissingCodes(ctx, request)

					if err != nil {
						t.Fatalf("failed to delete resource via test SDK client: %v", err)
					}
				},
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigDirectory:          config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
				},
				Destroy: true,
			},
		},
	})
}

func TestXEntityMissingCodesResourceRead(t *testing.T) {
	t.Parallel()

	// TODO: Update to t.Context() instead of ctx when Go module uses go 1.24
	ctx := context.Background()
	id := "id-value"
	responseOverlay := map[string]any{
		"id": id,
	}

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint:        "POST /v0/xentitymissingcodes",
				ResponseOverlay: responseOverlay,
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint:           "GET /v0/xentitymissingcodes/{id}",
				NotFoundStatusCode: http.StatusGone,
				ResponseOverlay:    responseOverlay,
			},
		},
		Update: tfmockserver.Endpoints{
			{
				Endpoint:        "PUT /v0/xentitymissingcodes/{id}",
				ResponseOverlay: responseOverlay,
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/xentitymissingcodes/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_x_entity_missing_codes.test"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigDirectory:          config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("id"),
						knownvalue.StringExact(id),
					),
				},
			},
			{
				PreConfig: func() {
					client := sdk.New(sdk.WithServerURL(mockServer.URL))
					request := operations.DeleteXEntityMissingCodesRequest{
						ID: id,
					}
					_, err := client.DeleteXEntityMissingCodes(ctx, request)

					if err != nil {
						t.Fatalf("failed to delete resource via test SDK client: %v", err)
					}
				},
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigDirectory:          config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						// No error means planning was successful.
						plancheck.ExpectResourceAction(resourceAddress, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}
