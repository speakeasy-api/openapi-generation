package validation

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/speakeasy-api/openapi/linter"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/validation"
	"gopkg.in/yaml.v3"
)

// The MissingExamples rule checks that examples are provided for parameters, request bodies,
// responses, and component schemas where possible to improve API documentation
type MissingExamples struct{}

var _ Rule = (*MissingExamples)(nil)

func (r *MissingExamples) ID() string {
	return "generator-missing-examples"
}

func (r *MissingExamples) Category() string {
	return "validation"
}

func (r *MissingExamples) Summary() string {
	return "Ensure request, response, and schema examples are provided."
}

func (r *MissingExamples) HowToFix() string {
	return "Add examples to request bodies, responses, parameters, and component schemas where applicable."
}

func (r *MissingExamples) Description() string {
	return "Examples should be provided where possible to improve API documentation and SDK usability. Missing examples reduce clarity for consumers."
}

func (r *MissingExamples) Link() string {
	return ""
}

func (r *MissingExamples) DefaultSeverity() validation.Severity {
	return validation.SeverityHint
}

func (r *MissingExamples) Versions() []string {
	return nil // Applies to all versions
}

func (r *MissingExamples) Run(ctx context.Context, docInfo *linter.DocumentInfo[*openapi.OpenAPI], config *linter.RuleConfig) []error {
	if docInfo == nil || docInfo.Index == nil {
		return nil
	}

	var validationErrors []error
	seen := make(map[*yaml.Node]bool)

	// Check media types (from request bodies and responses) for missing examples
	// The Index.MediaTypes contains all media types from both request bodies and responses
	for _, mediaTypeNode := range docInfo.Index.MediaTypes {
		mediaType := mediaTypeNode.Node
		if mediaType == nil {
			continue
		}

		schemaRef := mediaType.GetSchema()
		if schemaRef == nil {
			continue
		}

		// Check if example exists at media type level
		// MediaType.Example is values.Value (*yaml.Node)
		// MediaType.Examples is *sequencedmap.Map[string, *ReferencedExample]
		hasMediaTypeExample := mediaType.Example != nil || (mediaType.Examples != nil && mediaType.Examples.Len() > 0)

		if !hasMediaTypeExample {
			// For references, check if the component schema being referenced has an example
			// For inline schemas, check the schema directly
			schema := schemaRef.GetSchema()
			if schema != nil {
				// Schema.Example is values.Value (*yaml.Node)
				// Schema.Examples is []values.Value ([]*yaml.Node)
				hasSchemaExample := schema.Example != nil || len(schema.Examples) > 0
				if !hasSchemaExample {
					node := schemaRef.GetRootNode()
					if node != nil && !seen[node] {
						seen[node] = true

						// Determine context from location - is this in a request body or response?
						contextKey := determineMediaTypeContext(mediaTypeNode)
						validationErrors = append(validationErrors, &validation.Error{
							Rule:            r.ID(),
							Severity:        r.DefaultSeverity(),
							Node:            node,
							UnderlyingError: fmt.Errorf("missing example for `%s`. Consider adding an example", contextKey),
						})
					}
				}
			}
		}
	}

	// Check inline parameters for missing examples
	for _, paramNode := range docInfo.Index.InlineParameters {
		paramRef := paramNode.Node
		if paramRef == nil {
			continue
		}

		param := paramRef.GetObject()
		if param == nil {
			continue
		}

		// Parameter.Example is values.Value (*yaml.Node)
		// Parameter.Examples is *sequencedmap.Map[string, *ReferencedExample]
		hasParamExample := param.Example != nil || (param.Examples != nil && param.Examples.Len() > 0)

		if !hasParamExample {
			// Check if parameter uses content instead of schema
			paramContent := param.GetContent()
			if paramContent.Len() > 0 {
				// Parameters with content are handled via MediaTypes index
				continue
			}

			// Check schema for example
			schemaRef := param.GetSchema()
			if schemaRef == nil {
				continue
			}

			// Check the referenced or inline schema for examples
			schema := schemaRef.GetSchema()
			if schema != nil {
				// Schema.Example is values.Value (*yaml.Node)
				// Schema.Examples is []values.Value ([]*yaml.Node)
				hasSchemaExample := schema.Example != nil || len(schema.Examples) > 0
				if !hasSchemaExample {
					node := schemaRef.GetRootNode()
					if node != nil && !seen[node] {
						seen[node] = true
						validationErrors = append(validationErrors, &validation.Error{
							Rule:            r.ID(),
							Severity:        r.DefaultSeverity(),
							Node:            node,
							UnderlyingError: errors.New("missing example for parameter. Consider adding an example"),
						})
					}
				}
			}
		}
	}

	// Check component parameters for missing examples
	for _, paramNode := range docInfo.Index.ComponentParameters {
		paramRef := paramNode.Node
		if paramRef == nil {
			continue
		}

		param := paramRef.GetObject()
		if param == nil {
			continue
		}

		// Parameter.Example is values.Value (*yaml.Node)
		// Parameter.Examples is *sequencedmap.Map[string, *ReferencedExample]
		hasParamExample := param.Example != nil || (param.Examples != nil && param.Examples.Len() > 0)

		if !hasParamExample {
			// Check if parameter uses content instead of schema
			paramContent := param.GetContent()
			if paramContent.Len() > 0 {
				// Parameters with content are handled via MediaTypes index
				continue
			}

			// Check schema for example
			schemaRef := param.GetSchema()
			if schemaRef == nil {
				continue
			}

			// Check the referenced or inline schema for examples
			schema := schemaRef.GetSchema()
			if schema != nil {
				// Schema.Example is values.Value (*yaml.Node)
				// Schema.Examples is []values.Value ([]*yaml.Node)
				hasSchemaExample := schema.Example != nil || len(schema.Examples) > 0
				if !hasSchemaExample {
					node := schemaRef.GetRootNode()
					if node != nil && !seen[node] {
						seen[node] = true
						validationErrors = append(validationErrors, &validation.Error{
							Rule:            r.ID(),
							Severity:        r.DefaultSeverity(),
							Node:            node,
							UnderlyingError: errors.New("missing example for parameter. Consider adding an example"),
						})
					}
				}
			}
		}
	}

	// Check component schemas for missing examples
	// Use Index.ComponentSchemas to catch all schemas including those from external documents
	for _, schemaNode := range docInfo.Index.ComponentSchemas {
		schemaRef := schemaNode.Node
		if schemaRef == nil {
			continue
		}

		schema := schemaRef.GetSchema()
		if schema == nil {
			continue
		}

		// Schema.Example is values.Value (*yaml.Node)
		// Schema.Examples is []values.Value ([]*yaml.Node)
		hasSchemaExample := schema.Example != nil || len(schema.Examples) > 0

		if !hasSchemaExample {
			// Get the root node for proper error reporting
			node := schemaRef.GetRootNode()
			if node != nil && !seen[node] {
				seen[node] = true
				validationErrors = append(validationErrors, &validation.Error{
					Rule:            r.ID(),
					Severity:        r.DefaultSeverity(),
					Node:            node,
					UnderlyingError: errors.New("missing example for component. Consider adding an example"),
				})
			}
		}
	}

	return validationErrors
}

// determineMediaTypeContext determines if a media type is in a request body or response
// based on the index node location information
func determineMediaTypeContext(mediaTypeNode *openapi.IndexNode[*openapi.MediaType]) string {
	// Check the location path to determine if this is in a request body or response
	if mediaTypeNode.Location != nil {
		locationPath := mediaTypeNode.Location.ToJSONPointer().String()

		// Check if the path contains "requestBody" or "responses"
		// Examples:
		// - #/paths/~1test/get/requestBody/content/application~1json
		// - #/paths/~1test/get/responses/200/content/application~1json
		if strings.Contains(locationPath, "/requestBody/") || strings.Contains(locationPath, "/requestBody") {
			return "requestBody"
		}
		if strings.Contains(locationPath, "/responses/") || strings.Contains(locationPath, "/responses") {
			return "responses"
		}
	}

	// Default to responses if we can't determine
	return "responses"
}
