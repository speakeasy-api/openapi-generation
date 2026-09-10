package generate

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/internal/resolution"
	"github.com/speakeasy-api/openapi-generation/v2/internal/security"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/pointer"
	"github.com/speakeasy-api/openapi/sequencedmap"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type securityUsage struct {
	count        int
	requirements *openapi.SecurityRequirement
}

func (g *Generator) handleGlobalSecurity(ctx context.Context, docInfo *document.DocumentInfo, a *ast.AST) (*ast.Security, *sequencedmap.Map[string, *openapi.ReferencedSecurityScheme], error) {
	if g.subsystem.Config.Generation.Fixes.SecurityFeb2025 {
		return security.HandleGlobalSecurity(ctx, docInfo, a, security.HandleGlobalSecurityOptions{
			Opts: security.Opts{
				Subsystem: g.subsystem,
				Schemas:   g.schemas,
				DocInfo:   docInfo,
			},
		})
	}

	securitySchemes := docInfo.Doc.GetComponents().GetSecuritySchemes()

	mostUsedOperationSecurity, err := g.getMostUsedOperationSecurity(ctx, docInfo)
	if err != nil {
		return nil, nil, err
	}

	globalSecurityOptional, err := g.isGlobalSecurityOptional(ctx, docInfo, mostUsedOperationSecurity)
	if err != nil {
		return nil, nil, err
	}

	globalSecurity, err := g.handleSecurity(ctx, docInfo, ast.ContextStack{}, docInfo.Doc.Security, securitySchemes, globalSecurityOptional, nil, nil, ast.ScopeShared, true)
	if err != nil {
		return nil, nil, err
	}

	if (globalSecurity == nil || globalSecurity.Security == nil) && mostUsedOperationSecurity != nil && g.subsystem.Config.HoistGlobalSecurity() {
		globalSecurity, err = g.handleSecurity(ctx, docInfo, ast.ContextStack{}, []*openapi.SecurityRequirement{mostUsedOperationSecurity}, securitySchemes, globalSecurityOptional, nil, nil, ast.ScopeShared, true)
		if err != nil {
			return nil, nil, err
		}
		if globalSecurity != nil {
			g.log.Info("No global security found, hoisting most used operation security to global level", zap.Any("requirements", globalSecurity.Requirements))
		}
	}

	if globalSecurity != nil && a.MainSDK != nil {
		a.MainSDK.Security = globalSecurity.Security
		a.MainSDK.SecurityConfig = globalSecurity.SecurityConfig
	}

	flattenGlobalSecurity := g.subsystem.Config.GetLanguageConfigValue("flattenGlobalSecurity")

	if globalSecurity != nil && globalSecurity.Security != nil && len(globalSecurity.Security.Type.Fields) == 1 && g.subsystem.Features.IsFeatureSupported(ctx, features.FeatureGlobalSecurityFlattening) && flattenGlobalSecurity != nil && flattenGlobalSecurity.(bool) {
		g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureGlobalSecurityFlattening)
	}

	return globalSecurity, securitySchemes, nil
}

