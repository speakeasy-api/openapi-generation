package extensions

import (
	"fmt"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi/jsonschema/oas3"
	"gopkg.in/yaml.v3"
)

func (e *Extensions) GetEnumNames(schema *oas3.Schema) ([]string, map[any]string, error) {
	if schema.GetExtensions().Len() == 0 {
		return nil, nil, nil
	}

	enumExtension, ok := e.findExtension(schema.GetExtensions(), ExtEnums)
	if !ok {
		return nil, nil, nil
	}

	var names []string
	err := enumExtension.Decode(&names)
	if err == nil {
		return names, nil, nil
	}

	var nameMap map[any]string
	err = enumExtension.Decode(&nameMap)
	if err != nil {
		return nil, nil, errors.NewValidationError("failed to unmarshal "+ExtEnums.Name(), enumExtension, err)
	}

	return nil, nameMap, nil
}

func (e *Extensions) GetEnumDescriptions(schema *oas3.Schema) ([]string, map[any]string, error) {
	if schema.GetExtensions().Len() == 0 {
		return nil, nil, nil
	}

	descriptionExtension, ok := e.findExtension(schema.GetExtensions(), ExtEnumDescriptions)
	if !ok {
		return nil, nil, nil
	}

	var descriptions []string
	if err := descriptionExtension.Decode(&descriptions); err == nil {
		return descriptions, nil, nil
	}

	var descriptionMap map[any]string
	if err := descriptionExtension.Decode(&descriptionMap); err != nil {
		return nil, nil, errors.NewValidationError("failed to unmarshal "+ExtEnumDescriptions.Name(), descriptionExtension, err)
	}

	return nil, descriptionMap, nil
}

func (e *Extensions) GetEnumGroups(schema *oas3.Schema) ([]string, map[string]string, error) {
	if schema.GetExtensions().Len() == 0 {
		return nil, nil, nil
	}

	node, ok := e.findExtension(schema.GetExtensions(), ExtEnumGroups)
	if !ok {
		return nil, nil, nil
	}

	values := schema.GetEnum()
	switch node.Kind {
	case yaml.SequenceNode:
		if len(node.Content) != len(values) {
			return nil, nil, errors.NewValidationError("`x-speakeasy-enum-groups` array must be the same length as enum values array", node, nil)
		}
		groups := make([]string, len(node.Content))
		for i, item := range node.Content {
			if item == nil || item.Kind != yaml.ScalarNode || item.Tag != "!!str" {
				return nil, nil, errors.NewValidationError("`x-speakeasy-enum-groups` entries must be strings", item, nil)
			}
			groups[i] = strings.TrimSpace(item.Value)
		}
		return groups, nil, nil
	case yaml.MappingNode:
		if len(node.Content)%2 != 0 {
			return nil, nil, errors.NewValidationError("`x-speakeasy-enum-groups` must be a map of enum values to group titles", node, nil)
		}
		knownValues := make(map[string]struct{}, len(values))
		for _, value := range values {
			if value != nil && value.Kind == yaml.ScalarNode && value.Tag != "!!null" {
				knownValues[value.Value] = struct{}{}
			}
		}
		groups := make(map[string]string, len(node.Content)/2)
		seen := make(map[string]struct{}, len(node.Content)/2)
		for i := 0; i < len(node.Content); i += 2 {
			key, title := node.Content[i], node.Content[i+1]
			if key == nil || key.Kind != yaml.ScalarNode || key.Tag == "!!null" || title == nil || title.Kind != yaml.ScalarNode || title.Tag != "!!str" {
				return nil, nil, errors.NewValidationError("`x-speakeasy-enum-groups` keys must be enum values and titles must be strings", node, nil)
			}
			if _, ok := knownValues[key.Value]; !ok {
				return nil, nil, errors.NewValidationError(fmt.Sprintf("`x-speakeasy-enum-groups` map contains key `%s` that does not exist in enum values", key.Value), key, nil)
			}
			if _, exists := seen[key.Value]; exists {
				return nil, nil, errors.NewValidationError(fmt.Sprintf("`x-speakeasy-enum-groups` map contains duplicate key `%s`", key.Value), key, nil)
			}
			seen[key.Value] = struct{}{}
			if trimmed := strings.TrimSpace(title.Value); trimmed != "" {
				groups[key.Value] = trimmed
			}
		}
		return nil, groups, nil
	default:
		return nil, nil, errors.NewValidationError("`x-speakeasy-enum-groups` must be either an array or a map", node, nil)
	}
}

func (e *Extensions) IsOpenEnum(schema *oas3.Schema) (bool, error) {
	if schema.GetExtensions().Len() == 0 {
		return false, nil
	}

	options := map[string]bool{"": false, "disallow": false, "allow": true}

	value, err := getExtensionValueWithValidation(e.GetResolvedName(ExtUnknownValues), schema.GetExtensions(), "disallow", func(value string) error {
		if _, found := options[value]; !found {
			return fmt.Errorf("%s: value is not allowed", value)
		}

		return nil
	})
	if err != nil {
		return false, err
	}

	return options[value], nil
}

func (e *Extensions) GetEnumFormat(schema *oas3.Schema) (string, error) {
	if schema.GetExtensions().Len() == 0 {
		return "", nil
	}

	value, err := getExtensionValue(e.GetResolvedName(ExtEnumFormat), schema.GetExtensions(), "")
	if err != nil {
		return "", err
	}

	return value, nil
}
