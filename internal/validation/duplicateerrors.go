package validation

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/sanitization"
	"github.com/speakeasy-api/openapi/hashing"
	"github.com/speakeasy-api/openapi/jsonschema/oas3"
	"github.com/speakeasy-api/openapi/linter"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/validation"
)

type DuplicateErrors struct{}

var _ Rule = (*DuplicateErrors)(nil)

func (r *DuplicateErrors) ID() string {
	return "generator-duplicate-errors"
}

func (r *DuplicateErrors) Category() string {
	return "validation"
}

func (r *DuplicateErrors) Summary() string {
	return "Ensure no duplicated error schemas found."
}

func (r *DuplicateErrors) HowToFix() string {
	return "Consolidate identical error schemas for the same status code or change the schema so duplicate error types are not generated."
}

func (r *DuplicateErrors) Description() string {
	return "Detect duplicate error schemas that would produce multiple SDK error types. When their generated names collide, error handling becomes ambiguous."
}

func (r *DuplicateErrors) Link() string {
	return ""
}

func (r *DuplicateErrors) DefaultSeverity() validation.Severity {
	return validation.SeverityHint
}

func (r *DuplicateErrors) Versions() []string {
	return nil // Applies to all versions
}

// Pattern matches error response schema paths to extract status code.
// Location.ToJSONPointer().String() does not include a leading "#".
// Example: /paths/~/users/get/responses/400/content/application~1json/schema
var errorPattern = regexp.MustCompile(`^/paths/[^/]+/(?:get|put|post|delete|patch|trace|head|options)/responses/([45](?:\d\d|XX))/content/[^/]+/schema$`)

func getErrorStatusCodeFromPath(path string) string {
	matches := errorPattern.FindStringSubmatch(path)
	if len(matches) != 2 {
		return ""
	}
	return matches[1]
}

func (r *DuplicateErrors) Run(ctx context.Context, docInfo *linter.DocumentInfo[*openapi.OpenAPI], config *linter.RuleConfig) []error {
	if docInfo == nil || docInfo.Index == nil {
		return nil
	}

	var schemasByHash = map[string][]*openapi.IndexNode[*oas3.JSONSchemaReferenceable]{}

	// Iterate through all inline schemas
	for _, indexNode := range docInfo.Index.InlineSchemas {
		schemaRef := indexNode.Node
		if schemaRef == nil {
			continue
		}

		schema := schemaRef.GetSchema()
		if schema == nil {
			continue
		}

		// Extract status code from the path
		statusCode := getErrorStatusCodeFromPath(indexNode.Location.ToJSONPointer().String())
		if statusCode == "" {
			continue
		}

		// Check if this is an object, enum, or composite schema
		types := schema.GetType()
		isObject := false
		for _, t := range types {
			if string(t) == "object" {
				isObject = true
				break
			}
		}

		enumValues := schema.GetEnum()
		isEnum := len(enumValues) > 0

		isOneOf := schema.GetOneOf() != nil
		isAnyOf := schema.GetAnyOf() != nil
		isAllOf := schema.GetAllOf() != nil

		if !isObject && !isEnum && !isOneOf && !isAnyOf && !isAllOf {
			continue
		}

		// Create hash of schema content with status code
		hash := "status: " + statusCode + hashing.Hash(schema.GetRootNode())
		schemasByHash[hash] = append(schemasByHash[hash], indexNode)
	}

	// Find duplicates and generate errors
	validationErrors := make([]error, 0, len(schemasByHash))
	for _, schemas := range schemasByHash {
		if len(schemas) <= 1 {
			continue
		}

		statusCode := getErrorStatusCodeFromPath(schemas[0].Location.ToJSONPointer().String())
		errorName := sanitization.HumanizeStatusCode(statusCode)
		if !strings.Contains(errorName, "Error") {
			errorName += "Error"
		}

		lines := []string{}
		for _, indexNode := range schemas {
			node := indexNode.Node.GetSchema().GetRootNode()
			if node != nil && node.Line > 0 {
				lines = append(lines, strconv.Itoa(node.Line))
			}
		}

		message := fmt.Sprintf("`%s` duplicated `%d` times at lines[%s]", errorName, len(schemas), strings.Join(lines, ","))

		first := schemas[0]
		firstNode := first.Node.GetSchema().GetRootNode()

		validationErrors = append(validationErrors, &validation.Error{
			Rule:            r.ID(),
			Severity:        r.DefaultSeverity(),
			Node:            firstNode,
			UnderlyingError: errors.New(message),
		})
	}

	return validationErrors
}
