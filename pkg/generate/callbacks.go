package generate

import (
	"context"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/internal/resolution"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	oas "github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/sequencedmap"
	"go.uber.org/zap"
)

type handleCallbacksParams struct {
	GlobalNameOverrides []*extensions.NameOverride
	DocInfo             *document.DocumentInfo
}

func (g *Generator) handleCallbacks(ctx context.Context, parentName string, callbacks *sequencedmap.Map[string, *oas.ReferencedCallback], params handleCallbacksParams) ([]*ast.TypeDef, error) {
	callbackTypes := []*ast.TypeDef{}

	if callbacks == nil {
		return callbackTypes, nil
	}

	for id, cr := range callbacks.AllOrdered(g.subsystem.Config.GetSequencedMapIterationOrder()) {
		cb, err := resolution.Resolve(ctx, cr, params.DocInfo)
		if err != nil {
			return nil, err
		}

		exts := cb.GetExtensions()

		ignore, err := g.subsystem.Extensions.Ignore(exts)
		if err != nil {
			return nil, err
		}
		if ignore {
			g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureIgnores)
			continue
		}

		callbackErrors, err := g.subsystem.Extensions.HandleErrors(exts)
		if err != nil {
			return nil, err
		}
		if callbackErrors != nil {
			g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureErrors)
		}

		if cb.Len() > 1 {
			return nil, errors.NewUnsupportedError("multiple path items in callback are not supported", cb.GetRootNode())
		}

		for _, pi := range cb.All() {
			item, err := resolution.Resolve(ctx, pi, params.DocInfo)
			if err != nil {
				return nil, err
			}

			exts := item.GetExtensions()

			ignore, err := g.subsystem.Extensions.Ignore(exts)
			if err != nil {
				return nil, err
			}
			if ignore {
				g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureIgnores)
				continue
			}

			pathItemErrors, err := g.subsystem.Extensions.HandleErrors(exts)
			if err != nil {
				return nil, err
			}
			if pathItemErrors != nil {
				g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureErrors)
			}

			pathItemErrors = extensions.MergeErrors(callbackErrors, pathItemErrors)

			for _, op := range item.AllOrdered(g.subsystem.Config.GetSequencedMapIterationOrder()) {
				exts := op.GetExtensions()

				ignore, err := g.subsystem.Extensions.Ignore(exts)
				if err != nil {
					return nil, err
				}
				if ignore {
					g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureIgnores)
					continue
				}

				opErrors, err := g.subsystem.Extensions.HandleErrors(exts)
				if err != nil {
					return nil, err
				}
				if opErrors != nil {
					g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureErrors)
				}

				opErrors = extensions.MergeErrors(pathItemErrors, opErrors)

				opID := op.GetOperationID()
				if opID == "" {
					opID = id
				}
				opID = parentName + "_" + opID

				ctx := logging.WithFields(ctx, zap.String("operation", opID), zap.String("type", "callbacks"))
				operations, err := g.handleOperation(ctx, handleOpParams{
					OpID:  opID,
					Op:    op,
					Scope: ast.ScopeCallbacks,
					RequestParams: handleReqParams{
						RequestBody:         op.RequestBody,
						PathItemParameters:  item.Parameters,
						OpID:                opID,
						Op:                  op,
						GlobalNameOverrides: params.GlobalNameOverrides,
						Errors:              opErrors,
						DocInfo:             params.DocInfo,
					},
				})
				if err != nil {
					if g.skip(ctx, errors.SkippedEntity{Name: opID, Type: "operation"}, err) {
						continue
					}
					return nil, err
				}

				for _, operation := range operations {
					if operation.Operation.Request != nil {
						callbackTypes = append(callbackTypes, operation.Operation.Request.Field.Type)
					}
					callbackTypes = append(callbackTypes, operation.Operation.Response.Type)
				}

				if !g.validationOnly {
					logging.From(ctx).Debug("operation valid")
				}
			}
		}
	}

	if len(callbackTypes) > 0 {
		g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureCallbacks)
	}

	return callbackTypes, nil
}
