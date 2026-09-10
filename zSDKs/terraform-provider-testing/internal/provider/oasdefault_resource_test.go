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

func TestOASDefaultResourceLifecycle(t *testing.T) {
	t.Parallel()
	//t.Skip("TODO: fix the provider to produce the correct code, once provider is fixed this test should pass. Issue: When terraform plans the resource again (after creating it) it expects the plan to produce 0 changes instead some changes show up as a result the test fails.")

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/oasdefaults",
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/oasdefault/{id}",
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/oasdefault/{id}",
			},
		},
	}
	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_oas_default.my_oasdefault"

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
					"server_url": config.StringVariable(mockServer.URL),
				},
				// TODO: Response-only defaults are not currently implemented
				// in Terraform generation. ExpectNonEmptyPlan should never be
				// necessary (its always a resource implementation bug), however
				// the existing Go SDK will still always set default values, so
				// the issue is unavoidable until the schema generation is
				// fixed. We do still want this testing to run and verify
				// existing implementation details though.
				ExpectNonEmptyPlan: true,
				// Check computed values.
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("bool_request_and_response"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("bool_request"),
						knownvalue.Bool(false),
					),
					// TODO: Response-only defaults are not currently implemented.
					// statecheck.ExpectKnownValue(
					// 	resourceAddress,
					// 	tfjsonpath.New("bool_response"),
					// 	knownvalue.Bool(false),
					// ),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("float32_request_and_response"),
						knownvalue.Float32Exact(12.12),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("float32_request"),
						knownvalue.Float32Exact(1.1),
					),
					// TODO: Response-only defaults are not currently implemented.
					// statecheck.ExpectKnownValue(
					// 	resourceAddress,
					// 	tfjsonpath.New("float32_response"),
					// 	knownvalue.Float32Exact(2.2),
					// ),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("float64_request_and_response"),
						knownvalue.Float64Exact(12.12),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("float64_request"),
						knownvalue.Float64Exact(1.1),
					),
					// TODO: Response-only defaults are not currently implemented.
					// statecheck.ExpectKnownValue(
					// 	resourceAddress,
					// 	tfjsonpath.New("float64_response"),
					// 	knownvalue.Float64Exact(2.2),
					// ),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("id"),
						knownvalue.StringExact("default"),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("int32_request_and_response"),
						knownvalue.Int32Exact(12),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("int32_request"),
						knownvalue.Int32Exact(1),
					),
					// TODO: Response-only defaults are not currently implemented.
					// statecheck.ExpectKnownValue(
					// 	resourceAddress,
					// 	tfjsonpath.New("int32_response"),
					// 	knownvalue.Int32Exact(2),
					// ),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("int64_request_and_response"),
						knownvalue.Int64Exact(12),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("int64_request"),
						knownvalue.Int64Exact(1),
					),
					// TODO: Response-only defaults are not currently implemented.
					// statecheck.ExpectKnownValue(
					// 	resourceAddress,
					// 	tfjsonpath.New("int64_response"),
					// 	knownvalue.Int64Exact(2),
					// ),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_bool_request_and_response_empty"),
						knownvalue.ListSizeExact(0),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_bool_request_and_response_value"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.Bool(true),
							knownvalue.Bool(false),
						}),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_bool_request_empty"),
						knownvalue.ListSizeExact(0),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_bool_request_value"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.Bool(true),
							knownvalue.Bool(false),
						}),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_bool_response_empty"),
						knownvalue.ListSizeExact(0),
					),
					// TODO: Response-only defaults are not currently implemented.
					// statecheck.ExpectKnownValue(
					// 	resourceAddress,
					// 	tfjsonpath.New("list_bool_response_value"),
					// 	knownvalue.ListExact([]knownvalue.Check{
					// 		knownvalue.Bool(true),
					// 		knownvalue.Bool(false),
					// 	}),
					// ),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_float32_request_and_response_empty"),
						knownvalue.ListSizeExact(0),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_float32_request_and_response_value"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.Float32Exact(1.1),
							knownvalue.Float32Exact(2.2),
						}),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_float32_request_empty"),
						knownvalue.ListSizeExact(0),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_float32_request_value"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.Float32Exact(3.3),
							knownvalue.Float32Exact(4.4),
						}),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_float32_response_empty"),
						knownvalue.ListSizeExact(0),
					),
					// TODO: Response-only defaults are not currently implemented.
					// statecheck.ExpectKnownValue(
					// 	resourceAddress,
					// 	tfjsonpath.New("list_float32_response_value"),
					// 	knownvalue.ListExact([]knownvalue.Check{
					// 		knownvalue.Float32Exact(5.5),
					// 		knownvalue.Float32Exact(6.6),
					// 	}),
					// ),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_float64_request_and_response_empty"),
						knownvalue.ListSizeExact(0),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_float64_request_and_response_value"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.Float64Exact(1.1),
							knownvalue.Float64Exact(2.2),
						}),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_float64_request_empty"),
						knownvalue.ListSizeExact(0),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_float64_request_value"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.Float64Exact(3.3),
							knownvalue.Float64Exact(4.4),
						}),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_float64_response_empty"),
						knownvalue.ListSizeExact(0),
					),
					// TODO: Response-only defaults are not currently implemented.
					// statecheck.ExpectKnownValue(
					// 	resourceAddress,
					// 	tfjsonpath.New("list_float64_response_value"),
					// 	knownvalue.ListExact([]knownvalue.Check{
					// 		knownvalue.Float64Exact(5.5),
					// 		knownvalue.Float64Exact(6.6),
					// 	}),
					// ),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_int32_request_and_response_empty"),
						knownvalue.ListSizeExact(0),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_int32_request_and_response_value"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.Int32Exact(1),
							knownvalue.Int32Exact(2),
						}),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_int32_request_empty"),
						knownvalue.ListSizeExact(0),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_int32_request_value"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.Int32Exact(3),
							knownvalue.Int32Exact(4),
						}),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_int32_response_empty"),
						knownvalue.ListSizeExact(0),
					),
					// TODO: Response-only defaults are not currently implemented.
					// statecheck.ExpectKnownValue(
					// 	resourceAddress,
					// 	tfjsonpath.New("list_int32_response_value"),
					// 	knownvalue.ListExact([]knownvalue.Check{
					// 		knownvalue.Int32Exact(5),
					// 		knownvalue.Int32Exact(6),
					// 	}),
					// ),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_int64_request_and_response_empty"),
						knownvalue.ListSizeExact(0),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_int64_request_and_response_value"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.Int64Exact(1),
							knownvalue.Int64Exact(2),
						}),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_int64_request_empty"),
						knownvalue.ListSizeExact(0),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_int64_request_value"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.Int64Exact(3),
							knownvalue.Int64Exact(4),
						}),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_int64_response_empty"),
						knownvalue.ListSizeExact(0),
					),
					// TODO: Response-only defaults are not currently implemented.
					// statecheck.ExpectKnownValue(
					// 	resourceAddress,
					// 	tfjsonpath.New("list_int64_response_value"),
					// 	knownvalue.ListExact([]knownvalue.Check{
					// 		knownvalue.Int64Exact(5),
					// 		knownvalue.Int64Exact(6),
					// 	}),
					// ),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_integer_request_and_response_empty"),
						knownvalue.ListSizeExact(0),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_integer_request_and_response_value"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.Int64Exact(1),
							knownvalue.Int64Exact(2),
						}),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_integer_request_empty"),
						knownvalue.ListSizeExact(0),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_integer_request_value"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.Int64Exact(3),
							knownvalue.Int64Exact(4),
						}),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_integer_response_empty"),
						knownvalue.ListSizeExact(0),
					),
					// TODO: Response-only defaults are not currently implemented.
					// statecheck.ExpectKnownValue(
					// 	resourceAddress,
					// 	tfjsonpath.New("list_integer_response_value"),
					// 	knownvalue.ListExact([]knownvalue.Check{
					// 		knownvalue.Int64Exact(5),
					// 		knownvalue.Int64Exact(6),
					// 	}),
					// ),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_string_request_and_response_empty"),
						knownvalue.ListSizeExact(0),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_string_request_and_response_value"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("one"),
							knownvalue.StringExact("two"),
						}),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_string_request_empty"),
						knownvalue.ListSizeExact(0),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_string_request_value"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("three"),
							knownvalue.StringExact("four"),
						}),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("list_string_response_empty"),
						knownvalue.ListSizeExact(0),
					),
					// TODO: Response-only defaults are not currently implemented.
					// statecheck.ExpectKnownValue(
					// 	resourceAddress,
					// 	tfjsonpath.New("list_string_response_value"),
					// 	knownvalue.ListExact([]knownvalue.Check{
					// 		knownvalue.StringExact("five"),
					// 		knownvalue.StringExact("six"),
					// 	}),
					// ),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("number_request"),
						knownvalue.Float64Exact(1.1),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("number_request_and_response"),
						knownvalue.Float64Exact(12.12),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("object_request_and_response_null"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("object_request_null"),
						knownvalue.Null(),
					),
					// TODO: Response-only defaults are not currently implemented.
					// statecheck.ExpectKnownValue(
					// 	resourceAddress,
					// 	tfjsonpath.New("object_response_null"),
					// 	knownvalue.Null(),
					// ),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("set_bool_request_and_response_empty"),
						knownvalue.SetSizeExact(0),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("set_bool_request_and_response_value"),
						knownvalue.SetExact([]knownvalue.Check{
							knownvalue.Bool(true),
							knownvalue.Bool(false),
						}),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("set_bool_request_empty"),
						knownvalue.SetSizeExact(0),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("set_bool_request_value"),
						knownvalue.SetExact([]knownvalue.Check{
							knownvalue.Bool(true),
							knownvalue.Bool(false),
						}),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("set_bool_response_empty"),
						knownvalue.SetSizeExact(0),
					),
					// TODO: Response-only defaults are not currently implemented.
					// statecheck.ExpectKnownValue(
					// 	resourceAddress,
					// 	tfjsonpath.New("set_bool_response_value"),
					// 	knownvalue.SetExact([]knownvalue.Check{
					// 		knownvalue.Bool(true),
					// 		knownvalue.Bool(false),
					// 	}),
					// ),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("set_float32_request_and_response_empty"),
						knownvalue.SetSizeExact(0),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("set_float32_request_and_response_value"),
						knownvalue.SetExact([]knownvalue.Check{
							knownvalue.Float32Exact(1.1),
							knownvalue.Float32Exact(2.2),
						}),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("set_float32_request_empty"),
						knownvalue.SetSizeExact(0),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("set_float32_request_value"),
						knownvalue.SetExact([]knownvalue.Check{
							knownvalue.Float32Exact(3.3),
							knownvalue.Float32Exact(4.4),
						}),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("set_float32_response_empty"),
						knownvalue.SetSizeExact(0),
					),
					// TODO: Response-only defaults are not currently implemented.
					// statecheck.ExpectKnownValue(
					// 	resourceAddress,
					// 	tfjsonpath.New("set_float32_response_value"),
					// 	knownvalue.SetExact([]knownvalue.Check{
					// 		knownvalue.Float32Exact(5.5),
					// 		knownvalue.Float32Exact(6.6),
					// 	}),
					// ),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("set_float64_request_and_response_empty"),
						knownvalue.SetSizeExact(0),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("set_float64_request_and_response_value"),
						knownvalue.SetExact([]knownvalue.Check{
							knownvalue.Float64Exact(1.1),
							knownvalue.Float64Exact(2.2),
						}),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("set_float64_request_empty"),
						knownvalue.SetSizeExact(0),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("set_float64_request_value"),
						knownvalue.SetExact([]knownvalue.Check{
							knownvalue.Float64Exact(3.3),
							knownvalue.Float64Exact(4.4),
						}),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("set_float64_response_empty"),
						knownvalue.SetSizeExact(0),
					),
					// TODO: Response-only defaults are not currently implemented.
					// statecheck.ExpectKnownValue(
					// 	resourceAddress,
					// 	tfjsonpath.New("set_float64_response_value"),
					// 	knownvalue.SetExact([]knownvalue.Check{
					// 		knownvalue.Float64Exact(5.5),
					// 		knownvalue.Float64Exact(6.6),
					// 	}),
					// ),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("set_int32_request_and_response_empty"),
						knownvalue.SetSizeExact(0),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("set_int32_request_and_response_value"),
						knownvalue.SetExact([]knownvalue.Check{
							knownvalue.Int32Exact(1),
							knownvalue.Int32Exact(2),
						}),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("set_int32_request_empty"),
						knownvalue.SetSizeExact(0),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("set_int32_request_value"),
						knownvalue.SetExact([]knownvalue.Check{
							knownvalue.Int32Exact(3),
							knownvalue.Int32Exact(4),
						}),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("set_int32_response_empty"),
						knownvalue.SetSizeExact(0),
					),
					// TODO: Response-only defaults are not currently implemented.
					// statecheck.ExpectKnownValue(
					// 	resourceAddress,
					// 	tfjsonpath.New("set_int32_response_value"),
					// 	knownvalue.SetExact([]knownvalue.Check{
					// 		knownvalue.Int32Exact(5),
					// 		knownvalue.Int32Exact(6),
					// 	}),
					// ),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("set_int64_request_and_response_empty"),
						knownvalue.SetSizeExact(0),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("set_int64_request_and_response_value"),
						knownvalue.SetExact([]knownvalue.Check{
							knownvalue.Int64Exact(1),
							knownvalue.Int64Exact(2),
						}),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("set_int64_request_empty"),
						knownvalue.SetSizeExact(0),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("set_int64_request_value"),
						knownvalue.SetExact([]knownvalue.Check{
							knownvalue.Int64Exact(3),
							knownvalue.Int64Exact(4),
						}),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("set_int64_response_empty"),
						knownvalue.SetSizeExact(0),
					),
					// TODO: Response-only defaults are not currently implemented.
					// statecheck.ExpectKnownValue(
					// 	resourceAddress,
					// 	tfjsonpath.New("set_int64_response_value"),
					// 	knownvalue.SetExact([]knownvalue.Check{
					// 		knownvalue.Int64Exact(5),
					// 		knownvalue.Int64Exact(6),
					// 	}),
					// ),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("set_integer_request_and_response_empty"),
						knownvalue.SetSizeExact(0),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("set_integer_request_and_response_value"),
						knownvalue.SetExact([]knownvalue.Check{
							knownvalue.Int64Exact(1),
							knownvalue.Int64Exact(2),
						}),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("set_integer_request_empty"),
						knownvalue.SetSizeExact(0),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("set_integer_request_value"),
						knownvalue.SetExact([]knownvalue.Check{
							knownvalue.Int64Exact(3),
							knownvalue.Int64Exact(4),
						}),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("set_integer_response_empty"),
						knownvalue.SetSizeExact(0),
					),
					// TODO: Response-only defaults are not currently implemented.
					// statecheck.ExpectKnownValue(
					// 	resourceAddress,
					// 	tfjsonpath.New("set_integer_response_value"),
					// 	knownvalue.SetExact([]knownvalue.Check{
					// 		knownvalue.Int64Exact(5),
					// 		knownvalue.Int64Exact(6),
					// 	}),
					// ),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("set_string_request_and_response_empty"),
						knownvalue.SetSizeExact(0),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("set_string_request_and_response_value"),
						knownvalue.SetExact([]knownvalue.Check{
							knownvalue.StringExact("one"),
							knownvalue.StringExact("two"),
						}),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("set_string_request_empty"),
						knownvalue.SetSizeExact(0),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("set_string_request_value"),
						knownvalue.SetExact([]knownvalue.Check{
							knownvalue.StringExact("three"),
							knownvalue.StringExact("four"),
						}),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("set_string_response_empty"),
						knownvalue.SetSizeExact(0),
					),
					// TODO: Response-only defaults are not currently implemented.
					// statecheck.ExpectKnownValue(
					// 	resourceAddress,
					// 	tfjsonpath.New("set_string_response_value"),
					// 	knownvalue.SetExact([]knownvalue.Check{
					// 		knownvalue.StringExact("five"),
					// 		knownvalue.StringExact("six"),
					// 	}),
					// ),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("string_request_and_response"),
						knownvalue.StringExact("three"),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("string_request"),
						knownvalue.StringExact("one"),
					),
					// TODO: Response-only defaults are not currently implemented.
					// statecheck.ExpectKnownValue(
					// 	resourceAddress,
					// 	tfjsonpath.New("string_response"),
					// 	knownvalue.StringExact("two"),
					// ),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("string_sanitization"),
						knownvalue.StringExact("Example value with {{ templating }} character sequences"),
					),
				},
			},
			// Testing framework implicitly verifies resource delete.
		},
	})
}

