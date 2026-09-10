package security

import (
	"context"
	"fmt"
	"slices"
	"strconv"

	"github.com/ettle/strcase"
	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/internal/resolution"
	"github.com/speakeasy-api/openapi-generation/v2/internal/schemas"
	"github.com/speakeasy-api/openapi-generation/v2/internal/subsystem"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/sequencedmap"
	"gopkg.in/yaml.v3"
)

type Opts struct {
	Subsystem *subsystem.Subsystem
	Schemas   *schemas.Schemas
	DocInfo   *document.DocumentInfo
}

type determineSecurityOpts struct {
	Opts
	contextStack          ast.ContextStack
	securityReqs          []*openapi.SecurityRequirement
	securitySchemes       *sequencedmap.Map[string, *openapi.ReferencedSecurityScheme]
	optional              ast.SecurityOptionalityReason
	globalSecurityMatcher GlobalSecurityMatcher
	scope                 ast.Scope
	global                bool
	docInfo               *document.DocumentInfo
	nameOverrides         SchemeNameOverrides
}

// determineSecurity will scan the security requirements of a security block and determine the type of security used and generate the appropriate AST
// security blocks can be configured in a number of ways and be provided at the global or per operation level
// some examples of security blocks are:
// ```
// security: [] # <-- no security, if used at the operation level it can be used to disable security for the operation
// ```
//
// ```
// security:
//   - ApiKey: [] # <-- single security requirement with a single option, this is required for the operation to be executed
//
// ```
//
// ```
// security:
//   - ApiKey: []
//   - {} # <-- this is an empty security requirement which marks this security block as optional
//
// ```
//
// ```
// security:
//   - {} # <-- this is an empty security requirement which marks this security block as optional, can be used at the operation level to mark the global security as optional for this operation (a security block at the global level has no effect)
//
// ```
//
// ```
// # multiple security requirements with single options, either can be used but security is required for the operation to be executed. This is the `OR` pattern of security
// security:
//   - ApiKey: []
//   - BasicAuth: []
//
// ```
//
// ```
// # Single security requirement with multiple options both must be used for the operation to be executed. This is the `AND` pattern of security
// security:
//   - ApiKey: []
//     BasicAuth: []
//
// ```
//
// ```
// # Mix of security requirement single and multiple options, either top level security requirement can be used but if the first one is used then both options must be provided.
// security:
//   - ApiKey: []
//     BasicAuth: []
//   - OAuth2:
//   - read:messages # <-- security options can define a list of scopes required by the operation to be granted by the authentication, mainly used for oauth2 and similar schemes, by default it is an empty list
//   - write:messages
//
// ```
func determineSecurity(ctx context.Context, opts determineSecurityOpts) (*ast.Security, error) {
	securityReqs := opts.securityReqs
	if securityReqs == nil {
		// No security provided, so nothing to do
		return nil, nil
	}

	nameOverrides, err := BuildSchemeNameOverrides(opts.Subsystem.Extensions, opts.securitySchemes)
	if err != nil {
		return nil, err
	}
	if len(nameOverrides.originalToOverride) > 0 || len(nameOverrides.overrideToOriginal) > 0 {
		opts.nameOverrides = nameOverrides
	}

	if IsSecurityDisabled(securityReqs) {
		return &ast.Security{
			Security: nil,
			SecurityConfig: ast.SecurityConfig{
				OAuth2Config: ast.OAuth2Config{},
				Disabled:     true,
			},
		}, nil
	}

	typeName := getSecurityTypeName(opts.contextStack)
	secType := ast.NewType(&ast.TypeDef{
		Name:  typeName,
		Type:  ast.DataTypeClass,
		Scope: opts.scope,
	}, opts.contextStack)

	numSchemes, doAllSecurityReqContainSingleSchemes := getSchemeDetails(securityReqs)

	if numSchemes == 0 {
		// If there are no schemes amongst any of the security requirements then we can just treat security as optional
		return &ast.Security{
			Security: nil,
			SecurityConfig: ast.SecurityConfig{
				OAuth2Config: ast.OAuth2Config{},
			},
		}, nil
	}

	oauth2Config, err := getOAuth2Config(ctx, securityReqs, &opts)
	if err != nil {
		return nil, err
	}

	// For operation-level security, check if all requirements match global security schemes. If so, hoist the operation security.
	// In this case global security is used but the subset of fields allowed for the operation is passed to HoistedSecurityConfig.Fields
	// by respecting the order in which the security requirements are defined at the operation level.
	if !opts.global && opts.globalSecurityMatcher != nil {
		opReqs, optional := resolveSecurityRequirements(securityReqs, opts.nameOverrides)
		equivalent, hoistedFields := opts.globalSecurityMatcher(opReqs, optional)
		if hoistedFields != nil {
			return &ast.Security{
				SecurityConfig: ast.SecurityConfig{
					OAuth2Config: oauth2Config,
					HoistedSecurityConfig: &ast.HoistedSecurityConfig{
						Equivalent: equivalent,
						Fields:     hoistedFields,
					},
				},
			}, nil
		}
	}

	optional := opts.optional
	if isSecurityOptional(securityReqs) {
		optional = ast.SecOptReasonOptionalScheme
	}

	fields := ast.Fields{}

	// Each security requirement in a block represents a single security option that can be used independently of each other. This is the `OR` pattern of security as you can use one or the other but not both
	for i, req := range securityReqs {
		// If this req has no requirements then it is just marking this security block as optional and there is nothing to add to the security object
		if req.Len() == 0 {
			continue
		}

		// x-speakeasy-name-override does not apply to `AND` configurations
		optionNameOverride := ""
		if req.Len() == 1 {
			for schemeKey := range req.Keys() {
				overrideKey := opts.nameOverrides.ResolveSchemeKey(schemeKey)
				if overrideKey != schemeKey {
					optionNameOverride = overrideKey
				}
				break
			}
		}

		// All names get suffixes if there are multiple options
		name := typeName + "Option"
		if optionNameOverride != "" {
			name = optionNameOverride
		} else if numSchemes > 1 {
			name += strconv.Itoa(i + 1)
		}

		optionObj := ast.NewType(&ast.TypeDef{
			Name:  name,
			Type:  ast.DataTypeClass,
			Scope: opts.scope,
		}, opts.contextStack)

		// If this security requirement has multiple schemes used then both are required when using this option. This is the `AND` pattern of security.
		for schemeKey := range req.AllOrdered(opts.Subsystem.Config.GetSequencedMapIterationOrder()) {
			node := getSecurityRequirementNode(req, schemeKey)
			resolvedKey := opts.nameOverrides.ResolveSchemeKey(schemeKey)

			if err := addSecurityField(ctx, optionObj, addSecurityFieldOpts{
				Opts:            opts.Opts,
				securitySchemes: opts.securitySchemes,
				schemeKey:       resolvedKey,
				originalKey:     schemeKey,
				node:            node,
				optional:        optional != ast.SecOptReasonNotOptional && numSchemes == 1, // It is only optional if we have only one option otherwise it is required within this option (whether ANDed with another scheme or not)
				docInfo:         opts.docInfo,
				nameOverrides:   opts.nameOverrides,
			}); err != nil {
				return nil, err
			}
		}

		if req.Len() > 1 {
			for _, f := range optionObj.Fields {
				f.Annotations.Get(ast.AnnotationTypeSecurity).(*ast.SecurityAnnotation).Composite = true
			}
		}

		if numSchemes > 1 && !doAllSecurityReqContainSingleSchemes {
			// If we have multiple multi scheme options then we need to add the option object to the security object as a wrapper for any schemes
			fieldName := fmt.Sprintf("Option%d", i+1)
			originalName := fieldName
			if optionNameOverride != "" {
				fieldName = optionNameOverride
				originalName = optionNameOverride
			}
			fields = fields.MustAddField(&ast.FieldDef{
				Name:         fieldName,
				OriginalName: originalName,
				Type:         opts.Subsystem.Register.RegisterType(ctx, optionObj, true),
				Optional:     true,
				Annotations: []ast.Annotation{
					&ast.SecurityAnnotation{Option: true, SecurityOption: true},
					&ast.NeedsCasingAnnotation{},
				},
			}, opts.Subsystem.Config.MaintainOpenAPIOrder())
		} else {
			// If all the security requirements contain a single scheme then we can just add the option object to the security object, flattening the options object away
			for _, f := range optionObj.Fields {
				// If we have multiple single schemes they are all optional as any can be used
				if doAllSecurityReqContainSingleSchemes && numSchemes > 1 {
					f.Optional = true
				}

				// Mark the field as a security option to show it is one of many options
				if numSchemes > 1 {
					f.Annotations.Get(ast.AnnotationTypeSecurity).(*ast.SecurityAnnotation).SecurityOption = true
				}

				fields = fields.MustAddField(f, opts.Subsystem.Config.MaintainOpenAPIOrder())
			}
		}
	}

	// No fields were actually found, return any oauth scopes that may have been collected
	if len(fields) == 0 {
		return &ast.Security{
			Security: nil,
			SecurityConfig: ast.SecurityConfig{
				OAuth2Config: oauth2Config,
			},
		}, nil
	}

	// We'll lose extension information if we're flattening a single security type to top level fields.
	if len(fields) == 1 {
		secType.Extensions = fields[0].Type.Extensions
	}
	// Some objects may have been created with too much nesting (as they only contain single fields) so this will flatten them to simplify the objects
	secType.Fields = flattenSecurityFields(fields, &opts.Opts)

	var annotations []ast.Annotation
	if opts.global {
		opts.Subsystem.Features.RecordFeatureUsage(ctx, features.FeatureGlobalSecurity)
	} else {
		annotations = append(annotations, &ast.OperationSecurityAnnotation{})
	}

	requirements, _ := resolveSecurityRequirements(securityReqs, opts.nameOverrides)

	return &ast.Security{
		Security: &ast.FieldDef{
			Name:         "Security",
			OriginalName: "security",
			Type:         opts.Subsystem.Register.RegisterType(ctx, secType, true),
			Optional:     optional != ast.SecOptReasonNotOptional,
			Annotations:  append(annotations, &ast.NeedsCasingAnnotation{}),
		},
		Requirements: requirements,
		SecurityConfig: ast.SecurityConfig{
			OAuth2Config:      oauth2Config,
			OptionalityReason: optional,
		},
	}, nil
}

