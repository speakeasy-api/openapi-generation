package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-testing/internal/provider"
	"github.com/hashicorp/terraform-provider-testing/internal/tfmockserver"
)

func TestUnsoundReadonlyMatchedOpDataSource(t *testing.T) {
	t.Parallel()

	name := "test-resource"
	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/unsound-readonly-matched-op",
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/unsound-readonly-matched-op/{entityId}",
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/unsound-readonly-matched-op/{entityId}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	dataResourceAddress := "data.testing_unsound_readonly_matched_op.test"
	managedResourceAddress := "testing_unsound_readonly_matched_op.test"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
					"name":       config.StringVariable(name),
				},
				// Verify that the data source and managed resource have the same field values
				// Exclude "id" from comparison as it's checked automatically by Terraform
				ConfigStateChecks: provider.AllModelFieldCompareChecks(
					provider.UnsoundReadonlyMatchedOpDataSourceModel{},
					managedResourceAddress,
					dataResourceAddress,
					[]string{"id"},
				),
			},
		},
	})
}