func TestOASDefaultResourceWithCustomValues(t *testing.T) {
	t.Parallel()
	t.Skip("TODO: fix the provider to produce the correct code, once provider is fixed this test should pass. Issue: When terraform plans the resource again (after creating it) it expects the plan to produce 0 changes instead some changes show up as a result the test fails.")

	// Start mock server
	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/oasdefaults",
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/oasdefault/{id}",
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/oasdefault/{id}",
			},
		},
	}
	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_oas_default.my_oasdefault"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Verifies resource create and read with custom values.
			{
				// Specifies the directory to look for the terraform configuration files.
				ConfigDirectory: config.TestNameDirectory(),
				// Setup provider for this test
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
				},
				// Check computed values.
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("id"),
						knownvalue.StringRegexp(regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					),
					// Test that custom values override defaults
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("string_request"),
						knownvalue.StringExact("custom_string"),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("int32_request"),
						knownvalue.Float64Exact(42.0),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("float32_request"),
						knownvalue.Float64Exact(3.14),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("number_request"),
						knownvalue.Float64Exact(2.718),
					),
					// Test that defaults still apply to fields not specified
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("string_request_and_response"),
						knownvalue.StringExact("three"),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("int32_request_and_response"),
						knownvalue.Float64Exact(12.0),
					),
				},
			},
			// Testing framework implicitly verifies resource delete.
		},
	})
}