func getSecurityRequirementNode(req *openapi.SecurityRequirement, schemeKey string) *yaml.Node {
	node := req.GetCore().GetMapKeyNodeOrRoot(schemeKey, req.GetRootNode())
	if node == nil {
		return nil
	}
	return node
}

func IsSecurityDisabled(securityReqs []*openapi.SecurityRequirement) bool {
	/* Check if security has been explicitly disabled
	for example:
	```
	security: []
	```
	or all requirements are empty (no schemes), which also disables security:
	```
	security:
	 - {}
	```
	*/
	if len(securityReqs) == 0 {
		return true
	}
	for _, req := range securityReqs {
		if req.Len() > 0 {
			return false
		}
	}
	return true
}

func isSecurityOptional(securityReqs []*openapi.SecurityRequirement) bool {
	/*Check if the security block contains any optional security requirements
	for example:
	```
	security:
	 - apiKey: []
	 - {}
	```
	*/
	for _, sec := range securityReqs {
		if sec.Len() == 0 {
			return true
		}
	}

	return false
}

// getSecurityTypeName will return the name of the container security object depending on the context
func getSecurityTypeName(contextStack ast.ContextStack) string {
	typeName := "Security"

	// If its operation security we will prefix with Operation name always for naming consistency
	if contextStack.HasFrameOfType(ast.ContextTypeOperation) {
		opFrame := contextStack.FindLastFrameOfType(ast.ContextTypeOperation)
		typeName = opFrame.Identifier + "_" + typeName
		opFrame.Used = true
	}

	return typeName
}

