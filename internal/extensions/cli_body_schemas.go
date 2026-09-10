package extensions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi/jsonschema/oas3"
	"github.com/speakeasy-api/openapi/references"
	"gopkg.in/yaml.v3"
)

func decodeYAMLNode(node *yaml.Node) (any, error) {
	var v any
	if err := node.Decode(&v); err != nil {
		return nil, err
	}
	return normalizeJSONValue(v), nil
}

// normalizeJSONValue rewrites YAML-decoded values into JSON-marshalable
// shapes: mappings with non-string keys (YAML permits booleans, numbers, and
// dates as keys) become map[string]any with stringified keys.
func normalizeJSONValue(v any) any {
	switch typed := v.(type) {
	case map[string]any:
		for k, val := range typed {
			typed[k] = normalizeJSONValue(val)
		}
		return typed
	case map[any]any:
		out := make(map[string]any, len(typed))
		for k, val := range typed {
			out[fmt.Sprintf("%v", k)] = normalizeJSONValue(val)
		}
		return out
	case []any:
		for i, val := range typed {
			typed[i] = normalizeJSONValue(val)
		}
		return typed
	default:
		return v
	}
}

var cliComponentRefValue = regexp.MustCompile(`^#/components/schemas/([^"#/]+)((?:/[^"#]*)?)$`)

// cliSchemaDataKeywords carry payload data inside a schema, not subschemas:
// a "$ref" appearing under them (an example whose VALUE contains a $ref
// field, say) is content and must never be collected or rewritten.
var cliSchemaDataKeywords = map[string]bool{
	"example":  true,
	"examples": true,
	"default":  true,
	"const":    true,
	"enum":     true,
}

// cliSchemaMapKeywords hold maps whose VALUES are subschemas while their
// keys are arbitrary names (a property may legitimately be called
// "example").
var cliSchemaMapKeywords = map[string]bool{
	"properties":        true,
	"patternProperties": true,
	"$defs":             true,
	"definitions":       true,
	"dependentSchemas":  true,
}

// walkSchemaRefs visits every schema-position "$ref" string in a decoded
// schema value, replacing it with fn's return value. Data keywords and
// x- extensions are skipped entirely; unknown keywords are conservatively
// treated as data.
func walkSchemaRefs(schema any, fn func(ref string) string) {
	schemaMap, ok := schema.(map[string]any)
	if !ok {
		if list, ok := schema.([]any); ok {
			for _, item := range list {
				walkSchemaRefs(item, fn)
			}
		}
		return
	}
	for k, v := range schemaMap {
		switch {
		case k == "$ref":
			if ref, ok := v.(string); ok {
				schemaMap[k] = fn(ref)
			}
		case cliSchemaMapKeywords[k]:
			if m, ok := v.(map[string]any); ok {
				for _, sub := range m {
					walkSchemaRefs(sub, fn)
				}
			}
		case cliSchemaDataKeywords[k] || strings.HasPrefix(k, "x-"):
			// data: never descend
		default:
			// Applicator keywords (items, allOf, anyOf, oneOf, not,
			// if/then/else, contains, prefixItems, additionalProperties,
			// propertyNames, unevaluated*) all hold subschemas directly or
			// in arrays; descending into anything else is harmless only
			// when it really is a schema, so restrict to known shapes.
			switch k {
			case "items", "additionalProperties", "not", "if", "then", "else",
				"contains", "propertyNames", "unevaluatedItems",
				"unevaluatedProperties", "additionalItems", "contentSchema":
				walkSchemaRefs(v, fn)
			case "allOf", "anyOf", "oneOf", "prefixItems":
				walkSchemaRefs(v, fn)
			}
		}
	}
}

