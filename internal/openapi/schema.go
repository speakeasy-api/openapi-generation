package openapi

import (
	"context"
	"fmt"
	"iter"
	"reflect"
	"slices"

	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi-generation/v2/internal/resolution"
	"github.com/speakeasy-api/openapi/extensions"
	"github.com/speakeasy-api/openapi/hashing"
	"github.com/speakeasy-api/openapi/jsonschema/oas3"
	"github.com/speakeasy-api/openapi/yml"
	"gopkg.in/yaml.v3"
)

const xReferenceOverride = "x-speakeasy-reference-override"

func IsReferenceForCircularReferences(schema *oas3.JSONSchema[oas3.Referenceable]) bool {
	return RefForCircularReferences(schema) != ""
}

func RefForCircularReferences(schema *oas3.JSONSchema[oas3.Referenceable]) string {
	// First check if there's a reference override extension in the resolved schema
	resolved := schema.GetResolvedSchema()
	if resolved != nil && resolved.IsSchema() {
		ext, ok := resolved.GetSchema().GetExtensions().Get(xReferenceOverride)
		if ok {
			return ext.Value
		}
	}

	// Always return the reference if it exists, regardless of resolved schema state
	return schema.GetRef().String()
}

func TrackRefForCircularReferences(schema *oas3.Schema, ref string) {
	if schema.Extensions == nil {
		schema.Extensions = extensions.New()
	}

	schema.GetExtensions().Set(xReferenceOverride, yml.CreateStringNode(ref))
}

func ClearRefForCircularReferences(schema *oas3.Schema) {
	if schema.Extensions == nil {
		return
	}

	schema.GetExtensions().Delete(xReferenceOverride)
}

func GetTypePropertyNode(schema *oas3.JSONSchema[oas3.Concrete]) *yaml.Node {
	if schema.IsBool() {
		return schema.GetRootNode()
	}

	return schema.GetSchema().GetCore().Type.GetKeyNodeOrRoot(schema.GetSchema().GetRootNode())
}

func IsComplex(ctx context.Context, schema *oas3.JSONSchema[oas3.Concrete], docInfo *document.DocumentInfo) (bool, error) {
	if schema.IsBool() {
		return false, nil
	}

	typ, subTypes, _ := GetResolvedType(ctx, schema)
	if len(subTypes) > 0 {
		return true, nil
	}

	switch typ {
	case "allOf":
		return areSubSchemasComplex(ctx, schema.GetSchema().GetAllOf(), docInfo)
	case "oneOf":
		return areSubSchemasComplex(ctx, schema.GetSchema().GetOneOf(), docInfo)
	case "anyOf":
		return areSubSchemasComplex(ctx, schema.GetSchema().GetAnyOf(), docInfo)
	case "object":
		return true, nil
	case "map":
		fallthrough
	case "array":
		return areSubSchemasComplex(ctx, []*oas3.JSONSchema[oas3.Referenceable]{schema.GetSchema().GetItems()}, docInfo)
	default:
		return false, nil
	}
}

func GetResolvedType(ctx context.Context, schema *oas3.JSONSchema[oas3.Concrete]) (string, []string, bool) {
	if schema.IsBool() {
		return "any", nil, false
	}

	s := schema.GetSchema()

	nullable := s.GetNullable()

	subTypes := []string{}

	for _, t := range s.GetType() {
		switch t {
		case oas3.SchemaTypeNull:
			nullable = true
		case "": // treat empty type as any
			subTypes = append(subTypes, "any")
		default:
			subTypes = append(subTypes, string(t))
		}
	}

	if len(s.AllOf) > 0 {
		return "allOf", nil, nullable
	}

	if len(s.AnyOf) > 0 {
		return "anyOf", nil, nullable
	}

	if len(s.OneOf) > 0 {
		return "oneOf", nil, nullable
	}

	if len(subTypes) == 0 {
		// If we have no types but its a enum we will handle as a string enum
		if len(s.Enum) > 0 {
			return "string", nil, nullable
		}

		// If we have a const value, infer the type from it (matching old schema builder logic)
		if s.Const != nil {
			var constValue any
			if err := s.Const.Decode(&constValue); err == nil && constValue != nil {
				switch constValue.(type) {
				case bool:
					return "boolean", nil, nullable
				case float64:
					return "number", nil, nullable
				case string:
					return "string", nil, nullable
				case int:
					return "integer", nil, nullable
				}
			}
		}

		if s.Properties.Len() > 0 {
			// Missing type but found properties treating as object
			return "object", nil, nullable
		}

		if s.AdditionalProperties != nil {
			// Missing type but found additionalProperties treating as object
			return "object", nil, nullable
		}

		if s.Items != nil {
			// Missing type but found items treating as array
			return "array", nil, nullable
		}

		// Likely a `type: 'null'` or no type set treating as any
		return "any", nil, nullable
	}

	if len(subTypes) == 1 {
		return subTypes[0], nil, nullable
	}

	return "oneOf", subTypes, nullable
}

