package provider

import (
	"reflect"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/echoprovider"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

// Returns a mapping of provider type names to provider server implementations,
// suitable for acceptance testing via the
// resource.TestCase.ProtoV6ProtocolFactories field.
func GetTestProviders() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		// echoprovider is used for ephemeral resource testing
		"echo":    echoprovider.NewProviderServer(),
		"testing": providerserver.NewProtocol6WithError(New("test")()),
	}
}

// ExtractFieldNamesFromModel extracts field names from a struct model using reflection.
// Used to extract field names from terraform resource or data source model struct.
func ExtractFieldNamesFromModel(model interface{}, excludeFields []string) []string {
	var fieldNames []string

	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)

		// Get the tfsdk tag
		tfsdkTag := field.Tag.Get("tfsdk")
		if tfsdkTag == "" {
			continue
		}

		// Skip the fields in the excludeFields list
		if slices.Contains(excludeFields, tfsdkTag) {
			continue
		}

		fieldNames = append(fieldNames, tfsdkTag)
	}

	return fieldNames
}

func AllModelFieldCompareChecks(model interface{}, managedResourceAddress, dataResourceAddress string, excludeFields []string) []statecheck.StateCheck {
	fieldNames := ExtractFieldNamesFromModel(model, excludeFields)
	var checks []statecheck.StateCheck
	for _, field := range fieldNames {
		checks = append(checks, statecheck.CompareValuePairs(
			managedResourceAddress,
			tfjsonpath.New(field),
			dataResourceAddress,
			tfjsonpath.New(field),
			compare.ValuesSame(),
		))
	}
	return checks
}

type TestModel struct {
	ID         string            `tfsdk:"id"`
	String     string            `tfsdk:"string"`
	Number     int               `tfsdk:"number"`
	ListString []string          `tfsdk:"list_string"`
	MapString  map[string]string `tfsdk:"map_string"`
	Bool       bool              `tfsdk:"bool"`
}

func TestExtractFieldNamesFromModel(t *testing.T) {
	model := &TestModel{}
	fieldNames := ExtractFieldNamesFromModel(model, []string{"id"})

	// Verify that we get field names
	if len(fieldNames) == 0 {
		t.Fatal("Expected to extract field names from model, but got none")
	}

	// Verify that "id" is not included (should be excluded)
	for _, field := range fieldNames {
		if field == "id" {
			t.Error("Expected 'id' field to be excluded from field names")
		}
	}

	// Verify that some expected fields are present
	expectedFields := []string{"bool", "string", "number", "list_string", "map_string"}
	for _, expected := range expectedFields {
		found := false
		for _, field := range fieldNames {
			if field == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected field '%s' not found in extracted field names", expected)
		}
	}

	t.Logf("Extracted %d field names: %v", len(fieldNames), fieldNames)
}
