package validation_test

import (
	"bytes"
	"context"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/internal/validation"
	"github.com/speakeasy-api/openapi/openapi"
)

// filterByRule returns the error strings produced by the given rule ID.
func filterByRule(errs []error, ruleID string) []string {
	var got []string
	for _, e := range errs {
		if strings.Contains(e.Error(), ruleID) {
			got = append(got, e.Error())
		}
	}
	return got
}

// validateSpec is a test helper that parses a spec and validates it
// This replaces the old Validate() method that accepted raw bytes
func validateSpec(v *validation.Validator, ctx context.Context, schema []byte, schemaPath string, target types.Target) *validation.Result {
	// Parse the document
	doc, validationErrs, err := openapi.Unmarshal(ctx, bytes.NewReader(schema))
	if err != nil {
		return validation.NewResult(nil, []error{err}, "", nil)
	}

	// Create DocumentInfo
	docInfo := &document.DocumentInfo{
		Doc:        doc,
		Schema:     schema,
		SchemaPath: schemaPath,
	}

	// Call ValidateDocument with parse errors
	return v.ValidateDocument(ctx, docInfo, target, validationErrs...)
}
