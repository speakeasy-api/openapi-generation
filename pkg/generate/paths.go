package generate

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"sync/atomic"

	"github.com/speakeasy-api/openapi-generation/v2/internal/analytics"
	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi-generation/v2/internal/env"
	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/internal/licensing"
	"github.com/speakeasy-api/openapi-generation/v2/internal/openapi"
	"github.com/speakeasy-api/openapi-generation/v2/internal/resolution"
	"github.com/speakeasy-api/openapi-generation/v2/internal/security"
	"github.com/speakeasy-api/openapi-generation/v2/internal/utils"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	oas "github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/sequencedmap"
	"go.uber.org/zap"
)

type handlePathsParams struct {
	Paths               *oas.Paths
	SecuritySchemes     *sequencedmap.Map[string, *oas.ReferencedSecurityScheme]
	GlobalNameOverrides []*extensions.NameOverride
	GlobalRetries       *extensions.Retries
	GlobalTimeout       *int64
	AnalyticsData       *analytics.Data
	GlobalSecurity      *ast.Security
	Globals             *ast.TypeDef
	MaxMethodParams     *int
	FlattenRequest      *bool
	DocInfo             *document.DocumentInfo
	IsWebhooks          bool
	WebhookSecurity     *extensions.WebhookSecurity
}

const cliRetryMethodPolicySafeMethods = "safe-methods"

func inheritedRetriesFilteredByMethod(target string, retryMethodPolicy string) bool {
	return target == "cli" && retryMethodPolicy == cliRetryMethodPolicySafeMethods
}

func (g *Generator) retryMethodPolicy() string {
	retryMethodPolicy, _ := g.subsystem.Config.GetLanguageConfigValue("retryMethodPolicy").(string)
	return retryMethodPolicy
}

func resolveMethodRetries(target string, retryMethodPolicy string, httpMethod oas.HTTPMethod, globalRetries, operationRetries *extensions.Retries, overrideGlobalRetries bool) *extensions.Retries {
	methodRetries := globalRetries
	if overrideGlobalRetries {
		return operationRetries
	}

	if !inheritedRetriesFilteredByMethod(target, retryMethodPolicy) {
		return methodRetries
	}

	switch {
	case httpMethod.Is(http.MethodGet),
		httpMethod.Is(http.MethodHead),
		httpMethod.Is(http.MethodOptions),
		httpMethod.Is(http.MethodTrace):
		return methodRetries
	default:
		return nil
	}
}

