package validation

import (
	"context"
	"fmt"
	"mime"
	"strconv"
	"strings"

	"github.com/speakeasy-api/openapi/linter"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/validation"
	"gopkg.in/yaml.v3"
)

type ValidateResponses struct{}

var _ Rule = (*ValidateResponses)(nil)

func (r *ValidateResponses) ID() string {
	return "generator-validate-responses"
}

func (r *ValidateResponses) Category() string {
	return "validation"
}

func (r *ValidateResponses) Summary() string {
	return "Validate response status codes and content types are valid."
}

func (r *ValidateResponses) HowToFix() string {
	return "Use valid HTTP status codes (or wildcards like 2XX) and valid MIME types in response content entries."
}

func (r *ValidateResponses) Description() string {
	return "Validate response status codes and content types are syntactically valid. This ensures MIME types can be mapped correctly during generation."
}

func (r *ValidateResponses) Link() string {
	return ""
}

func (r *ValidateResponses) DefaultSeverity() validation.Severity {
	return validation.SeverityError
}

func (r *ValidateResponses) Versions() []string {
	return nil // Applies to all versions
}

func (r *ValidateResponses) Run(ctx context.Context, docInfo *linter.DocumentInfo[*openapi.OpenAPI], config *linter.RuleConfig) []error {
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

		// Get responses
		responses := operation.GetResponses()
		if responses == nil {
			continue
		}

		// Iterate through all status codes (including default)
		for statusCode, responseRef := range responses.All() {
			// Validate status code format
			if strings.Contains(strings.ToUpper(statusCode), "XX") {
				// Wildcard status code like 2XX, 3XX, etc.
				_, err := strconv.Atoi(statusCode[:1])
				if err != nil {
					// Get the node for error reporting
					var node *yaml.Node
					if responseRef != nil && responseRef.GetCore() != nil {
						node = responseRef.GetRootNode()
					}

					validationErrors = append(validationErrors, &validation.Error{
						Rule:            r.ID(),
						Severity:        r.DefaultSeverity(),
						Node:            node,
						UnderlyingError: fmt.Errorf("status codes must be valid HTTP status code or wildcards in the form of `1XX`, `2XX`, etc - found `%s`", statusCode),
					})
					continue
				}
			}

			// Get the actual response object (handles $ref automatically)
			response := responseRef.GetObject()
			if response == nil {
				continue
			}

			// Get content map
			content := response.GetContent()
			if content == nil || content.Len() == 0 {
				continue
			}

			// Iterate through content types
			for contentType, mediaType := range content.All() {
				// Validate the content type as a valid MIME type
				_, _, err := mime.ParseMediaType(contentType)
				if err != nil {
					// Get the node for error reporting - use the media type node
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
	}

	return validationErrors
}
