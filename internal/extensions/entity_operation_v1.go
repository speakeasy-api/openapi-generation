package extensions

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/terraform"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi/openapi"
	"gopkg.in/yaml.v3"
)

// Collection of valid entity operations types across targets and entity
// types. This currently includes Terraform data, ephemeral, and managed
// resource operation types.
var ValidEntityOperationV1ConfigOperationTypes = []string{
	"close",  // Ephemeral resources
	"create", // Managed resources
	"delete", // Managed resources
	"invoke", // Actions
	"open",   // Ephemeral resources
	"read",   // Data and managed resources
	"update", // Managed resources
}

// Describes the parsed x-speakeasy-entity-operation extension configuration.
// The data is normalized into collections of target-specific entity operations.
type EntityOperationV1 struct {
	TerraformActions            []EntityOperationV1Config `json:"terraform_actions" yaml:"terraform_actions"`
	TerraformDataResources      []EntityOperationV1Config `json:"terraform_data_resources" yaml:"terraform_data_resources"`
	TerraformEphemeralResources []EntityOperationV1Config `json:"terraform_ephemeral_resources" yaml:"terraform_ephemeral_resources"`
	TerraformManagedResources   []EntityOperationV1Config `json:"terraform_managed_resources" yaml:"terraform_managed_resources"`
}

// Describes an individual entity operation parsed from
// Entity#OperationType[,OperationType...][#Order] string.
type EntityOperationV1Config struct {
	// Name of the entity for this operation.
	Entity string `json:"entity" yaml:"entity"`

	// Types of the entity operation.
	OperationTypes []string `json:"operation_types" yaml:"operation_types"`

	// Optional ordering of the operation compared to other definitions of the
	// same entity and operation type.
	Order *int `json:"order,omitempty" yaml:"order,omitempty"`

	// Optional SDK options for this entity operation.
	Options *EntityOperationV1Options `json:"options,omitempty" yaml:"options,omitempty"`
}

// Describes SDK options for entity operations.
type EntityOperationV1Options struct {
	// Polling configuration for this entity operation.
	Polling *EntityOperationV1Polling `json:"polling,omitempty" yaml:"polling,omitempty"`

	// Patch configuration for update operations.
	Patch *EntityOperationV1Patch `json:"patch,omitempty" yaml:"patch,omitempty"`
}

// Describes polling configuration for entity operations.
type EntityOperationV1Polling struct {
	// Overrides the number of seconds before the first request.
	DelaySeconds *int `json:"delaySeconds,omitempty" yaml:"delaySeconds,omitempty"`

	// Overrides the number of seconds between requests.
	IntervalSeconds *int `json:"intervalSeconds,omitempty" yaml:"intervalSeconds,omitempty"`

	// Overrides the number of requests to limit polling.
	LimitCount *int `json:"limitCount,omitempty" yaml:"limitCount,omitempty"`

	// Name of the polling option to use.
	Name string `json:"name" yaml:"name"`
}

// Describes patch configuration for update operations.
type EntityOperationV1Patch struct {
	// Style of patch semantics to use for updates.
	// Valid values: "only-send-changed-attributes"
	Style string `json:"style" yaml:"style"`
}

// Parses given Entity#OperationType[,OperationType...][#Order] string into an
// EntityOperationV1Config.
func ParseEntityOperationV1ConfigString(input string) (*EntityOperationV1Config, error) {
	parts := strings.Split(input, "#")

	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid entity operation format: %s, expected Entity#OperationType[,OperationType...][#Order]", input)
	}

	entity := parts[0]
	operationTypes := strings.Split(parts[1], ",")

	for index, opType := range operationTypes {
		// Legacy support for "get" operation type as alias to "read".
		if opType == "get" {
			opType = "read"
			operationTypes[index] = opType
		}

		if !slices.Contains(ValidEntityOperationV1ConfigOperationTypes, opType) {
			return nil, fmt.Errorf("invalid entity operation type: %s, valid types are: %s", opType, strings.Join(ValidEntityOperationV1ConfigOperationTypes, ", "))
		}
	}

	var order *int

	if len(parts) > 2 {
		o, err := strconv.Atoi(parts[2])

		if err != nil || o <= 0 {
			return nil, fmt.Errorf("invalid entity operation order value: %s", parts[2])
		}

		order = &o
	}

	return &EntityOperationV1Config{
		Entity:         entity,
		OperationTypes: operationTypes,
		Order:          order,
	}, nil
}

// Returns a string representation of the EntityOperationV1Config in the
// Entity#OperationType[,OperationType...][#Order] format.
func (c EntityOperationV1Config) String() string {
	if c.Order != nil {
		return fmt.Sprintf("%s#%s#%d", c.Entity, strings.Join(c.OperationTypes, ","), *c.Order)
	}

	return fmt.Sprintf("%s#%s", c.Entity, strings.Join(c.OperationTypes, ","))
}

