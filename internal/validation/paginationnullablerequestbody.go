package validation

import (
	"context"
	"fmt"

	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/speakeasy-api/openapi/jsonschema/oas3"
	"github.com/speakeasy-api/openapi/linter"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/validation"
	"gopkg.in/yaml.v3"
)

// PaginationNullableRequestBody rejects a nullable paginated body with required properties.
type PaginationNullableRequestBody struct{}

var _ Rule = (*PaginationNullableRequestBody)(nil)

func (r *PaginationNullableRequestBody) ID() string {
	return "generator-validate-pagination-nullable-request-body"
}

func (r *PaginationNullableRequestBody) Category() string {
	return "validation"
}

func (r *PaginationNullableRequestBody) Summary() string {
	return "Paginated operations must not take a nullable request body with required properties."
}

func (r *PaginationNullableRequestBody) HowToFix() string {
	return "Drop `nullable: true` from the request body schema, make every property optional so a body can be constructed for pages after the first, or remove `x-speakeasy-pagination` from the operation."
}

func (r *PaginationNullableRequestBody) Description() string {
	return "Pagination sets its input on the request body and resends it, so pages after the first need a body object. A top-level-nullable body with required properties cannot be constructed when the caller passes null."
}

func (r *PaginationNullableRequestBody) Link() string {
	return ""
}

func (r *PaginationNullableRequestBody) DefaultSeverity() validation.Severity {
	return validation.SeverityError
}

func (r *PaginationNullableRequestBody) Versions() []string {
	return nil
}

func (r *PaginationNullableRequestBody) Run(_ context.Context, docInfo *linter.DocumentInfo[*openapi.OpenAPI], _ *linter.RuleConfig) []error {
	if docInfo == nil || docInfo.Index == nil {
		return nil
	}

	var validationErrors []error

	for _, opIndexNode := range docInfo.Index.Operations {
		operation := opIndexNode.Node
		if operation == nil {
			continue
		}

		paginationExtension, ok := operation.GetExtensions().Get("x-speakeasy-pagination")
		if !ok {
			continue
		}

		if !paginationReadsRequestBody(paginationExtension) {
			continue
		}

		requestBodyRef := operation.GetRequestBody()
		if requestBodyRef == nil {
			continue
		}

		requestBody := requestBodyRef.GetObject()
		if requestBody == nil {
			continue
		}

		content := requestBody.GetContent()
		if content == nil || content.Len() == 0 {
			continue
		}

		for _, mediaType := range content.All() {
			if mediaType == nil {
				continue
			}

			schemaRef := mediaType.GetSchema()
			if schemaRef == nil {
				continue
			}

			schemas := make([]*oas3.Schema, 0, 2)
			if localSchema := schemaRef.GetSchema(); localSchema != nil {
				schemas = append(schemas, localSchema)
			}
			if resolvedSchemaRef := schemaRef.GetResolvedSchema(); resolvedSchemaRef != nil {
				if resolvedSchema := resolvedSchemaRef.GetSchema(); resolvedSchema != nil && (len(schemas) == 0 || resolvedSchema != schemas[0]) {
					schemas = append(schemas, resolvedSchema)
				}
			}
			if len(schemas) == 0 {
				continue
			}

			nullable := false
			seen := make(map[string]bool)
			var required []string
			for _, schema := range schemas {
				if isNullableSchema(schema) {
					nullable = true
				}
				for _, name := range schema.GetRequired() {
					if !seen[name] {
						seen[name] = true
						required = append(required, name)
					}
				}
			}
			if !nullable || len(required) == 0 {
				continue
			}

			schema := schemas[len(schemas)-1]

			validationErrors = append(validationErrors, &validation.Error{
				Rule:     r.ID(),
				Severity: r.DefaultSeverity(),
				Node:     schema.GetRootNode(),
				UnderlyingError: fmt.Errorf(
					"paginated operation has a nullable request body with required properties (%v) - pages after the first need a body object, which cannot be constructed when the caller passes null",
					required,
				),
			})
		}
	}

	return validationErrors
}

// Parameter-only pagination never materializes a body, so it is exempt.
func paginationReadsRequestBody(extension *yaml.Node) bool {
	if extension == nil {
		return false
	}

	var pagination extensions.Pagination
	if err := extension.Decode(&pagination); err != nil {
		return false
	}

	for _, input := range pagination.Inputs {
		if input.In == extensions.PaginationInputInTypeRequestBody {
			return true
		}
	}

	return false
}

// 3.1 spells nullability as a `null` member of the type array.
func isNullableSchema(schema *oas3.Schema) bool {
	if schema.GetNullable() {
		return true
	}

	for _, t := range schema.GetType() {
		if t == "null" {
			return true
		}
	}

	return false
}