func (g *Generator) handlePaths(ctx context.Context, params handlePathsParams) (*sequencedmap.Map[string, []ast.Operation], error) {
	if params.Paths == nil {
		return sequencedmap.New[string, []ast.Operation](), nil
	}
	retryMethodPolicy := g.retryMethodPolicy()

	pathsErrors, err := g.subsystem.Extensions.HandleErrors(params.Paths.GetExtensions())
	if err != nil {
		return nil, err
	}
	if pathsErrors != nil {
		g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureErrors)
	}

	usesUserAgentHeader := false
	operations := sequencedmap.New[string, []ast.Operation]()
	globalSecurityMatcher := security.NewGlobalSecurityMatcher(params.GlobalSecurity)

	for path, pi := range params.Paths.AllOrdered(g.subsystem.Config.GetSequencedMapIterationOrder()) {
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

		pathItemErrors = extensions.MergeErrors(pathsErrors, pathItemErrors)

		pathServers, err := g.handleServers(ctx, item.Servers, ast.ScopeOperations)
		if err != nil {
			if g.skip(ctx, errors.SkippedEntity{Name: path, Type: "path"}, err) {
				continue
			}
			return nil, err
		}

		if params.AnalyticsData.ServerURL == "" && pathServers != nil {
			params.AnalyticsData.ServerURL = pathServers.GetDefaultURL(false)

			if params.AnalyticsData.ServerURL == "" && len(pathServers.Servers) > 0 {
				params.AnalyticsData.ServerURL = pathServers.Servers[0].URL
			}
		}

		// TODO: I hate that we have to do this but libopenapi doesn't actually maintain the document order of operations so we need to sort things the same way libopenapi does to avoid introducing unnecessary diffs
		// We should look to remove this in the future when we can justify the order changes be reintroducing the below line for iteration instead
		// for httpMethod, op := range item.AllOrdered(g.subsystem.Config.GetSequencedMapIterationOrder()) {
		for httpMethod, op := range openapi.SortOperationsLikeLibOpenAPI(item, g.subsystem.Config) {
			exts := op.GetExtensions()

			if shouldSkipOperationBasedOnEnvDebugFilter(op.GetOperationID()) {
				continue
			}

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

			if op.RequestBody != nil && httpMethod.Is(http.MethodGet) {
				if !g.subsystem.Features.IsFeatureSupported(ctx, features.FeatureGetRequestBodies) {
					if g.skip(ctx, errors.SkippedEntity{Name: path, Type: "path"}, errors.NewUnsupportedError(fmt.Sprintf("request bodies are not supported for GET operations in %s clients", g.target), op.GetCore().RequestBody.GetKeyNodeOrRoot(op.GetRootNode()))) {
						continue
					}
				} else {
					g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureGetRequestBodies)
				}
			}

			opID := op.GetOperationID()
			if opID == "" {
				opID = fmt.Sprintf("%s_%s", httpMethod, path)
				if params.IsWebhooks && httpMethod == "post" {
					opID = path // Most webhooks are POSTs - so don't include in the fallback operation ID
				}
			}

			ctx := logging.WithFields(ctx, zap.String("operation", opID), zap.String("type", "paths"))

			var methodServers *ast.Servers

			if op.Servers != nil {
				methodServers, err = g.handleServers(ctx, op.GetServers(), ast.ScopeOperations)
				if err != nil {
					return nil, err
				}
			}

			if methodServers == nil {
				methodServers = pathServers
			}

			if methodServers != nil {
				g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureMethodServerURLs)
			}

			if params.AnalyticsData.ServerURL == "" && methodServers != nil {
				params.AnalyticsData.ServerURL = methodServers.GetDefaultURL(false)
			}

			var entityOperation *extensions.EntityOperationV1

			// Only process the entity operation extensions if target accepts
			// entity operations. If more targets accept entity operations in
			// the future, this could be changed into a feature supported check.
			if g.target.Target == "terraform" {
				var err error

				entityOperation, err = g.subsystem.Extensions.HandleEntityOperationExtension(op)
				if err != nil {
					return nil, err
				}
			}

			var pollingExt *extensions.Polling

			if g.subsystem.Features.IsFeatureSupported(ctx, features.FeatureOperationPolling) {
				pollingExt, err = g.subsystem.Extensions.HandlePollingExtension(op.Extensions)
				if err != nil {
					return nil, err
				}
			}

			pagination, err := g.subsystem.Extensions.HandleOperationPaginationExtension(ctx, op, params.DocInfo)
			if err != nil {
				return nil, err
			}

			if pagination != nil {
				switch {
				case pagination.Type == extensions.PaginationTypeURL && !g.subsystem.Features.IsFeatureSupported(ctx, features.FeatureURLBasedPagination):
					// If the URL-based pagination feature is not supported in this target, then we should ignore the pagination extension.
					pagination = nil
				case pollingExt != nil && (g.target.Target == "go" || g.target.Target == "terraform"):
					// The Go method templates hand polled operations a request-scoped
					// context, so a pagination Next closure cannot outlive the call and
					// is not emitted; drop the extension so the output matches.
					logging.LogWarning(ctx, fmt.Sprintf("%s extension is not supported alongside %s in %s clients, pagination is dropped for polled operations", extensions.ExtPagination.Name(), extensions.ExtPolling.Name(), g.target.Target), features.ErrUnsupportedFeature)
					pagination = nil
				default:
					g.subsystem.Features.RecordFeatureUsage(ctx, features.FeaturePagination)
				}
			}

			opRetries, overrideGlobalRetries, err := g.subsystem.Extensions.HandleOperationRetryExtension(op)
			if err != nil {
				return nil, err
			}

			var goOptionalMethodArguments *extensions.GoOptionalMethodArguments
			if g.target.Target == "go" {
				goOptionalMethodArguments, err = g.subsystem.Extensions.GetGoOptionalMethodArguments(op.GetExtensions())
				if err != nil {
					return nil, err
				}
			}

			var reactHook *extensions.ReactHook
			if !params.IsWebhooks && g.reactQueryHooksFeatureEnabled(ctx) {
				reactHook, err = g.subsystem.Extensions.HandleReactHookExtension(op)
				if err != nil {
					return nil, err
				}

				if reactHook == nil || !reactHook.Disabled {
					g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureReactQueryHooks)
				}
			}

			var timeout *int64
			if g.subsystem.Features.IsFeatureSupported(ctx, features.FeatureOperationTimeout) {
				localTimeout, overrideGlobalTimeout, err := g.subsystem.Extensions.HandleOperationTimeoutExtension(op)
				if err != nil {
					return nil, err
				}

				timeout = params.GlobalTimeout
				if overrideGlobalTimeout {
					timeout = localTimeout
				}

				if timeout != nil {
					g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureOperationTimeout)
				}
			}

			sseOverload, err := g.determineSSEOverload(ctx, op, params.DocInfo)
			if err != nil {
				return nil, err
			}

			docsRateLimits, err := g.subsystem.Extensions.HandleDocsRateLimitExtension(op)
			if err != nil {
				return nil, err
			}

			methodRetries := resolveMethodRetries(g.target.Target, retryMethodPolicy, httpMethod, params.GlobalRetries, opRetries, overrideGlobalRetries)
			if methodRetries != nil {
				g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureRetries)
			}

			scope := ast.ScopeOperations
			if params.IsWebhooks {
				scope = ast.ScopeWebhooks
			}

			resolvedOperations, err := g.handleOperation(ctx, handleOpParams{
				OpID:  opID,
				Op:    op,
				Scope: scope,
				RequestParams: handleReqParams{
					RequestBody:         op.GetRequestBody(),
					SecuritySchemes:     params.SecuritySchemes,
					Servers:             methodServers,
					Retries:             methodRetries,
					Timeout:             timeout,
					DocsRateLimits:      docsRateLimits,
					Pagination:          pagination,
					PathItemParameters:  item.Parameters,
					OpID:                opID,
					Op:                  op,
					GlobalNameOverrides: params.GlobalNameOverrides,
					Globals:             params.Globals,
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

			for _, resolvedOp := range resolvedOperations {
				operation := resolvedOp.Operation

				if operation.UsesUserAgentHeader {
					usesUserAgentHeader = true
				}

				isUsageExample, err := g.subsystem.Extensions.IsUsageExample(op.GetExtensions())
				if err != nil {
					return nil, err
				}

				var usageExample *extensions.UsageExampleConfig
				if isUsageExample {
					usageExample, err = g.subsystem.Extensions.GetUsageConfig(op.GetExtensions())
					if err != nil {
						return nil, err
					}
					g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureExamples)
				}

				opMethodName, overrideGlobalMethodName, err := g.subsystem.Extensions.HandleOperationMethodNameExtension(op)
				if err != nil {
					return nil, err
				}

				publicExports, err := g.subsystem.Extensions.HandlePublicExportsExtension(op.GetExtensions())
				if err != nil {
					return nil, err
				}

				methodNameOverride := ""
				for _, globalNameOverride := range params.GlobalNameOverrides {
					if globalNameOverride.OperationId != "" {
						re := regexp.MustCompile(globalNameOverride.OperationId)
						if re.MatchString(opID) {
							methodNameOverride = globalNameOverride.GlobalMethodNameOverride
						}
					}
				}
				if overrideGlobalMethodName {
					methodNameOverride = opMethodName.Name
					g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureNameOverrides)
				}

				callbacks, err := g.handleCallbacks(ctx, opID, op.GetCallbacks(), handleCallbacksParams{
					GlobalNameOverrides: params.GlobalNameOverrides,
					DocInfo:             params.DocInfo,
				})
				if err != nil {
					return nil, err
				}

				var globalSecurity *ast.FieldDef
				var operationSecurity *ast.FieldDef
				var oauth2Config *ast.OAuth2Config
				var hoistConfig *ast.HoistedSecurityConfig

				if params.GlobalSecurity != nil {
					globalSecurity = params.GlobalSecurity.Security
				}

				security, err := g.handleSecurity(ctx, params.DocInfo, resolvedOp.ContextStack, op.Security, params.SecuritySchemes, ast.SecOptReasonNotOptional, params.GlobalSecurity, globalSecurityMatcher, ast.ScopeOperations, false)
				if err != nil {
					return nil, err
				}

				if security != nil {
					if security.Security != nil {
						// Disable global security if method overrides it
						globalSecurity = nil
						operationSecurity = security.Security
						g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureMethodSecurity)
					} else if !security.SecurityConfig.IsSubsetOfGlobalSecurity() {
						// Operation explicitly declared security but it resolved to
						// nothing (disabled via security: [] or optional via security: [{}]).
						// Clear global security so usage examples don't show credentials.
						// Note: matching global security (where operation schemes match global)
						// must NOT clear globalSecurity, as the operation still uses it.
						globalSecurity = nil
					}

					// Always override the OAuth2 configuration when operation-level
					// security is explicitly defined. This includes security matching
					// global with scopes, as well as disabled (security: []) or optional
					// (security: [{}]) blocks where OAuth2Config is empty. Using
					// the explicit config prevents the template from falling back
					// to global OAuth2 scopes for operations that have opted out.
					oauth2Config = &security.SecurityConfig.OAuth2Config

					hoistConfig = security.SecurityConfig.HoistedSecurityConfig
				}

				if operation.Request != nil {
					// Remove request if it has no fields
					if operation.Request.Params == nil && operation.Request.RequestBody == nil {
						operation.Request = nil
					}
				}

				operationComments, err := g.handleOperationComments(ctx, op)
				if err != nil {
					return nil, err
				}

				isTestingEnabled, err := g.subsystem.Extensions.IsTestingEnabled(op.GetExtensions())
				if err != nil {
					return nil, err
				}

				var globals *ast.TypeDef
				if len(resolvedOp.GlobalParams) > 0 {
					globals = g.subsystem.Register.RegisterType(ctx, ast.NewType(&ast.TypeDef{
						Name:   resolvedOp.Operation.ID + "Globals",
						Type:   ast.DataTypeClass,
						Fields: resolvedOp.GlobalParams,
						Scope:  ast.ScopeOperations,
					}, ast.ContextStack{}), true)
				}

				maxMethodParams, err := g.getMaxMethodParams(ctx, params, op)
				if err != nil {
					return nil, err
				}

				requestFlattening, err := g.requestFlatteningEnabled(ctx, params, op, operation.Request)
				if err != nil {
					return nil, err
				}

				var mcp *extensions.MCP

				mcp, err = g.subsystem.Extensions.HandleMCPExtension(op)
				if err != nil {
					return nil, err
				}

				if g.mcpServerFeatureEnabled(ctx, handleMCPParams{
					isWebhooks:        params.IsWebhooks,
					globalSecurity:    globalSecurity,
					operationSecurity: operationSecurity,
				}) && (mcp == nil || !mcp.Disabled) {
					g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureMCPServer)
				}

				var polling *ast.Polling

				if pollingExt != nil {
					polling, err = operation.NewPollingFromExtensionsPolling(pollingExt)
					if err != nil {
						return nil, err
					}
				}

				entityMissingCodes, err := g.subsystem.Extensions.HandleEntityMissingCodesExtension(op.GetExtensions())
				if err != nil {
					return nil, err
				}

				if polling != nil {
					g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureOperationPolling)
				}

				operationExtensions := &ast.OperationExtensions{
					All:                       utils.ExtensionsToGoMap(op.GetExtensions()),
					DocsRateLimits:            docsRateLimits,
					EntityOperation:           entityOperation,
					EntityMissingCodes:        entityMissingCodes,
					GoOptionalMethodArguments: goOptionalMethodArguments,
					MCP:                       mcp,
					MethodNameOverride:        methodNameOverride,
					Pagination:                pagination,
					Polling:                   polling,
					PublicExports:             publicExports,
					ReactHook:                 reactHook,
					Retries:                   methodRetries,
					SSEOverload:               sseOverload,
					Timeout:                   timeout,
					UsageExample:              usageExample,
				}

				o := ast.Operation{
					BaseOperation:          *operation,
					Callbacks:              callbacks,
					Comments:               operationComments,
					Extensions:             operationExtensions,
					Globals:                globals,
					GlobalSecurity:         globalSecurity,
					HoistedSecurityConfig:  hoistConfig,
					Location:               &ast.OpenAPILocation{Node: op.GetRootNode()},
					MaxMethodParams:        maxMethodParams,
					Method:                 string(httpMethod),
					OAuth2Config:           oauth2Config,
					Path:                   path,
					Scope:                  ast.ScopeSDK,
					Security:               operationSecurity,
					Servers:                methodServers,
					Tags:                   op.GetTags(),
					TestExplicitlyDisabled: isTestingEnabled != nil && !*isTestingEnabled,
				}

				if params.IsWebhooks {
					o.Webhook = &ast.Webhook{
						Key:      path,
						Security: params.WebhookSecurity,
					}
					o.Security = nil
					o.GlobalSecurity = nil
					o.OAuth2Config = nil
				}

				// It's unlikely that the user wants an optional request body
				if o.Request != nil && o.Request.RequestBody != nil && o.Request.RequestBody.Optional {
					msg := fmt.Sprintf("request body is optional for operation: '%s' did you mean to add `required: true`?", fmt.Sprintf("%s %s", strings.ToUpper(o.Method), path))
					err := errors.NewValidationWarning(msg, op.GetCore().RequestBody.GetKeyNodeOrRoot(op.GetRootNode()), nil)
					logging.LogWarning(ctx, "validation warning", err)
				}

				fo := g.subsystem.Config.GetLanguageConfigValue("flatteningOrder")
				if fo == nil {
					fo = ""
				}

				flatteningOrder, ok := fo.(string)
				if !ok {
					return nil, fmt.Errorf("flatteningOrder must be a string, got %T", fo)
				}

				var paramsFirst bool
				switch flatteningOrder {
				case "parameters-first":
					paramsFirst = true
				case "", "body-first":
					paramsFirst = false
				default:
					return nil, fmt.Errorf("unrecognised flatteningOrder value: %v", flatteningOrder)
				}

				constFieldsAlwaysOptional := g.subsystem.Config.GetLanguageConfigValue("constFieldsAlwaysOptional")

				args, err := ast.NewArguments(&o, operationSecurity, ast.ArgumentsOptions{
					ParamFlatteningParamsFirst: paramsFirst,
					MaxMethodParams:            maxMethodParams,
					FlattenRequest:             requestFlattening,
					ConstFieldsAlwaysOptional:  toBoolPtr(constFieldsAlwaysOptional),
				})
				if args.Warning != nil {
					err := errors.NewValidationWarning(args.Warning.Error(), op.GetRootNode(), nil)
					logging.LogWarning(ctx, "validation warning", err)
				}

				if err != nil {
					return nil, err
				}
				o.Arguments = args

				// Record flattening usage
				if g.subsystem.Features.IsFeatureSupported(ctx, features.FeatureFlattening) {
					if maxMethodParams > 0 && o.Request != nil && o.Request.Field.Type.Type == ast.DataTypeClass && o.Request.Params != nil && len(o.Request.Field.Type.Fields) <= maxMethodParams {
						g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureFlattening)
					}
				}

				if args.Flattening == "all" || args.Flattening == "body" {
					g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureFlatRequests)
				}

				// Record acceptHeaders usage
				if g.subsystem.Features.IsFeatureSupported(ctx, features.FeatureAcceptHeaders) && len(o.GetAcceptTypes()) > 1 {
					g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureAcceptHeaders)
				}

				tags, err := g.getOperationTags(ctx, op)
				if err != nil {
					return nil, err
				}

				resolvedTags := []string{}

				if len(tags) == 0 {
					resolvedTags = append(tags, "")
				} else {
					for _, tag := range tags {
						if !slices.Contains(resolvedTags, tag) {
							resolvedTags = append(resolvedTags, tag)
						}
					}
				}

				for _, tag := range resolvedTags {
					ops, ok := operations.Get(tag)
					if !ok {
						ops = []ast.Operation{}
					}

					if g.subsystem.Config.MaintainOpenAPIOrder() {
						operations.Set(tag, append(ops, o))
					} else {
						operations.Set(tag, utils.AppendSorted(ops, o, func(i, j ast.Operation) bool {
							return i.GetID() >= j.GetID()
						}))
					}
				}
			}

			if !g.validationOnly {
				logging.From(ctx).Debug("operation valid")
			}
		}
	}

	if usesUserAgentHeader {
		for tag, ops := range operations.All() {
			for i := range ops {
				ops[i].UsesUserAgentHeader = true
				operations.Set(tag, ops)
			}
		}
	}

	return operations, nil
}

