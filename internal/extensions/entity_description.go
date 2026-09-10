package extensions

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Describes the parsed x-speakeasy-entity-description extension configuration.
// The data is normalized into values for target-specific entities.
type EntityDescription struct {
	// Entity description for Terraform action.
	TerraformAction string `json:"terraform_action" yaml:"terraform_action"`

	// Entity description for Terraform data resource.
	TerraformDataResource string `json:"terraform_data_resource" yaml:"terraform_data_resource"`

	// Entity description for Terraform ephemeral resource.
	TerraformEphemeralResource string `json:"terraform_ephemeral_resource" yaml:"terraform_ephemeral_resource"`

	// Entity description for Terraform managed resource.
	TerraformManagedResource string `json:"terraform_managed_resource" yaml:"terraform_managed_resource"`
}

// Clone creates a deep copy of the EntityDescription
func (e *EntityDescription) Clone() *EntityDescription {
	if e == nil {
		return nil
	}

	return &EntityDescription{
		TerraformAction:            e.TerraformAction,
		TerraformDataResource:      e.TerraformDataResource,
		TerraformEphemeralResource: e.TerraformEphemeralResource,
		TerraformManagedResource:   e.TerraformManagedResource,
	}
}

// Merges the given EntityDescription into this EntityDescription. The algorithm
// adds data from the given EntityDescription to this EntityDescription where it
// is undefined. Where there is conflicting data, this EntityDescription's data
// is preserved or otherwise delegated to data-specific merge functionality. It
// does not remove any data from this EntityDescription.
func (e *EntityDescription) Merge(other *EntityDescription) {
	if e == nil || other == nil {
		return
	}

	if e.TerraformAction == "" {
		e.TerraformAction = other.TerraformAction
	}

	if e.TerraformDataResource == "" {
		e.TerraformDataResource = other.TerraformDataResource
	}

	if e.TerraformEphemeralResource == "" {
		e.TerraformEphemeralResource = other.TerraformEphemeralResource
	}

	if e.TerraformManagedResource == "" {
		e.TerraformManagedResource = other.TerraformManagedResource
	}
}

// Handles parsing of the x-speakeasy-entity-description extension from the
// given OpenAPI extensions map.
func (e *Extensions) HandleEntityDescriptionExtension(extensions OAExtensions) (*EntityDescription, error) {
	if extensions.Len() == 0 {
		return nil, nil
	}

	yamlNode, ok := e.findExtension(extensions, ExtEntityDescription)

	if !ok {
		return nil, nil
	}

	result, err := e.parseEntityDescription(yamlNode)
	if err != nil {
		return nil, fmt.Errorf("failed to parse extension value: %w", err)
	}

	return result, nil
}

// Parses the given YAML node into an EntityDescription structure.
func (e *Extensions) parseEntityDescription(node *yaml.Node) (*EntityDescription, error) {
	switch node.Kind {
	case yaml.ScalarNode:
		value := node.Value
		result := &EntityDescription{
			TerraformAction:            value,
			TerraformDataResource:      value,
			TerraformEphemeralResource: value,
			TerraformManagedResource:   value,
		}

		return result, nil
	default:
		return nil, fmt.Errorf("unsupported YAML node kind: %v", node.Kind)
	}
}