func (g *Generator) isGlobalSecurityOptional(ctx context.Context, docInfo *document.DocumentInfo, mostUsedOperationSecurity *openapi.SecurityRequirement) (ast.SecurityOptionalityReason, error) {
	// If we have both global and operation level security and the operation level isn't present in the global level we need to make the global optional

	globalSecurityOptional := ast.SecOptReasonNotOptional

	securitySchemes := docInfo.Doc.GetComponents().GetSecuritySchemes()
	nameOverrides, err := security.BuildSchemeNameOverrides(g.subsystem.Extensions, securitySchemes)
	if err != nil {
		return ast.SecurityOptionalityReason(""), err
	}

	globalSchemes := []string{}
	for _, rec := range docInfo.Doc.Security {
		if rec.Len() == 0 {
			globalSecurityOptional = ast.SecOptReasonOptionalScheme
		}

		for scheme := range rec.Keys() {
			globalSchemes = append(globalSchemes, nameOverrides.ResolveSchemeKey(scheme))
		}
	}

	if mostUsedOperationSecurity != nil && len(globalSchemes) == 0 {
		for scheme := range mostUsedOperationSecurity.Keys() {
			globalSchemes = append(globalSchemes, nameOverrides.ResolveSchemeKey(scheme))
		}
	}

	if len(globalSchemes) == 0 {
		return ast.SecOptReasonNotOptional, nil
	}

	for pathItem := range docInfo.Doc.Paths.Values() {
		pi, err := resolution.Resolve(ctx, pathItem, docInfo)
		if err != nil {
			return ast.SecurityOptionalityReason(""), err
		}

		for op := range pi.Values() {
			// If the operation has specifically disable security then global security needs to be optional
			if op.Security != nil && security.IsSecurityDisabled(op.Security) {
				globalSecurityOptional = ast.SecOptReasonOperationOverride
				break
			}

			for _, rec := range op.Security {
				for scheme := range rec.Keys() {
					// If the operation has schemes not found in the global schemes then global security needs to be optional
					if !slices.Contains(globalSchemes, nameOverrides.ResolveSchemeKey(scheme)) {
						globalSecurityOptional = ast.SecOptReasonOperationOverride
						break
					}
				}
			}
		}
	}

	if globalSecurityOptional == ast.SecOptReasonNotOptional && g.subsystem.Features.IsFeatureEnabled(ctx, features.FeatureEnvVarSecurityUsage) {
		globalSecurityOptional = ast.SecOptReasonEnvVar
	}

	return globalSecurityOptional, nil
}