func (g *Generator) getMaxMethodParams(ctx context.Context, params handlePathsParams, op *oas.Operation) (int, error) {
	if !g.subsystem.Features.IsFeatureSupported(ctx, features.FeatureFlattening) {
		return 0, nil
	}

	// Attempt to set the MaxMethodParams based on a number of configurable sources
	// 1. The per language config in gen.yaml
	// 2. The global extension in the OpenAPI document
	// 3. The operation extension in the OpenAPI document
	maxMethodParams := 0
	defaultMaxMethodParams := 4
	configMaxMethodParams := g.subsystem.Config.GetLanguageConfigValue("maxMethodParams")
	if configMaxMethodParams != nil {
		switch mmp := configMaxMethodParams.(type) {
		case int:
			maxMethodParams = mmp
		case int64:
			maxMethodParams = int(mmp)
		default:
			return 0, fmt.Errorf("maxMethodParams must be an int or int64, got %s", reflect.TypeOf(mmp))
		}
		defaultMaxMethodParams = maxMethodParams
	}
	configMaxMethodParamsOverriden := false
	if params.MaxMethodParams != nil {
		maxMethodParams = *params.MaxMethodParams
		configMaxMethodParamsOverriden = true
	}
	opMaxMethodParams, err := g.subsystem.Extensions.GetMaxMethodParams(op.GetExtensions())
	if err != nil {
		return 0, err
	}
	if opMaxMethodParams != nil {
		maxMethodParams = *opMaxMethodParams
		configMaxMethodParamsOverriden = true
	}
	if configMaxMethodParamsOverriden && defaultMaxMethodParams != 4 {
		logging.From(ctx).Info(fmt.Sprintf("maxMethodParams from `gen.yaml` is overridden by the OpenAPI document with the %s extension", g.subsystem.Extensions.GetResolvedName(extensions.ExtMaxMethodParams)), zap.Int("maxMethodParams", maxMethodParams))
	}
	return maxMethodParams, nil
}

