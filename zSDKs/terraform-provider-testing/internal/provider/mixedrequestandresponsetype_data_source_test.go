package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-testing/internal/provider"
	"github.com/hashicorp/terraform-provider-testing/internal/tfmockserver"
)

func TestMixedRequestAndResponseTypeDataSource(t *testing.T) {
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
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/mixedRequestAndResponseType/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	dataResourceAddress := "data.testing_mixed_request_and_response_type.test"
	managedResourceAddress := "testing_mixed_request_and_response_type.test"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
				},
				ConfigStateChecks: provider.AllModelFieldCompareChecks(
					provider.MixedRequestAndResponseTypeDataSourceModel{},
					managedResourceAddress,
					dataResourceAddress,
					[]string{"id"}, // Exclude ID as Terraform validates it automatically
				),
			},
		},
	})
}
