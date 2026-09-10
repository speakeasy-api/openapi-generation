package extensions

import (
	"errors"
	"fmt"
	"slices"

	"gopkg.in/yaml.v3"
)

// Describes the parsed x-speakeasy-terraform-custom-default extension
// configuration.
type TerraformCustomDefault struct {
	// Go package imports required for the custom default.
	Imports []string `json:"imports" yaml:"imports"`

	// Code rendered into the schema to instantiate the custom default
	// implementation.
	SchemaDefinition string `json:"schemaDefinition" yaml:"schemaDefinition"`
}

// Clone creates a deep copy of the TerraformCustomDefault
func (t *TerraformCustomDefault) Clone() *TerraformCustomDefault {
	if t == nil {
		return nil
	}

	cloned := &TerraformCustomDefault{
		Imports:          slices.Clone(t.Imports),
		SchemaDefinition: t.SchemaDefinition,
	}

	return cloned
}

// Handles parsing of the x-speakeasy-terraform-custom-default extension from the
// given OpenAPI extensions map.
func (e *Extensions) HandleTerraformCustomDefaultExtension(extensions OAExtensions) (*TerraformCustomDefault, error) {
	if extensions.Len() == 0 {
		return nil, nil
	}

	yamlNode, ok := e.findExtension(extensions, ExtTerraformCustomDefault)

	if !ok {
		return nil, nil
	}

	result, err := e.parseTerraformCustomDefault(yamlNode)
	if err != nil {
		return nil, fmt.Errorf("failed to parse extension value: %w", err)
	}

	return result, nil
}

// Parses the given YAML node into an TerraformCustomDefault structure.
func (e *Extensions) parseTerraformCustomDefault(node *yaml.Node) (*TerraformCustomDefault, error) {
	switch node.Kind {
	case yaml.MappingNode:
		var result TerraformCustomDefault

		if err := node.Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode mapping node: %w", err)
		}

		if result.SchemaDefinition == "" {
			return nil, errors.New("schemaDefinition must be a non-empty string")
		}

		return &result, nil
	default:
		return nil, fmt.Errorf("unsupported YAML node kind: %v", node.Kind)
	}
}
