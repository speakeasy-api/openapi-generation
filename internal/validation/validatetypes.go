package validation

import (
	"context"
	stderrors "errors"
	"strings"

	internalOpenapi "github.com/speakeasy-api/openapi-generation/v2/internal/openapi"
	"github.com/speakeasy-api/openapi/jsonschema/oas3"
	"github.com/speakeasy-api/openapi/linter"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/validation"
)

// ValidateTypes validates array-specific schema constraints for SDK generation.
// Note: Basic type validation (checking if type is "string", "integer", etc.) is handled
// by the built-in OpenAPI schema validator, so we only validate array-specific rules here.
type ValidateTypes struct{}

var _ Rule = (*ValidateTypes)(nil)

func (r *ValidateTypes) ID() string {
	return "generator-validate-types"
}

func (r *ValidateTypes) Category() string {
	return "validation"
}

func (r *ValidateTypes) Summary() string {
	return "Ensure schema types and formats are supported for generation."
}

func (r *ValidateTypes) HowToFix() string {
	return "For array schemas, provide items in OpenAPI 3.0 and avoid boolean items unless using prefixItems or contains in OpenAPI 3.1."
}

func (r *ValidateTypes) Description() string {
	return "Ensure schema data types and formats are supported for generation. Unsupported types lead to broken or unusable SDK code."
}

func (r *ValidateTypes) Link() string {
	return ""
}

func (r *ValidateTypes) DefaultSeverity() validation.Severity {
	return validation.SeverityError
}

func (r *ValidateTypes) Versions() []string {
	return nil // Applies to all versions
}

func (r *ValidateTypes) Run(ctx context.Context, docInfo *linter.DocumentInfo[*openapi.OpenAPI], config *linter.RuleConfig) []error {
	if docInfo == nil || docInfo.Document == nil || docInfo.Index == nil {
		return nil
	}

	doc := docInfo.Document
	var errors []error

	// Iterate over all schemas in the document
	for _, indexNode := range docInfo.Index.GetAllSchemas() {
		schemaRef := indexNode.Node
		if schemaRef == nil {
			continue
		}

		schema := schemaRef.GetSchema()
		if schema == nil {
			continue
		}

		// Check if this is an array type schema
		types := schema.GetType()
		isArray := false
		for _, t := range types {
			if strings.ToLower(string(t)) == "array" {
				isArray = true
				break
			}
		}

		if isArray {
			errors = append(errors, r.validateArrays(doc, schema)...)
		}
	}

	return errors
}

func (r *ValidateTypes) validateArrays(doc *openapi.OpenAPI, schema *oas3.Schema) []error {
	var errors []error

	prefixItems := schema.GetPrefixItems()
	contains := schema.GetContains()
	items := schema.GetItems()

	// Only validate when neither prefixItems nor contains are present
	if prefixItems == nil && contains == nil {
		if items != nil {
			// OAS 3.1: items as boolean is only valid when used with prefixItems
			// This is a constraint not checked by the base schema validator
			if items.IsBool() {
				itemsNode := schema.GetPropertyNode("Items")
				if itemsNode == nil {
					itemsNode = schema.GetRootNode()
				}
				errors = append(errors, &validation.Error{
					Rule:            r.ID(),
					Severity:        r.DefaultSeverity(),
					Node:            itemsNode,
					UnderlyingError: stderrors.New("`items` as `boolean` only valid when used with `prefixItems`"),
				})
			}
		} else if internalOpenapi.DetermineOpenAPIVersion(doc.GetOpenAPI()) != internalOpenapi.Version31 {
			// OAS 3.0.X: arrays must have an items attribute
			// This is required for SDK generation in OAS 3.0.X
			typeNode := schema.GetPropertyNode("Type")
			if typeNode == nil {
				typeNode = schema.GetRootNode()
			}
			errors = append(errors, &validation.Error{
				Rule:            r.ID(),
				Severity:        r.DefaultSeverity(),
				Node:            typeNode,
				UnderlyingError: stderrors.New("`items` attribute required for arrays in OpenAPI 3.0.X documents"),
			})
		}
	}

	return errors
}
