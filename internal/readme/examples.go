package readme

import (
	"slices"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
)

type opPredicate func(*ast.SDK, *ast.Operation) bool

// OperationsForUsageSnippets returns usage example operations that match any of
// the given criteria.
//
// The operationId will select any operation with a matching operationId as an
// example.
//
// The namespace will select all operations within a given namespace.
//
// The rootExample will select the first main usage example.
//
// The all flag will select all available usage examples, overriding the operationId and namespace parameters.
func OperationsForUsageSnippets(
	s *ast.SDK,
	operationIds []string,
	namespace string,
	rootExample bool,
	all bool,
	exampleRequestBodyJSON string,
	paramToExampleValues map[string]string,
) ([]ast.UsageContext, error) {
	var operations []ast.UsageContext
	var scopes []ast.UsageExampleScope

	if rootExample {
		if example := GrabDefaultUsageExample(s); example != nil {
			operations = append(operations, *example)
		}
	}
	if all {
		allPredicate := func(_ *ast.SDK, _ *ast.Operation) bool {
			return true
		}

		preferred, extra := selectExampleOperationsRecurse(s, scopes, allPredicate, 0, false, false)
		operations = append(operations, preferred...)
		operations = append(operations, extra...)
	} else {
		if len(operationIds) > 0 {
			idPredicate := func(_ *ast.SDK, operation *ast.Operation) bool {
				return slices.Contains(operationIds, operation.ID)
			}

			preferred, extra := selectExampleOperationsRecurse(s, scopes, idPredicate, 0, false, false)

			// If an explicit list of operationIds was provided, then we want to return
			// all of them, ignoring the preferred/extra distinction.
			operations = append(operations, preferred...)
			operations = append(operations, extra...)
		}
		if namespace != "" {
			nsPredicate := func(sdk *ast.SDK, operation *ast.Operation) bool {
				groupFrames := sdk.Type.ContextStack.GetGroups()

				groups := make([]string, len(groupFrames))
				for i, frame := range groupFrames {
					groups[i] = frame.Identifier
				}

				originalGroupIdentifier := ""
				if len(groups) > 0 {
					originalGroupIdentifier = strings.Join(groups, ".") + "."
				}

				originalGroupIdentifier += sdk.Type.Name
				return originalGroupIdentifier == namespace
			}

			preferred, extra := selectExampleOperationsRecurse(s, scopes, nsPredicate, 0, false, false)
			operations = append(operations, preferred...)
			operations = append(operations, extra...)
		}
	}

	if len(operations) == 0 {
		return nil, errors.ErrGeneration.Wrap(errors.New("no operations found"))
	}

	// If example overrides were provided, populate them into the operation examples
	if exampleRequestBodyJSON != "" || len(paramToExampleValues) > 0 {
		for _, op := range operations {
			if slices.Contains(operationIds, op.Operation.ID) {
				example, err := ast.NewExampleFromJSON("example-override", "Example override to use for usage snippet generation", exampleRequestBodyJSON)
				if err != nil {
					return nil, errors.ErrGeneration.Wrap(err)
				}

				// We override the Examples array rather than appending because otherwise the usage snippet
				// defaults to using the Faker example data.
				op.Operation.Request.Examples = []*ast.Example{example}

				if len(paramToExampleValues) == 0 {
					continue
				}

				// Now override the examples for any params that were passed in.
				allParams := append(op.Operation.Request.Params.QueryParams, op.Operation.Request.Params.PathParams...)
				allParams = append(allParams, op.Operation.Request.Params.HeaderParams...)
				for _, param := range allParams {
					paramName := param.Field.Name
					if val, ok := paramToExampleValues[paramName]; ok {
						example, err := ast.NewExampleFromString("example-override", "Example override to use for usage snippet generation", val)
						if err != nil {
							return nil, errors.ErrGeneration.Wrap(err)
						}
						param.Examples = []*ast.Example{example}
					}
				}
			}
		}
	}

	return operations, nil
}

