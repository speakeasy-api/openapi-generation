package schemas

import (
	"context"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/handlers"
	"github.com/speakeasy-api/openapi/jsonschema/oas3"
)

func (s *Schemas) handleSchemaComments(
	ctx context.Context,
	schema *oas3.Schema,
	parentComponentDescription string,
) (*ast.Comment, error) {
	if schema.GetDescription() == "" &&
		schema.GetExternalDocs() == nil &&
		schema.Deprecated == nil &&
		parentComponentDescription == "" {
		return nil, nil
	}

	comments := &ast.Comment{}

	if schema.GetDescription() != "" {
		comments.Description = schema.GetDescription()
	}
	if comments.Description == "" && parentComponentDescription != "" {
		comments.Description = parentComponentDescription
	}

	deprecated := schema.GetDeprecated()

	var err error
	comments, err = handlers.HandleDeprecated(ctx, comments, deprecated, schema.GetExtensions(), s.Subsystem)
	if err != nil {
		return nil, err
	}

	if schema.GetExternalDocs().GetURL() != "" {
		comments.ExternalDocs = &ast.ExternalDocs{
			URL: schema.GetExternalDocs().GetURL(),
		}

		if schema.GetExternalDocs().GetDescription() != "" {
			comments.ExternalDocs.Description = schema.GetExternalDocs().GetDescription()
		}
	}

	extComments, err := s.Subsystem.Extensions.GetCustomDocs(schema.GetExtensions())
	if err != nil {
		return nil, err
	}

	if extComments != nil {
		outComments := make(map[string]*ast.ExtendedComment, len(extComments))
		for k, v := range extComments {
			outComments[k] = &ast.ExtendedComment{
				Summary:     v.Summary,
				Description: v.Description,
			}
		}
		comments.ExtendedComments = outComments
	}

	return comments, nil
}
