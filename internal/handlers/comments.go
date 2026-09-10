package handlers

import (
	"context"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/internal/subsystem"
)

type Features interface {
	RecordFeatureUsage(ctx context.Context, feature features.Feature)
}

func HandleDeprecated(ctx context.Context, comments *ast.Comment, deprecated bool, exts extensions.OAExtensions, subsystem *subsystem.Subsystem) (*ast.Comment, error) {
	if deprecated {
		if comments == nil {
			comments = &ast.Comment{}
		}

		deprecationReplacement, err := subsystem.Extensions.GetDeprecationReplacement(exts)
		if err != nil {
			return nil, err
		}

		deprecationMessage, err := subsystem.Extensions.GetDeprecationMessage(exts)
		if err != nil {
			return nil, err
		}
		if deprecationReplacement != "" || deprecationMessage != "" {
			subsystem.Features.RecordFeatureUsage(ctx, features.FeatureDeprecations)
		}

		comments.Deprecated = true
		comments.DeprecationReplacement = deprecationReplacement
		comments.DeprecationMessage = deprecationMessage
	}

	return comments, nil
}
