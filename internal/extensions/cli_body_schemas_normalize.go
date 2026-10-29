package extensions

import (
	"strings"
)

// walkSchemaObjects calls fn on every schema object in a decoded schema,
// parents before children, then descends into the (possibly rewritten)
// object. It follows the same keyword rules as walkSchemaRefs.
func walkSchemaObjects(schema any, fn func(map[string]any)) {
	switch typed := schema.(type) {
	case []any:
		for _, item := range typed {
			walkSchemaObjects(item, fn)
		}
		return
	case map[string]any:
		fn(typed)
		for k, v := range typed {
			switch {
			case cliSchemaMapKeywords[k]:
				if m, ok := v.(map[string]any); ok {
					for _, sub := range m {
						walkSchemaObjects(sub, fn)
					}
				}
			case cliSchemaDataKeywords[k] || strings.HasPrefix(k, "x-"):
			default:
				switch k {
				case "items", "additionalProperties", "not", "if", "then", "else",
					"contains", "propertyNames", "unevaluatedItems",
					"unevaluatedProperties", "additionalItems", "contentSchema",
					"allOf", "anyOf", "oneOf", "prefixItems":
					walkSchemaObjects(v, fn)
				}
			}
		}
	}
}

// cliSchemaAnnotationKeywords stay on the outer schema when a nullable schema
// without a type is wrapped in anyOf.
var cliSchemaAnnotationKeywords = map[string]bool{
	"title":       true,
	"description": true,
	"default":     true,
	"example":     true,
	"examples":    true,
	"deprecated":  true,
	"readOnly":    true,
	"writeOnly":   true,
}

// convertOpenAPI30Schema rewrites the OpenAPI 3.0 keywords whose meaning
// differs in JSON Schema 2020-12: nullable and boolean exclusive bounds.
func convertOpenAPI30Schema(schema map[string]any) {
	convertExclusiveBound(schema, "exclusiveMinimum", "minimum")
	convertExclusiveBound(schema, "exclusiveMaximum", "maximum")

	nullable, _ := schema["nullable"].(bool)
	delete(schema, "nullable")
	if !nullable {
		return
	}
	if enum, ok := schema["enum"].([]any); ok && !containsNil(enum) {
		schema["enum"] = append(enum, nil)
	}
	switch typ := schema["type"].(type) {
	case string:
		if typ != "null" {
			schema["type"] = []any{typ, "null"}
		}
		return
	case []any:
		if !containsValue(typ, "null") {
			schema["type"] = append(typ, "null")
		}
		return
	}
	inner := map[string]any{}
	for k, v := range schema {
		if cliSchemaAnnotationKeywords[k] || strings.HasPrefix(k, "x-") {
			continue
		}
		inner[k] = v
		delete(schema, k)
	}
	if len(inner) == 0 {
		return
	}
	schema["anyOf"] = []any{inner, map[string]any{"type": "null"}}
}

func convertExclusiveBound(schema map[string]any, exclusiveKey, boundKey string) {
	exclusive, ok := schema[exclusiveKey].(bool)
	if !ok {
		return
	}
	delete(schema, exclusiveKey)
	if bound, hasBound := schema[boundKey]; exclusive && hasBound {
		schema[exclusiveKey] = bound
		delete(schema, boundKey)
	}
}

// stripReadOnlyProperties removes readOnly properties, and their required
// entries, from every object schema in a bundle. A property is readOnly when
// it says so itself or references a bundled definition that does.
func stripReadOnlyProperties(bundle any) {
	root, _ := bundle.(map[string]any)
	defs, _ := root["$defs"].(map[string]any)
	isReadOnly := func(prop any) bool {
		m, ok := prop.(map[string]any)
		if !ok {
			return false
		}
		if readOnly, _ := m["readOnly"].(bool); readOnly {
			return true
		}
		ref, _ := m["$ref"].(string)
		name, found := strings.CutPrefix(ref, "#/$defs/")
		if !found || strings.Contains(name, "/") {
			return false
		}
		def, _ := defs[name].(map[string]any)
		readOnly, _ := def["readOnly"].(bool)
		return readOnly
	}
	walkSchemaObjects(bundle, func(schema map[string]any) {
		props, ok := schema["properties"].(map[string]any)
		if !ok {
			return
		}
		removed := map[string]bool{}
		for name, prop := range props {
			if isReadOnly(prop) {
				removed[name] = true
				delete(props, name)
			}
		}
		if len(removed) == 0 {
			return
		}
		required, ok := schema["required"].([]any)
		if !ok {
			return
		}
		kept := make([]any, 0, len(required))
		for _, name := range required {
			if s, isString := name.(string); isString && removed[s] {
				continue
			}
			kept = append(kept, name)
		}
		if len(kept) == 0 {
			delete(schema, "required")
		} else {
			schema["required"] = kept
		}
	})
}

func containsNil(values []any) bool {
	for _, v := range values {
		if v == nil {
			return true
		}
	}
	return false
}

func containsValue(values []any, want string) bool {
	for _, v := range values {
		if s, ok := v.(string); ok && s == want {
			return true
		}
	}
	return false
}
