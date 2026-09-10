package validation

import (
	"context"
	stderrors "errors"

	"github.com/speakeasy-api/openapi/linter"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/validation"
	"gopkg.in/yaml.v3"
)

// ValidateDocument ensures the document has paths or webhooks with at least one entry
type ValidateDocument struct{}

var _ Rule = (*ValidateDocument)(nil)

func (r *ValidateDocument) ID() string {
	return "speakeasy-validate-document"
}

func (r *ValidateDocument) Category() string {
	return "validation"
}

func (r *ValidateDocument) Summary() string {
	return "Validate a document contains at least one path or webhook."
}

func (r *ValidateDocument) HowToFix() string {
	return "Define at least one path or webhook entry in the document."
}

func (r *ValidateDocument) Description() string {
	return "Document must have a paths or webhooks object with at least one entry. Empty specs cannot be generated into usable SDKs."
}

func (r *ValidateDocument) Link() string {
	return ""
}

func (r *ValidateDocument) DefaultSeverity() validation.Severity {
	return validation.SeverityError
}

func (r *ValidateDocument) Versions() []string {
	return nil // Applies to all versions
}

func (r *ValidateDocument) Run(ctx context.Context, docInfo *linter.DocumentInfo[*openapi.OpenAPI], config *linter.RuleConfig) []error {
	var errors []error
	doc := docInfo.Document

	// Check if document has paths or webhooks with at least one entry
	hasPaths := doc.Paths != nil && doc.Paths.Len() > 0
	hasWebhooks := doc.Webhooks != nil && doc.Webhooks.Len() > 0

	if !hasPaths && !hasWebhooks {
		// Get the root node for line number reporting
		var node *yaml.Node
		if doc.Paths != nil {
			node = doc.Paths.GetRootNode()
		}
		if node == nil {
			node = doc.GetRootNode()
		}

		errors = append(errors, &validation.Error{
			Rule:            r.ID(),
			Severity:        r.DefaultSeverity(),
			Node:            node,
			UnderlyingError: stderrors.New("document must have a `paths` or `webhooks` object with at least one entry"),
		})
	}

	return errors
}
