package validation

import (
	"context"
	"fmt"
	"mime"

	"github.com/speakeasy-api/openapi/linter"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/validation"
	"gopkg.in/yaml.v3"
)

type ValidateRequests struct{}

var _ Rule = (*ValidateRequests)(nil)

func (r *ValidateRequests) ID() string {
	return "generator-validate-requests"
}

func (r *ValidateRequests) Category() string {
	return "validation"
}

func (r *ValidateRequests) Summary() string {
	return "Validate request body content types are valid MIME types."
}

func (r *ValidateRequests) HowToFix() string {
	return "Use valid MIME types for request body content entries (for example, application/json)."
}

func (r *ValidateRequests) Description() string {
	return "Validate request body content types are valid MIME types. This ensures payloads can be mapped consistently by generators."
}

func (r *ValidateRequests) Link() string {
	return ""
}

func (r *ValidateRequests) DefaultSeverity() validation.Severity {
	return validation.SeverityError
}

func (r *ValidateRequests) Versions() []string {
	return nil // Applies to all versions
}

func (r *ValidateRequests) Run(ctx context.Context, docInfo *linter.DocumentInfo[*openapi.OpenAPI], config *linter.RuleConfig) []error {
	if docInfo == nil || docInfo.Document == nil || docInfo.Index == nil {
		return nil
	}

	var validationErrors []error

	// Iterate through all operations
	for _, opIndexNode := range docInfo.Index.Operations {
		operation := opIndexNode.Node
		if operation == nil {
			continue
		}

		// Get request body
		requestBodyRef := operation.GetRequestBody()
		if requestBodyRef == nil {
			continue
		}

		// Get the actual request body object (handles $ref automatically)
		requestBody := requestBodyRef.GetObject()
		if requestBody == nil {
			continue
		}

		// Get content map
		content := requestBody.GetContent()
		if content == nil || content.Len() == 0 {
			continue
		}

		// Iterate through content types
		for contentType, mediaType := range content.All() {
			// Validate the content type as a valid MIME type
			_, _, err := mime.ParseMediaType(contentType)
			if err != nil {
				// Get the node for error reporting - use the media type key node
				var node *yaml.Node
				if mediaType != nil && mediaType.GetCore() != nil {
					node = mediaType.GetRootNode()
				}

				validationErrors = append(validationErrors, &validation.Error{
					Rule:            r.ID(),
					Severity:        r.DefaultSeverity(),
					Node:            node,
					UnderlyingError: fmt.Errorf("content type not valid - found `%s`: `%s`", contentType, err.Error()),
				})
			}
		}
	}

	return validationErrors
}
