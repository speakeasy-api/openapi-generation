package ast_post_processing

import (
	"context"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/ast_post_processing/pre_apply_union_discriminators"
	"github.com/speakeasy-api/openapi-generation/v2/internal/env"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"go.uber.org/zap"
)

// PreApplyUnionDiscriminators sets const values on fields that are always used with the same discriminator value
// This simplifies the discriminator into a const field value when a type is consistently used across unions
func PreApplyUnionDiscriminators(ctx context.Context, a *ast.AST, cfg map[string]any) {
	logger := logging.From(ctx)
	if !env.DebugPreApplyUnionDiscriminators() {
		logger = &logging.ZapLogger{Logger: zap.NewNop()}
	}

	preApplyUnionDiscriminators := false
	if preApplyUnionDiscriminatorsConfig, ok := cfg["preApplyUnionDiscriminators"].(bool); ok {
		preApplyUnionDiscriminators = preApplyUnionDiscriminatorsConfig
	}

	logger.Debug("PreApplyUnionDiscriminators started")

	tracker := pre_apply_union_discriminators.NewTracker(logger, preApplyUnionDiscriminators)

	for operation := range a.MainSDK.WalkOperations() {
		for typeDef := range OperationRequestTypes(operation) {
			tracker.WalkType(typeDef)
		}

		for typeDef := range OperationResponseTypes(operation) {
			tracker.WalkType(typeDef)
		}
	}

	for _, ops := range a.Webhooks.All() {
		for _, operation := range ops {
			for typeDef := range OperationRequestTypes(&operation) {
				tracker.WalkType(typeDef)
			}

			for typeDef := range OperationResponseTypes(&operation) {
				tracker.WalkType(typeDef)
			}
		}
	}

	tracker.ApplyDiscriminators()
}
