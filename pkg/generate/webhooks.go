package generate

import (
	"context"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/internal/licensing"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/sequencedmap"
	"gopkg.in/yaml.v3"
)

func (g *Generator) handleWebhooks(ctx context.Context, docInfo *document.DocumentInfo, params handlePathsParams) (*sequencedmap.Map[string, []ast.Operation], error) {
	params.IsWebhooks = true

	if docInfo.Doc.GetWebhooks().Len() == 0 {
		return nil, nil
	}

	params.Paths = &openapi.Paths{Map: docInfo.Doc.Webhooks}

	webhooksConfig, err := g.subsystem.Extensions.HandleWebhooksExtension(docInfo.Doc.GetExtensions())
	if err != nil {
		return nil, err
	}

	if webhooksConfig != nil {
		params.WebhookSecurity = webhooksConfig.Security
	}

	webhookOperations, err := g.handlePaths(ctx, params)
	if err != nil {
		return nil, err
	}

	if webhookOperations.Len() > 0 {
		var webhookOperationsNode *yaml.Node
		for _, pi := range docInfo.Doc.Webhooks.All() {
			webhookOperationsNode = pi.GetRootNode()
		}

		// Validate access to the webhooks feature - backwards compatibility
		accessErr := licensing.ValidateAccountHasFeatureAccess(ctx, features.FeatureWebhooks)
		if accessErr != nil {
			errMsg := "skipping webhooks: " + accessErr.Error()
			logging.LogWarning(ctx, "unsupported", errors.NewUnsupportedError(errMsg, webhookOperationsNode))
			//nolint:nilerr
			return sequencedmap.New[string, []ast.Operation](), nil
		}
		g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureWebhooks)

		// Validate access to the newer webhook handlers feature
		accessErr = licensing.ValidateAccountHasFeatureAccess(ctx, features.FeatureWebhookHandlers)
		supported := g.subsystem.Features.IsFeatureSupported(ctx, features.FeatureWebhookHandlers)
		if accessErr != nil {
			errMsg := "skipping webhooks: " + accessErr.Error()
			logging.LogWarning(ctx, "unsupported", errors.NewUnsupportedError(errMsg, webhookOperationsNode))
		}
		if accessErr == nil && supported {
			g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureWebhookHandlers)
		}
		if accessErr == nil && !supported {
			logging.LogWarning(ctx, "unsupported", errors.NewUnsupportedError("webhooks is enabled but webhook handlers are not supported by this target", webhookOperationsNode))
		}
	}

	return webhookOperations, nil
}