// GrabDefaultUsageExample finds the first main usage example.
func GrabDefaultUsageExample(s *ast.SDK) *ast.UsageContext {
	ues := SelectExampleOperations(s, []ast.UsageExampleScope{}, 1, true, false)
	if len(ues) > 0 {
		return &ues[0]
	}
	return nil
}

// SelectExampleOperations first collects a pool of "possible" usage examples
// that satisfy all of the requirements associated with the `OpFilter` tags
// passed inside `scopes` (see getOperationPredicate for more details).
// If `scopes` is an empty slice, all available usage examples will be collected.
//
// Then, we look for "preferred" usage examples by checking the tags registered
// under the `x-speakeasy-usage-example` extension. An operation is considered
// preferred if it contains all the `OpFilter` tags passed inside `scopes`.
// If scopes is an empty slice, examples tagged as `usage` as well as all
// *untagged* examples will be grabbed instead.
//
// The limit parameter indicates that searching should stop after at least `limit`
// preferred examples have been found and caps the amount of usage examples returned.
// If not enough preferred examples were found compared to the given limit,
// additional operations will be picked from the pool of possible usage examples
// to try and reach the limit.
//
// Limit has a couple special values:
//   - if limit is set to -1, all *possible* examples will be returned.
//   - if limit is set to 0, all *preferred* examples will be returned. If none are
//     found (e.g. the x-speakeasy-usage-example extension hasn't been used at all),
//     then *one* of the possible usage examples will be returned.
func SelectExampleOperations(
	sdk *ast.SDK,
	scopes []ast.UsageExampleScope,
	limit int,
	isMainExample bool,
	shouldIncludeServerSelection bool,
) []ast.UsageContext {
	predicate := getOperationPredicate(sdk, scopes)
	preferred, extra := selectExampleOperationsRecurse(sdk, scopes, predicate, limit, isMainExample, shouldIncludeServerSelection)

	if limit < 0 {
		return append(preferred, extra...)
	}

	if limit == 0 {
		if len(preferred) > 0 {
			return preferred
		}

		if len(extra) > 0 {
			return extra[:1]
		}

		return []ast.UsageContext{}
	}

	toAdd := limit - len(preferred)
	if toAdd <= 0 {
		return preferred[:limit]
	}

	extra = deduplicateExampleOperations(extra)
	for i := 0; i < min(toAdd, len(extra)); i++ {
		preferred = append(preferred, extra[i])
	}

	return preferred
}

func deduplicateExampleOperations(usages []ast.UsageContext) []ast.UsageContext {
	output := []ast.UsageContext{}
	usageMap := map[string]struct{}{}
	for _, usage := range usages {
		if _, ok := usageMap[usage.Operation.ID]; !ok {
			usageMap[usage.Operation.ID] = struct{}{}
			output = append(output, usage)
		}
	}
	return output
}

func selectExampleOperationsRecurse(
	sdk *ast.SDK,
	scopes []ast.UsageExampleScope,
	predicate opPredicate,
	limit int,
	isMainExample bool,
	shouldIncludeServerSelection bool,
) ([]ast.UsageContext, []ast.UsageContext) {
	possible := getPossibleExampleOperations(sdk, scopes, predicate, isMainExample, shouldIncludeServerSelection)
	preferred, extra := selectPreferredExampleOperations(possible, limit)

	for _, subSDK := range sdk.SubSDKs {
		if limit > 0 && len(preferred) >= limit {
			return preferred, extra
		}
		subPreferred, subExtra := selectExampleOperationsRecurse(
			subSDK,
			scopes,
			predicate,
			limit,
			isMainExample,
			shouldIncludeServerSelection,
		)

		preferred = append(preferred, subPreferred...)
		extra = append(extra, subExtra...)
	}

	preferred = deduplicateExampleOperations(preferred)

	return preferred, extra
}

func selectPreferredExampleOperations(
	usages []ast.UsageContext,
	limit int,
) ([]ast.UsageContext, []ast.UsageContext) {
	if limit < 0 {
		return usages, []ast.UsageContext{}
	}

	preferred := []ast.UsageContext{}
	extra := []ast.UsageContext{}
	for _, usage := range usages {
		if isPreferredOperation(usage.Operation, usage.Scopes) {
			preferred = append(preferred, usage)
		} else if usage.Operation.Comments == nil || !usage.Operation.Comments.Deprecated {
			extra = append(extra, usage)
		}
	}

	return preferred, extra
}

