package validation

import (
	"context"
	"errors"

	"github.com/speakeasy-api/openapi/linter"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/validation"
)

// Pagination checks if pagination might be supported by operations based on parameter names
type Pagination struct{}

var _ Rule = (*Pagination)(nil)

func (r *Pagination) ID() string {
	return "generator-pagination"
}

func (r *Pagination) Category() string {
	return "validation"
}

func (r *Pagination) Summary() string {
	return "Detect pagination patterns and suggest pagination extensions."
}

func (r *Pagination) HowToFix() string {
	return "Add the x-speakeasy-pagination extension to operations that support pagination parameters."
}

func (r *Pagination) Description() string {
	return "Detect operations that appear to support pagination based on request and response patterns. This drives pagination helpers in generated SDKs."
}

func (r *Pagination) Link() string {
	return ""
}

func (r *Pagination) DefaultSeverity() validation.Severity {
	return validation.SeverityHint
}

func (r *Pagination) Versions() []string {
	return nil // Applies to all versions
}

func (r *Pagination) Run(ctx context.Context, docInfo *linter.DocumentInfo[*openapi.OpenAPI], config *linter.RuleConfig) []error {
	if docInfo == nil || docInfo.Index == nil {
		return nil
	}

	var validationErrors []error

	// Iterate all operations
	for _, opNode := range docInfo.Index.Operations {
		operation := opNode.Node
		if operation == nil {
			continue
		}

		// Check if operation has potential pagination parameters
		hasPotentialPagination := false
		parameters := operation.GetParameters()

		for _, paramRef := range parameters {
			param := paramRef.GetObject()
			if param == nil {
				continue
			}

			paramName := param.GetName()
			// Check for common pagination parameter names
			if paramName == "page" || paramName == "limit" {
				hasPotentialPagination = true
				break
			}
		}

		// If operation might support pagination, check for extension
		if hasPotentialPagination {
			extensions := operation.GetExtensions()
			_, hasPaginationExt := extensions.Get("x-speakeasy-pagination")

			if !hasPaginationExt {
				// Report hint suggesting pagination extension
				operationNode := operation.GetRootNode()
				if operationNode != nil {
					validationErrors = append(validationErrors, &validation.Error{
						Rule:            r.ID(),
						Severity:        r.DefaultSeverity(),
						Node:            operationNode,
						UnderlyingError: errors.New("pagination might be supported by this operation - consider adding `x-speakeasy-pagination` extension"),
					})
				}
			}
		}
	}

	return validationErrors
}