// Handles parsing of the x-speakeasy-entity-operation extension from the
// given OpenAPI operation.
func (e *Extensions) HandleEntityOperationExtension(operation *openapi.Operation) (*EntityOperationV1, error) {
	result, err := e.handleEntityOperationV1Extension(operation.GetExtensions())
	if err != nil {
		return nil, fmt.Errorf("failed to handle x-speakeasy-entity-operation extension: %w", err)
	}

	return result, nil
}

// Handles parsing of the x-speakeasy-entity-operation extension from the
// given OpenAPI extensions map.
func (e *Extensions) handleEntityOperationV1Extension(extensions OAExtensions) (*EntityOperationV1, error) {
	if extensions.Len() == 0 {
		return nil, nil
	}

	yamlNode, ok := e.findExtension(extensions, ExtEntityOperation)

	if !ok {
		return nil, nil
	}

	result, err := e.parseEntityOperationV1(yamlNode)
	if err != nil {
		return nil, fmt.Errorf("failed to parse extension value: %w", err)
	}

	return result, nil
}

// Parses the given YAML node into an EntityOperationV1 structure.
func (e *Extensions) parseEntityOperationV1(node *yaml.Node) (*EntityOperationV1, error) {
	switch node.Kind {
	case yaml.ScalarNode:
		configs, err := e.parseEntityOperationV1ConfigsFromNode(node)
		if err != nil {
			return nil, errors.NewValidationError(fmt.Sprintf("failed to parse %s string", ExtEntityOperation.Name()), node, err)
		}
		return e.categorizeEntityOperationV1Configs(configs), nil

	case yaml.SequenceNode:
		configs, err := e.parseEntityOperationV1ConfigsFromNode(node)
		if err != nil {
			return nil, errors.NewValidationError(fmt.Sprintf("failed to parse %s array", ExtEntityOperation.Name()), node, err)
		}
		return e.categorizeEntityOperationV1Configs(configs), nil

	case yaml.MappingNode:
		return e.parseEntityOperationV1MappingNode(node)

	default:
		return nil, fmt.Errorf("unsupported YAML node kind: %v", node.Kind)
	}
}

// parseEntityOperationV1ConfigFromMapping parses a mapping node containing entityOperation and options
func (e *Extensions) parseEntityOperationV1ConfigFromMapping(node *yaml.Node) (*EntityOperationV1Config, error) {
	if node.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("expected mapping node, got %v", node.Kind)
	}

	var entityOperationStr string
	var options *EntityOperationV1Options

	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]

		switch keyNode.Value {
		case "entityOperation":
			if valueNode.Kind != yaml.ScalarNode {
				return nil, errors.New("entityOperation must be a scalar string")
			}
			entityOperationStr = valueNode.Value
		case "options":
			if valueNode.Kind != yaml.MappingNode {
				return nil, errors.New("options must be a mapping")
			}
			opts, err := e.parseEntityOperationV1Options(valueNode)
			if err != nil {
				return nil, fmt.Errorf("failed to parse options: %w", err)
			}
			options = opts
		default:
			return nil, fmt.Errorf("unknown key in entity operation mapping: %s", keyNode.Value)
		}
	}

	if entityOperationStr == "" {
		return nil, errors.New("entityOperation is required in mapping form")
	}

	config, err := ParseEntityOperationV1ConfigString(entityOperationStr)
	if err != nil {
		return nil, err
	}

	config.Options = options

	return config, nil
}

// parseEntityOperationV1Options parses the options mapping node
func (e *Extensions) parseEntityOperationV1Options(node *yaml.Node) (*EntityOperationV1Options, error) {
	if node.Kind != yaml.MappingNode {
		return nil, errors.New("expected mapping node for options")
	}

	options := &EntityOperationV1Options{}

	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]

		switch keyNode.Value {
		case "polling":
			if valueNode.Kind != yaml.MappingNode {
				return nil, errors.New("polling must be a mapping")
			}
			polling, err := e.parseEntityOperationV1Polling(valueNode)
			if err != nil {
				return nil, fmt.Errorf("failed to parse polling: %w", err)
			}
			options.Polling = polling
		case "patch":
			if valueNode.Kind != yaml.MappingNode {
				return nil, errors.New("patch must be a mapping")
			}
			patch, err := e.parseEntityOperationV1Patch(valueNode)
			if err != nil {
				return nil, fmt.Errorf("failed to parse patch: %w", err)
			}
			options.Patch = patch
		default:
			return nil, fmt.Errorf("unknown key in options: %s", keyNode.Value)
		}
	}

	return options, nil
}

