package extensions

import (
	"fmt"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Handles parsing of the x-speakeasy-response-filter extension from the
// given OpenAPI extensions map.
func (e *Extensions) HandleResponseFilterExtension(extensions OAExtensions) (*bool, error) {
	if extensions.Len() == 0 {
		return nil, nil
	}

	yamlNode, ok := e.findExtension(extensions, ExtResponseFilter)

	if !ok {
		return nil, nil
	}

	result, err := parseResponseFilter(yamlNode)
	if err != nil {
		return nil, fmt.Errorf("failed to parse extension value: %w", err)
	}

	return result, nil
}

// Parses the given YAML node into *bool.
func parseResponseFilter(node *yaml.Node) (*bool, error) {
	switch node.Kind {
	case yaml.ScalarNode:
		boolValue, err := strconv.ParseBool(node.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse boolean value: %w", err)
		}

		return &boolValue, nil
	default:
		return nil, fmt.Errorf("unsupported YAML node kind: %v", node.Kind)
	}
}
