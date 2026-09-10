package validation

import (
	"context"
	"fmt"

	"github.com/speakeasy-api/openapi-generation/v2/internal/validation/sanitization"
	"github.com/speakeasy-api/openapi/linter"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/validation"
)

// DuplicateSchemaName checks that schema names are unique when converted to class names
type DuplicateSchemaName struct{}

var _ Rule = (*DuplicateSchemaName)(nil)

func (r *DuplicateSchemaName) ID() string {
	return "generator-duplicate-schema-name"
}

func (r *DuplicateSchemaName) Category() string {
	return "collision"
}

func (r *DuplicateSchemaName) Summary() string {
	return "Ensure no duplicate schema names collide after naming rules."
}

func (r *DuplicateSchemaName) HowToFix() string {
	return "Rename schemas so their sanitized class names are unique after naming rules."
}

func (r *DuplicateSchemaName) Description() string {
	return "Schema names must be unique when converted to class names. Collisions lead to overwritten or ambiguous SDK models."
}

func (r *DuplicateSchemaName) Link() string {
	return ""
}

func (r *DuplicateSchemaName) DefaultSeverity() validation.Severity {
	return validation.SeverityWarning
}

func (r *DuplicateSchemaName) Versions() []string {
	return nil // Applies to all versions
}

func (r *DuplicateSchemaName) Run(ctx context.Context, docInfo *linter.DocumentInfo[*openapi.OpenAPI], config *linter.RuleConfig) []error {
	if docInfo == nil || docInfo.Document == nil {
		return nil
	}

	doc := docInfo.Document
	components := doc.GetComponents()
	if components == nil {
		return nil
	}

	schemas := components.GetSchemas()
	if schemas == nil {
		return nil
	}

	componentsCore := components.GetCore()
	componentsRoot := components.GetRootNode()

	var errors []error
	uniqueSchemaNames := []nameReference{}

	// Iterate over all schemas and check for collisions
	for schemaName := range schemas.All() {
		// Get the key node for error reporting
		keyNode := componentsCore.Schemas.GetMapKeyNodeOrRoot(schemaName, componentsRoot)

		// Sanitize the schema name according to all target language rules
		sanitizedSchemaNameResult := sanitization.GetSanitizedClassNameResult(schemaName)

		// Check if this sanitized name conflicts with any existing schema
		if sanitizedSchemaName, origSchemaName, orig := findNameConflict(sanitizedSchemaNameResult, uniqueSchemaNames); orig != nil {
			errors = append(errors, &validation.Error{
				Rule:     r.ID(),
				Severity: r.DefaultSeverity(),
				Node:     keyNode,
				UnderlyingError: fmt.Errorf(
					"`%s` (`%s`) will collide with `%s` (`%s`) [line `%d`] when converted to class name",
					schemaName,
					sanitizedSchemaName,
					orig.Name,
					origSchemaName,
					orig.Line,
				),
			})
		} else {
			// No conflict, add to list of unique names
			uniqueSchemaNames = append(uniqueSchemaNames, nameReference{
				Name:   schemaName,
				Line:   keyNode.Line,
				Result: sanitizedSchemaNameResult,
			})
		}
	}

	return errors
}
