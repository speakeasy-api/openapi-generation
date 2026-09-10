package validation

import (
	"context"
	"errors"

	"github.com/speakeasy-api/openapi/linter"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/validation"
)

// Retries checks if retry configuration is present
type Retries struct{}

var _ Rule = (*Retries)(nil)

func (r *Retries) ID() string {
	return "generator-retries"
}

func (r *Retries) Category() string {
	return "validation"
}

func (r *Retries) Summary() string {
	return "Ensure retries are configured for safe operations."
}

func (r *Retries) HowToFix() string {
	return "Configure retries using the x-speakeasy-retries extension globally or per operation."
}

func (r *Retries) Description() string {
	return "Retries should be configured for operations that are safe to retry. This enables resilient SDK behavior without hidden failures."
}

func (r *Retries) Link() string {
	return ""
}

func (r *Retries) DefaultSeverity() validation.Severity {
	return validation.SeverityHint
}

func (r *Retries) Versions() []string {
	return nil // Applies to all versions
}

func (r *Retries) Run(ctx context.Context, docInfo *linter.DocumentInfo[*openapi.OpenAPI], config *linter.RuleConfig) []error {
	if docInfo == nil || docInfo.Document == nil {
		return nil
	}

	doc := docInfo.Document

	// Check for global retries extension
	if doc.Extensions != nil {
		if _, hasGlobalRetries := doc.Extensions.Get("x-speakeasy-retries"); hasGlobalRetries {
			return nil // Global retries configured
		}
	}

	// Check if any operation has retries extension
	if docInfo.Index != nil {
		for _, opNode := range docInfo.Index.Operations {
			operation := opNode.Node
			if operation == nil {
				continue
			}

			extensions := operation.GetExtensions()
			if _, hasRetries := extensions.Get("x-speakeasy-retries"); hasRetries {
				return nil // At least one operation has retries
			}
		}
	}

	// No retries configured anywhere - report hint
	var rootNode = doc.GetRootNode()
	if rootNode == nil {
		return nil
	}

	return []error{
		&validation.Error{
			Rule:            r.ID(),
			Severity:        r.DefaultSeverity(),
			Node:            rootNode,
			UnderlyingError: errors.New("retries should be configured - consider adding `x-speakeasy-retries` extension globally or per-operation"),
		},
	}
}
