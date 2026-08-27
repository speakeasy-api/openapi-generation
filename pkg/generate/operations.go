package generate

import (
	"context"
	"fmt"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi/openapi"
	config "github.com/speakeasy-api/sdk-gen-config"
)

type handleOpParams struct {
	OpID          string
	Scope         ast.Scope
	Op            *openapi.Operation
	RequestParams handleReqParams
}

type resolvedOperation struct {
	ContextStack ast.ContextStack
	Operation    *ast.BaseOperation
	GlobalParams ast.Fields
}

func (g *Generator) handleOperation(ctx context.Context, params handleOpParams) ([]resolvedOperation, error) {
	op := params.Op

	if op.GetResponses() == nil {
		return nil, errors.NewValidationError(fmt.Sprintf("operation %q has no responses", params.OpID), op.GetCore().Responses.GetKeyNodeOrRoot(op.GetRootNode()), nil)
	}

	params.RequestParams.Scope = params.Scope

	contextStack := ast.ContextStack{}

	tags, err := g.getOperationTags(ctx, op)
	if err != nil {
		return nil, err
	}

	// TODO should this step be forgotten for operations outside paths?
	for _, tag := range tags {
		contextStack = append(contextStack, ast.ContextFrame{
			Type:       ast.ContextTypeOperationTag,
			Identifier: tag,
		})
	}

	contextStack.AppendOperation(params.OpID)

	reqs, err := g.handleRequests(ctx, contextStack, params.RequestParams)
	if err != nil {
		return nil, err
	}
	// If there are no requests, we still need to generate an operation
	if len(reqs) == 0 {
		reqs = []request{
			{
				UsesUserAgentHeader: false,
				Request:             nil,
			},
		}
	}

	operations := []resolvedOperation{}

	for _, r := range reqs {
		req := r.Request

		// Clone the operation, one for each serialization method
		opContextStack := make(ast.ContextStack, len(contextStack))
		copy(opContextStack, contextStack)

		id := params.OpID
		serializationMethodSuffix := ""
		if len(reqs) > 1 && req != nil && req.RequestBody != nil && req.RequestBody.SerializationMethod != nil {
			if !g.subsystem.Config.NameResolutionAtLeast(config.NameResolutionShortest) || *req.RequestBody.SerializationMethod != ast.SerializationMethodJSON {
				// Only add the serialization method suffix if it's not JSON
				id = fmt.Sprintf("%s_%s", id, *req.RequestBody.SerializationMethod)
				serializationMethodSuffix = fmt.Sprintf("_%s", *req.RequestBody.SerializationMethod)
			}
		}

		opContextStack.UpdateOperation(id)

		var serializationMethod *ast.SerializationMethod
		if req != nil && req.RequestBody != nil && req.RequestBody.SerializationMethod != nil {
			serializationMethod = req.RequestBody.SerializationMethod
		}

		opOut := ast.BaseOperation{
			ID:                        id,
			OriginalID:                params.OpID,
			UsesUserAgentHeader:       r.UsesUserAgentHeader,
			Request:                   req,
			SerializationMethod:       serializationMethod,
			SerializationMethodSuffix: serializationMethodSuffix,
		}

		res, err := g.handleResponses(ctx, op.Responses, opContextStack, handleResponsesParams{
			Scope:      params.Scope,
			Pagination: params.RequestParams.Pagination,
			ErrsExt:    params.RequestParams.Errors,
			DocInfo:    params.RequestParams.DocInfo,
		})
		if err != nil {
			return nil, err
		}
		opOut.Response = res

		operations = append(operations, resolvedOperation{
			ContextStack: opContextStack,
			Operation:    &opOut,
			GlobalParams: r.GlobalParams,
		})
	}

	if len(operations) == 0 {
		return nil, nil
	}

	return operations, nil
}

func (g *Generator) getOperationTags(ctx context.Context, op *openapi.Operation) ([]string, error) {
	groups, err := g.subsystem.Extensions.GetGroups(op)
	if err != nil {
		return nil, err
	}

	tags := op.Tags

	if groups != nil {
		tags = groups
		g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureGroups)
	}

	return tags, nil
}
