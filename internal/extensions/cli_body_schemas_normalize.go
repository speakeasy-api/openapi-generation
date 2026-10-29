package extensions

import (
	"strconv"
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
// without a type is wrapped in anyOf. Definition containers stay there too so
// that #/$defs references keep resolving against the bundle root.
var cliSchemaAnnotationKeywords = map[string]bool{
	"$schema":     true,
	"$id":         true,
	"$comment":    true,
	"$defs":       true,
	"definitions": true,
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
// it says so itself, references a schema that is, is an allOf with a readOnly
// member, or is an anyOf/oneOf whose non-null members are all readOnly.
// Required entries are removed wherever the object, its allOf members or the
// schemas they reference declare the property readOnly.
func stripReadOnlyProperties(bundle any) {
	s := readOnlyStripper{root: bundle, resolving: map[string]bool{}}
	walkSchemaObjects(bundle, func(schema map[string]any) {
		names := map[string]bool{}
		s.collectReadOnlyNames(schema, names)
		if len(names) == 0 {
			return
		}
		removeRequired(schema, names)
		if members, ok := schema["allOf"].([]any); ok {
			for _, member := range members {
				if m, isMap := member.(map[string]any); isMap {
					removeRequired(m, names)
				}
			}
		}
	})
	walkSchemaObjects(bundle, func(schema map[string]any) {
		props, ok := schema["properties"].(map[string]any)
		if !ok {
			return
		}
		for name, prop := range props {
			if s.isReadOnly(prop) {
				delete(props, name)
			}
		}
	})
}

type readOnlyStripper struct {
	root      any
	resolving map[string]bool
}

func (s *readOnlyStripper) withRef(schema map[string]any, fn func(target any)) {
	ref, _ := schema["$ref"].(string)
	if ref == "" || s.resolving[ref] {
		return
	}
	target, ok := resolveLocalPointer(s.root, ref)
	if !ok {
		return
	}
	s.resolving[ref] = true
	defer delete(s.resolving, ref)
	fn(target)
}

func (s *readOnlyStripper) isReadOnly(schema any) bool {
	m, ok := schema.(map[string]any)
	if !ok {
		return false
	}
	if readOnly, _ := m["readOnly"].(bool); readOnly {
		return true
	}
	result := false
	s.withRef(m, func(target any) { result = s.isReadOnly(target) })
	if result {
		return true
	}
	if members, ok := m["allOf"].([]any); ok {
		for _, member := range members {
			if s.isReadOnly(member) {
				return true
			}
		}
	}
	for _, key := range []string{"anyOf", "oneOf"} {
		members, ok := m[key].([]any)
		if !ok {
			continue
		}
		matched := 0
		for _, member := range members {
			if isNullOnlySchema(member) {
				continue
			}
			if !s.isReadOnly(member) {
				matched = -1
				break
			}
			matched++
		}
		if matched > 0 {
			return true
		}
	}
	return false
}

func (s *readOnlyStripper) collectReadOnlyNames(schema any, names map[string]bool) {
	m, ok := schema.(map[string]any)
	if !ok {
		return
	}
	if props, ok := m["properties"].(map[string]any); ok {
		for name, prop := range props {
			if s.isReadOnly(prop) {
				names[name] = true
			}
		}
	}
	s.withRef(m, func(target any) { s.collectReadOnlyNames(target, names) })
	if members, ok := m["allOf"].([]any); ok {
		for _, member := range members {
			s.collectReadOnlyNames(member, names)
		}
	}
}

func removeRequired(schema map[string]any, names map[string]bool) {
	required, ok := schema["required"].([]any)
	if !ok {
		return
	}
	kept := make([]any, 0, len(required))
	for _, name := range required {
		if str, isString := name.(string); isString && names[str] {
			continue
		}
		kept = append(kept, name)
	}
	if len(kept) == len(required) {
		return
	}
	if len(kept) == 0 {
		delete(schema, "required")
	} else {
		schema["required"] = kept
	}
}

func isNullOnlySchema(schema any) bool {
	m, ok := schema.(map[string]any)
	if !ok {
		return false
	}
	typ, _ := m["type"].(string)
	return typ == "null"
}

// resolveLocalPointer resolves a same-document JSON Pointer reference such as
// "#/$defs/Wrapper/properties/inner" against root.
func resolveLocalPointer(root any, ref string) (any, bool) {
	pointer, ok := strings.CutPrefix(ref, "#")
	if !ok {
		return nil, false
	}
	if pointer == "" {
		return root, true
	}
	if !strings.HasPrefix(pointer, "/") {
		return nil, false
	}
	current := root
	for _, token := range strings.Split(pointer[1:], "/") {
		token = strings.ReplaceAll(strings.ReplaceAll(token, "~1", "/"), "~0", "~")
		switch typed := current.(type) {
		case map[string]any:
			next, exists := typed[token]
			if !exists {
				return nil, false
			}
			current = next
		case []any:
			index, err := strconv.Atoi(token)
			if err != nil || index < 0 || index >= len(typed) {
				return nil, false
			}
			current = typed[index]
		default:
			return nil, false
		}
	}
	return current, true
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
