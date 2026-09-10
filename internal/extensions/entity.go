package extensions

import (
	"fmt"
	"slices"

	"gopkg.in/yaml.v3"
)

// Describes the parsed x-speakeasy-entity extension configuration.
type Entity struct {
	// All entity names described by the x-speakeasy-entity extension configuration.
	Names []string `json:"names" yaml:"names"`
}

// Returns a new Entity.
func NewEntity() *Entity {
	return &Entity{
		Names: make([]string, 0, 1),
	}
}

// Adds entity names to the names, if they do not already exist.
func (e *Entity) AddNames(names ...string) {
	if e.Names == nil {
		e.Names = make([]string, 0, len(names))
	}

	for _, name := range names {
		if slices.Contains(e.Names, name) {
			continue
		}

		e.Names = append(e.Names, name)
	}

	// Sort for consistency.
	slices.Sort(e.Names)
}

// Clone creates a deep copy of the Entity
func (e *Entity) Clone() *Entity {
	if e == nil {
		return nil
	}

	cloned := &Entity{
		Names: slices.Clone(e.Names),
	}

	return cloned
}

// Merges the given Entity into this Entity. The algorithm adds data from the
// given Entity to this Entity where it is undefined. It does not remove any
// data from this Entity.
func (e *Entity) Merge(other *Entity) {
	if e == nil || other == nil {
		return
	}

	e.AddNames(other.Names...)
}

// Handles parsing of the x-speakeasy-entity extension from the
// given OpenAPI extensions map.
func (e *Extensions) HandleEntityExtension(extensions OAExtensions) (*Entity, error) {
	if extensions.Len() == 0 {
		return nil, nil
	}

	yamlNode, ok := e.findExtension(extensions, ExtEntity)

	if !ok {
		return nil, nil
	}

	result, err := e.parseEntity(yamlNode)
	if err != nil {
		return nil, fmt.Errorf("failed to parse extension value: %w", err)
	}

	return result, nil
}

// Parses the given YAML node into an Entity structure.
func (e *Extensions) parseEntity(node *yaml.Node) (*Entity, error) {
	switch node.Kind {
	case yaml.ScalarNode:
		result := NewEntity()

		result.AddNames(node.Value)

		return result, nil
	case yaml.SequenceNode:
		names := make([]string, 0, len(node.Content))
		result := NewEntity()

		for _, item := range node.Content {
			if item.Kind != yaml.ScalarNode {
				return nil, fmt.Errorf("unsupported YAML sequence value kind: %v", item.Kind)
			}

			names = append(names, item.Value)
		}

		result.AddNames(names...)

		return result, nil
	default:
		return nil, fmt.Errorf("unsupported YAML node kind: %v", node.Kind)
	}
}
