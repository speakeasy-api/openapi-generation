package extensions

import (
	"fmt"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Describes the parsed x-speakeasy-entity-version extension configuration.
// The data is normalized into values for target-specific entities.
type EntityVersion struct {
	// Entity version for Terraform managed resource.
	TerraformManagedResource int64 `json:"terraform_managed_resource" yaml:"terraform_managed_resource"`
}

// Clone creates a deep copy of the EntityVersion
func (e *EntityVersion) Clone() *EntityVersion {
	if e == nil {
		return nil
	}

	return &EntityVersion{
		TerraformManagedResource: e.TerraformManagedResource,
	}
}

// Merges the given EntityVersion into this EntityVersion. The algorithm
// adds data from the given EntityVersion to this EntityVersion where it is
// undefined. Where there is conflicting data, this EntityVersion's data is
// preserved or otherwise delegated to data-specific merge functionality. It
// does not remove any data from this EntityVersion.
func (e *EntityVersion) Merge(other *EntityVersion) {
	if e == nil || other == nil {
		return
	}

	if e.TerraformManagedResource == 0 {
		e.TerraformManagedResource = other.TerraformManagedResource
	}
}

// Handles parsing of the x-speakeasy-entity-version extension from the
// given OpenAPI extensions map.
func (e *Extensions) HandleEntityVersionExtension(extensions OAExtensions) (*EntityVersion, error) {
	if extensions.Len() == 0 {
		return nil, nil
	}

	yamlNode, ok := e.findExtension(extensions, ExtEntityVersion)

	if !ok {
		return nil, nil
	}

	result, err := e.parseEntityVersion(yamlNode)
	if err != nil {
		return nil, fmt.Errorf("failed to parse extension value: %w", err)
	}

	return result, nil
}

// Parses the given YAML node into an EntityVersion structure.
func (e *Extensions) parseEntityVersion(node *yaml.Node) (*EntityVersion, error) {
	switch node.Kind {
	case yaml.ScalarNode:
		value := node.Value

		intValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse integer value: %w", err)
		}

		result := &EntityVersion{
			TerraformManagedResource: intValue,
		}

		return result, nil
	default:
		return nil, fmt.Errorf("unsupported YAML node kind: %v", node.Kind)
	}
}
