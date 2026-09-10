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

func TestMixedRequestAndResponseTypeResourceLifecycle(t *testing.T) {
	t.Parallel()

	// Declare repeated values as variables to avoid duplication
	testID := "test-id-123"
	integerWithUpdateRequestInt32 := int64(10)
	floatWithUpdateRequestFloat32 := 5.5

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/mixedRequestAndResponseType",
				ResponseOverlay: map[string]any{
					"id":                                testID,
					"integer_with_update_request_int32": integerWithUpdateRequestInt32,
					"float_with_update_request_float32": floatWithUpdateRequestFloat32,
				},
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/mixedRequestAndResponseType/{id}",
				ResponseOverlay: map[string]any{
					"id":                                testID,
					"integer_with_update_request_int32": integerWithUpdateRequestInt32,
					"float_with_update_request_float32": floatWithUpdateRequestFloat32,
				},
			},
		},
		Update: tfmockserver.Endpoints{
			{
				Endpoint: "PATCH /v0/mixedRequestAndResponseType/{id}",
				ResponseOverlay: map[string]any{
					"id":                                testID,
					"integer_with_update_request_int32": integerWithUpdateRequestInt32,
					"float_with_update_request_float32": floatWithUpdateRequestFloat32,
				},
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/mixedRequestAndResponseType/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_mixed_request_and_response_type.my_resource"

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
						knownvalue.StringExact(testID),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("integer_with_update_request_int32"),
						knownvalue.Int64Exact(integerWithUpdateRequestInt32),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("float_with_update_request_float32"),
						knownvalue.Float64Exact(floatWithUpdateRequestFloat32),
					),
				},
			},
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ResourceName:             resourceAddress,
				ImportState:              true,
				ImportStateId:            testID,
				ImportStateVerify:        true,
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
				},
			},
		},
	})
}

func TestMixedRequestAndResponseTypeResourceUpdate(t *testing.T) {
	t.Parallel()

	// Declare repeated values as variables to avoid duplication
	testID := "test-id-123"

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/mixedRequestAndResponseType",
				ResponseOverlay: map[string]any{
					"id": testID,
				},
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/mixedRequestAndResponseType/{id}",
				ResponseOverlay: map[string]any{
					"id": testID,
				},
			},
		},
		Update: tfmockserver.Endpoints{
			{
				Endpoint: "PATCH /v0/mixedRequestAndResponseType/{id}",
				ResponseOverlay: map[string]any{
					"id": testID,
				},
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/mixedRequestAndResponseType/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_mixed_request_and_response_type.my_resource"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url":                        config.StringVariable(mockServer.URL),
					"integer_with_update_request_int32": config.IntegerVariable(10),
					"float_with_update_request_float32": config.FloatVariable(5.5),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("id"),
						knownvalue.StringExact(testID),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("integer_with_update_request_int32"),
						knownvalue.Int64Exact(10),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("float_with_update_request_float32"),
						knownvalue.Float64Exact(5.5),
					),
				},
			},
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url":                        config.StringVariable(mockServer.URL),
					"integer_with_update_request_int32": config.IntegerVariable(20),
					"float_with_update_request_float32": config.FloatVariable(10.5),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("id"),
						knownvalue.StringExact(testID),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("integer_with_update_request_int32"),
						knownvalue.Int64Exact(20),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("float_with_update_request_float32"),
						knownvalue.Float64Exact(10.5),
					),
				},
			},
		},
	})
}
