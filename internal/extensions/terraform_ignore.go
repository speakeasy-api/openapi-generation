package extensions

import (
	"fmt"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Describes the parsed x-speakeasy-terraform-ignore extension configuration.
//
// When the extension is set to true, all values will be ignored.
type TerraformIgnore struct {
	// When enabled, the field will be ignored in Terraform data models.
	// Enabled when the ignore extension is set to true.
	DataModel bool `json:"dataModel" yaml:"dataModel"`

	// When enabled, the field will be ignored in Terraform schema definitions.
	// Enabled when the ignore extension is set to true.
	Schema bool `json:"schema" yaml:"schema"`
}

// Clone creates a deep copy of the TerraformIgnore
func (t *TerraformIgnore) Clone() *TerraformIgnore {
	if t == nil {
		return nil
	}

	cloned := &TerraformIgnore{
		DataModel: t.DataModel,
		Schema:    t.Schema,
	}

	return cloned
}

// Handles parsing of the x-speakeasy-terraform-ignore extension from the
// given OpenAPI extensions map.
func (e *Extensions) HandleTerraformIgnoreExtension(extensions OAExtensions) (*TerraformIgnore, error) {
	if extensions.Len() == 0 {
		return nil, nil
	}

	yamlNode, ok := e.findExtension(extensions, ExtTerraformIgnore)

	if !ok {
		return nil, nil
	}

	result, err := e.parseTerraformIgnore(yamlNode)
	if err != nil {
		return nil, fmt.Errorf("failed to parse extension value: %w", err)
	}

	return result, nil
}

// Parses the given YAML node into an TerraformIgnore structure.
func (e *Extensions) parseTerraformIgnore(node *yaml.Node) (*TerraformIgnore, error) {
	switch node.Kind {
	case yaml.ScalarNode:
		var stringValue string

		if err := node.Decode(&stringValue); err != nil {
			return nil, fmt.Errorf("failed to parse string: %w", err)
		}

		if boolValue, err := strconv.ParseBool(stringValue); err == nil {
			return &TerraformIgnore{
				DataModel: boolValue,
				Schema:    boolValue,
			}, nil
		}

		switch stringValue {
		case "schema":
			return &TerraformIgnore{
				DataModel: false,
				Schema:    true,
			}, nil
		default:
			return nil, fmt.Errorf("invalid string value for x-speakeasy-terraform-ignore extension: %s", stringValue)
		}
	default:
		return nil, fmt.Errorf("unsupported YAML node kind: %v", node.Kind)
	}
}