const sdkExampleUsageTag = "usage"

func isPreferredOperation(op *ast.Operation, scopes []ast.UsageExampleScope) bool {
	if op.Extensions.UsageExample == nil {
		return false
	}

	// Filter out scopes with empty OpFilter, which are not valid in this context
	filteredScopes := []ast.UsageExampleScope{}
	for _, scope := range scopes {
		if scope.OpFilter != "" {
			filteredScopes = append(filteredScopes, scope)
		}
	}

	tags := op.Extensions.UsageExample.Tags
	if len(filteredScopes) == 0 {
		return op.Extensions.UsageExample != nil &&
			(len(tags) == 0 || slices.Contains(tags, sdkExampleUsageTag))
	}

	ret := true
	for _, scope := range filteredScopes {
		ret = ret && slices.Contains(tags, scope.OpFilter)
	}
	return ret
}

func getOperationPredicate(
	sdk *ast.SDK,
	scopes []ast.UsageExampleScope,
) opPredicate {
	return func(s *ast.SDK, op *ast.Operation) bool {
		ret := true
		for _, scope := range scopes {
			currentScope := scope
			negate := false
			if strings.HasPrefix(scope.OpFilter, "!") {
				currentScope.OpFilter = strings.TrimPrefix(scope.OpFilter, "!")
				negate = true
			}

			pred := scopeOpPredicate(sdk, currentScope)

			if negate {
				ret = ret && !pred(s, op)
			} else {
				ret = ret && pred(s, op)
			}
		}
		return ret
	}
}

func scopeOpPredicate(
	sdk *ast.SDK,
	scope ast.UsageExampleScope,
) opPredicate {
	switch scope.OpFilter {
	case sdkExampleUsageTag:
		fallthrough
	case "all":
		return func(*ast.SDK, *ast.Operation) bool { return true }
	case "pagination":
		return paginationOpPredicate
	case "file-upload":
		return FileUploadOpPredicate
	case "global-parameters":
		return globalParameterOpPredicate(sdk)
	case "errors":
		return errorOpPredicate
	case "server":
		return serverOpPredicate(scope.IsGlobal)
	case "security":
		return securityOpPredicate(scope.IsGlobal)
	case "retries":
		return retriesOpPredicate
	case "eventstream":
		return eventStreamOpPredicate
	case "jsonl":
		return jsonlOpPredicate
	case "get-request":
		return getRequestPredicate
	case "post-request":
		return postRequestPredicate
	default:
		// UNKNOWN OPERATION: should it be an error?
		return func(*ast.SDK, *ast.Operation) bool { return false }
	}
}

func paginationOpPredicate(
	_ *ast.SDK,
	op *ast.Operation,
) bool {
	return op.Extensions.Pagination != nil
}

func FileUploadOpPredicate(
	_ *ast.SDK,
	op *ast.Operation,
) bool {
	if op.Request == nil || op.Request.RequestBody == nil {
		return false
	}

	if op.Request.RequestBody.Type.Type == ast.DataTypeBytes || op.Request.RequestBody.Type.Type == ast.DataTypeRequestStream {
		return true
	}

	if op.SerializationMethod == nil || *op.SerializationMethod != ast.SerializationMethodMultipart {
		return false
	}

	for _, field := range op.Request.RequestBody.Type.Fields {
		ann := field.Annotations.Get(ast.AnnotationTypeMultipartForm)
		if ann == nil {
			continue
		}

		mf, ok := ann.(*ast.MultipartFormAnnotation)
		if ok && mf.File {
			return true
		}
	}

	return false
}

func eventStreamOpPredicate(
	_ *ast.SDK,
	op *ast.Operation,
) bool {
	for _, resp := range op.Response.Responses {
		if resp.Error {
			continue
		}

		for _, c := range resp.Content {
			if c.Content == nil {
				continue
			}

			if c.SerializationMethod == string(ast.SerializationMethodEventStream) {
				return true
			}
		}
	}

	return false
}

