package validation

import (
	"context"
	"fmt"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi/jsonschema/oas3"
	"github.com/speakeasy-api/openapi/linter"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/validation"
	config "github.com/speakeasy-api/sdk-gen-config"
)

// CsharpOptionalNullableDecimal warns when a scalar optional+nullable `format: decimal`
// property is generated for C# in presence-aware mode. Such a field cannot be represented as
// OptionalNullable<decimal?> without precision loss (Json.NET parses the scalar number token
// as double before the converter runs), so the generator emits a plain `decimal?` instead. The
// unwrap decision itself lives in the C# templates (isScalarOptionalNullableDecimal); this rule
// only surfaces the author-facing warning.
type CsharpOptionalNullableDecimal struct {
	sdkConfig *config.Configuration
	target    types.Target
}

var _ Rule = (*CsharpOptionalNullableDecimal)(nil)
var _ configAwareRule = (*CsharpOptionalNullableDecimal)(nil)
var _ targetAwareRule = (*CsharpOptionalNullableDecimal)(nil)

func (r *CsharpOptionalNullableDecimal) SetConfig(cfg *config.Configuration) {
	r.sdkConfig = cfg
}

func (r *CsharpOptionalNullableDecimal) SetTarget(target types.Target) {
	r.target = target
}

func (r *CsharpOptionalNullableDecimal) ID() string {
	return "generator-csharp-optional-nullable-decimal"
}

func (r *CsharpOptionalNullableDecimal) Category() string {
	return "validation"
}

func (r *CsharpOptionalNullableDecimal) Summary() string {
	return "Warn when a scalar optional+nullable decimal is emitted as plain decimal? in C#."
}

func (r *CsharpOptionalNullableDecimal) HowToFix() string {
	return "Use a string-encoded decimal (`type: string, format: decimal`) to preserve both precision and presence awareness, or drop nullability/optionality."
}

func (r *CsharpOptionalNullableDecimal) Description() string {
	return "In C# presence-aware mode, a scalar `format: decimal` property that is both optional and nullable cannot be wrapped in OptionalNullable<decimal?> without silent precision loss, so the generator emits a plain nullable `decimal?` and drops the absent/null/set distinction."
}

func (r *CsharpOptionalNullableDecimal) Link() string {
	return ""
}

func (r *CsharpOptionalNullableDecimal) DefaultSeverity() validation.Severity {
	return validation.SeverityWarning
}

func (r *CsharpOptionalNullableDecimal) Versions() []string {
	return nil // Applies to all versions
}

func (r *CsharpOptionalNullableDecimal) Run(ctx context.Context, docInfo *linter.DocumentInfo[*openapi.OpenAPI], config *linter.RuleConfig) []error {
	if docInfo == nil || docInfo.Index == nil {
		return nil
	}

	// This is a C#-only concern, gated on presenceAwareJsonSerialization.
	if r.target.Template != "csharp" || !r.presenceAwareEnabled() {
		return nil
	}

	var validationErrors []error

	// Walk both component schemas and inline schemas (request/response bodies,
	// nested objects, etc.) — the template predicate applies per-field regardless
	// of where the object is defined, so the warning must too.
	for _, indexNode := range docInfo.Index.ComponentSchemas {
		// ComponentSchemas entries are always top-level `/components/schemas/<Name>`;
		// use the bare name for a clean, author-facing label.
		validationErrors = append(validationErrors, r.checkSchema(indexNode, schemaNameFromPointer(indexNode))...)
	}
	for _, indexNode := range docInfo.Index.InlineSchemas {
		// Inline schemas have no name; use the full JSON pointer to locate them.
		validationErrors = append(validationErrors, r.checkSchema(indexNode, strings.TrimPrefix(pointerString(indexNode), "/"))...)
	}

	return validationErrors
}

func pointerString(indexNode *openapi.IndexNode[*oas3.JSONSchemaReferenceable]) string {
	return string(indexNode.Location.ToJSONPointer())
}

func schemaNameFromPointer(indexNode *openapi.IndexNode[*oas3.JSONSchemaReferenceable]) string {
	ptr := pointerString(indexNode)
	if i := strings.LastIndex(ptr, "/"); i >= 0 {
		return ptr[i+1:]
	}
	return ptr
}

// checkSchema emits a warning for each scalar optional+nullable decimal property
// of the given schema node.
func (r *CsharpOptionalNullableDecimal) checkSchema(indexNode *openapi.IndexNode[*oas3.JSONSchemaReferenceable], schemaLabel string) []error {
	if indexNode == nil || indexNode.Node == nil {
		return nil
	}

	schema := indexNode.Node.GetSchema()
	if schema == nil {
		return nil
	}

	properties := schema.GetProperties()
	if properties == nil || properties.Len() == 0 {
		return nil
	}

	required := map[string]bool{}
	for _, name := range schema.GetRequired() {
		required[name] = true
	}

	var validationErrors []error

	for propertyName, propertyRef := range properties.All() {
		if propertyRef == nil {
			continue
		}

		propertySchema := propertyRef.GetSchema()
		if propertySchema == nil {
			continue
		}

		// optional: not in parent's required set.
		if required[propertyName] {
			continue
		}

		if !isScalarOptionalNullableDecimal(propertySchema) {
			continue
		}

		node := schema.GetCore().Properties.GetMapKeyNodeOrRoot(propertyName, propertySchema.GetRootNode())

		validationErrors = append(validationErrors, &validation.Error{
			Rule:     r.ID(),
			Severity: r.DefaultSeverity(),
			Node:     node,
			UnderlyingError: fmt.Errorf(
				"C#: '%s.%s' (format: decimal) cannot be both optional and nullable without precision loss: treating as plain nullable 'decimal?' instead. Consider using a string-encoded decimal: 'type: string, format: decimal' to preserve both precision and presence awareness.",
				schemaLabel,
				propertyName,
			),
		})
	}

	return validationErrors
}

// presenceAwareEnabled reports whether presenceAwareJsonSerialization is set for the C# target.
func (r *CsharpOptionalNullableDecimal) presenceAwareEnabled() bool {
	if r.sdkConfig == nil {
		return false
	}
	langCfg, ok := r.sdkConfig.Languages[r.target.Template]
	if !ok || langCfg.Cfg == nil {
		return false
	}
	enabled, _ := langCfg.Cfg["presenceAwareJsonSerialization"].(bool)
	return enabled
}

// isScalarOptionalNullableDecimal reports whether a schema is a scalar (non-array/non-map)
// `format: decimal` number that is nullable. Mirrors the template predicate of the same name:
// `type: string, format: decimal` (the escape hatch) and arrays/maps of decimal are excluded.
func isScalarOptionalNullableDecimal(schema *oas3.Schema) bool {
	if schema.GetFormat() != "decimal" {
		return false
	}

	isNumber := false
	isNullable := schema.GetNullable()
	for _, t := range schema.GetType() {
		switch t {
		case oas3.SchemaTypeNumber:
			isNumber = true
		case oas3.SchemaTypeNull:
			isNullable = true
		}
	}

	return isNumber && isNullable
}
