package ast_post_processing

import (
	"context"
	"fmt"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/ast_post_processing/infer_union_discriminator"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
)

// MarkUnionsOpen marks unions as open (tolerating unknown discriminator values)
// based on configuration and usage.
func MarkUnionsOpen(ctx context.Context, types []*ast.TypeDef, cfg map[string]interface{}) {
	config, ok := cfg["forwardCompatibleUnionsByDefault"]
	if !ok {
		return
	}

	for _, typeDef := range types {
		markDiscriminatedUnionAsOpenIfNeeded(typeDef, config)
		warnIfForwardCompatSilentlyDropped(ctx, typeDef, config)
	}
}

func markDiscriminatedUnionAsOpenIfNeeded(typeDef *ast.TypeDef, config interface{}) {
	if typeDef.Type != ast.DataTypeUnion {
		return
	}

	// Respect explicit x-speakeasy-unknown-values extension if present
	if typeDef.Extensions != nil && typeDef.Extensions.All != nil {
		if unknownValuesSetting, hasExplicitSetting := typeDef.Extensions.All["x-speakeasy-unknown-values"]; hasExplicitSetting {
			// Respect the explicit setting
			if str, ok := unknownValuesSetting.(string); ok && str == "allow" {
				typeDef.IsUnionOpen = true
			}
			// If "disallow" or any other value, leave IsUnionOpen as false
			return
		}
	}

	if typeDef.Discriminator == nil && config == "tagged-only" {
		return
	}

	if !typeDef.UsedInResponse {
		return
	}

	typeDef.IsUnionOpen = config == "tagged-and-untagged" || config == "true" || config == true ||
		(config == "tagged-only" && typeDef.Discriminator != nil)
}

func warnIfForwardCompatSilentlyDropped(ctx context.Context, typeDef *ast.TypeDef, config interface{}) {
	if typeDef.Type != ast.DataTypeUnion || typeDef.IsUnionOpen {
		return
	}
	if !typeDef.UsedInResponse {
		return
	}
	if typeDef.Discriminator != nil {
		return
	}
	if config != "tagged-only" {
		return
	}

	// An explicit x-speakeasy-unknown-values setting is a deliberate choice.
	if typeDef.Extensions != nil && typeDef.Extensions.All != nil {
		if _, hasExplicitSetting := typeDef.Extensions.All["x-speakeasy-unknown-values"]; hasExplicitSetting {
			return
		}
	}

	summary, ok := infer_union_discriminator.SummarizeMixedTagging(typeDef)
	if !ok {
		return
	}

	name := typeDef.Name
	if name == "" {
		name = "<unnamed>"
	}

	logging.LogWarning(
		ctx,
		fmt.Sprintf("union %q will not be forward-compatible", name),
		fmt.Errorf(
			"no single discriminator property could be inferred for union %q, so it is generated WITHOUT the Unknown* fallback; "+
				"unrecognized or partial variants from newer servers will fail to parse. %s. "+
				"Align all members on one required const-tagged property, or set x-speakeasy-unknown-values to silence this warning.",
			name, summary,
		),
	)
}