// getSchemeDetails will determine if any options are found amongst all the security requirements and whether all security requirements contain a single scheme
func getSchemeDetails(securityReqs []*openapi.SecurityRequirement) (int, bool) {
	numSchemes := 0
	doAllSecurityReqContainSingleSchemes := true

	for _, sec := range securityReqs {
		if sec.Len() > 0 {
			numSchemes++

			if sec.Len() > 1 {
				doAllSecurityReqContainSingleSchemes = false
			}
		}
	}

	return numSchemes, doAllSecurityReqContainSingleSchemes
}

// getOAuth2Config collects enabled OAuth2 flows from the security requirements.
// Each scheme retains its own flow configuration and required scopes, so multiple
// built-in OAuth2 flows can be represented in the generated SDK.
func getOAuth2Config(ctx context.Context, securityReqs []*openapi.SecurityRequirement, opts *determineSecurityOpts) (ast.OAuth2Config, error) {
	oauth2Config := ast.OAuth2Config{}
	for _, req := range securityReqs {
		for schemeKey, scopes := range req.All() {
			originalKey := opts.nameOverrides.ResolveOriginalKey(schemeKey)
			resolvedKey := opts.nameOverrides.ResolveSchemeKey(schemeKey)
			scheme, _ := opts.securitySchemes.Get(originalKey)
			config, err := GetOAuth2FlowConfig(ctx, resolvedKey, scheme, &opts.Opts)
			if err != nil {
				return nil, err
			}
			if config.Flow != ast.OAuth2FlowNone {
				if existing, ok := oauth2Config[resolvedKey]; ok && !AreOAuth2ScopeSetsEqual(existing.RequiredScopes, scopes) {
					node := req.GetCore().GetMapKeyNodeOrRoot(schemeKey, req.GetRootNode())
					return nil, errors.NewValidationError(
						fmt.Sprintf("multiple security requirements for OAuth2 scheme %q with different scopes are not currently supported.", resolvedKey),
						node,
						nil,
					)
				}

				config.RequiredScopes = scopes
				oauth2Config[resolvedKey] = *config
			}
		}
	}

	return oauth2Config, nil
}

