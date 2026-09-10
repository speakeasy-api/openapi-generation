package extensions

import (
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Handles parsing of the x-speakeasy-ignore extension from the given OpenAPI
// extensions map.
func (e *Extensions) HandleIgnoreExtension(extensions OAExtensions) (*bool, error) {
	if extensions.Len() == 0 {
		return nil, nil
	}

	yamlNode, ok := e.findExtension(extensions, ExtIgnore)

	if !ok {
		return nil, nil
	}

	result, err := e.parseIgnore(yamlNode)

	if err != nil {
		return nil, fmt.Errorf("failed to parse extension value: %w", err)
	}

	return result, nil
}

// Parses the given YAML node.
func (e *Extensions) parseIgnore(node *yaml.Node) (*bool, error) {
	switch node.Kind {
	case yaml.ScalarNode:
		// Handle null values
		if node.Tag == "!!null" || (node.Value == "" && node.Tag == "") {
			return nil, nil
		}

		value, err := strconv.ParseBool(node.Value)

		if err != nil {
			// Not a boolean, try parsing as comma-separated target list
			parts := strings.Split(node.Value, ",")

			for _, part := range parts {
				if strings.TrimSpace(part) == e.target.Target {
					result := true
					return &result, nil
				}
			}

			result := false
			return &result, nil
		}

		return &value, nil
	default:
		return nil, fmt.Errorf("unsupported YAML node kind: %v", node.Kind)
	}
}
