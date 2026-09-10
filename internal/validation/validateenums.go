package validation

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"

	"github.com/speakeasy-api/openapi-generation/v2/internal/validation/sanitization"
	"github.com/speakeasy-api/openapi/linter"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/validation"
	"gopkg.in/yaml.v3"
)

type ValidateEnums struct{}

var _ Rule = (*ValidateEnums)(nil)

func (r *ValidateEnums) ID() string {
	return "generator-validate-enums"
}

func (r *ValidateEnums) Category() string {
	return "validation"
}

func (r *ValidateEnums) Summary() string {
	return "Validate enums for type safety and generation."
}

func (r *ValidateEnums) HowToFix() string {
	return "Use string or integer enums with unique values, keep x-speakeasy-enums/descriptions aligned with enum values, and include null when enums are nullable."
}

func (r *ValidateEnums) Description() string {
	return "Validate enums for type generation by checking value types, uniqueness, and x-speakeasy-enums definitions. Also ensure nullable usage is consistent with the schema's type."
}

func (r *ValidateEnums) Link() string {
	return ""
}

func (r *ValidateEnums) DefaultSeverity() validation.Severity {
	return validation.SeverityError
}

func (r *ValidateEnums) Versions() []string {
	return nil // Applies to all versions
}

type enumNameReference struct {
	Node   *yaml.Node
	Result sanitization.Result
	Value  string
	Path   string
}