// AreOAuth2ScopeSetsEqual reports whether two OAuth2 scope lists contain the same distinct scopes.
// Scope order and duplicate entries do not change the security requirement's meaning.
func AreOAuth2ScopeSetsEqual(left, right []string) bool {
	left = slices.Compact(slices.Sorted(slices.Values(left)))
	right = slices.Compact(slices.Sorted(slices.Values(right)))
	return slices.Equal(left, right)
}

// GlobalSecurityMatcher checks whether an operation's security requirements match the global security options.
// If all requirements defined on the operation are already present at the global level, the operation security
// is no longer needed (i.e. can be hoisted). In this case the matching global security fields are returned in
// the priority order defined by the operation.
// If the operation cannot be hoisted, the matcher returns (false, nil).
// If the operation security is equivalent to the global security (i.e. no filtering or reordering is needed) the
// matcher returns (true, [..])
// Built once per generation via NewGlobalSecurityMatcher and reused across all operations. Nil when global=true.
type GlobalSecurityMatcher func(opReqs []ast.SecurityRequirement, optional bool) (equivalent bool, hoistedFields []ast.HoistedSecurityField)

// NewGlobalSecurityMatcher builds a GlobalSecurityMatcher based on the SDK's global security.
// Returns nil if globalSecurity is nil or has no requirements.
func NewGlobalSecurityMatcher(globalSecurity *ast.Security) GlobalSecurityMatcher {
	if globalSecurity == nil || len(globalSecurity.Requirements) == 0 {
		return nil
	}

	globalIsOptional := globalSecurity.SecurityConfig.OptionalityReason != ast.SecOptReasonNotOptional

	var globalFields ast.Fields
	if globalSecurity.Security != nil && globalSecurity.Security.Type != nil {
		globalFields = globalSecurity.Security.Type.Fields
	}

	// Map individual scheme keys to their corresponding global security field.
	hoistMap := buildGlobalSecurityHoistMap(globalFields)

	return func(opReqs []ast.SecurityRequirement, optional bool) (bool, []ast.HoistedSecurityField) {
		// If the operation security is optional but global is not, op-security cannot be hoisted.
		if optional && !globalIsOptional {
			return false, nil
		}

		var hoistedFields []ast.HoistedSecurityField
		for _, opReq := range opReqs {
			// Check if this requirement matches any global requirement.
			matched := slices.ContainsFunc(globalSecurity.Requirements, func(globalReq ast.SecurityRequirement) bool {
				return slices.Equal(opReq, globalReq)
			})
			if !matched {
				return false, nil
			}

			seen := make(map[string]struct{})
			for _, schemeKey := range opReq {
				if hf, ok := hoistMap[string(schemeKey)]; ok {
					if _, dup := seen[hf.Name]; !dup {
						seen[hf.Name] = struct{}{}
						hoistedFields = append(hoistedFields, hf)
					}
				}
			}
		}

		// Optionality must be checked separately because both operation and global requirements have had their {}
		// elements stripped by resolveSecurityRequirements()
		// When envVarPrefix is set however, a match is not required  since global security was made optional.
		optionalityMatches := globalSecurity.SecurityConfig.OptionalityReason == ast.SecOptReasonEnvVar ||
			optional == globalIsOptional
		reqsEquivalent := optionalityMatches && ast.AreEquivalentRequirements(opReqs, globalSecurity.Requirements)

		return reqsEquivalent, hoistedFields
	}
}

