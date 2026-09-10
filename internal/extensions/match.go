package extensions

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
)

// MatchConfig describes the x-speakeasy-match extension configuration.
// It can be used to alias/map parameters to entity fields or to specify
// that prior state values should be used for parameters in update operations.
type MatchConfig struct {
	// Path to match in the entity (e.g., "id", "object.id")
	// When specified as a scalar string, this field is populated.
	Path *string `json:"path,omitempty" yaml:"path,omitempty"`

	// Whether to use the prior state value for this parameter in update operations.
	UsePriorState bool `json:"usePriorState,omitempty" yaml:"usePriorState,omitempty"`
}

// Handles parsing of the x-speakeasy-match extension from the
// given OpenAPI extensions map.
func (e *Extensions) HandleMatchExtension(extensions OAExtensions) (*MatchConfig, error) {
	if extensions.Len() == 0 {
		return nil, nil
	}

	yamlNode, ok := e.findExtension(extensions, ExtMatch)

	if !ok {
		return nil, nil
	}

	result, err := e.parseMatch(yamlNode)
	if err != nil {
		return nil, fmt.Errorf("failed to parse extension value: %w", err)
	}

	return result, nil
}

// Parses the given YAML node.
func (e *Extensions) parseMatch(node *yaml.Node) (*MatchConfig, error) {
	switch node.Kind {
	case yaml.ScalarNode:
		// Handle null values
		if node.Tag == "!!null" || (node.Value == "" && node.Tag == "") {
			return nil, nil
		}
		// Backward compatibility: scalar string means path match
		value := node.Value
		return &MatchConfig{
			Path:          &value,
			UsePriorState: false,
		}, nil
	case yaml.MappingNode:
		var matchConfig MatchConfig
		if err := node.Decode(&matchConfig); err != nil {
			return nil, fmt.Errorf("failed to decode match config: %w", err)
		}
		// Validate that the config is meaningful - must have either a path or usePriorState
		if matchConfig.Path == nil && !matchConfig.UsePriorState {
			return nil, errors.New("x-speakeasy-match object must specify either 'path' or 'usePriorState: true'")
		}
		return &matchConfig, nil
	default:
		return nil, fmt.Errorf("unsupported YAML node kind: %v", node.Kind)
	}
}