func jsonlOpPredicate(
	_ *ast.SDK,
	op *ast.Operation,
) bool {
	for _, resp := range op.Response.Responses {
		if resp.Error {
			continue
		}

		for _, c := range resp.Content {
			if c.Content == nil {
				continue
			}

			if c.SerializationMethod == string(ast.SerializationMethodJsonL) {
				return true
			}
		}
	}

	return false
}

func retriesOpPredicate(
	_ *ast.SDK,
	op *ast.Operation,
) bool {
	if op.Extensions.Retries == nil {
		return false
	}

	if op.Extensions.Retries.Disabled != nil {
		return !*op.Extensions.Retries.Disabled
	}

	return true
}

func errorOpPredicate(
	_ *ast.SDK,
	op *ast.Operation,
) bool {
	for _, resp := range op.Response.Responses {
		if resp.Error && len(resp.Content) > 0 {
			return true
		}
	}
	return false
}

func securityOpPredicate(isGlobal bool) opPredicate {
	return func(_ *ast.SDK, op *ast.Operation) bool {
		if isGlobal {
			return op.Security == nil
		}
		return op.Security != nil && len(op.Security.Type.Fields) > 0
	}
}

func serverOpPredicate(isGlobal bool) opPredicate {
	return func(_ *ast.SDK, op *ast.Operation) bool {
		hasOperationServers := op.Servers != nil && len(op.Servers.Servers) > 0
		if !isGlobal {
			return hasOperationServers
		}
		// For global server examples, exclude operations that have their own
		// servers since the global server URL won't apply to them.
		return !hasOperationServers
	}
}

func globalParameterOpPredicate(
	sdk *ast.SDK,
) opPredicate {
	if sdk.Globals == nil {
		return func(*ast.SDK, *ast.Operation) bool { return false }
	}

	globalNames := make(map[string]struct{}, len(sdk.Globals.Fields))
	for _, param := range sdk.Globals.Fields {
		globalNames[param.Name] = struct{}{}
	}

	return func(_ *ast.SDK, op *ast.Operation) bool {
		if op.Request == nil {
			return false
		}

		if op.Request.Params == nil {
			return false
		}

		if op.Request.Params.QueryParams != nil {
			for _, param := range op.Request.Params.QueryParams {
				if _, ok := globalNames[param.Field.Name]; ok {
					return true
				}
			}
		}

		if op.Request.Params.PathParams != nil {
			for _, param := range op.Request.Params.PathParams {
				if _, ok := globalNames[param.Field.Name]; ok {
					return true
				}
			}
		}

		return false
	}
}

func getRequestPredicate(
	_ *ast.SDK,
	op *ast.Operation,
) bool {
	return strings.ToLower(op.Method) == "get" &&
		op.Response != nil &&
		len(op.Response.Responses) > 0
}

func postRequestPredicate(
	_ *ast.SDK,
	op *ast.Operation,
) bool {
	return strings.ToLower(op.Method) == "post"
}

func getPossibleExampleOperations(
	sdk *ast.SDK,
	scopes []ast.UsageExampleScope,
	pred opPredicate,
	isMainExample bool,
	shouldIncludeServerSelection bool,
) []ast.UsageContext {
	ucs := []ast.UsageContext{}
	for i := range sdk.Operations {
		op := sdk.Operations[i]
		if pred(sdk, op) {
			// safety for when it gets passed through to templates
			ue := op.Extensions.UsageExample
			if ue == nil {
				ue = &extensions.UsageExampleConfig{
					Tags: []string{},
				}
			}

			uc := ast.CreateUsageContext(sdk, op, ue, false)
			uc.Scopes = scopes
			uc.IsMainExample = isMainExample
			uc.PopulateGlobalParameterScopes(uc.ExampleName, shouldIncludeServerSelection)
			ucs = append(ucs, *uc)
		}
	}

	return ucs
}