type GlobalSecurityHoistMap map[string]ast.HoistedSecurityField

// buildGlobalSecurityHoistMap maps individual scheme keys to the global security field that contains them.
func buildGlobalSecurityHoistMap(fields ast.Fields) GlobalSecurityHoistMap {
	hoistMap := make(GlobalSecurityHoistMap)
	numFields := len(fields)
	group := 0
	for i, f := range fields {
		a := getSecurityAnnotation(f)
		if a == nil {
			// For example `Scopes` and `TokenURL` have no security annotation (oauth2:clientCredentials)
			continue
		}

		switch {
		case numFields > 1 && !a.SecurityOption:
			// Either a flattened scheme (e.g. Username/Password) or a security
			// class that is part of a AND group (e.g. clientCredentialsBasic).
			// For flattened schemes, only the first field is mapped.
			if _, exists := hoistMap[a.SchemeKey]; !exists {
				hoistMap[a.SchemeKey] = ast.HoistedSecurityField{Name: f.Name, Index: i, Group: group}
			}
		case a.Option && f.Type != nil:
			// Option Wrapper
			for _, inner := range f.Type.Fields {
				if k := getSecuritySchemeKey(inner); k != "" {
					hoistMap[k] = ast.HoistedSecurityField{Name: f.Name, Index: i, Group: group}
				}
			}
			group++
		default:
			// Either a single scheme of a security class in a OR configuration.
			hoistMap[a.SchemeKey] = ast.HoistedSecurityField{Name: f.Name, Index: i, Group: group}
			group++
		}
	}

	return hoistMap
}

// getSecurityAnnotation extracts a field's security annotation (if any)
func getSecurityAnnotation(f *ast.FieldDef) *ast.SecurityAnnotation {
	if f.Annotations == nil {
		return nil
	}
	if a := f.Annotations.Get(ast.AnnotationTypeSecurity); a != nil {
		return a.(*ast.SecurityAnnotation)
	}

	return nil
}