func TestOASDefaultResourceWithComputedResponse(t *testing.T) {
	t.Parallel()
	t.Skip("TODO: fix the provider to produce the correct code, once provider is fixed this test should pass. Issue: When terraform plans the resource again (after creating it) it expects the plan to produce 0 changes instead some changes show up as a result the test fails.")

	string_response := "computed-value"
	// Start mock server with response overlays to simulate dynamic response-only fields
	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/oasdefaults",
				ResponseOverlay: map[string]any{
					"string_response": string_response,
				},
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/oasdefault/{id}",
				ResponseOverlay: map[string]any{
					"string_response": string_response,
				},
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/oasdefault/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_oas_default.my_oasdefault"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Create resource and verify response-only fields
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				ConfigVariables: config.Variables{
					"server_url": config.StringVariable(mockServer.URL),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("string_response"),
						knownvalue.StringExact(string_response),
					),
				},
			},
		},
	})
}

func TestOASDefaultResourceWhenResourceValueDriftsOutsideOfTerraform(t *testing.T) {
	t.Parallel()

	t.Skip("TODO: fix the provider to produce the correct code, once provider is fixed this test should pass. Issue: When terraform plans the resource again (after creating it) it expects the plan to produce 0 changes instead some changes show up as a result the test fails.")

	endpoints := tfmockserver.ResourceEndpoints{
		Create: tfmockserver.Endpoints{
			{
				Endpoint: "POST /v0/oasdefaults",
				ResponseOverlay: map[string]any{
					"string_response": "example/a",
				},
			},
		},
		Get: tfmockserver.Endpoints{
			{
				Endpoint: "GET /v0/oasdefault/{id}",
				ResponseOverlay: map[string]any{
					"string_response": "example/b",
				},
			},
		},
		Delete: tfmockserver.Endpoints{
			{
				Endpoint: "DELETE /v0/oasdefault/{id}",
			},
		},
	}

	mockServer := tfmockserver.StartServer(endpoints, t)
	defer mockServer.Close()

	resourceAddress := "testing_oas_default.my_oasdefault"

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
						tfjsonpath.New("string_response"),
						knownvalue.StringExact("example/b"),
					),
				},
			},
		},
	})
}
