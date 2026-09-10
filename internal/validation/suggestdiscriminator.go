package validation

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/speakeasy-api/openapi/jsonschema/oas3"
	"github.com/speakeasy-api/openapi/linter"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/sequencedmap"
	"github.com/speakeasy-api/openapi/validation"
	"gopkg.in/yaml.v3"
)

// SuggestDiscriminator flags oneOf/anyOf unions of object schemas whose members are distinguished by
// a shared property with distinct constant values, but that have no explicit discriminator. Adding one
// lets the SDK select the variant directly instead of attempting to deserialize against each member.
type SuggestDiscriminator struct{}

var _ Rule = (*SuggestDiscriminator)(nil)

func (r *SuggestDiscriminator) ID() string {
	return "generator-suggest-discriminator"
}

func (r *SuggestDiscriminator) Category() string {
	return "validation"
}

func (r *SuggestDiscriminator) Summary() string {
	return "Suggest a discriminator for object unions distinguished by a constant property."
}

func (r *SuggestDiscriminator) HowToFix() string {
	return "Add `discriminator: { propertyName: <property> }` to the union so the SDK can select the matching variant directly."
}

func (r *SuggestDiscriminator) Description() string {
	return "A oneOf/anyOf union whose object members are distinguished by a shared property with distinct constant values can declare a discriminator. Without one, SDKs fall back to attempting deserialization against each member in turn, which is slower and more error prone."
}

func (r *SuggestDiscriminator) Link() string {
	return ""
}

func (r *SuggestDiscriminator) DefaultSeverity() validation.Severity {
	return validation.SeverityHint
}

func (r *SuggestDiscriminator) Versions() []string {
	return nil // Applies to all versions
}

func (r *SuggestDiscriminator) Run(ctx context.Context, docInfo *linter.DocumentInfo[*openapi.OpenAPI], config *linter.RuleConfig) []error {
	if docInfo == nil || docInfo.Index == nil {
		return nil
	}

	resolve := componentSchemaResolver(docInfo.Document)

	var validationErrors []error

	for _, indexNode := range docInfo.Index.GetAllSchemas() {
		if indexNode == nil || indexNode.Node == nil {
			continue
		}

		schema := indexNode.Node.GetSchema()
		if schema == nil || schema.GetDiscriminator() != nil {
			continue
		}

		members := schema.GetOneOf()
		nodeKey := "oneOf"
		if len(members) == 0 {
			members = schema.GetAnyOf()
			nodeKey = "anyOf"
		}
		if len(members) < 2 {
			continue
		}

		property, requiredOnAll := discriminatingProperty(members, resolve)
		if property == "" {
			continue
		}

		reportNode := findChildNode(schema.GetRootNode(), nodeKey)
		if reportNode == nil {
			reportNode = schema.GetRootNode()
		}

		message := fmt.Sprintf(
			"`%s` members are distinguished by distinct `%s` values; add `discriminator: { propertyName: %s }` so the SDK can select the variant directly",
			nodeKey,
			property,
			property,
		)
		if !requiredOnAll {
			message += fmt.Sprintf(" and mark `%s` as required on each member so it is always present to discriminate on", property)
		}

		validationErrors = append(validationErrors, &validation.Error{
			Rule:            r.ID(),
			Severity:        r.DefaultSeverity(),
			Node:            reportNode,
			UnderlyingError: errors.New(message),
		})
	}

	return validationErrors
}