func (r *ValidateEnums) Run(ctx context.Context, docInfo *linter.DocumentInfo[*openapi.OpenAPI], config *linter.RuleConfig) []error {
	if docInfo == nil || docInfo.Document == nil || docInfo.Index == nil {
		return nil
	}

	var validationErrors []error

	// Iterate all schemas and check for enums
	allSchemas := docInfo.Index.GetAllSchemas()
	for _, indexNode := range allSchemas {
		schemaRef := indexNode.Node
		if schemaRef == nil {
			continue
		}

		schema := schemaRef.GetSchema()
		if schema == nil {
			continue
		}

		// Check if schema has enum values
		enumValues := schema.GetEnum()
		if len(enumValues) == 0 {
			continue // Not an enum
		}

		schemaNode := schemaRef.GetRootNode()
		if schemaNode == nil {
			continue
		}

		// Determine nullable status
		nullable := false
		if schema.Nullable != nil && *schema.Nullable {
			nullable = true
		}

		// Check type array for "null" and non-null types
		typeNodes := schema.GetType()
		var nonNullTypes []string
		for _, typeNode := range typeNodes {
			typeStr := string(typeNode)
			if typeStr == "null" {
				nullable = true
			} else if typeStr != "" {
				nonNullTypes = append(nonNullTypes, typeStr)
			}
		}

		// Determine enum type - must have exactly one non-null type
		enumType := "string"
		if len(nonNullTypes) > 1 {
			// Error: multiple non-null types
			typeNode := schema.GetPropertyNode("Type")
			if typeNode == nil {
				typeNode = schemaNode
			}
			validationErrors = append(validationErrors, &validation.Error{
				Rule:            r.ID(),
				Severity:        r.DefaultSeverity(),
				Node:            typeNode,
				UnderlyingError: errors.New("only one type allowed for `enum`"),
			})
			continue
		} else if len(nonNullTypes) == 1 {
			enumType = nonNullTypes[0]
		}

		// Validate enum type
		if enumType != "string" && enumType != "integer" {
			typeNode := schema.GetPropertyNode("Type")
			if typeNode == nil {
				typeNode = schemaNode
			}
			validationErrors = append(validationErrors, &validation.Error{
				Rule:            r.ID(),
				Severity:        validation.SeverityWarning,
				Node:            typeNode,
				UnderlyingError: errors.New("only `enum` types of `string` or `integer` supported. Enum type won't be generated and will be treated as base type"),
			})
			continue
		}

		// Process enum values
		foundNull := false
		values := []string{}
		valueNodes := map[string]*yaml.Node{}

		for _, enumValue := range enumValues {
			if enumValue == nil || enumValue.Tag == "!!null" {
				foundNull = true
				continue
			}

			enumStringValue := enumValue.Value

			// Validate integer enums
			if enumType == "integer" {
				if _, err := strconv.ParseInt(enumStringValue, 10, 64); err != nil {
					validationErrors = append(validationErrors, &validation.Error{
						Rule:            r.ID(),
						Severity:        r.DefaultSeverity(),
						Node:            enumValue,
						UnderlyingError: fmt.Errorf("enum value `%s` is not an `integer`", enumStringValue),
					})
				}
			}

			// Track enum values and nodes (duplicate detection is handled by built-in semantic-duplicated-enum rule)
			if !slices.Contains(values, enumStringValue) {
				values = append(values, enumStringValue)
				valueNodes[enumStringValue] = enumValue
			}
		}

		// Process x-speakeasy-enums extension
		names := []enumNameReference{}
		extensions := schema.GetExtensions()

		if enumsExt, ok := extensions.Get("x-speakeasy-enums"); ok {
			enumsNode := enumsExt

			switch enumsNode.Kind {
			case yaml.SequenceNode:
				// Array format
				for _, content := range enumsNode.Content {
					// Skip null values, just like we do for enum values
					if content == nil || content.Tag == "!!null" {
						continue
					}

					names = append(names, enumNameReference{
						Node:   content,
						Result: sanitization.GetSanitizedEnumNameResult(content.Value),
						Value:  content.Value,
					})
				}

				if len(names) != len(values) {
					validationErrors = append(validationErrors, &validation.Error{
						Rule:            r.ID(),
						Severity:        r.DefaultSeverity(),
						Node:            enumsNode,
						UnderlyingError: errors.New("`x-speakeasy-enums` array must be the same length as enum values array"),
					})
				}
			case yaml.MappingNode:
				// Map format: value -> name mapping
				enumValueMap := make(map[string]string)
				mapNodeMap := make(map[string]*yaml.Node)

				// Parse the mapping
				for i := 0; i < len(enumsNode.Content); i += 2 {
					if i+1 < len(enumsNode.Content) {
						keyNode := enumsNode.Content[i]
						valueNode := enumsNode.Content[i+1]
						enumValueMap[keyNode.Value] = valueNode.Value
						mapNodeMap[keyNode.Value] = valueNode
					}
				}

				// Validate that all enum values exist in the map and collect names
				for _, enumValue := range values {
					if enumValue == "null" {
						continue
					}

					if customName, exists := enumValueMap[enumValue]; exists {
						names = append(names, enumNameReference{
							Node:   mapNodeMap[enumValue],
							Result: sanitization.GetSanitizedEnumNameResult(customName),
							Value:  customName,
						})
					} else {
						// Use the enum value itself if not found in map
						names = append(names, enumNameReference{
							Node:   valueNodes[enumValue],
							Result: sanitization.GetSanitizedEnumNameResult(enumValue),
							Value:  enumValue,
						})
					}
				}

				// Validate that all keys in the map correspond to actual enum values
				for mapKey := range enumValueMap {
					if !slices.Contains(values, mapKey) {
						validationErrors = append(validationErrors, &validation.Error{
							Rule:            r.ID(),
							Severity:        r.DefaultSeverity(),
							Node:            enumsNode,
							UnderlyingError: fmt.Errorf("`x-speakeasy-enums` map contains key `%s` that does not exist in enum values", mapKey),
						})
					}
				}
			default:
				validationErrors = append(validationErrors, &validation.Error{
					Rule:            r.ID(),
					Severity:        r.DefaultSeverity(),
					Node:            enumsNode,
					UnderlyingError: errors.New("`x-speakeasy-enums` must be either an array or a map"),
				})
			}
		}

		// Process x-speakeasy-enum-descriptions extension
		nameType := "name"

		if descriptionsExt, ok := extensions.Get("x-speakeasy-enum-descriptions"); ok {
			descriptionsNode := descriptionsExt

			switch descriptionsNode.Kind {
			case yaml.SequenceNode:
				if len(descriptionsNode.Content) != len(values) {
					validationErrors = append(validationErrors, &validation.Error{
						Rule:            r.ID(),
						Severity:        r.DefaultSeverity(),
						Node:            descriptionsNode,
						UnderlyingError: errors.New("`x-speakeasy-enum-descriptions` array must be the same length as enum values array"),
					})
				}
			case yaml.MappingNode:
				descriptionKeys := make(map[string]struct{}, len(descriptionsNode.Content)/2)
				for i := 0; i < len(descriptionsNode.Content); i += 2 {
					if i+1 >= len(descriptionsNode.Content) {
						break
					}
					keyNode := descriptionsNode.Content[i]
					key := keyNode.Value
					descriptionKeys[key] = struct{}{}

					if !slices.Contains(values, key) {
						validationErrors = append(validationErrors, &validation.Error{
							Rule:            r.ID(),
							Severity:        validation.SeverityWarning,
							Node:            keyNode,
							UnderlyingError: fmt.Errorf("`x-speakeasy-enum-descriptions` map contains key `%s` that does not exist in enum values", key),
						})
					}
				}

				for _, enumValue := range values {
					if enumValue == "null" {
						continue
					}

					if _, ok := descriptionKeys[enumValue]; !ok {
						validationErrors = append(validationErrors, &validation.Error{
							Rule:            r.ID(),
							Severity:        validation.SeverityWarning,
							Node:            valueNodes[enumValue],
							UnderlyingError: fmt.Errorf("`x-speakeasy-enum-descriptions` map missing description for enum value `%s`", enumValue),
						})
					}
				}
			default:
				validationErrors = append(validationErrors, &validation.Error{
					Rule:            r.ID(),
					Severity:        r.DefaultSeverity(),
					Node:            descriptionsNode,
					UnderlyingError: errors.New("`x-speakeasy-enum-descriptions` must be either an array or a map"),
				})
			}
		}

		// Generate names from values if no x-speakeasy-enums provided
		if len(names) == 0 {
			results := sanitization.GetSanitizedEnumNamesResult(values)

			for i, result := range results {
				if i < len(enumValues) && enumValues[i] != nil {
					names = append(names, enumNameReference{
						Node:   enumValues[i],
						Result: result,
						Value:  enumValues[i].Value,
					})
				}
			}

			nameType = "value"
		}

		// Check for name collisions
		uniqueEnums := []enumNameReference{}

		for _, nameRef := range names {
			if sanitizedName, origSanitizedName, orig := findEnumConflict(nameRef.Result, uniqueEnums); orig != nil {
				validationErrors = append(validationErrors, &validation.Error{
					Rule:            r.ID(),
					Severity:        r.DefaultSeverity(),
					Node:            nameRef.Node,
					UnderlyingError: fmt.Errorf("enum %s `%s` (`%s`) will collide with `%s` (`%s`) [line `%d`] when normalized, try using `x-speakeasy-enums`", nameType, nameRef.Value, sanitizedName, orig.Value, origSanitizedName, orig.Node.Line),
				})
			} else {
				uniqueEnums = append(uniqueEnums, nameRef)
			}
		}

		// Check for nullable without null value
		if nullable && !foundNull {
			enumNode := schema.GetPropertyNode("Enum")
			if enumNode == nil {
				enumNode = schemaNode
			}
			validationErrors = append(validationErrors, &validation.Error{
				Rule:            r.ID(),
				Severity:        validation.SeverityHint,
				Node:            enumNode,
				UnderlyingError: errors.New("enum is nullable but does not contain a `null` value"),
			})
		}
	}

	return validationErrors
}