// cliDeepCopyJSON deep-copies a decoded JSON value so in-place reference
// rewriting cannot mutate the shared component cache across operations. The
// decoder preserves numbers verbatim (json.Number): a plain round-trip
// through float64 would silently round integers beyond 2^53.
func cliDeepCopyJSON(v any) (any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	var out any
	if err := dec.Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

func cliIsExternalRef(ref string) bool {
	return references.Reference(ref).GetURI() != ""
}

type cliExternalSchema struct {
	value    any
	docPath  string
	document any
}

type cliExternalSchemaResolver struct {
	ctx     context.Context //nolint:containedctx // resolver is request-scoped; ctx feeds schema resolution
	options references.ResolveOptions
	rootDoc any
	schemas map[string]*cliExternalSchema
	targets map[string]map[string]string
}

func newCLIExternalSchemaResolver(ctx context.Context, docInfo *document.DocumentInfo) *cliExternalSchemaResolver {
	return &cliExternalSchemaResolver{
		ctx:     ctx,
		options: docInfo.GetResolutionOptions(ctx),
		rootDoc: docInfo.Doc,
		schemas: map[string]*cliExternalSchema{},
		targets: map[string]map[string]string{},
	}
}

func (r *cliExternalSchemaResolver) rootDocPath() string {
	return r.options.TargetLocation
}

func (r *cliExternalSchemaResolver) target(docPath, ref string) (string, bool) {
	absRef, ok := r.targets[docPath][ref]
	return absRef, ok
}

func (r *cliExternalSchemaResolver) resolve(ref, docPath string, doc any) (string, error) {
	if absRef, ok := r.target(docPath, ref); ok {
		return absRef, nil
	}
	schema := oas3.NewJSONSchemaFromReference(references.Reference(ref))
	opts := r.options
	opts.TargetLocation = docPath
	opts.TargetDocument = doc
	if _, err := schema.Resolve(r.ctx, opts); err != nil {
		return "", err
	}
	info := schema.GetReferenceResolutionInfo()
	for info != nil && info.Object != nil && info.Object.IsReference() {
		next := info.Object.GetReferenceResolutionInfo()
		if next == nil || next.Object == nil {
			break
		}
		info = next
	}
	resolved := schema.GetResolvedSchema()
	if info == nil || resolved == nil {
		return "", fmt.Errorf("reference %q did not resolve", ref)
	}
	absRef := string(info.AbsoluteReference)
	if r.targets[docPath] == nil {
		r.targets[docPath] = map[string]string{}
	}
	r.targets[docPath][ref] = absRef
	if _, ok := r.schemas[absRef]; ok {
		return absRef, nil
	}
	var value any
	switch {
	case resolved.IsBool():
		value = *resolved.GetBool()
	case resolved.GetSchema() != nil && resolved.GetSchema().GetRootNode() != nil:
		decoded, err := decodeYAMLNode(resolved.GetSchema().GetRootNode())
		if err != nil {
			return "", fmt.Errorf("decode %s: %w", absRef, err)
		}
		value = decoded
	default:
		return "", fmt.Errorf("reference %q resolved to an empty schema", ref)
	}
	entry := &cliExternalSchema{value: value, docPath: info.AbsoluteDocumentPath, document: info.ResolvedDocument}
	r.schemas[absRef] = entry
	var nestedErr error
	walkSchemaRefs(value, func(nested string) string {
		if nestedErr == nil {
			if _, err := r.resolve(nested, entry.docPath, entry.document); err != nil {
				nestedErr = fmt.Errorf("%s references %q: %w", absRef, nested, err)
			}
		}
		return nested
	})
	if nestedErr != nil {
		return "", nestedErr
	}
	return absRef, nil
}

func cliExternalDefName(absRef string) string {
	ref := references.Reference(absRef)
	if pointer := string(ref.GetJSONPointer()); pointer != "" {
		pointer = strings.TrimPrefix(pointer, "/components/schemas/")
		pointer = strings.TrimPrefix(pointer, "/")
		return strings.ReplaceAll(pointer, "/", "_")
	}
	base := path.Base(ref.GetURI())
	return strings.TrimSuffix(base, path.Ext(base))
}

func cliUniqueDefName(name string, taken map[string]any) string {
	if _, exists := taken[name]; !exists {
		return name
	}
	for n := 2; ; n++ {
		candidate := fmt.Sprintf("%s_%d", name, n)
		if _, exists := taken[candidate]; !exists {
			return candidate
		}
	}
}

// CollectCLIBodySchemas extracts a self-contained JSON Schema for every
// operation's application/json request body, keyed by operationId. Each
// schema bundles its transitive component dependencies under $defs with
// rewritten references, so a CLI can print an exact, machine-readable
// request-body schema (e.g. behind a --schema flag) with zero hand-authoring.
func (e *Extensions) CollectCLIBodySchemas(ctx context.Context, docInfo *document.DocumentInfo) (map[string]string, error) {
	if docInfo == nil || docInfo.Doc == nil || docInfo.Doc.Paths == nil {
		return nil, nil
	}
	doc := docInfo.Doc
	resolver := newCLIExternalSchemaResolver(ctx, docInfo)

	// Decode every component schema to a plain value once.
	components := map[string]any{}
	if doc.Components != nil && doc.Components.Schemas != nil {
		for name, schemaRef := range doc.Components.Schemas.All() {
			if schemaRef == nil {
				continue
			}
			if schemaRef.IsBool() {
				if boolSchema := schemaRef.GetBool(); boolSchema != nil {
					components[name] = *boolSchema
				}
				continue
			}
			if schemaRef.GetSchema() == nil {
				continue
			}
			node := schemaRef.GetSchema().GetRootNode()
			if node == nil {
				continue
			}
			decoded, err := decodeYAMLNode(node)
			if err != nil {
				continue
			}
			components[name] = decoded
		}
	}

	collectRefs := func(v any) (names []string, externals []string) {
		walkSchemaRefs(v, func(ref string) string {
			if m := cliComponentRefValue.FindStringSubmatch(ref); m != nil {
				names = append(names, m[1])
			} else if cliIsExternalRef(ref) {
				externals = append(externals, ref)
			}
			return ref
		})
		return names, externals
	}

	// OpenAPI 3.1 request schemas are JSON Schema draft 2020-12; 3.0 Schema
	// Objects are a distinct dialect (nullable, boolean exclusiveMinimum), so
	// claiming 2020-12 there would misdescribe the schema. Only 3.1+ bundles
	// carry the $schema marker.
	isJSONSchemaDialect := strings.HasPrefix(doc.OpenAPI, "3.1")

	out := map[string]string{}
	addSchema := func(operationID, schema string) error {
		if _, exists := out[operationID]; exists {
			return fmt.Errorf("operationId %q is used by more than one operation; operation ids must be unique to key request body schemas", operationID)
		}
		out[operationID] = schema
		return nil
	}
	for _, pathItemRef := range doc.Paths.All() {
		if pathItemRef == nil {
			continue
		}
		pathItem := pathItemRef.GetObject()
		if pathItem == nil || !pathItem.IsInitialized() {
			continue
		}
		ignorePath, err := e.Ignore(pathItem.GetExtensions())
		if err != nil {
			return nil, err
		}
		if ignorePath {
			continue
		}
		for _, op := range pathItem.All() {
			if op == nil || op.GetOperationID() == "" {
				continue
			}
			ignoreOp, err := e.Ignore(op.GetExtensions())
			if err != nil {
				return nil, err
			}
			if ignoreOp {
				continue
			}
			opID := op.GetOperationID()
			rbRef := op.GetRequestBody()
			if rbRef == nil {
				continue
			}
			rb := rbRef.GetObject()
			if rb == nil {
				continue
			}
			content := rb.GetContent()
			if content == nil {
				continue
			}
			mediaType, ok := content.Get("application/json")
			if !ok || mediaType == nil {
				continue
			}
			schemaRef := mediaType.GetSchema()
			if schemaRef == nil {
				continue
			}
			if schemaRef.IsBool() {
				boolSchema := schemaRef.GetBool()
				if boolSchema == nil {
					continue
				}
				encoded, err := json.Marshal(*boolSchema)
				if err != nil {
					return nil, fmt.Errorf("marshal body schema for %s: %w", opID, err)
				}
				if err := addSchema(opID, string(encoded)); err != nil {
					return nil, err
				}
				continue
			}
			if schemaRef.GetSchema() == nil {
				continue
			}
			node := schemaRef.GetSchema().GetRootNode()
			if node == nil {
				continue
			}
			root, err := decodeYAMLNode(node)
			if err != nil {
				// A body schema that exists but cannot be decoded must not
				// silently lose its --schema surface.
				return nil, fmt.Errorf("decode request body schema for %s: %w", opID, err)
			}

			// Transitive closure of referenced components. A reference to a
			// missing component must fail here: deferring to the post-rewrite
			// dangling check would let a missing name that shadows a local
			// $defs definition silently rebind to that definition.
			defs := map[string]any{}
			seen := map[string]bool{}
			externalRefs := map[string]bool{}
			var addExternal func(absRef string)
			addExternal = func(absRef string) {
				if externalRefs[absRef] {
					return
				}
				externalRefs[absRef] = true
				entry := resolver.schemas[absRef]
				walkSchemaRefs(entry.value, func(ref string) string {
					if target, ok := resolver.target(entry.docPath, ref); ok {
						addExternal(target)
					}
					return ref
				})
			}
			resolveExternals := func(externals []string) error {
				for _, ref := range externals {
					absRef, err := resolver.resolve(ref, resolver.rootDocPath(), resolver.rootDoc)
					if err != nil {
						return fmt.Errorf("request body schema for %s references %q, which cannot be resolved: %w", opID, ref, err)
					}
					addExternal(absRef)
				}
				return nil
			}
			pending, externals := collectRefs(root)
			if err := resolveExternals(externals); err != nil {
				return nil, err
			}
			for len(pending) > 0 {
				name := pending[0]
				pending = pending[1:]
				if seen[name] {
					continue
				}
				seen[name] = true
				component, found := components[name]
				if !found {
					return nil, fmt.Errorf("request body schema for %s references component %q, which does not exist in components.schemas", opID, name)
				}
				defs[name] = component
				names, externals := collectRefs(component)
				pending = append(pending, names...)
				if err := resolveExternals(externals); err != nil {
					return nil, err
				}
			}

			bundled := map[string]any{}
			if isJSONSchemaDialect {
				bundled["$schema"] = "https://json-schema.org/draft/2020-12/schema"
			}
			rootMap, ok := root.(map[string]any)
			if !ok {
				// A non-object root is still a schema (an empty node means
				// "anything"); emit it directly instead of silently dropping
				// the operation's schema surface.
				value := root
				if value == nil {
					value = true
				}
				encoded, err := json.Marshal(value)
				if err != nil {
					return nil, fmt.Errorf("marshal body schema for %s: %w", opID, err)
				}
				if err := addSchema(opID, string(encoded)); err != nil {
					return nil, err
				}
				continue
			}
			for k, v := range rootMap {
				bundled[k] = v
			}

			// Preserve root-local definitions because their names and references
			// are part of the document's public schema surface. OpenAPI 3.1
			// keeps #/$defs and #/components/schemas as distinct namespaces,
			// so a component may legally share a local definition's name;
			// bundle such components under a deterministic numbered alias.
			mergedDefs := map[string]any{}
			if rootDefs, exists := rootMap["$defs"]; exists {
				localDefs, isMap := rootDefs.(map[string]any)
				if !isMap && (len(defs) > 0 || len(externalRefs) > 0) {
					return nil, fmt.Errorf("request body schema for %s has a non-object $defs, so referenced components cannot be bundled into it", opID)
				}
				if isMap {
					for name, definition := range localDefs {
						mergedDefs[name] = definition
					}
				}
			}
			renames := map[string]string{}
			names := make([]string, 0, len(defs))
			for name := range defs {
				names = append(names, name)
			}
			sort.Strings(names)
			for _, name := range names {
				target := name
				if _, exists := mergedDefs[name]; exists {
					for n := 2; ; n++ {
						candidate := fmt.Sprintf("%s_%d", name, n)
						if _, taken := mergedDefs[candidate]; taken {
							continue
						}
						if _, taken := defs[candidate]; taken {
							continue
						}
						target = candidate
						break
					}
					renames[name] = target
				}
				mergedDefs[target] = defs[name]
			}
			externalAliases := map[string]string{}
			externalDefs := map[string]*cliExternalSchema{}
			absRefs := make([]string, 0, len(externalRefs))
			for absRef := range externalRefs {
				absRefs = append(absRefs, absRef)
			}
			sort.Strings(absRefs)
			for _, absRef := range absRefs {
				alias := cliUniqueDefName(cliExternalDefName(absRef), mergedDefs)
				externalAliases[absRef] = alias
				externalDefs[alias] = resolver.schemas[absRef]
				mergedDefs[alias] = resolver.schemas[absRef].value
			}
			if len(mergedDefs) > 0 {
				bundled["$defs"] = mergedDefs
			}

			// Rewrite component references to the bundled $defs (through the
			// collision aliases) on a deep copy — the decoded components are
			// shared across operations — then verify the bundle is
			// internally complete: every reference the emitted schema makes
			// must have a definition, otherwise the printed schema would
			// dangle. Both passes walk schema positions only, so a $ref
			// field inside example payload data is left untouched.
			copied, err := cliDeepCopyJSON(bundled)
			if err != nil {
				return nil, fmt.Errorf("marshal body schema for %s: %w", opID, err)
			}
			externalAlias := func(docPath, ref string) (string, bool) {
				absRef, ok := resolver.target(docPath, ref)
				if !ok {
					return "", false
				}
				alias, ok := externalAliases[absRef]
				return alias, ok
			}
			rootRewrite := func(ref string) string {
				if m := cliComponentRefValue.FindStringSubmatch(ref); m != nil {
					name := m[1]
					if alias, ok := renames[name]; ok {
						name = alias
					}
					return "#/$defs/" + name + m[2]
				}
				if alias, ok := externalAlias(resolver.rootDocPath(), ref); ok {
					return "#/$defs/" + alias
				}
				return ref
			}
			copiedMap := copied.(map[string]any)
			copiedDefs, _ := copiedMap["$defs"].(map[string]any)
			delete(copiedMap, "$defs")
			walkSchemaRefs(copiedMap, rootRewrite)
			if copiedDefs != nil {
				copiedMap["$defs"] = copiedDefs
				for name, definition := range copiedDefs {
					entry, external := externalDefs[name]
					if !external {
						walkSchemaRefs(definition, rootRewrite)
						continue
					}
					walkSchemaRefs(definition, func(ref string) string {
						if alias, ok := externalAlias(entry.docPath, ref); ok {
							return "#/$defs/" + alias
						}
						return ref
					})
				}
			}
			var dangling error
			walkSchemaRefs(copied, func(ref string) string {
				if rest, ok := strings.CutPrefix(ref, "#/$defs/"); ok && dangling == nil {
					name, _, _ := strings.Cut(rest, "/")
					if _, ok := mergedDefs[name]; !ok {
						dangling = fmt.Errorf("request body schema for %s references component %q, which does not exist in components.schemas", opID, name)
					}
				}
				return ref
			})
			if dangling != nil {
				return nil, dangling
			}
			encoded, err := json.Marshal(copied)
			if err != nil {
				return nil, fmt.Errorf("marshal body schema for %s: %w", opID, err)
			}
			if err := addSchema(opID, string(encoded)); err != nil {
				return nil, err
			}
		}
	}
	if len(out) == 0 {
		return nil, nil
	}
	return out, nil
}
