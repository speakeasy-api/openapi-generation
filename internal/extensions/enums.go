package extensions

import (
	"fmt"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi/jsonschema/oas3"
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