// parseEntityOperationV1Polling parses the polling mapping node
func (e *Extensions) parseEntityOperationV1Polling(node *yaml.Node) (*EntityOperationV1Polling, error) {
	if node.Kind != yaml.MappingNode {
		return nil, errors.New("expected mapping node for polling")
	}

	polling := &EntityOperationV1Polling{}

	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]

		switch keyNode.Value {
		case "name":
			if valueNode.Kind != yaml.ScalarNode {
				return nil, errors.New("polling name must be a scalar string")
			}
			polling.Name = valueNode.Value
		case "delaySeconds":
			if valueNode.Kind != yaml.ScalarNode {
				return nil, errors.New("polling delaySeconds must be a scalar integer")
			}
			val, err := strconv.Atoi(valueNode.Value)
			if err != nil {
				return nil, fmt.Errorf("invalid polling delaySeconds value: %s", valueNode.Value)
			}
			if val < 0 {
				return nil, errors.New("polling delaySeconds must be non-negative")
			}
			polling.DelaySeconds = &val
		case "intervalSeconds":
			if valueNode.Kind != yaml.ScalarNode {
				return nil, errors.New("polling intervalSeconds must be a scalar integer")
			}
			val, err := strconv.Atoi(valueNode.Value)
			if err != nil {
				return nil, fmt.Errorf("invalid polling intervalSeconds value: %s", valueNode.Value)
			}
			if val <= 0 {
				return nil, errors.New("polling intervalSeconds must be positive")
			}
			polling.IntervalSeconds = &val
		case "limitCount":
			if valueNode.Kind != yaml.ScalarNode {
				return nil, errors.New("polling limitCount must be a scalar integer")
			}
			val, err := strconv.Atoi(valueNode.Value)
			if err != nil {
				return nil, fmt.Errorf("invalid polling limitCount value: %s", valueNode.Value)
			}
			if val <= 0 {
				return nil, errors.New("polling limitCount must be positive")
			}
			polling.LimitCount = &val
		default:
			return nil, fmt.Errorf("unknown key in polling: %s", keyNode.Value)
		}
	}

	if polling.Name == "" {
		return nil, errors.New("polling name is required")
	}

	return polling, nil
}

// parseEntityOperationV1Patch parses the patch mapping node
func (e *Extensions) parseEntityOperationV1Patch(node *yaml.Node) (*EntityOperationV1Patch, error) {
	if node.Kind != yaml.MappingNode {
		return nil, errors.New("expected mapping node for patch")
	}

	patch := &EntityOperationV1Patch{}

	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]

		switch keyNode.Value {
		case "style":
			if valueNode.Kind != yaml.ScalarNode {
				return nil, errors.New("patch style must be a scalar string")
			}
			patch.Style = valueNode.Value
		default:
			return nil, fmt.Errorf("unknown key in patch: %s", keyNode.Value)
		}
	}

	if patch.Style == "" {
		return nil, errors.New("patch style is required")
	}

	// Validate style value
	validStyles := []string{"only-send-changed-attributes"}
	if !slices.Contains(validStyles, patch.Style) {
		return nil, fmt.Errorf("invalid patch style: %s, valid styles are: %s", patch.Style, strings.Join(validStyles, ", "))
	}

	return patch, nil
}

// parseEntityOperationV1ConfigsFromNode parses EntityOperationV1Config from either scalar or sequence node
func (e *Extensions) parseEntityOperationV1ConfigsFromNode(node *yaml.Node) ([]EntityOperationV1Config, error) {
	switch node.Kind {
	case yaml.ScalarNode:
		config, err := ParseEntityOperationV1ConfigString(node.Value)
		if err != nil {
			return nil, err
		}
		return []EntityOperationV1Config{*config}, nil

	case yaml.SequenceNode:
		configs := make([]EntityOperationV1Config, 0, len(node.Content))
		for _, item := range node.Content {
			// Each item can be either a scalar string or a mapping
			switch item.Kind {
			case yaml.ScalarNode:
				config, err := ParseEntityOperationV1ConfigString(item.Value)
				if err != nil {
					return nil, err
				}
				configs = append(configs, *config)
			case yaml.MappingNode:
				config, err := e.parseEntityOperationV1ConfigFromMapping(item)
				if err != nil {
					return nil, err
				}
				configs = append(configs, *config)
			default:
				return nil, fmt.Errorf("unsupported node kind in sequence: %v", item.Kind)
			}
		}
		return configs, nil

	default:
		return nil, fmt.Errorf("unsupported node kind for entity operation config: %v", node.Kind)
	}
}

