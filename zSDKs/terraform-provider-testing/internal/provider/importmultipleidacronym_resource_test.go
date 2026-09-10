package provider_test

import (
	"encoding/json"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-testing/internal/provider"
	"github.com/hashicorp/terraform-provider-testing/internal/tfmockserver"
)

func TestImportMultipleIDAcronymResourceLifecycle(t *testing.T) {
	t.Parallel()

	apiId := "api-123"
	portalId := "portal-456"
	requestBodyProperty := "test-request-body"
	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/import-multiple-id-acronym/{apiId}/{portalId}",
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/import-multiple-id-acronym/{apiId}/{portalId}",
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/import-multiple-id-acronym/{apiId}/{portalId}",
			},
		},
	}

	// Start mock server with custom endpoints
	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_import_multiple_id_acronym.my_importmultipleidacronym"

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
					"server_url":            config.StringVariable(mockServer.URL),
					"api_id":                config.StringVariable(apiId),
					"portal_id":             config.StringVariable(portalId),
					"request_body_property": config.StringVariable(requestBodyProperty),
				},
				// Check computed values.
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("api_id"),
						knownvalue.StringExact(apiId),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("portal_id"),
						knownvalue.StringExact(portalId),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("request_body_property"),
						knownvalue.StringExact(requestBodyProperty),
					),
				},
			},
			// Test import functionality with JSON format
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url":            config.StringVariable(mockServer.URL),
					"api_id":                config.StringVariable(apiId),
					"portal_id":             config.StringVariable(portalId),
					"request_body_property": config.StringVariable(requestBodyProperty),
				},
				ResourceName: resourceAddress,
				ImportState:  true,
				// ImportStateId: `{"api_id": "api-123", "portal_id": "portal-456"}`,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources[resourceAddress]
					jsonBytes, err := json.Marshal(map[string]string{
						"api_id":    rs.Primary.Attributes["api_id"],
						"portal_id": rs.Primary.Attributes["portal_id"],
					})
					return string(jsonBytes), err
				},
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "api_id",
				// Note: request_body_property is present in request-only and hence not importable
				ImportStateVerifyIgnore: []string{"request_body_property"},
			},
			// Testing framework implicitly verifies resource delete.
		},
	})
}

func TestImportMultipleIDAcronymResourceImportValidation(t *testing.T) {
	t.Parallel()
	apiId := "api-123"
	portalId := "portal-456"
	// Define endpoints for importmultipleidacronym resource
	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/import-multiple-id-acronym/{apiId}/{portalId}",
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/import-multiple-id-acronym/{apiId}/{portalId}",
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/import-multiple-id-acronym/{apiId}/{portalId}",
			},
		},
	}

	// Start mock server
	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_import_multiple_id_acronym.my_importmultipleidacronym"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Create resource first
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("api_id"),
						knownvalue.StringExact(apiId),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("portal_id"),
						knownvalue.StringExact(portalId),
					),
				},
			},
			// Test import with invalid JSON format
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
					"api_id":     config.StringVariable(apiId),
					"portal_id":  config.StringVariable(portalId),
				},
				ResourceName:  resourceAddress,
				ImportState:   true,
				ImportStateId: "not-json-format",
				ExpectError:   regexp.MustCompile("The import ID is not valid"),
			},
			// Test import with missing api_id field
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
					"api_id":     config.StringVariable(apiId),
					"portal_id":  config.StringVariable(portalId),
				},
				ResourceName: resourceAddress,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources[resourceAddress]
					jsonBytes, err := json.Marshal(map[string]string{
						"portal_id": rs.Primary.Attributes["portal_id"],
					})
					return string(jsonBytes), err
				},
				ExpectError: regexp.MustCompile("The field api_id is required"),
			},
			// Test import with missing portal_id field
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
					"api_id":     config.StringVariable(apiId),
					"portal_id":  config.StringVariable(portalId),
				},
				ResourceName: resourceAddress,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources[resourceAddress]
					jsonBytes, err := json.Marshal(map[string]string{
						"api_id": rs.Primary.Attributes["api_id"],
					})
					return string(jsonBytes), err
				},
				ExpectError: regexp.MustCompile("The field portal_id is required"),
			},
		},
	})
}
