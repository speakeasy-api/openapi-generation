package security

import (
	"context"
	"slices"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/internal/resolution"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/pointer"
	"github.com/speakeasy-api/openapi/sequencedmap"
	"go.uber.org/zap"
)

type HandleGlobalSecurityOptions struct {
	Opts
}

func HandleGlobalSecurity(ctx context.Context, docInfo *document.DocumentInfo, a *ast.AST, opts HandleGlobalSecurityOptions) (*ast.Security, *sequencedmap.Map[string, *openapi.ReferencedSecurityScheme], error) {
	var securitySchemes *sequencedmap.Map[string, *openapi.ReferencedSecurityScheme]
	if docInfo.Doc.Components != nil {
		securitySchemes = docInfo.Doc.Components.SecuritySchemes
	}

	hoistingCandidate, err := getHoistingCandidate(ctx, docInfo, &opts.Opts)
	if err != nil {
		return nil, nil, err
	}

	globalSecurityOptional, err := isGlobalSecurityOptional(ctx, docInfo, hoistingCandidate, &opts.Opts)
	if err != nil {
		return nil, nil, err
	}

	globalSecurity, err := determineSecurity(ctx, determineSecurityOpts{
		Opts:                  opts.Opts,
		contextStack:          ast.ContextStack{},
		securityReqs:          docInfo.Doc.Security,
		securitySchemes:       securitySchemes,
		optional:              pointer.Value(globalSecurityOptional),
		globalSecurityMatcher: nil,
		scope:                 ast.ScopeShared,
		global:                true,
		docInfo:               docInfo,
	})
	if err != nil {
		return nil, nil, err
	}

	// If we have no global security and a valid hoisting candidate then we hoist that security to the global level
	if (globalSecurity == nil || globalSecurity.Security == nil) && hoistingCandidate != nil && opts.Subsystem.Config.HoistGlobalSecurity() {
		globalSecurity, err = determineSecurity(ctx, determineSecurityOpts{
			Opts:                  opts.Opts,
			contextStack:          ast.ContextStack{},
			securityReqs:          []*openapi.SecurityRequirement{hoistingCandidate},
			securitySchemes:       securitySchemes,
			optional:              pointer.Value(globalSecurityOptional),
			globalSecurityMatcher: nil,
			scope:                 ast.ScopeShared,
			global:                true,
			docInfo:               docInfo,
		})
		if err != nil {
			return nil, nil, err
		}
		if globalSecurity != nil {
			logging.From(ctx).Info("No global security found, hoisting most used operation security to global level", zap.Any("requirements", globalSecurity.Requirements))
		}
	}

	if globalSecurity != nil && a != nil && a.MainSDK != nil {
		a.MainSDK.Security = globalSecurity.Security
		a.MainSDK.SecurityConfig = globalSecurity.SecurityConfig
	}

	flattenGlobalSecurity := opts.Subsystem.Config.GetLanguageConfigValue("flattenGlobalSecurity")

	if globalSecurity != nil && globalSecurity.Security != nil && len(globalSecurity.Security.Type.Fields) == 1 && opts.Subsystem.Features.IsFeatureSupported(ctx, features.FeatureGlobalSecurityFlattening) && flattenGlobalSecurity != nil && flattenGlobalSecurity.(bool) {
		opts.Subsystem.Features.RecordFeatureUsage(ctx, features.FeatureGlobalSecurityFlattening)
	}

	return globalSecurity, securitySchemes, nil
}

// isGlobalSecurityOptional will determine if global security has to be marked as optional and the reason why
// This will normally occur when an operation overrides the global security requirements with its own,
// therefore the global security requirements need to be marked as optional so they don't have to be provided
// when the operation is executed
func isGlobalSecurityOptional(ctx context.Context, docInfo *document.DocumentInfo, hoistingCandidate *openapi.SecurityRequirement, opts *Opts) (*ast.SecurityOptionalityReason, error) {
	globalSecurityOptional := ast.SecOptReasonNotOptional

	// Collect all the global scheme names
	globalSchemes := []string{}
	var securitySchemes *sequencedmap.Map[string, *openapi.ReferencedSecurityScheme]
	if docInfo.Doc.Components != nil {
		securitySchemes = docInfo.Doc.Components.SecuritySchemes
	}
	nameOverrides, err := BuildSchemeNameOverrides(opts.Subsystem.Extensions, securitySchemes)
	if err != nil {
		return nil, err
	}

	for _, req := range docInfo.Doc.Security {
		/* If we find an empty security requirement then security has been explicitly marked as optional
		for example:
		```
		security:
		  - ApiKey: []
		  - {} // <-- this is an empty security requirement which marks this security block as optional
		```
		*/
		if req.Len() == 0 {
			globalSecurityOptional = ast.SecOptReasonOptionalScheme
		}

		for scheme := range req.All() {
			globalSchemes = append(globalSchemes, nameOverrides.ResolveSchemeKey(scheme))
		}
	}

	// Consider the hoisted scheme if there are no global schemes
	if hoistingCandidate != nil && len(globalSchemes) == 0 {
		if hoistingCandidate.Len() == 0 {
			globalSecurityOptional = ast.SecOptReasonOptionalScheme
		}

		for scheme := range hoistingCandidate.All() {
			globalSchemes = append(globalSchemes, nameOverrides.ResolveSchemeKey(scheme))
		}
	}

	// If we have no global schemes then we just return the default
	if len(globalSchemes) == 0 {
		return pointer.From(ast.SecOptReasonNotOptional), nil
	}

	if globalSecurityOptional != ast.SecOptReasonNotOptional {
		return pointer.From(globalSecurityOptional), nil
	}

	for pi := range docInfo.Doc.Paths.Values() {
		pathItem, err := resolution.Resolve(ctx, pi, docInfo)
		if err != nil {
			return nil, err
		}
		for op := range pathItem.Values() {
			if op.Security == nil {
				continue
			}

			// If the operation has specifically disable security then global security needs to be optional
			if IsSecurityDisabled(op.Security) {
				globalSecurityOptional = ast.SecOptReasonOperationOverride
				break
			}

			hasOptions := false

			for _, req := range op.Security {
				if req.Len() > 0 {
					hasOptions = true
				}
				for scheme := range req.Keys() {
					// If the operation has schemes not found in the global schemes then global security needs to be optional
					if !slices.Contains(globalSchemes, nameOverrides.ResolveSchemeKey(scheme)) {
						globalSecurityOptional = ast.SecOptReasonOperationOverride
						break
					}
				}
			}

			// If there are no options amongst any of the security requirements then we can just treat security as optional
			if !hasOptions {
				globalSecurityOptional = ast.SecOptReasonOptionalScheme
			}
		}
	}

	// If the we have enable security env vars then the security can come from the environment and doesn't need to be provided
	// via the constructor of the SDK
	if globalSecurityOptional == ast.SecOptReasonNotOptional && opts.Subsystem.Features.IsFeatureEnabled(ctx, features.FeatureEnvVarSecurityUsage) {
		globalSecurityOptional = ast.SecOptReasonEnvVar
	}

	return pointer.From(globalSecurityOptional), nil
}
