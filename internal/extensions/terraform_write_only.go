package extensions

import (
	"fmt"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Handles parsing of the x-speakeasy-terraform-write-only extension from the
// given OpenAPI extensions map.
func (e *Extensions) HandleTerraformWriteOnlyExtension(extensions OAExtensions) (*bool, error) {
	if extensions.Len() == 0 {
		return nil, nil
	}

	yamlNode, ok := e.findExtension(extensions, ExtTerraformWriteOnly)

	if !ok {
		return nil, nil
	}

	result, err := e.parseTerraformWriteOnly(yamlNode)
	if err != nil {
		return nil, fmt.Errorf("failed to parse extension value: %w", err)
	}

	return result, nil
}

// Parses the given YAML node into *bool.
func (e *Extensions) parseTerraformWriteOnly(node *yaml.Node) (*bool, error) {
	switch node.Kind {
	case yaml.ScalarNode:
		value := node.Value

		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse boolean value: %w", err)
		}

		result := &boolValue

		return result, nil
	default:
		return nil, fmt.Errorf("unsupported YAML node kind: %v", node.Kind)
	}
}