// getSecuritySchemeKey extracts the SchemeKey from a field's security annotation.
func getSecuritySchemeKey(f *ast.FieldDef) string {
	a := getSecurityAnnotation(f)
	if a == nil {
		return ""
	}

	return a.SchemeKey
}

// resolveSecurityRequirements converts OpenAPI security requirements into comparable ast.SecurityRequirements.
// Name overrides are resolved and scheme keys are sorted within each requirement, so that the result can be
// used to compare operation-level and global security via slices.Equal.
// Empty requirements ({}) are skipped — their presence is signaled via the returned isOptional bool.
func resolveSecurityRequirements(securityReqs []*openapi.SecurityRequirement, nameOverrides SchemeNameOverrides) (reqs []ast.SecurityRequirement, isOptional bool) {
	reqs = make([]ast.SecurityRequirement, 0, len(securityReqs))
	for _, req := range securityReqs {
		if req.Len() == 0 {
			isOptional = true
			continue
		}
		schemeKeys := make(ast.SecurityRequirement, 0, req.Len())
		for schemeKey := range req.Keys() {
			schemeKeys = append(schemeKeys, ast.SecurityScheme(nameOverrides.ResolveSchemeKey(schemeKey)))
		}
		slices.Sort(schemeKeys)
		reqs = append(reqs, schemeKeys)
	}
	return reqs, isOptional
}

// GetOAuth2FlowConfig will return the OAuth2 flow configuration for a given security scheme,
// which includes available scopes and whether the flow is enabled or not.
func GetOAuth2FlowConfig(ctx context.Context, schemeKey string, s *openapi.ReferencedSecurityScheme, opts *Opts) (*ast.OAuth2FlowConfig, error) {
	scheme, err := resolution.Resolve(ctx, s, opts.DocInfo)
	if err != nil {
		return nil, err
	}

	flow := ast.OAuth2FlowNone
	if scheme == nil {
		return &ast.OAuth2FlowConfig{Flow: flow}, nil
	}

	enabled := true
	var scopeMap *sequencedmap.Map[string, string]
	if scheme.GetType() == "oauth2" {
		switch {
		case scheme.Flows.ClientCredentials != nil:
			// The `client-credentials` flow is supported through a built-in hook
			flow = ast.OAuth2FlowClientCredentials
			enabled = IsOAuth2ClientCredentialsEnabled(ctx, opts.Subsystem)
			scopeMap = scheme.Flows.ClientCredentials.Scopes
		case scheme.Flows.Password != nil:
			// The `password` flow is supported through a built-in hook
			flow = ast.OAuth2FlowPassword
			enabled = IsOAuth2PasswordEnabled(ctx, opts.Subsystem)
			scopeMap = scheme.Flows.Password.Scopes
		case scheme.Flows.Implicit != nil:
			// Customers are expected to set up their own hook for the `implicit` flow
			flow = ast.OAuth2FlowImplicit
			scopeMap = scheme.Flows.Implicit.Scopes
		case scheme.Flows.AuthorizationCode != nil:
			// The `authorizationCode` flow should be implemented using x-speakeasy-custom-security-scheme
			flow = ast.OAuth2FlowAuthorizationCode
			scopeMap = scheme.Flows.AuthorizationCode.Scopes
		}
	}

	// Collect available scopes, even if the flow is disabled
	availableScopes := []ast.OAuth2Scope{}
	if scopeMap != nil {
		for name, description := range scopeMap.All() {
			availableScopes = append(availableScopes, ast.OAuth2Scope{
				Name: name,
				Comments: ast.Comment{
					Description: description,
					Summary:     "",
				},
			})
		}
	}

	comments := ast.Comment{
		Summary:     fmt.Sprintf("Available scopes for the %s OAuth 2.0 scheme (%s flow).", schemeKey, strcase.ToCamel(string(flow))),
		Description: scheme.GetDescription(),
	}

	return &ast.OAuth2FlowConfig{
		Flow:            flow,
		Enabled:         enabled,
		Comments:        comments,
		AvailableScopes: availableScopes,
	}, nil
}
