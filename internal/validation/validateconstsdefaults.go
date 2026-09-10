package validation

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/speakeasy-api/openapi/jsonschema/oas3"
	"github.com/speakeasy-api/openapi/linter"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/validation"
	"github.com/speakeasy-api/openapi/values"
	"gopkg.in/yaml.v3"
)

type ValidateConstsDefaults struct{}

var _ Rule = (*ValidateConstsDefaults)(nil)

func (r *ValidateConstsDefaults) ID() string {
	return "generator-validate-consts-defaults"
}

func (r *ValidateConstsDefaults) Category() string {
	return "validation"
}

func (r *ValidateConstsDefaults) Summary() string {
	return "Validate const and default values match their schema type."
}

func (r *ValidateConstsDefaults) HowToFix() string {
	return "Update const/default values to match the schema type (and enum values), and only use null when the schema is nullable."
}

func (r *ValidateConstsDefaults) Description() string {
	return "Validate const and default values match their declared schema type. Type mismatches lead to invalid clients or runtime errors."
}

func (r *ValidateConstsDefaults) Link() string {
	return ""
}

func (r *ValidateConstsDefaults) DefaultSeverity() validation.Severity {
	return validation.SeverityWarning
}

func (r *ValidateConstsDefaults) Versions() []string {
	return nil // Applies to all versions
}

func (r *ValidateConstsDefaults) Run(ctx context.Context, docInfo *linter.DocumentInfo[*openapi.OpenAPI], config *linter.RuleConfig) []error {
	if docInfo == nil || docInfo.Index == nil {
		return nil
	}

	var validationErrors []error

	// Iterate through all schemas
	allSchemas := docInfo.Index.GetAllSchemas()
	for _, indexNode := range allSchemas {
		if indexNode == nil || indexNode.Node == nil {
			continue
		}

		schemaRef := indexNode.Node
		schema := schemaRef.GetSchema()
		if schema == nil {
			continue
		}

		// Get types from schema
		types := schema.GetType()
		if len(types) == 0 {
			continue
		}

		// Check nullable
		nullable := schema.Nullable != nil && *schema.Nullable

		// Check if we have multiple types (type array)
		if len(types) == 1 {
			// Single type value
			validationErrors = append(validationErrors, r.validateTypeValue(schema, string(types[0]), nullable)...)
		} else {
			// Type array - check if any type in the array contains "null"
			containsNull := false
			for _, t := range types {
				if string(t) == "null" {
					containsNull = true
					break
				}
			}
			validationErrors = append(validationErrors, r.validateTypeArray(schema, types, containsNull || nullable)...)
		}
	}

	return validationErrors
}

func (r *ValidateConstsDefaults) validateTypeValue(schema *oas3.Schema, typ string, nullable bool) []error {
	var validationErrors []error

	// Get const and default values directly from schema
	constVal := schema.Const
	defaultVal := schema.Default

	// Validate const value if present
	if constVal != nil {
		var val any
		constValNullValue := constVal.Tag == "!!null"

		if !constValNullValue {
			_ = constVal.Decode(&val)
		}

		// Get the key node for the const field for error reporting
		constNode := schema.GetPropertyNode("Const")
		if constNode == nil {
			constNode = constVal // Fallback to value node
		}

		validationErrors = append(validationErrors, r.validateDefaultConstVal("const", strings.ToLower(typ), schema, constNode, val, constValNullValue, nullable)...)
	}

	// Validate default value if present
	if defaultVal != nil {
		var val any
		defaultValNullValue := defaultVal.Tag == "!!null"

		if !defaultValNullValue {
			_ = defaultVal.Decode(&val)
		}

		// Get the key node for the default field for error reporting
		defaultNode := schema.GetPropertyNode("Default")
		if defaultNode == nil {
			defaultNode = defaultVal // Fallback to value node
		}

		validationErrors = append(validationErrors, r.validateDefaultConstVal("default", strings.ToLower(typ), schema, defaultNode, val, defaultValNullValue, nullable)...)
	}

	return validationErrors
}

func (r *ValidateConstsDefaults) validateDefaultConstVal(defaultConstType, typ string, schema *oas3.Schema, valNode *yaml.Node, val any, nullVal, nullable bool) []error {
	var validationErrors []error

	if nullVal {
		if !nullable && typ != "null" {
			return append(validationErrors, &validation.Error{
				Rule:            r.ID(),
				Severity:        r.DefaultSeverity(),
				Node:            valNode,
				UnderlyingError: fmt.Errorf("`%s` value `null` cannot be used with non-nullable type `%s`", defaultConstType, typ),
			})
		}
		return validationErrors
	}

	printVal := fmt.Sprintf("`%v`", val)
	if val == "" || reflect.TypeOf(val).Kind() == reflect.String {
		printVal = fmt.Sprintf("`%v`", val)
	}

	typeMismatchError := &validation.Error{
		Rule:            r.ID(),
		Severity:        r.DefaultSeverity(),
		Node:            valNode,
		UnderlyingError: fmt.Errorf("`%s` value %s does not match type `%s`", defaultConstType, printVal, typ),
	}
	enumMismatchError := &validation.Error{
		Rule:            r.ID(),
		Severity:        r.DefaultSeverity(),
		Node:            valNode,
		UnderlyingError: fmt.Errorf("`%s` value %s not present in `enum` values", defaultConstType, printVal),
	}

	// Get enum values from schema
	enumValues := schema.Enum

	switch typ {
	case "string":
		if _, ok := val.(string); !ok {
			validationErrors = append(validationErrors, typeMismatchError)
		} else if !r.validateValAgainstEnum(val, enumValues) {
			validationErrors = append(validationErrors, enumMismatchError)
		}
	case "integer":
		switch val.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			if !r.validateValAgainstEnum(val, enumValues) {
				validationErrors = append(validationErrors, enumMismatchError)
			}
		default:
			validationErrors = append(validationErrors, typeMismatchError)
		}
	case "number":
		switch val.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		case float32, float64:
		default:
			validationErrors = append(validationErrors, typeMismatchError)
		}
	case "boolean":
		if _, ok := val.(bool); !ok {
			validationErrors = append(validationErrors, typeMismatchError)
		}
	case "null":
		if !nullVal {
			validationErrors = append(validationErrors, typeMismatchError)
		}
	}

	return validationErrors
}

func (r *ValidateConstsDefaults) validateValAgainstEnum(val any, enumValues []values.Value) bool {
	if len(enumValues) == 0 {
		return true
	}

	found := false

	for _, enumValNode := range enumValues {
		var enumVal any
		_ = enumValNode.Decode(&enumVal)

		if reflect.DeepEqual(val, enumVal) {
			found = true
			break
		}
	}

	return found
}

func (r *ValidateConstsDefaults) validateTypeArray(schema *oas3.Schema, types []oas3.SchemaType, containsNull bool) []error {
	var validationErrors []error

	matchesSomeType := false

	for _, typeName := range types {
		typ := string(typeName)
		if strings.ToLower(typ) != "null" {
			f := r.validateTypeValue(schema, typ, containsNull)
			if len(f) == 0 {
				matchesSomeType = true
				break
			}

			validationErrors = append(validationErrors, f...)
		}
	}

	if matchesSomeType {
		return []error{}
	}

	return validationErrors
}
