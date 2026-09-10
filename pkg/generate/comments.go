package generate

import (
	"context"
	"fmt"

	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi/openapi"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/handlers"
	"github.com/speakeasy-api/openapi/extensions"
)

func (g *Generator) handleDocumentComments(ctx context.Context, doc *openapi.OpenAPI, a *ast.AST) (*ast.Comment, error) {
	if (doc.GetInfo().GetTitle() == "" && doc.GetInfo().GetDescription() == "" && doc.GetInfo().GetSummary() == "") &&
		(doc.GetExternalDocs() == nil || (doc.GetExternalDocs().GetURL() == "" && doc.GetExternalDocs().GetDescription() == "")) {
		return nil, nil
	}

	comment := &ast.Comment{}

	titleAdded := false

	if doc.GetInfo().GetSummary() != "" {
		if doc.GetInfo().GetTitle() != "" {
			comment.Summary = fmt.Sprintf("%s: %s", doc.GetInfo().GetTitle(), doc.GetInfo().GetSummary())
			titleAdded = true
		} else {
			comment.Summary = doc.GetInfo().GetSummary()
		}
	}

	if doc.GetInfo().GetDescription() != "" {
		if !titleAdded && doc.GetInfo().GetTitle() != "" {
			comment.Description = fmt.Sprintf("%s: %s", doc.GetInfo().GetTitle(), doc.GetInfo().GetDescription())
		} else {
			comment.Description = doc.GetInfo().GetDescription()
		}
	}

	if doc.GetExternalDocs().GetURL() != "" {
		comment.ExternalDocs = &ast.ExternalDocs{
			URL: doc.GetExternalDocs().GetURL(),
		}

		if doc.GetExternalDocs().GetDescription() != "" {
			comment.ExternalDocs.Description = doc.GetExternalDocs().GetDescription()
		}
	}

	extComments, err := g.handleDocsExtension(ctx, doc.GetExtensions())
	if err != nil {
		return nil, err
	}

	comment.ExtendedComments = extComments

	if a != nil && a.MainSDK != nil {
		a.MainSDK.Comments = comment
	}

	return comment, err
}

func (g *Generator) handleOperationComments(ctx context.Context, operation *openapi.Operation) (*ast.Comment, error) {
	if operation.GetSummary() == "" && operation.GetDescription() == "" && operation.GetExternalDocs() == nil && !operation.GetDeprecated() {
		return nil, nil
	}

	comment, err := handlers.HandleDeprecated(ctx, nil, operation.GetDeprecated(), operation.GetExtensions(), g.subsystem)
	if err != nil {
		return nil, err
	}

	if comment == nil {
		comment = &ast.Comment{}
	}

	if operation.GetSummary() != "" {
		comment.Summary = operation.GetSummary()
	}

	if operation.GetDescription() != "" || operation.GetSummary() != "" {
		if operation.GetSummary() == "" {
			comment.Summary = operation.GetDescription()
		} else {
			comment.Description = operation.GetDescription()
		}
	}

	if operation.GetExternalDocs() != nil {
		if operation.GetExternalDocs().GetURL() != "" {
			comment.ExternalDocs = &ast.ExternalDocs{
				URL: operation.GetExternalDocs().GetURL(),
			}

			if operation.GetExternalDocs().GetDescription() != "" {
				comment.ExternalDocs.Description = operation.GetExternalDocs().GetDescription()
			}
		}
	}

	extComments, err := g.handleDocsExtension(ctx, operation.GetExtensions())
	if err != nil {
		return nil, err
	}

	comment.ExtendedComments = extComments

	return comment, nil
}

func (g *Generator) handleDocsExtension(ctx context.Context, ext *extensions.Extensions) (map[string]*ast.ExtendedComment, error) {
	if ext == nil {
		return nil, nil
	}

	comments, err := g.subsystem.Extensions.GetCustomDocs(ext)
	if err != nil {
		return nil, err
	}

	if len(comments) > 0 {
		g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureDocs)

		outComments := make(map[string]*ast.ExtendedComment, len(comments))
		for k, v := range comments {
			outComments[k] = &ast.ExtendedComment{
				Summary:     v.Summary,
				Description: v.Description,
			}
		}

		return outComments, nil
	}

	return nil, nil
}

func (g *Generator) handleTagComments(ctx context.Context, tag *openapi.Tag) (*ast.Comment, error) {
	if tag == nil || (tag.GetDescription() == "" && tag.GetExternalDocs() == nil && tag.GetSummary() == "") {
		return nil, nil
	}

	comment := &ast.Comment{
		Summary:     tag.GetSummary(),
		Description: tag.GetDescription(),
	}

	if tag.GetExternalDocs() != nil {
		if tag.GetExternalDocs().GetURL() != "" {
			comment.ExternalDocs = &ast.ExternalDocs{
				URL:         tag.GetExternalDocs().GetURL(),
				Description: tag.GetExternalDocs().GetDescription(),
			}
		}
	}

	extComments, err := g.handleDocsExtension(ctx, tag.GetExtensions())
	if err != nil {
		return nil, err
	}

	comment.ExtendedComments = extComments

	return comment, nil
}
