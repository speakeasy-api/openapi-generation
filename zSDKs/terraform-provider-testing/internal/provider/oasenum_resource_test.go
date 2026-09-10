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
)

func TestOASEnumResourceLifecycle(t *testing.T) {
	t.Parallel()

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/oasenums",
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/oasenum/{id}",
			},
		},
		Update: tfmockserver.Endpoints{
			{
				Endpoint: "PATCH /v0/oasenum/{id}",
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/oasenum/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_oas_enum.my_oasenum"

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
						tfjsonpath.New("string_request"),
						knownvalue.StringExact("one"),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("int32_enum_request_int32_enum_response"),
						knownvalue.Int32Exact(1),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("int64_enum_request_int64_enum_response"),
						knownvalue.Int64Exact(2),
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

func TestOASEnumResourceUpdate(t *testing.T) {
	t.Parallel()

	// Start mock server

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/oasenums",
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/oasenum/{id}",
			},
		},
		Update: tfmockserver.Endpoints{
			{
				Endpoint: "PATCH /v0/oasenum/{id}",
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/oasenum/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_oas_enum.my_oasenum"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Create initial resource
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url":     config.StringVariable(mockServer.URL),
					"string_request": config.StringVariable("one"),
					"int32_enum":     config.IntegerVariable(1),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("id"),
						knownvalue.StringExact(tfmockserver.StoreKey),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("string_request"),
						knownvalue.StringExact("one"),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("int32_enum_request_int32_enum_response"),
						knownvalue.Int32Exact(1),
					),
				},
			},
			// Update resource
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url":     config.StringVariable(mockServer.URL),
					"string_request": config.StringVariable("two"),
					"int32_enum":     config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("id"),
						knownvalue.StringExact(tfmockserver.StoreKey),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("string_request"),
						knownvalue.StringExact("two"),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("int32_enum_request_int32_enum_response"),
						knownvalue.Int32Exact(2),
					),
				},
			},
		},
	})
}

func TestOASEnumResourceWithResponseFields(t *testing.T) {
	t.Parallel()

	// Start mock server with response overlay to simulate response-only fields

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/oasenums",
				ResponseOverlay: map[string]any{
					"string_response":       "one",
					"number_response":       1.1,
					"float32_response":      float32(2.2),
					"float64_response":      2.2,
					"list_string_response":  []string{"one", "two"},
					"set_string_response":   []string{"one", "two"},
					"list_integer_response": []int64{1, 2},
					"set_integer_response":  []int64{1, 2},
				},
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/oasenum/{id}",
				ResponseOverlay: map[string]any{
					"string_response":       "one",
					"number_response":       1.1,
					"float32_response":      float32(2.2),
					"float64_response":      2.2,
					"list_string_response":  []string{"one", "two"},
					"set_string_response":   []string{"one", "two"},
					"list_integer_response": []int64{1, 2},
					"set_integer_response":  []int64{1, 2},
				},
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/oasenum/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_oas_enum.my_oasenum"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Verifies resource create and read with response-only fields
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
						tfjsonpath.New("string_response"),
						knownvalue.StringExact("one"),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("number_response"),
						knownvalue.Float64Exact(1.1),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("float32_response"),
						knownvalue.Float32Exact(2.2),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("float64_response"),
						knownvalue.Float64Exact(2.2),
					),
					// Test list and set response fields
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_string_response"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("one"),
							knownvalue.StringExact("two"),
						}),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_integer_response"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.Int64Exact(1),
							knownvalue.Int64Exact(2),
						}),
					),
				},
			},
		},
	})
}

func TestOASEnumResourceValidation(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Test enum validation - invalid string value
			{
				Config: `resource "testing_oas_enum" "my_oasenum" {
					string_request = "invalid_value"
				}`,
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ExpectError:              regexp.MustCompile("Invalid Attribute Value Match"),
			},
			// Test enum validation - invalid int32 value
			{
				Config: `resource "testing_oas_enum" "my_oasenum" {
					int32_enum_request_int32_enum_response = 99
				}`,
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ExpectError:              regexp.MustCompile("Invalid Attribute Value Match"),
			},
			// Test enum validation - invalid int64 value
			{
				Config: `resource "testing_oas_enum" "my_oasenum" {
					int64_enum_request_int64_enum_response = 99
				}`,
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ExpectError:              regexp.MustCompile("Invalid Attribute Value Match"),
			},
			// Test nested object validation - invalid nested string enum
			{
				Config: `resource "testing_oas_enum" "my_oasenum" {
					object_request = {
						object_string = "invalid_nested_value"
					}
				}`,
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ExpectError:              regexp.MustCompile("Invalid Attribute Value Match"),
			},
			// Test nested object validation - invalid nested integer enum
			{
				Config: `resource "testing_oas_enum" "my_oasenum" {
					object_request = {
						object_integer = 99
					}
				}`,
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ExpectError:              regexp.MustCompile("Invalid Attribute Value Match"),
			},
			// Test list nested validation - invalid nested enum in list
			{
				Config: `resource "testing_oas_enum" "my_oasenum" {
					list_nested_request = [{
						list_nested_string = "invalid_list_value"
					}]
				}`,
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ExpectError:              regexp.MustCompile("Invalid Attribute Value Match"),
			},
			// Test set nested validation - invalid nested enum in set
			{
				Config: `resource "testing_oas_enum" "my_oasenum" {
					set_nested_request = [{
						set_nested_integer = 99
					}]
				}`,
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ExpectError:              regexp.MustCompile("Invalid Attribute Value Match"),
			},
		},
	})
}

func TestOASEnumResourceLongEnum(t *testing.T) {
	t.Parallel()

	// Start mock server

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/oasenums",
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/oasenum/{id}",
			},
		},
		Update: tfmockserver.Endpoints{
			{
				Endpoint: "PATCH /v0/oasenum/{id}",
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/oasenum/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_oas_enum.my_oasenum"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Test long enum values
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
					"long_enum":  config.IntegerVariable(1000000000),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("id"),
						knownvalue.StringExact(tfmockserver.StoreKey),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("int32_long_enum"),
						knownvalue.Int32Exact(1000000000),
					),
				},
			},
		},
	})
}

func TestOASEnumResourceBoundaryValues(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.GetTestProviders(),
		Steps: []resource.TestStep{
			// Test boundary values for long enum - invalid value below range
			{
				Config: `resource "testing_oas_enum" "my_oasenum" {
					int32_long_enum = 999999999
				}`,
				ExpectError: regexp.MustCompile("Invalid Attribute Value Match"),
			},
			// Test boundary values for long enum - invalid value above range
			{
				Config: `resource "testing_oas_enum" "my_oasenum" {
					int32_long_enum = 1000000007
				}`,
				ExpectError: regexp.MustCompile("Invalid Attribute Value Match"),
			},
		},
	})
}
