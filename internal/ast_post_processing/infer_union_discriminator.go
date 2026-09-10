package ast_post_processing

import (
	"context"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/ast_post_processing/infer_union_discriminator"
	"github.com/speakeasy-api/openapi-generation/v2/internal/env"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"go.uber.org/zap"
)

// InferUnionDiscriminators analyzes all union types and infers discriminators where possible
// schemaPath is used for logging with clickable file:line:col links
func InferUnionDiscriminators(ctx context.Context, types []*ast.TypeDef, schemaPath string, cfg map[string]any) {
	// Check if inference is disabled via config
	if inferUnionDiscriminators, ok := cfg["inferUnionDiscriminators"].(bool); ok && !inferUnionDiscriminators {
		return
	}

	logger := logging.From(ctx)
	if !env.DebugInferDiscriminators() {
		logger = &logging.ZapLogger{Logger: zap.NewNop()}
	}
	for _, typeDef := range types {
		if typeDef.Type != ast.DataTypeUnion {
			continue
		}

		infer_union_discriminator.ProcessUnionType(logger, typeDef, schemaPath)
	}
}