func (g *Generator) requestFlatteningEnabled(ctx context.Context, params handlePathsParams, op *oas.Operation, request *ast.Request) (bool, error) {
	if !g.subsystem.Features.IsFeatureSupported(ctx, features.FeatureFlattening) || !g.subsystem.Features.IsFeatureSupported(ctx, features.FeatureFlatRequests) {
		return false, nil
	}

	if request != nil && request.Field != nil && (request.Field.Optional || request.Field.Nullable) {
		// Disabling flattening for optional/nullable request bodies otherwise they can't be omitted
		return false, nil
	}

	// Attempt to set FlattenRequest based on a number of configurable sources
	// 1. The per language config in gen.yaml
	// 2. The global extension in the OpenAPI document
	// 3. The operation extension in the OpenAPI document
	configKey := "flattenRequests"
	flatten := false
	defaultFlatten := false
	configFlatten := g.subsystem.Config.GetLanguageConfigValue(configKey)
	if configFlatten != nil {
		switch f := configFlatten.(type) {
		case bool:
			flatten = f
		default:
			return false, fmt.Errorf("%s must be a boolean, got %s", configKey, reflect.TypeOf(f))
		}
		defaultFlatten = flatten
	}
	configOverriden := false
	if params.FlattenRequest != nil {
		flatten = *params.FlattenRequest
		configOverriden = true
	}
	opFlatten, err := g.subsystem.Extensions.GetFlattenRequest(op.GetExtensions())
	if err != nil {
		return false, err
	}
	if opFlatten != nil {
		flatten = *opFlatten
		configOverriden = true
	}
	if configOverriden && defaultFlatten != flatten {
		logging.From(ctx).Info(fmt.Sprintf("%s from `gen.yaml` is overridden by the OpenAPI document with the %s extension", configKey, g.subsystem.Extensions.GetResolvedName(extensions.ExtFlattenRequest)), zap.Bool(configKey, flatten))
	}
	return flatten, nil
}

