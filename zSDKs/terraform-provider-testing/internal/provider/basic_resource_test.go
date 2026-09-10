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

func TestBasicResourceLifecycle(t *testing.T) {
	t.Parallel()

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/basics",
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/basic/{id}",
			},
		},
		Update: tfmockserver.Endpoints{
			{
				Endpoint: "PATCH /v0/basic/{id}",
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/basic/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_basic.my_basic"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Verifies resource create and read.
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
						knownvalue.StringExact(tfmockserver.StoreKey),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-basic-resource"),
					),
				},
			},
			// Import resource
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
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

func TestBasicResourceUpdate(t *testing.T) {
	t.Parallel()

	// Start mock server

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/basics",
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/basic/{id}",
			},
		},
		Update: tfmockserver.Endpoints{
			{
				Endpoint: "PATCH /v0/basic/{id}",
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/basic/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_basic.my_basic"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Create initial resource
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
					"name":       config.StringVariable("initial-name"),
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
						knownvalue.StringExact("initial-name"),
					),
				},
			},
			// Update resource
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
					"name":       config.StringVariable("updated-name"),
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
						knownvalue.StringExact("updated-name"),
					),
				},
			},
		},
	})
}

func TestBasicResourceDeleteNotFound(t *testing.T) {
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
				Endpoint:        "POST /v0/basics",
				ResponseOverlay: responseOverlay,
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint:        "GET /v0/basic/{id}",
				ResponseOverlay: responseOverlay,
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint:           "DELETE /v0/basic/{id}",
				NotFoundStatusCode: http.StatusNotFound,
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_basic.test"

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
					request := operations.DeleteBasicRequest{
						ID: id,
					}
					_, err := client.DeleteBasic(ctx, request)

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

func TestBasicResourceReadNotFound(t *testing.T) {
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
				Endpoint:        "POST /v0/basics",
				ResponseOverlay: responseOverlay,
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint:           "GET /v0/basic/{id}",
				NotFoundStatusCode: http.StatusNotFound,
				ResponseOverlay:    responseOverlay,
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/basic/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_basic.test"

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
					request := operations.DeleteBasicRequest{
						ID: id,
					}
					_, err := client.DeleteBasic(ctx, request)

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
