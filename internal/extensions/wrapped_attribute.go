package extensions

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Handles parsing of the x-speakeasy-wrapped-attribute extension from the
// given OpenAPI extensions map.
func (e *Extensions) HandleWrappedAttributeExtension(extensions OAExtensions) (*string, error) {
	if extensions.Len() == 0 {
		return nil, nil
	}

	yamlNode, ok := e.findExtension(extensions, ExtWrappedAttribute)

	if !ok {
		return nil, nil
	}

	result, err := e.parseWrappedAttribute(yamlNode)

	if err != nil {
		return nil, fmt.Errorf("failed to parse extension value: %w", err)
	}

	return result, nil
}

// Parses the given YAML node.
func (e *Extensions) parseWrappedAttribute(node *yaml.Node) (*string, error) {
	switch node.Kind {
	case yaml.ScalarNode:
		// Handle null values
		if node.Tag == "!!null" || (node.Value == "" && node.Tag == "") {
			return nil, nil
		}
		value := node.Value
		return &value, nil
	default:
		return nil, fmt.Errorf("unsupported YAML node kind: %v", node.Kind)
	}
}
