package schemas

import (
	"strings"

	"github.com/speakeasy-api/openapi/jsonschema/oas3"
	"github.com/speakeasy-api/openapi/values"
	"gopkg.in/yaml.v3"
)

// number is a constraint that matches any numeric type.
type number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~float32 | ~float64
}

// leastRestrictiveMax returns the largest value from the slice, or nil if any
// element is nil (meaning no restriction on that schema).
func leastRestrictiveMax[T number](values []*T) *T {
	var result *T

	for _, v := range values {
		if v == nil {
			return nil
		}

		if result == nil || *v > *result {
			result = v
		}
	}

	return result
}

// leastRestrictiveMin returns the smallest value from the slice, or nil if any
// element is nil (meaning no restriction on that schema).
func leastRestrictiveMin[T number](values []*T) *T {
	var result *T

	for _, v := range values {
		if v == nil {
			return nil
		}

		if result == nil || *v < *result {
			result = v
		}
	}

	return result
}

// mergeDescriptions returns the parent description if non-nil, otherwise
// returns the first non-nil description from the sub-schema descriptions.
func mergeDescriptions(parentDescription *string, descriptions []*string) *string {
	if parentDescription != nil {
		return parentDescription
	}

	for _, d := range descriptions {
		// NOTE: This algorithm prefers the first non-nil description, however
		// other strategies could be used such as concatenation.
		if d != nil {
			return d
		}
	}

	return nil
}

// mergePatterns combines patterns with alternation (e.g., "(A|B|C)"), or
// returns nil if any element is nil (meaning no restriction on that schema).
// Duplicate patterns are removed while preserving order.
func mergePatterns(patterns []*string) *string {
	seen := make(map[string]struct{}, len(patterns))
	uniquePatterns := make([]string, 0, len(patterns))

	for _, p := range patterns {
		if p == nil {
			return nil
		}

		if _, exists := seen[*p]; !exists {
			seen[*p] = struct{}{}
			uniquePatterns = append(uniquePatterns, *p)
		}
	}

	if len(uniquePatterns) == 0 {
		return nil
	}

	if len(uniquePatterns) == 1 {
		return &uniquePatterns[0]
	}

	result := "(" + strings.Join(uniquePatterns, "|") + ")"

	return &result
}

// mergeEnumValues combines the enum values allowed by each member of a
// flattened union into a single enum. Duplicates are removed while preserving
// order, and nil entries (null members) are dropped.
func mergeEnumValues(enumValues []values.Value) []values.Value {
	seen := make(map[string]struct{}, len(enumValues))
	merged := make([]values.Value, 0, len(enumValues))

	for _, v := range enumValues {
		if v == nil {
			continue
		}

		key := v.Tag + "\x00" + v.Value
		if _, exists := seen[key]; exists {
			continue
		}

		seen[key] = struct{}{}
		merged = append(merged, v)
	}

	if len(merged) == 0 {
		return nil
	}

	return merged
}

// subSchemaExamples holds example and examples from a sub-schema.
type subSchemaExamples struct {
	Example  values.Value
	Examples []values.Value
}

// exampleMatchesSchemaType returns true if the example's YAML type is compatible
// with the schema's declared type(s). Returns true if example is nil or schema
// has no declared types (permissive when information is missing).
func exampleMatchesSchemaType(example values.Value, schemaTypes []oas3.SchemaType) bool {
	if example == nil || len(schemaTypes) == 0 {
		return true // No example or untyped schema - allow propagation
	}

	for _, schemaType := range schemaTypes {
		switch schemaType {
		case "string":
			if example.Tag == "!!str" {
				return true
			}
		case "integer":
			if example.Tag == "!!int" {
				return true
			}
		case "number":
			if example.Tag == "!!int" || example.Tag == "!!float" {
				return true
			}
		case "boolean":
			if example.Tag == "!!bool" {
				return true
			}
		case "array":
			if example.Kind == yaml.SequenceNode {
				return true
			}
		case "object":
			if example.Kind == yaml.MappingNode {
				return true
			}
		case "null":
			if example.Tag == "!!null" {
				return true
			}
		}
	}

	return false
}

// filterExamplesBySchemaType returns only the examples that match the schema's
// declared type(s). Returns nil if no examples match.
func filterExamplesBySchemaType(examples []values.Value, schemaTypes []oas3.SchemaType) []values.Value {
	if len(examples) == 0 || len(schemaTypes) == 0 {
		return examples
	}

	var filtered []values.Value
	for _, ex := range examples {
		if exampleMatchesSchemaType(ex, schemaTypes) {
			filtered = append(filtered, ex)
		}
	}

	return filtered
}

// mergeExamples returns the parent example/examples if present, otherwise
// returns the first non-empty example or examples from the sub-schemas.
// Priority: parentExample > parentExamples > first sub-schema example > first sub-schema examples
func mergeExamples(parentExample values.Value, parentExamples []values.Value, subExamples []subSchemaExamples) (values.Value, []values.Value) {
	if parentExample != nil {
		return parentExample, nil
	}

	if len(parentExamples) > 0 {
		return nil, parentExamples
	}

	for _, sub := range subExamples {
		if sub.Example != nil {
			return sub.Example, nil
		}

		if len(sub.Examples) > 0 {
			return nil, sub.Examples
		}
	}

	return nil, nil
}