// categorizeEntityOperationV1Configs categorizes operation configs into actions, data, ephemeral, and managed resources
func (e *Extensions) categorizeEntityOperationV1Configs(configs []EntityOperationV1Config) *EntityOperationV1 {
	result := &EntityOperationV1{
		TerraformActions:            make([]EntityOperationV1Config, 0),
		TerraformDataResources:      make([]EntityOperationV1Config, 0),
		TerraformEphemeralResources: make([]EntityOperationV1Config, 0),
		TerraformManagedResources:   make([]EntityOperationV1Config, 0),
	}

	for _, config := range configs {
		if e.hasEntityOperationV1ActionOperations(config.OperationTypes) {
			result.TerraformActions = append(result.TerraformActions, config)
		}

		if e.hasEntityOperationV1DataResourceOperations(config.OperationTypes) {
			result.TerraformDataResources = append(result.TerraformDataResources, config)
		}

		if e.hasEntityOperationV1EphemeralResourceOperations(config.OperationTypes) {
			result.TerraformEphemeralResources = append(result.TerraformEphemeralResources, config)
		}

		if e.hasEntityOperationV1ManagedResourceOperations(config.OperationTypes) {
			result.TerraformManagedResources = append(result.TerraformManagedResources, config)
		}
	}

	return result
}

// hasEntityOperationV1ActionOperations checks if operation types contain valid action operations
func (e *Extensions) hasEntityOperationV1ActionOperations(operationTypes []string) bool {
	return slices.ContainsFunc(operationTypes, func(opType string) bool {
		return slices.Contains(terraform.ActionOperationTypes, terraform.ActionOperationType(opType))
	})
}

// hasEntityOperationV1DataResourceOperations checks if operation types contain valid data resource operations
func (e *Extensions) hasEntityOperationV1DataResourceOperations(operationTypes []string) bool {
	return slices.ContainsFunc(operationTypes, func(opType string) bool {
		return slices.Contains(terraform.DataResourceOperationTypes, terraform.DataResourceOperationType(opType))
	})
}

// hasEntityOperationV1EphemeralResourceOperations checks if operation types contain valid ephemeral resource operations
func (e *Extensions) hasEntityOperationV1EphemeralResourceOperations(operationTypes []string) bool {
	return slices.ContainsFunc(operationTypes, func(opType string) bool {
		return slices.Contains(terraform.EphemeralResourceOperationTypes, terraform.EphemeralResourceOperationType(opType))
	})
}

// hasEntityOperationV1ManagedResourceOperations checks if operation types contain valid managed resource operations
func (e *Extensions) hasEntityOperationV1ManagedResourceOperations(operationTypes []string) bool {
	return slices.ContainsFunc(operationTypes, func(opType string) bool {
		return slices.Contains(terraform.ManagedResourceOperationTypes, terraform.ManagedResourceOperationType(opType))
	})
}

// parseEntityOperationV1MappingNode handles parsing of mapping nodes with terraform-datasource/terraform-resource keys
func (e *Extensions) parseEntityOperationV1MappingNode(node *yaml.Node) (*EntityOperationV1, error) {
	result := &EntityOperationV1{
		TerraformActions:            make([]EntityOperationV1Config, 0),
		TerraformDataResources:      make([]EntityOperationV1Config, 0),
		TerraformEphemeralResources: make([]EntityOperationV1Config, 0),
		TerraformManagedResources:   make([]EntityOperationV1Config, 0),
	}

	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]

		// Skip null values
		if valueNode.Kind == yaml.ScalarNode && valueNode.Tag == "!!null" {
			continue
		}

		configs, err := e.parseEntityOperationV1ConfigsFromNode(valueNode)
		if err != nil {
			return nil, errors.NewValidationError(fmt.Sprintf("failed to parse %s %s", ExtEntityOperation.Name(), keyNode.Value), valueNode, err)
		}

		switch keyNode.Value {
		case "terraform-action":
			for _, config := range configs {
				if e.hasEntityOperationV1ActionOperations(config.OperationTypes) {
					result.TerraformActions = append(result.TerraformActions, config)
				}
			}
		case "terraform-datasource":
			for _, config := range configs {
				if e.hasEntityOperationV1DataResourceOperations(config.OperationTypes) {
					result.TerraformDataResources = append(result.TerraformDataResources, config)
				}
			}
		case "terraform-ephemeral-resource":
			for _, config := range configs {
				if e.hasEntityOperationV1EphemeralResourceOperations(config.OperationTypes) {
					result.TerraformEphemeralResources = append(result.TerraformEphemeralResources, config)
				}
			}
		case "terraform-resource":
			for _, config := range configs {
				if e.hasEntityOperationV1ManagedResourceOperations(config.OperationTypes) {
					result.TerraformManagedResources = append(result.TerraformManagedResources, config)
				}
			}
		}
	}

	return result, nil
}
