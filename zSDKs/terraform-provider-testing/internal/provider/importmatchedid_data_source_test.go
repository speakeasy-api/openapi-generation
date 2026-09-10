package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-testing/internal/provider"
	"github.com/hashicorp/terraform-provider-testing/internal/tfmockserver"
)

func TestImportMatchedIDDataSource(t *testing.T) {
	t.Parallel()
	t.Skip("TODO: There seems to be an issue with x-speakeasy-match. entityID is required parameter in the openapi spec but in terraform code, its set as computed true. This is the only endpoint where we are using x-speakeasy-match to override the entityID name to ID. So I am guessing x-speakeasy-match is causing the problem. The problem is that when the data source calls the mock server it calls the endpoint /v0/import-matched-id/{parentEntityId}/entity/ . This is the wrong endpoint the correct endpoint is /v0/import-matched-id/{parentEntityId}/entity/{entityId}.")

	entityID := "entity-123"
	parentEntityID := "parent-entity-456"
	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/import-matched-id/{parentEntityId}/entity",
				ResponseOverlay: map[string]any{
					"id": entityID,
				},
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/import-matched-id/{parentEntityId}/entity/{entityId}",
				ResponseOverlay: map[string]any{
					"id": entityID,
				},
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/import-matched-id/{parentEntityId}/entity/{entityId}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	dataResourceAddress := "data.testing_import_matched_id.test"
	managedResourceAddress := "testing_import_matched_id.test"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Verifies resource create and read.
			{
				// Specifies the directory to look for the terraform configuration files.
				// By default it will look in the folder `testdata/<name of the test>`
				ConfigDirectory: config.TestNameDirectory(),
				// Setup provider for this test
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url":       config.StringVariable(mockServer.URL),
					"parent_entity_id": config.StringVariable(parentEntityID),
				},
				// Verify that the data source and managed resource have the same field values
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						managedResourceAddress,
						tfjsonpath.New("id"),
						dataResourceAddress,
						tfjsonpath.New("id"),
						compare.ValuesSame(),
					),
					statecheck.CompareValuePairs(
						managedResourceAddress,
						tfjsonpath.New("parent_entity_id"),
						dataResourceAddress,
						tfjsonpath.New("parent_entity_id"),
						compare.ValuesSame(),
					),
				},
			},
		},
	})
}