// discriminatingProperty returns a property name that is present on every member with a distinct
// constant value, or "" if none qualifies. A `type` property is preferred when it qualifies. The
// second result reports whether that property is already required on every member; when false the
// caller should also advise marking it required, since a discriminator property must be present to
// select a variant.
func discriminatingProperty(members []*oas3.JSONSchemaReferenceable, resolve func(*oas3.JSONSchemaReferenceable) *oas3.Schema) (string, bool) {
	memberConsts := make([]map[string]string, 0, len(members))
	memberRequired := make([]map[string]bool, 0, len(members))

	for _, member := range members {
		schema := resolve(member)
		if schema == nil {
			return "", false
		}

		properties := schema.GetProperties()
		if properties == nil || properties.Len() == 0 {
			return "", false
		}

		consts := map[string]string{}
		for propertyName, propertyRef := range properties.All() {
			if value, ok := singleConstValue(resolve(propertyRef)); ok {
				consts[propertyName] = value
			}
		}
		memberConsts = append(memberConsts, consts)

		required := map[string]bool{}
		for _, propertyName := range schema.GetRequired() {
			required[propertyName] = true
		}
		memberRequired = append(memberRequired, required)
	}

	if qualifiesAsDiscriminator("type", memberConsts) {
		return "type", requiredOnEveryMember("type", memberRequired)
	}

	candidates := make([]string, 0, len(memberConsts[0]))
	for propertyName := range memberConsts[0] {
		if propertyName != "type" && qualifiesAsDiscriminator(propertyName, memberConsts) {
			candidates = append(candidates, propertyName)
		}
	}
	if len(candidates) == 0 {
		return "", false
	}

	sort.Strings(candidates)
	return candidates[0], requiredOnEveryMember(candidates[0], memberRequired)
}

// requiredOnEveryMember reports whether the property is listed as required on every member.
func requiredOnEveryMember(propertyName string, memberRequired []map[string]bool) bool {
	for _, required := range memberRequired {
		if !required[propertyName] {
			return false
		}
	}
	return true
}

// qualifiesAsDiscriminator reports whether the property has a distinct constant value on every member.
func qualifiesAsDiscriminator(propertyName string, memberConsts []map[string]string) bool {
	seen := map[string]bool{}

	for _, consts := range memberConsts {
		value, ok := consts[propertyName]
		if !ok || seen[value] {
			return false
		}
		seen[value] = true
	}

	return len(seen) == len(memberConsts)
}

// singleConstValue returns the single constant value a schema constrains itself to via `const` or a
// single-value `enum`, or ("", false) if it is not constant.
func singleConstValue(schema *oas3.Schema) (string, bool) {
	if schema == nil {
		return "", false
	}

	if c := schema.GetConst(); c != nil && c.Kind == yaml.ScalarNode {
		return c.Value, true
	}

	if enum := schema.GetEnum(); len(enum) == 1 && enum[0] != nil && enum[0].Kind == yaml.ScalarNode {
		return enum[0].Value, true
	}

	return "", false
}

// componentSchemaResolver returns a function that resolves a (possibly referenced) schema to its
// concrete schema, following local `#/components/schemas/...` references. External references resolve
// to nil. The resolver is reused across members so the components lookup is built once.
func componentSchemaResolver(doc *openapi.OpenAPI) func(*oas3.JSONSchemaReferenceable) *oas3.Schema {
	var components *sequencedmap.Map[string, *oas3.JSONSchemaReferenceable]
	if doc != nil {
		if c := doc.GetComponents(); c != nil {
			components = c.GetSchemas()
		}
	}

	var resolve func(ref *oas3.JSONSchemaReferenceable, depth int) *oas3.Schema
	resolve = func(ref *oas3.JSONSchemaReferenceable, depth int) *oas3.Schema {
		if ref == nil || depth > 10 {
			return nil
		}

		if !ref.IsReference() {
			return ref.GetSchema()
		}

		name := localComponentName(ref.GetReference().String())
		if name == "" || components == nil {
			return nil
		}

		target, ok := components.Get(name)
		if !ok || target == ref {
			return nil
		}

		return resolve(target, depth+1)
	}

	return func(ref *oas3.JSONSchemaReferenceable) *oas3.Schema {
		return resolve(ref, 0)
	}
}

// localComponentName extracts the schema name from a local component reference, or "" otherwise.
func localComponentName(ref string) string {
	const prefix = "#/components/schemas/"
	if strings.HasPrefix(ref, prefix) {
		return strings.TrimPrefix(ref, prefix)
	}
	return ""
}

// findChildNode returns the value node for the given key in a YAML mapping node, or nil if absent.
func findChildNode(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Content == nil {
		return nil
	}

	for i := 0; i < len(node.Content)-1; i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}

	return nil
}