var comparisonIgnoreFields = []string{
	"schema",
	"ref",
	"refForCircularReferences",
}

type speakeasyExtensions interface {
	IsExtensionMergable(extName string) bool
	IsExtensionIdentifying(extName string) bool
}

func IsEmpty(schema *oas3.Schema, onlyConsiderTypeModifyingFields bool, e speakeasyExtensions) bool {
	if schema == nil {
		return true
	}

	sType := reflect.TypeOf(schema).Elem()

	typeModifyingFields := []string{
		"AdditionalItems",
		"AdditionalProperties",
		"AdditionalPropertiesAllowed",
		"AllOf",
		"AnyOf",
		"Enum",
		"Items",
		"OneOf",
		"Nullable",
		"PrefixItems",
		"Properties",
		"ReadOnly",
		"Required",
		"Type",
		"UnevaluatedItems",
		"UnevaluatedProperties",
		"WriteOnly",
	}
	skipFields := []string{
		"Extensions",
	}

	for extension := range schema.GetExtensions().Keys() {
		// It is not empty if we have an identifying extension (one that affects schema identity)
		if e.IsExtensionIdentifying(extension) {
			return false
		}
	}

	for i := 0; i < sType.NumField(); i++ {
		field := sType.Field(i)
		fieldVal := reflect.Indirect(reflect.ValueOf(schema)).Field(i)

		if field.Anonymous || !field.IsExported() || slices.Contains(comparisonIgnoreFields, field.Name) {
			continue
		}

		if onlyConsiderTypeModifyingFields && !slices.Contains(typeModifyingFields, field.Name) || slices.Contains(skipFields, field.Name) {
			continue
		}

		switch field.Type.Kind() {
		case reflect.Pointer:
			if !fieldVal.IsNil() {
				v := fieldVal.Interface()

				if m, ok := v.(sequencedMap); ok {
					if m.Len() > 0 {
						return false
					}
				} else {
					return false
				}
			}
		case reflect.Slice:
			if fieldVal.Len() > 0 {
				return false
			}
		case reflect.Map:
			if fieldVal.Len() > 0 {
				return false
			}
		case reflect.Struct:
			if !fieldVal.IsZero() {
				return false
			}
		case reflect.String:
			if fieldVal.String() != "" {
				return false
			}
		case reflect.Bool:
			if fieldVal.Bool() {
				return false
			}
		case reflect.Float64:
			if fieldVal.Float() != 0 {
				return false
			}
		case reflect.Int64:
			if fieldVal.Int() != 0 {
				return false
			}
		case reflect.Interface:
			if !fieldVal.IsNil() {
				return false
			}
		default:
			panic(fmt.Sprintf("unknown type %s", field.Type.Kind()))
		}
	}

	return true
}

// HoistAllOfExtensions collects extensions from a schema and its allOf children.
// The schema's own extensions are always collected. The childPredicate filters
// which extensions are hoisted from allOf children (including through $refs).
func HoistAllOfExtensions(ctx context.Context, s *oas3.JSONSchema[oas3.Referenceable], docInfo *document.DocumentInfo, visited map[string]bool, childPredicate func(string) bool) *extensions.Extensions {
	var exts *extensions.Extensions

	schema, err := resolution.Resolve(ctx, s, docInfo)
	if err != nil {
		return exts
	}

	// Top-level extensions are collected unconditionally.
	if schema.GetExtensions() != nil {
		exts = extensions.New()
		for name, ext := range schema.GetExtensions().All() {
			exts.Set(name, ext)
		}
	}

	if schema.IsBool() || len(schema.GetSchema().GetAllOf()) == 0 {
		return exts
	}

	curHash := hashing.Hash(schema)
	if visited == nil {
		visited = map[string]bool{}
	}

	if _, ok := visited[curHash]; ok {
		// We've already visited this schema, so we can skip it
		return exts
	}
	visited[curHash] = true

	for _, child := range schema.GetSchema().GetAllOf() {
		childExtensions := HoistAllOfExtensions(ctx, child, docInfo, visited, childPredicate)
		if childExtensions != nil {
			if exts == nil {
				exts = extensions.New()
			}

			for k, v := range childExtensions.All() {
				if childPredicate(k) {
					exts.Set(k, v)
				}
			}
		}
	}

	return exts
}

func areSubSchemasComplex(ctx context.Context, subSchemas []*oas3.JSONSchema[oas3.Referenceable], docInfo *document.DocumentInfo) (bool, error) {
	for _, ss := range subSchemas {
		if ss == nil {
			continue
		}

		schema, err := resolution.Resolve(ctx, ss, docInfo)
		if err != nil {
			return false, err
		}

		isComplex, err := IsComplex(ctx, schema, docInfo)
		if err != nil {
			return false, err
		}

		if isComplex {
			return true, nil
		}
	}

	return false, nil
}

type sequencedMap interface {
	Len() int
	AllUntyped() iter.Seq2[any, any]
	GetAny(any) (any, bool)
}