func shouldSkipOperationBasedOnEnvDebugFilter(opID string) bool {
	opFilter := env.DebugOperationFilter()
	if opFilter == "" {
		return false
	}
	split := strings.Split(opFilter, ",")
	for _, f := range split {
		if opID == f {
			return false
		}
	}

	return true
}

func (g *Generator) reactQueryHooksFeatureEnabled(ctx context.Context) bool {
	if !g.subsystem.Features.IsFeatureSupported(ctx, features.FeatureReactQueryHooks) {
		return false
	}

	configFlag := g.subsystem.Config.GetLanguageConfigValue("enableReactQuery")
	if configFlag == nil {
		return false
	}
	if enabled, ok := configFlag.(bool); !ok || !enabled {
		return false
	}

	if !licensing.AccountHasFeatureAccess(ctx, features.FeatureReactQueryHooks) {
		return false
	}

	return true
}

var mcpFormatWarningLogged atomic.Bool

type handleMCPParams struct {
	isWebhooks        bool
	globalSecurity    *ast.FieldDef
	operationSecurity *ast.FieldDef
}

func (g *Generator) mcpServerFeatureEnabled(ctx context.Context, params handleMCPParams) bool {
	configFlag := g.subsystem.Config.GetLanguageConfigValue("enableMCPServer")
	if configFlag == nil {
		return false
	}
	if enabled, ok := configFlag.(bool); !ok || !enabled {
		return false
	}

	if params.isWebhooks {
		return false
	}

	// MCP tools can't call SDK methods with per-operation security at this time
	if params.operationSecurity != nil {
		logging.LogWarning(
			ctx,
			"mcp tool generation skipped",
			errors.New("per-operation security is not supported in MCP server generation"),
		)
		return false
	}

	if !g.subsystem.Features.IsFeatureSupported(ctx, features.FeatureMCPServer) {
		return false
	}

	format, err := g.getResponseFormat(ctx)
	if err != nil || format != responseFormatFlat {
		if mcpFormatWarningLogged.CompareAndSwap(false, true) {
			logging.LogWarning(
				ctx,
				"mcp server generation disabled",
				fmt.Errorf("'responseFormat: %s' is not supported yet", format),
			)
		}

		return false
	}

	return true
}

// determineSSEOverload resolves the SSE-overload stream discriminator for an
// operation. Returns nil if the operation does not have SSE overload enabled.
func (g *Generator) determineSSEOverload(ctx context.Context, op *oas.Operation, docInfo *document.DocumentInfo) (*extensions.SSEOverloadConfig, error) {
	sseOverload, err := g.subsystem.Extensions.GetSSEOverload(op.Extensions)
	if err != nil {
		return nil, err
	}

	// Explicit x-speakeasy-sse-overload supports body or query stream discriminators.
	if sseOverload {
		return g.subsystem.Extensions.OperationCanHaveSSEOverload(ctx, op, docInfo)
	}

	// Inference remains body-only for backward compatibility.
	if g.subsystem.Config.Generation.InferSSEOverload {
		ref, err := g.subsystem.Extensions.OperationCanInferSSEOverload(ctx, op, docInfo)
		if err != nil {
			return nil, err
		}
		if ref != nil {
			return ref, nil
		}
	}

	return nil, nil
}

func toBoolPtr(data interface{}) *bool {
	switch v := data.(type) {
	case bool:
		return &v
	case *bool:
		return v
	default:
		return nil
	}
}
