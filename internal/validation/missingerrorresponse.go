package validation

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/speakeasy-api/openapi/linter"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/validation"
)

// The MissingErrorResponse rule checks that all operations have at least one error response defined
// An error response is defined as one of the following:
//   - A response with a wildcard status code (4XX, 5XX, or default)
//   - A response with a status code in the range 400-499
//   - A response with a status code in the range 500-599
type MissingErrorResponse struct{}

var _ Rule = (*MissingErrorResponse)(nil)

func (r *MissingErrorResponse) ID() string {
	return "generator-missing-error-response"
}

func (r *MissingErrorResponse) Category() string {
	return "validation"
}

func (r *MissingErrorResponse) Summary() string {
	return "Operation is missing an error response."
}

func (r *MissingErrorResponse) HowToFix() string {
	return "Add a default response or a 4XX/5XX response for each operation to document error cases."
}

func (r *MissingErrorResponse) Description() string {
	return "Error responses should be defined for all operations to document failure cases. This makes client error handling predictable and consistent."
}

func (r *MissingErrorResponse) Link() string {
	return ""
}

func (r *MissingErrorResponse) DefaultSeverity() validation.Severity {
	return validation.SeverityHint
}

func (r *MissingErrorResponse) Versions() []string {
	return nil // Applies to all versions
}

func (r *MissingErrorResponse) Run(ctx context.Context, docInfo *linter.DocumentInfo[*openapi.OpenAPI], config *linter.RuleConfig) []error {
	if docInfo == nil || docInfo.Index == nil {
		return nil
	}

	var validationErrors []error

	// Iterate through all operations
	for _, opIndexNode := range docInfo.Index.Operations {
		operation := opIndexNode.Node
		if operation == nil {
			continue
		}

		// Get the responses for this operation
		responses := operation.GetResponses()
		if responses == nil {
			continue
		}

		// Check if at least one error status code exists
		foundErrorStatusCode := false

		// Check for default response (separate from status code map)
		if responses.GetDefault() != nil {
			foundErrorStatusCode = true
		}

		// Check all status code responses
		if !foundErrorStatusCode {
			for statusCode := range responses.All() {
				// Status code must be 3 characters for wildcard or numeric check
				if len(statusCode) != 3 {
					continue
				}

				// Check for wildcard status codes (4XX, 5XX)
				if strings.Contains(strings.ToUpper(statusCode), "XX") && (statusCode[0] == '4' || statusCode[0] == '5') {
					foundErrorStatusCode = true
					break
				}

				// Check for numeric status codes in 400-599 range
				code, err := strconv.Atoi(statusCode)
				if err != nil {
					continue
				}

				// Verify it's a valid HTTP status code
				resolvedHttpCode := http.StatusText(code)
				if resolvedHttpCode == "OK" || resolvedHttpCode == "" {
					continue
				}

				// If we got here, it's a valid error status code
				foundErrorStatusCode = true
				break
			}
		}

		// If no error status code found, add validation error
		if !foundErrorStatusCode {
			validationErrors = append(validationErrors, &validation.Error{
				Rule:            r.ID(),
				Severity:        r.DefaultSeverity(),
				Node:            responses.GetRootNode(),
				UnderlyingError: errors.New("an error response should be defined for all operations"),
			})
		}
	}

	return validationErrors
}
