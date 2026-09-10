package validation

import (
	"context"
	"errors"
	"fmt"

	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/speakeasy-api/openapi/jsonschema/oas3"
	"github.com/speakeasy-api/openapi/linter"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/validation"
)

// ValidateBase64InputMode validates the x-speakeasy-base64-input-mode extension.
// The extension is opt-in and only meaningful on string schemas that already
// carry format:byte or contentEncoding:base64. This rule warns (never blocks)
// when the extension is misapplied so spec authors get clear feedback while
// generation continues to produce a usable SDK (the application path treats
// invalid usage as a no-op).
type ValidateBase64InputMode struct{}

var _ Rule = (*ValidateBase64InputMode)(nil)

func (r *ValidateBase64InputMode) ID() string {
	return "generator-validate-base64-input-mode"
}

func (r *ValidateBase64InputMode) Category() string {
	return "validation"
}

func (r *ValidateBase64InputMode) Summary() string {
	return "Validate x-speakeasy-base64-input-mode usage on schemas."
}

func (r *ValidateBase64InputMode) HowToFix() string {
	return "Apply x-speakeasy-base64-input-mode only to type:string schemas that also set format:byte or contentEncoding:base64, and use the value \"file\"."
}

func (r *ValidateBase64InputMode) Description() string {
	return "Ensure the x-speakeasy-base64-input-mode extension is applied to string schemas with format:byte or contentEncoding:base64 and uses a supported mode value."
}

func (r *ValidateBase64InputMode) Link() string {
	return ""
}

func (r *ValidateBase64InputMode) DefaultSeverity() validation.Severity {
	return validation.SeverityWarning
}

func (r *ValidateBase64InputMode) Versions() []string {
	return nil
}

func (r *ValidateBase64InputMode) Run(_ context.Context, docInfo *linter.DocumentInfo[*openapi.OpenAPI], _ *linter.RuleConfig) []error {
	if docInfo == nil || docInfo.Index == nil {
		return nil
	}

	const extName = "x-speakeasy-base64-input-mode"
	allowedModes := map[string]bool{extensions.Base64InputModeFile: true}

	var validationErrors []error

	for _, indexNode := range docInfo.Index.GetAllSchemas() {
		schemaRef := indexNode.Node
		if schemaRef == nil {
			continue
		}
		schema := schemaRef.GetSchema()
		if schema == nil {
			continue
		}

		ext, ok := schema.GetExtensions().Get(extName)
		if !ok || ext == nil {
			continue
		}

		if !allowedModes[ext.Value] {
			validationErrors = append(validationErrors, &validation.Error{
				Rule:            r.ID(),
				Severity:        r.DefaultSeverity(),
				Node:            ext,
				UnderlyingError: fmt.Errorf("%s value %q is not allowed (supported: %s)", extName, ext.Value, extensions.Base64InputModeFile),
			})
			continue
		}

		nonNullTypes := schemaNonNullTypes(schema.GetType())
		hasStringType := len(nonNullTypes) == 1 && nonNullTypes[0] == "string"
		if !hasStringType {
			typeNode := schema.GetPropertyNode("Type")
			if typeNode == nil {
				typeNode = ext
			}
			validationErrors = append(validationErrors, &validation.Error{
				Rule:            r.ID(),
				Severity:        r.DefaultSeverity(),
				Node:            typeNode,
				UnderlyingError: errors.New(extName + " requires a type:string schema; extension will be ignored"),
			})
			continue
		}

		if schema.GetFormat() != "byte" && schema.GetContentEncoding() != "base64" {
			validationErrors = append(validationErrors, &validation.Error{
				Rule:            r.ID(),
				Severity:        r.DefaultSeverity(),
				Node:            ext,
				UnderlyingError: errors.New(extName + " requires format:byte or contentEncoding:base64; extension will be ignored"),
			})
		}
	}

	return validationErrors
}

func schemaNonNullTypes(typeNodes []oas3.SchemaType) []string {
	var nonNull []string
	for _, t := range typeNodes {
		s := string(t)
		if s == "" || s == "null" {
			continue
		}
		nonNull = append(nonNull, s)
	}
	return nonNull
}
