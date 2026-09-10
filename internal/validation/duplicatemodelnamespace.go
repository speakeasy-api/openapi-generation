package validation

import (
	"context"
	"fmt"

	ext "github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/speakeasy-api/openapi-generation/v2/internal/validation/sanitization"
	"github.com/speakeasy-api/openapi/linter"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/validation"
)

// DuplicateModelNamespace checks that x-speakeasy-model-namespace values are unique when
// converted to folder/package names according to target language naming conventions.
type DuplicateModelNamespace struct{}

var _ Rule = (*DuplicateModelNamespace)(nil)

func (r *DuplicateModelNamespace) ID() string {
	return "generator-duplicate-model-namespace"
}

func (r *DuplicateModelNamespace) Category() string {
	return "collision"
}

func (r *DuplicateModelNamespace) Summary() string {
	return "Ensure no duplicate model namespace names collide after naming rules."
}

func (r *DuplicateModelNamespace) HowToFix() string {
	return "Rename x-speakeasy-model-namespace values so they remain unique after language-specific naming conventions are applied."
}

func (r *DuplicateModelNamespace) Description() string {
	return "Model namespace values (x-speakeasy-model-namespace) must be unique when converted to folder or package names. " +
		"Different namespace values that only differ by casing (e.g., federation_lookup vs federationLookup) " +
		"can collide when sanitized for a target language, leading to schemas being placed in the same namespace folder."
}

func (r *DuplicateModelNamespace) Link() string {
	return ""
}

func (r *DuplicateModelNamespace) DefaultSeverity() validation.Severity {
	return validation.SeverityError
}

func (r *DuplicateModelNamespace) Versions() []string {
	return nil // Applies to all versions
}

func (r *DuplicateModelNamespace) Run(ctx context.Context, docInfo *linter.DocumentInfo[*openapi.OpenAPI], config *linter.RuleConfig) []error {
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
	uniqueNamespaces := []nameReference{}

	// Iterate over all schemas and collect unique namespace values
	for schemaName, schema := range schemas.All() {
		if schema == nil {
			continue
		}

		extensions := schema.GetExtensions()
		if extensions == nil || extensions.Len() == 0 {
			continue
		}

		namespaceNode, ok := extensions.Get(ext.ExtModelNamespace.Name())
		if !ok || namespaceNode == nil {
			continue
		}

		namespace := namespaceNode.Value
		if namespace == "" {
			continue
		}

		// Use the schema key node for error reporting
		keyNode := componentsCore.Schemas.GetMapKeyNodeOrRoot(schemaName, componentsRoot)

		// Sanitize the namespace value according to all target language rules
		sanitizedResult := sanitization.GetSanitizedNamespaceResult(namespace)

		// Check if this sanitized namespace conflicts with any existing namespace
		if sanitizedName, origName, orig := findNameConflict(sanitizedResult, uniqueNamespaces); orig != nil {
			// Only report if the original namespace values are different
			// (same namespace value is fine - that's how schemas share a namespace)
			if orig.Name != namespace {
				errors = append(errors, &validation.Error{
					Rule:     r.ID(),
					Severity: r.DefaultSeverity(),
					Node:     keyNode,
					UnderlyingError: fmt.Errorf(
						"model namespace `%s` (`%s`) will collide with `%s` (`%s`) [line `%d`] when converted to folder/package name",
						namespace,
						sanitizedName,
						orig.Name,
						origName,
						orig.Line,
					),
				})
			}
		} else {
			// No conflict, add to list of unique namespaces
			uniqueNamespaces = append(uniqueNamespaces, nameReference{
				Name:   namespace,
				Line:   keyNode.Line,
				Result: sanitizedResult,
			})
		}
	}

	return errors
}
