package openapi

import (
	"iter"
	"slices"

	"github.com/speakeasy-api/openapi-generation/v2/internal/configuration"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/sequencedmap"
)

func SortOperationsLikeLibOpenAPI(pathItem *openapi.PathItem, config *configuration.Config) iter.Seq2[openapi.HTTPMethod, *openapi.Operation] {
	// If we aren't "trying" to maintain openapi order just use the specified order
	if config.GetSequencedMapIterationOrder() != sequencedmap.OrderAdded {
		return chainAdditionalOperations(pathItem, pathItem.AllOrdered(config.GetSequencedMapIterationOrder()), config)
	}

	// The below code matches the implementation of libopenapi https://github.com/pb33f/libopenapi/blob/main/datamodel/high/v3/path_item.go
	type op struct {
		method openapi.HTTPMethod
		op     *openapi.Operation
		line   int
	}

	getLine := func(method openapi.HTTPMethod, defaultLine int) int {
		node := pathItem.GetCore().GetMapKeyNodeOrRoot(string(method), nil)
		if node == nil {
			return defaultLine
		}
		return node.Line
	}

	ops := []op{}

	// Add items in order they appear but if the line number is not found (as maybe they were added dynamically) then order using a decreasing index
	// This is just matching the implementation in libopenapi to the letter to ensure we don't introduce churn between our library and theirs
	if pathItem.Get() != nil {
		ops = append(ops, op{method: openapi.HTTPMethodGet, op: pathItem.Get(), line: getLine(openapi.HTTPMethodGet, -8)})
	}

	if pathItem.Put() != nil {
		ops = append(ops, op{method: openapi.HTTPMethodPut, op: pathItem.Put(), line: getLine(openapi.HTTPMethodPut, -7)})
	}

	if pathItem.Post() != nil {
		ops = append(ops, op{method: openapi.HTTPMethodPost, op: pathItem.Post(), line: getLine(openapi.HTTPMethodPost, -6)})
	}

	if pathItem.Delete() != nil {
		ops = append(ops, op{method: openapi.HTTPMethodDelete, op: pathItem.Delete(), line: getLine(openapi.HTTPMethodDelete, -5)})
	}

	if pathItem.Options() != nil {
		ops = append(ops, op{method: openapi.HTTPMethodOptions, op: pathItem.Options(), line: getLine(openapi.HTTPMethodOptions, -4)})
	}

	if pathItem.Head() != nil {
		ops = append(ops, op{method: openapi.HTTPMethodHead, op: pathItem.Head(), line: getLine(openapi.HTTPMethodHead, -3)})
	}

	if pathItem.Patch() != nil {
		ops = append(ops, op{method: openapi.HTTPMethodPatch, op: pathItem.Patch(), line: getLine(openapi.HTTPMethodPatch, -2)})
	}

	if pathItem.Trace() != nil {
		ops = append(ops, op{method: openapi.HTTPMethodTrace, op: pathItem.Trace(), line: getLine(openapi.HTTPMethodTrace, -1)})
	}

	if pathItem.Query() != nil {
		ops = append(ops, op{method: openapi.HTTPMethodQuery, op: pathItem.Query(), line: getLine(openapi.HTTPMethodQuery, 0)})
	}

	// Include additionalOperations (OpenAPI 3.2+ custom HTTP methods)
	// Use a high default line so they sort after standard operations when line numbers are unavailable.
	if additionalOps := pathItem.GetAdditionalOperations(); additionalOps != nil {
		defaultLine := 1<<31 - 1 - additionalOps.Len() // math.MaxInt32 minus count to prevent overflow
		for methodName, operation := range additionalOps.All() {
			if operation != nil {
				ops = append(ops, op{method: openapi.HTTPMethod(methodName), op: operation, line: getLine(openapi.HTTPMethod(methodName), defaultLine)})
				defaultLine++
			}
		}
	}

	sortedOps := sequencedmap.New[openapi.HTTPMethod, *openapi.Operation]()

	slices.SortStableFunc(ops, func(a op, b op) int {
		return a.line - b.line
	})

	for _, op := range ops {
		sortedOps.Set(op.method, op.op)
	}

	return sortedOps.All()
}

// chainAdditionalOperations wraps an iterator of standard operations to also
// yield any additionalOperations (OpenAPI 3.2+ custom HTTP methods) from the path item.
// The additional operations are iterated using the same configured order as standard operations.
func chainAdditionalOperations(pathItem *openapi.PathItem, standardOps iter.Seq2[openapi.HTTPMethod, *openapi.Operation], config *configuration.Config) iter.Seq2[openapi.HTTPMethod, *openapi.Operation] {
	additionalOps := pathItem.GetAdditionalOperations()
	if additionalOps == nil {
		return standardOps
	}

	return func(yield func(openapi.HTTPMethod, *openapi.Operation) bool) {
		for method, op := range standardOps {
			if !yield(method, op) {
				return
			}
		}
		for methodName, op := range additionalOps.AllOrdered(config.GetSequencedMapIterationOrder()) {
			if op != nil {
				if !yield(openapi.HTTPMethod(methodName), op) {
					return
				}
			}
		}
	}
}

func GetLibOpenAPIExplodeDefault(p *openapi.Parameter) bool {
	if p.GetStyle() != openapi.SerializationStyleDeepObject || p.Explode != nil {
		return p.GetExplode()
	}

	// LibOpenAPI defaults to true for deepObject style parameters where explode is not actually relevant for deepObject style and should default to false but retaining behavior for backwards compatibility https://spec.openapis.org/oas/v3.1.1.html#parameter-explode
	return true
}