func (g *Generator) handleSecurity(ctx context.Context, docInfo *document.DocumentInfo, contextStack ast.ContextStack, securityRecs []*openapi.SecurityRequirement, securitySchemes *sequencedmap.Map[string, *openapi.ReferencedSecurityScheme], optional ast.SecurityOptionalityReason, globalSecurity *ast.Security, globalSecurityMatcher security.GlobalSecurityMatcher, scope ast.Scope, global bool) (*ast.Security, error) {
	if g.subsystem.Config.Generation.Fixes.SecurityFeb2025 && !global {
		return security.HandlePerOpSecurity(ctx, security.HandlePerOpSecurityOpts{
			Opts: security.Opts{
				Subsystem: g.subsystem,
				Schemas:   g.schemas,
				DocInfo:   docInfo,
			},
			ContextStack:          contextStack,
			SecurityReqs:          securityRecs,
			SecuritySchemes:       securitySchemes,
			Optional:              optional,
			GlobalSecurityMatcher: globalSecurityMatcher,
			Scope:                 scope,
			DocInfo:               docInfo,
		})
	}

	if securityRecs == nil {
		return nil, nil
	} else if security.IsSecurityDisabled(securityRecs) {
		return &ast.Security{
			Security: nil,
			SecurityConfig: ast.SecurityConfig{
				OAuth2Config: ast.OAuth2Config{},
				Disabled:     true,
			},
		}, nil
	}

	nameOverrides, err := security.BuildSchemeNameOverrides(g.subsystem.Extensions, securitySchemes)
	if err != nil {
		return nil, err
	}

	typeName := "Security"

	// If its operation security we will prefix with Operation name always for naming consistency
	if contextStack.HasFrameOfType(ast.ContextTypeOperation) {
		opFrame := contextStack.FindLastFrameOfType(ast.ContextTypeOperation)
		typeName = opFrame.Identifier + "_" + typeName
		opFrame.Used = true
	}

	secType := ast.NewType(&ast.TypeDef{
		Name:  typeName,
		Type:  ast.DataTypeClass,
		Scope: scope,
	}, contextStack)

	options := 0
	singleOptions := true

	for _, sec := range securityRecs {
		if sec.Len() == 0 {
			optional = ast.SecOptReasonOptionalScheme
		} else {
			options++

			if sec.Len() > 1 {
				singleOptions = false
			}
		}
	}

	if options == 0 {
		for scheme := range securitySchemes.AllOrdered(g.subsystem.Config.GetSequencedMapIterationOrder()) {
			resolvedKey := nameOverrides.ResolveSchemeKey(scheme)
			if err := g.addSecurityField(ctx, docInfo, securitySchemes, resolvedKey, scheme, secType, optional != ast.SecOptReasonNotOptional, nil); err != nil {
				return nil, err
			}
		}
	}

	fields := ast.Fields{}

	oauth2Config := ast.OAuth2Config{}
	requirements := make([]ast.SecurityRequirement, 0, len(securityRecs))

	var globalReqs []ast.SecurityRequirement
	if globalSecurity != nil {
		globalReqs = globalSecurity.Requirements
	}

	allMatchGlobal := len(globalReqs) > 0

	// Check if this security is the same as the global security
	isOpSecurityOptional := false
	for _, rec := range securityRecs {
		if rec.Len() == 0 {
			isOpSecurityOptional = true
			continue
		}
		var reqKeys ast.SecurityRequirement
		for schemeKey, scopes := range rec.All() {
			resolvedKey := nameOverrides.ResolveSchemeKey(schemeKey)
			originalKey := nameOverrides.ResolveOriginalKey(schemeKey)
			reqKeys = append(reqKeys, ast.SecurityScheme(resolvedKey))
			scheme, ok := securitySchemes.Get(originalKey)

			node := rec.GetCore().GetOrZero(schemeKey).GetKeyNodeOrRoot(rec.GetRootNode())
			if !ok {
				// Log a warning if the security scheme doesn't exist in
				// the security definitions. This is a warning because returning
				// a validation error would cause execution to break, which would be a
				// breaking change for pre-existing customers.
				logging.LogWarning(ctx, "validation warning", errors.NewValidationWarning(fmt.Sprintf("security scheme '%s' not found", schemeKey), node, nil))
			}

			opts := security.Opts{Subsystem: g.subsystem, DocInfo: docInfo}
			config, err := security.GetOAuth2FlowConfig(ctx, resolvedKey, scheme, &opts)
			if err != nil {
				return nil, err
			}
			if config.Flow != ast.OAuth2FlowNone {
				if existing, ok := oauth2Config[resolvedKey]; ok && !security.AreOAuth2ScopeSetsEqual(existing.RequiredScopes, scopes) {
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
		slices.Sort(reqKeys)

		if len(globalReqs) > 0 && !slices.ContainsFunc(globalReqs, func(globalReq ast.SecurityRequirement) bool {
			return slices.Equal(reqKeys, globalReq)
		}) {
			allMatchGlobal = false
			break
		}
		requirements = append(requirements, reqKeys)

	}

	// For operation-level security, check if all requirements match global security schemes. If so, hoist the operation security.
	// In this case global security is used but the subset of fields allowed for the operation is passed to Hoist.Fields
	// by respecting the order in which the security requirements are defined at the operation level.
	if allMatchGlobal && globalSecurityMatcher != nil {
		equivalent, hoistedFields := globalSecurityMatcher(requirements, isOpSecurityOptional)
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

	// OR
	for i, sec := range securityRecs {
		if sec.Len() == 0 {
			continue
		}

		optionSchemeKey := ""
		if sec.Len() == 1 {
			for schemeKey := range sec.Keys() {
				overrideKey := nameOverrides.ResolveSchemeKey(schemeKey)
				if overrideKey != schemeKey {
					optionSchemeKey = overrideKey
				}
				break
			}
		}

		// All names get suffixes if there are multiple options
		name := typeName + "Option"
		if optionSchemeKey != "" {
			name = optionSchemeKey
		} else if options > 1 {
			name += strconv.Itoa(i + 1)
		}

		optionObj := ast.NewType(&ast.TypeDef{
			Name:  name,
			Type:  ast.DataTypeClass,
			Scope: scope,
		}, contextStack)

		// AND
		for schemeKey := range sec.AllOrdered(g.subsystem.Config.GetSequencedMapIterationOrder()) {
			resolvedKey := nameOverrides.ResolveSchemeKey(schemeKey)
			originalKey := nameOverrides.ResolveOriginalKey(schemeKey)
			s, ok := securitySchemes.Get(originalKey)
			var scheme *openapi.SecurityScheme
			if ok {
				var err error
				scheme, err = resolution.Resolve(ctx, s, docInfo)
				if err != nil {
					return nil, err
				}
			}

			clientCredentialsFlow := ok && scheme.Type == "oauth2" && scheme.Flows.ClientCredentials != nil && security.IsOAuth2ClientCredentialsEnabled(ctx, g.subsystem)
			passwordFlow := ok && scheme.Type == "oauth2" && scheme.Flows.Password != nil && security.IsOAuth2PasswordEnabled(ctx, g.subsystem)
			isOAuthFlow := clientCredentialsFlow || passwordFlow

			if global || !isOAuthFlow {
				node := sec.GetCore().GetOrZero(schemeKey).GetKeyNodeOrRoot(sec.GetRootNode())

				if err := g.addSecurityField(ctx, docInfo, securitySchemes, resolvedKey, originalKey, optionObj, optional != ast.SecOptReasonNotOptional && options == 1, node); err != nil {
					return nil, err
				}
			}
		}

		if sec.Len() > 1 {
			for _, f := range optionObj.Fields {
				f.Annotations.Get(ast.AnnotationTypeSecurity).(*ast.SecurityAnnotation).Composite = true
			}
		}

		if options > 1 && !singleOptions {
			fieldName := fmt.Sprintf("Option%d", i+1)
			originalName := fieldName
			if optionSchemeKey != "" {
				fieldName = optionSchemeKey
				originalName = optionSchemeKey
			}
			fields = fields.MustAddField(&ast.FieldDef{
				Name:         fieldName,
				OriginalName: originalName,
				Type:         g.subsystem.Register.RegisterType(ctx, optionObj, true),
				Optional:     true,
				Annotations: []ast.Annotation{
					&ast.SecurityAnnotation{Option: true, SecurityOption: true},
					&ast.NeedsCasingAnnotation{},
				},
			}, g.subsystem.Config.MaintainOpenAPIOrder())
		} else {
			for _, f := range optionObj.Fields {
				if singleOptions && options > 1 {
					f.Optional = true
				}

				if options > 1 {
					f.Annotations.Get(ast.AnnotationTypeSecurity).(*ast.SecurityAnnotation).SecurityOption = true
				}

				fields = fields.MustAddField(f, g.subsystem.Config.MaintainOpenAPIOrder())
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
	secType.Fields = g.flatten(fields)

	var annotations []ast.Annotation
	if global {
		g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureGlobalSecurity)
	} else {
		annotations = append(annotations, &ast.OperationSecurityAnnotation{})
	}

	return &ast.Security{
		Security: &ast.FieldDef{
			Name:         "Security",
			OriginalName: "security",
			Type:         g.subsystem.Register.RegisterType(ctx, secType, true),
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

func (g *Generator) flatten(fields ast.Fields) ast.Fields {
	containsOptions := false

	flattenedFields := ast.Fields{}

	for _, field := range fields {
		newField := *field

		anno := newField.Annotations.Get(ast.AnnotationTypeSecurity)
		var secAnno *ast.SecurityAnnotation
		if anno != nil {
			secAnno = anno.(*ast.SecurityAnnotation)
		}
		if secAnno != nil && secAnno.Option {
			newField.Type.Fields = g.flattenSchemes(newField.Type.Fields)
			containsOptions = true
			flattenedFields = flattenedFields.MustAddField(&newField, g.subsystem.Config.MaintainOpenAPIOrder())
		} else {
			flattenedFields = flattenedFields.MustAddField(&newField, g.subsystem.Config.MaintainOpenAPIOrder())
		}
	}

	if !containsOptions {
		flattenedFields = g.flattenSchemes(flattenedFields)
	}

	return flattenedFields
}

func (g *Generator) flattenSchemes(fields ast.Fields) ast.Fields {
	flattenedFields := ast.Fields{}

	for _, scheme := range fields {
		if len(fields) > 1 && len(scheme.Type.Fields) > 1 {
			flattenedFields = flattenedFields.MustAddField(scheme, g.subsystem.Config.MaintainOpenAPIOrder())
		} else {
			for _, schemeField := range scheme.Type.Fields {
				// We need to make copies to avoid modifying the original fields
				newSchemeField := *schemeField
				newSchemeField.Annotations = make(ast.Annotations, len(schemeField.Annotations))
				copy(newSchemeField.Annotations, schemeField.Annotations)

				for _, a := range scheme.Annotations {
					idx, existing := newSchemeField.Annotations.Find(a)
					if existing != nil && existing.IsType(ast.AnnotationTypeSecurity) {
						// We need to make a copy to avoid modifying the original field
						secAnno := *existing.(*ast.SecurityAnnotation)
						secAnno.Merge(a.(*ast.SecurityAnnotation))
						newSchemeField.Annotations[idx] = &secAnno
					} else if existing == nil && a.IsType(ast.AnnotationTypeNeedsCasing) {
						// Only copy NeedsCasingAnnotation from scheme that doesn't exist on schemeField
						// Don't copy SecurityAnnotation as it shouldn't be on sub-fields
						newSchemeField.Annotations.Append(a)
					}
				}

				if len(scheme.Type.Fields) == 1 {
					newSchemeField.Name = scheme.Name
					newSchemeField.OriginalName = scheme.OriginalName
				}

				if scheme.Optional {
					newSchemeField.Optional = true
				}

				flattenedFields = flattenedFields.MustAddField(&newSchemeField, g.subsystem.Config.MaintainOpenAPIOrder())
			}
		}
	}

	return flattenedFields
}

func (g *Generator) addSecurityField(ctx context.Context, docInfo *document.DocumentInfo, securitySchemes *sequencedmap.Map[string, *openapi.ReferencedSecurityScheme], schemeKey string, originalKey string, optionType *ast.TypeDef, optional bool, node *yaml.Node) error {
	if originalKey == "" {
		originalKey = schemeKey
	}

	s, err := g.getSecurityScheme(ctx, securitySchemes, originalKey, node)
	if err != nil {
		return err
	}
	scheme, err := resolution.Resolve(ctx, s, docInfo)
	if err != nil {
		return err
	}

	var annotation *ast.SecurityAnnotation

	schemeType := string(scheme.GetType())
	switch schemeType {
	case "apiKey":
		annotation = &ast.SecurityAnnotation{Scheme: true, SecType: schemeType, SubType: scheme.GetIn().String()}
	case "openIdConnect":
		annotation = &ast.SecurityAnnotation{Scheme: true, SecType: schemeType}
	case "oauth2":
		builtinFlow := ast.OAuth2FlowNone
		// 'clientCredentials' and 'password' flows are considered "built-in" in the sense that their associated hooks are templated as part of the SDK,
		// as opposed to 'implicit' and 'authorizationCode' flows which rely on a custom hook and 'x-speakeasy-custom-security-scheme' respectively.
		if scheme.Flows.ClientCredentials != nil && security.IsOAuth2ClientCredentialsEnabled(ctx, g.subsystem) {
			builtinFlow = ast.OAuth2FlowClientCredentials
		} else if scheme.Flows.Password != nil && security.IsOAuth2PasswordEnabled(ctx, g.subsystem) {
			builtinFlow = ast.OAuth2FlowPassword
		}

		subType := ""
		if builtinFlow != ast.OAuth2FlowNone {
			subType = string(builtinFlow)
		}
		annotation = &ast.SecurityAnnotation{Scheme: true, SecType: scheme.GetType().String(), SubType: subType}
	case "http":
		// Normalize the scheme name to lowercase for case insensitivity and
		// simplified downstream logic.
		httpScheme := strings.ToLower(scheme.GetScheme())
		annotation = &ast.SecurityAnnotation{Scheme: true, SecType: scheme.GetType().String(), SubType: httpScheme}
	}

	if annotation != nil {
		annotation.SchemeKey = schemeKey
	}

	securityOpts := &security.Opts{
		Schemas:   g.schemas,
		Subsystem: g.subsystem,
		DocInfo:   docInfo,
	}

	typ, err := security.GetSecuritySchemeTypeDef(ctx, scheme, schemeKey, docInfo, securityOpts)
	if err != nil {
		return err
	}

	optionType.Fields = optionType.Fields.MustAddField(&ast.FieldDef{
		Name:         schemeKey,
		OriginalName: schemeKey,
		Type:         typ,
		Annotations: []ast.Annotation{
			annotation,
			&ast.NeedsCasingAnnotation{},
		},
		Optional: optional,
	}, g.subsystem.Config.MaintainOpenAPIOrder())
	return nil
}

func (g *Generator) getMostUsedOperationSecurity(_ context.Context, docInfo *document.DocumentInfo) (*openapi.SecurityRequirement, error) {
	if docInfo.Doc.Paths == nil {
		return nil, nil
	}

	usedSchemes := sequencedmap.New[string, securityUsage]()

	for _, pathItem := range docInfo.Doc.GetPaths().AllOrdered(g.subsystem.Config.GetSequencedMapIterationOrder()) {
		pi, err := resolution.Resolve(context.Background(), pathItem, docInfo)
		if err != nil {
			return nil, err
		}

		for _, op := range pi.AllOrdered(g.subsystem.Config.GetSequencedMapIterationOrder()) {
			for _, sec := range op.Security {
				schemeKeys := []string{}
				for scheme := range sec.Keys() {
					schemeKeys = append(schemeKeys, scheme)
				}
				slices.Sort(schemeKeys)

				key := strings.Join(schemeKeys, "-")

				schemaUsage, ok := usedSchemes.Get(key)
				if !ok {
					schemaUsage = securityUsage{
						count:        0,
						requirements: sec,
					}
				}

				usedSchemes.Set(key, securityUsage{
					count:        schemaUsage.count + 1,
					requirements: schemaUsage.requirements,
				})
			}
		}
	}

	mostUsedCount := 0
	var mostUsedSecurity *openapi.SecurityRequirement

	for usage := range usedSchemes.Values() {
		if usage.count > mostUsedCount {
			mostUsedCount = usage.count
			mostUsedSecurity = usage.requirements
		}
	}

	return mostUsedSecurity, nil
}

func (g *Generator) getSecurityScheme(ctx context.Context, securitySchemes *sequencedmap.Map[string, *openapi.ReferencedSecurityScheme], schemeKey string, node *yaml.Node) (*openapi.ReferencedSecurityScheme, error) {
	scheme, ok := securitySchemes.Get(schemeKey)
	if !ok {
		for key := range securitySchemes.Keys() {
			if strings.EqualFold(key, schemeKey) {
				return nil, errors.NewValidationError(fmt.Sprintf("security scheme %q not found, did you mean %q (case sensitive)?", schemeKey, key), node, nil)
			}
		}

		logging.LogWarning(ctx, "security scheme not found, defaulting to apiKey using Authorization header", errors.NewValidationError("security scheme not found", node, fmt.Errorf("%s not found", schemeKey)))

		scheme = &openapi.ReferencedSecurityScheme{
			Object: &openapi.SecurityScheme{
				Type: "apiKey",
				In:   pointer.From(openapi.SecuritySchemeInHeader),
				Name: pointer.From("Authorization"),
			},
		}
	}

	return scheme, nil
}
