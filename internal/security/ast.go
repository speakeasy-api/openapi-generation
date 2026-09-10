package security

import (
	"context"
	"fmt"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/speakeasy-api/openapi-generation/v2/internal/resolution"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/pointer"
	"github.com/speakeasy-api/openapi/sequencedmap"
	"github.com/speakeasy-api/openapi/yml"

	"github.com/ettle/strcase"
	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/internal/handlers"
	"github.com/speakeasy-api/openapi-generation/v2/internal/oauth2"
	"github.com/speakeasy-api/openapi-generation/v2/internal/schemas"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"gopkg.in/yaml.v3"
)

type addSecurityFieldOpts struct {
	Opts
	securitySchemes *sequencedmap.Map[string, *openapi.ReferencedSecurityScheme]
	schemeKey       string
	originalKey     string
	node            *yaml.Node
	optional        bool
	docInfo         *document.DocumentInfo
	nameOverrides   SchemeNameOverrides
}

func addSecurityField(ctx context.Context, optionObj *ast.TypeDef, opts addSecurityFieldOpts) error {
	schemeKey := opts.schemeKey
	lookupKey := schemeKey
	if opts.originalKey != "" {
		lookupKey = opts.originalKey
	}
	originalKey := opts.nameOverrides.ResolveOriginalKey(lookupKey)
	s, err := getSecurityScheme(opts.securitySchemes, originalKey, opts.node)
	if err != nil {
		return err
	}
	scheme, err := resolution.Resolve(ctx, s, opts.docInfo)
	if err != nil {
		return err
	}

	exts := map[string]string{}
	var annotation *ast.SecurityAnnotation

	switch scheme.GetType() {
	case openapi.SecuritySchemeTypeAPIKey:
		annotation = &ast.SecurityAnnotation{Scheme: true, SecType: scheme.GetType().String(), SubType: string(scheme.GetIn())}
	case openapi.SecuritySchemeTypeOpenIDConnect:
		annotation = &ast.SecurityAnnotation{Scheme: true, SecType: scheme.GetType().String()}
	case openapi.SecuritySchemeTypeOAuth2:
		builtinFlow := ast.OAuth2FlowNone
		// 'clientCredentials' and 'password' flows are considered "built-in" in the sense that their associated hooks are
		// templated as part of the SDK, as opposed to 'implicit' and 'authorizationCode' flows which rely on a custom hooks.
		if scheme.Flows.ClientCredentials != nil && IsOAuth2ClientCredentialsEnabled(ctx, opts.Subsystem) {
			builtinFlow = ast.OAuth2FlowClientCredentials
			clientCredentialsTokenAuthType, err := opts.Subsystem.Extensions.GetTokenServerAuthentication(scheme.Flows.ClientCredentials)
			if err != nil {
				return err
			}
			exts[opts.Subsystem.Extensions.GetResolvedName(extensions.ExtTokenEndpointAuth)] = clientCredentialsTokenAuthType // client_secret_basic, client_secret_post
		} else if scheme.Flows.Password != nil && IsOAuth2PasswordEnabled(ctx, opts.Subsystem) {
			builtinFlow = ast.OAuth2FlowPassword
		}

		subType := ""
		if builtinFlow != ast.OAuth2FlowNone {
			subType = string(builtinFlow)
		}
		annotation = &ast.SecurityAnnotation{Scheme: true, SecType: scheme.GetType().String(), SubType: subType}

	case openapi.SecuritySchemeTypeHTTP:
		// Normalize the scheme name to lowercase for case insensitivity and
		// simplified downstream logic.
		httpScheme := strings.ToLower(scheme.GetScheme())
		annotation = &ast.SecurityAnnotation{Scheme: true, SecType: scheme.GetType().String(), SubType: httpScheme}
	}

	if annotation != nil {
		annotation.SchemeKey = schemeKey
	}

	typ, err := GetSecuritySchemeTypeDef(ctx, scheme, schemeKey, opts.docInfo, &opts.Opts)
	if err != nil {
		return err
	}

	for k, v := range exts {
		typ.Extensions.All[k] = v
	}

	optionObj.Fields = optionObj.Fields.MustAddField(&ast.FieldDef{
		Name:         schemeKey,
		OriginalName: schemeKey,
		Type:         typ,
		Annotations: []ast.Annotation{
			annotation,
			&ast.NeedsCasingAnnotation{},
		},
		Optional: opts.optional,
	}, opts.Subsystem.Config.MaintainOpenAPIOrder())
	return nil
}

// getSecurityScheme will return the security scheme for the given key and default to an apiKey scheme if not found
func getSecurityScheme(securitySchemes *sequencedmap.Map[string, *openapi.ReferencedSecurityScheme], schemeKey string, node *yaml.Node) (*openapi.ReferencedSecurityScheme, error) {
	scheme, ok := securitySchemes.Get(schemeKey)
	if !ok {
		for key := range securitySchemes.Keys() {
			if strings.EqualFold(key, schemeKey) {
				return nil, errors.NewValidationError(fmt.Sprintf("security scheme %q not found, did you mean %q (case sensitive)?", schemeKey, key), node, nil)
			}
		}

		scheme = &openapi.ReferencedSecurityScheme{
			Object: &openapi.SecurityScheme{
				Type: "apiKey",
				In:   pointer.From[openapi.SecuritySchemeIn]("header"),
				Name: pointer.From("Authorization"),
			},
		}
	}

	return scheme, nil
}

// Returns the registered TypeDef of the security scheme.
func GetSecuritySchemeTypeDef(ctx context.Context, scheme *openapi.SecurityScheme, schemeKey string, docInfo *document.DocumentInfo, opts *Opts) (*ast.TypeDef, error) {
	example := opts.Subsystem.Extensions.GetSecuritySchemeExample(scheme)
	if example != nil {
		opts.Subsystem.Features.RecordFeatureUsage(ctx, features.FeatureExamples)
	}

	schemeName := "Scheme" + strcase.ToGoPascal(schemeKey)
	contextStack := ast.ContextStack{}
	contextStack.Append(ast.ContextTypeRefName, schemeName)

	// Add model namespace to context stack for type differentiation
	modelNamespace, _ := opts.Subsystem.Extensions.GetModelNamespace(scheme.GetExtensions())
	if modelNamespace != "" {
		contextStack.AppendModelNamespace(modelNamespace)
	}

	comments := &ast.Comment{Description: scheme.GetDescription()}
	comments, err := handlers.HandleDeprecated(ctx, comments, scheme.GetDeprecated(), scheme.GetExtensions(), opts.Subsystem)
	if err != nil {
		return nil, err
	}

	schemeType := ast.NewType(&ast.TypeDef{
		Name:        schemeName,
		Type:        ast.DataTypeClass,
		Scope:       ast.ScopeShared,
		IsComponent: true,
		Comments:    comments,
	}, contextStack)

	// Store namespace in extensions for namer resolution
	if modelNamespace != "" {
		if schemeType.Extensions == nil {
			schemeType.Extensions = &ast.TypeDefExtensions{All: make(map[string]any)}
		}
		schemeType.Extensions.ModelNamespace = &modelNamespace
	}

	switch scheme.GetType() {
	case openapi.SecuritySchemeTypeAPIKey:
		switch scheme.GetIn() {
		case openapi.SecuritySchemeInHeader, openapi.SecuritySchemeInQuery, openapi.SecuritySchemeInCookie:
			schemeType.Fields = schemeType.Fields.MustAddField(getApiKeyFieldDef(scheme, example), opts.Subsystem.Config.MaintainOpenAPIOrder())
		default:
			return nil, errors.NewValidationError("invalid apiKey location "+scheme.GetIn().String(), scheme.GetCore().In.GetKeyNodeOrRoot(scheme.GetRootNode()), nil)
		}
	case openapi.SecuritySchemeTypeOpenIDConnect:
		schemeType.Fields = schemeType.Fields.MustAddField(getAuthorizationFieldDef(scheme, example), opts.Subsystem.Config.MaintainOpenAPIOrder())
	case openapi.SecuritySchemeTypeOAuth2:
		fields := ast.Fields{}

		switch {
		case scheme.GetFlows().ClientCredentials != nil && IsOAuth2ClientCredentialsEnabled(ctx, opts.Subsystem):
			opts.Subsystem.Features.RecordFeatureUsage(ctx, features.FeatureOAuth2ClientCredentials)
			overridableScopes, _ := opts.Subsystem.Extensions.GetOverridableOAuth2Scopes(scheme.GetFlows().ClientCredentials)
			fields = oauth2.NewClientCredentialsFields(scheme, scheme.GetFlows().ClientCredentials, example, overridableScopes)
			// Store the overridable scopes extension value for template access
			if schemeType.Extensions == nil {
				schemeType.Extensions = &ast.TypeDefExtensions{All: make(map[string]any)}
			}
			schemeType.Extensions.OverridableOAuth2Scopes = overridableScopes
		case scheme.GetFlows().Password != nil && IsOAuth2PasswordEnabled(ctx, opts.Subsystem):
			opts.Subsystem.Features.RecordFeatureUsage(ctx, features.FeatureOAuth2Password)
			fields = oauth2.NewPasswordFields(ctx, opts.Subsystem.Register, schemeKey, scheme, scheme.GetFlows().Password, example)
		default:
			fields = append(fields, getAuthorizationFieldDef(scheme, example))
		}

		for _, f := range fields {
			schemeType.Fields = schemeType.Fields.MustAddField(f, opts.Subsystem.Config.MaintainOpenAPIOrder())
		}
	case openapi.SecuritySchemeTypeHTTP:
		switch strings.ToLower(scheme.GetScheme()) {
		case "basic":
			fields := getBasicAuthFields(scheme, example)
			for _, f := range fields {
				schemeType.Fields = schemeType.Fields.MustAddField(f, opts.Subsystem.Config.MaintainOpenAPIOrder())
			}
		case "bearer":
			schemeType.Fields = schemeType.Fields.MustAddField(getAuthorizationFieldDef(scheme, example), opts.Subsystem.Config.MaintainOpenAPIOrder())
		case "custom":
			if isCustomSecuritySchemesAvailable(ctx, opts.Subsystem) {
				opts.Subsystem.Features.RecordFeatureUsage(ctx, features.FeatureCustomSecuritySchemes)
				var err error
				schemeType, err = getCustomSecurityScheme(ctx, schemeType, scheme, docInfo, opts)
				if err != nil {
					return nil, err
				}
				break
			} else {
				logging.LogWarning(ctx, "custom security schemes are not supported in this target", errors.NewUnsupportedError("custom security schemes not supported falling back to http basic", scheme.GetCore().Scheme.GetValueNodeOrRoot(scheme.GetRootNode())))

				fields := getBasicAuthFields(scheme, example)
				for _, f := range fields {
					schemeType.Fields = schemeType.Fields.MustAddField(f, opts.Subsystem.Config.MaintainOpenAPIOrder())
				}
			}
		default:
			return nil, errors.NewValidationError(fmt.Sprintf("invalid http scheme '%s'", scheme.GetScheme()), scheme.GetCore().Scheme.GetValueNodeOrRoot(scheme.GetRootNode()), nil)
		}
	default:
		return nil, errors.NewValidationError(fmt.Sprintf("invalid security scheme type '%s'", scheme.GetType()), scheme.GetCore().Type.GetValueNodeOrRoot(scheme.GetRootNode()), nil)
	}

	schemeType = opts.Subsystem.Register.RegisterType(ctx, schemeType, true)

	return schemeType, nil
}

func getApiKeyFieldDef(scheme *openapi.SecurityScheme, example *yaml.Node) *ast.FieldDef {
	examples := []*ast.Example{}
	if example != nil {
		examples = append(examples, ast.NewExample("", "", example))
	}

	typeDef := ast.NewType(&ast.TypeDef{
		Type:     ast.DataTypeString,
		Comments: &ast.Comment{Description: scheme.GetDescription()},
		Examples: examples,
	}, nil)

	if typeDef.Comments.Description == "" {
		typeDef.Comments.Description = "API Key"
	}

	return &ast.FieldDef{
		Name:         "apiKey",
		OriginalName: "apiKey",
		Type:         typeDef,
		Annotations: []ast.Annotation{
			&ast.SecurityAnnotation{FieldName: scheme.GetName()},
			&ast.NeedsCasingAnnotation{},
		},
	}
}

func getAuthorizationFieldDef(scheme *openapi.SecurityScheme, example *yaml.Node) *ast.FieldDef {
	examples := []*ast.Example{}
	if example != nil {
		examples = append(examples, ast.NewExample("", "", example))
	}

	typeDef := ast.NewType(&ast.TypeDef{
		Type:     ast.DataTypeString,
		Comments: &ast.Comment{Description: scheme.GetDescription()},
		Examples: examples,
	}, nil)

	if typeDef.Comments.Description == "" {
		switch scheme.Type {
		case "http":
			typeDef.Comments.Description = "HTTP " + strings.ToUpper(scheme.GetScheme()[:1]) + scheme.GetScheme()[1:]
		case "oauth2":
			typeDef.Comments.Description = "OAuth2 Authorization"
		case "openIdConnect":
			typeDef.Comments.Description = "OpenID Connect Authorization"
		default:
			typeDef.Comments.Description = "Authorization"
		}
	}

	return &ast.FieldDef{
		Name:         "Authorization",
		OriginalName: "authorization",
		Type:         typeDef,
		Annotations: []ast.Annotation{
			&ast.SecurityAnnotation{FieldName: "Authorization"},
			&ast.NeedsCasingAnnotation{},
		},
	}
}

func getBasicAuthFields(scheme *openapi.SecurityScheme, example *yaml.Node) []*ast.FieldDef {
	userNameExample := ""
	passwordExample := ""

	if example != nil {
		var ex string
		_ = example.Decode(&ex)

		examples := strings.Split(ex, ";")

		if len(examples) == 2 {
			userNameExample = examples[0]
			passwordExample = examples[1]
		}
	}

	description := scheme.GetDescription()

	if description == "" {
		description = "HTTP Basic"
	}

	return []*ast.FieldDef{
		{
			Name:         "Username",
			OriginalName: "username",
			Type: ast.NewType(&ast.TypeDef{
				Type: ast.DataTypeString,
				Comments: &ast.Comment{
					Description: description + " username",
				},
				Examples: []*ast.Example{ast.NewExample("", "", yml.CreateStringNode(userNameExample))},
			}, nil),
			Annotations: []ast.Annotation{
				&ast.SecurityAnnotation{FieldName: "username"},
				&ast.NeedsCasingAnnotation{},
			},
		},
		{
			Name:         "Password",
			OriginalName: "password",
			Type: ast.NewType(&ast.TypeDef{
				Type: ast.DataTypeString,
				Comments: &ast.Comment{
					Description: description + " password",
				},
				Examples: []*ast.Example{ast.NewExample("", "", yml.CreateStringNode(passwordExample))},
			}, nil),
			Annotations: []ast.Annotation{
				&ast.SecurityAnnotation{FieldName: "password"},
				&ast.NeedsCasingAnnotation{},
			},
		},
	}
}

func getCustomSecurityScheme(ctx context.Context, schemeType *ast.TypeDef, scheme *openapi.SecurityScheme, docInfo *document.DocumentInfo, opts *Opts) (*ast.TypeDef, error) {
	cssConfig, err := opts.Subsystem.Extensions.HandleCustomSecurityConfig(ctx, scheme.GetExtensions())
	if err != nil {
		return nil, err
	}

	if cssConfig == nil {
		return nil, nil
	}

	schema, err := resolution.Resolve(ctx, cssConfig.Schema, docInfo)
	if err != nil {
		return nil, err
	}
	if schema.IsSchema() {
		schema.GetSchema().Title = pointer.From(schemeType.Name)
	}

	f, err := opts.Schemas.HandleSchema(ctx, schemas.Params{
		Schema:              cssConfig.Schema,
		ContextStack:        schemeType.ContextStack,
		Scope:               ast.ScopeShared,
		SerializationMethod: ast.SerializationMethodJSON,
		DocInfo:             opts.DocInfo,
	})
	if err != nil {
		return nil, err
	}
	if f == nil {
		return nil, errors.NewValidationError("empty schema", cssConfig.GetCore().Schema.GetKeyNodeOrRoot(cssConfig.Schema.GetRootNode()), errors.New("schema is required"))
	}
	newType := f.Type
	if newType.Type != ast.DataTypeClass {
		return nil, errors.NewValidationError("invalid schema type", cssConfig.GetCore().Schema.GetKeyNodeOrRoot(cssConfig.Schema.GetRootNode()), errors.New("schema must be an object with properties"))
	}
	if len(newType.Fields) == 0 {
		return nil, errors.NewValidationError("invalid schema type", cssConfig.GetCore().Schema.GetKeyNodeOrRoot(cssConfig.Schema.GetRootNode()), errors.New("schema must be an object with at least one property"))
	}
	for _, field := range newType.Fields {
		field.Annotations.Append(&ast.SecurityAnnotation{FieldName: field.OriginalName})
		field.Annotations.Append(&ast.NeedsCasingAnnotation{})
	}

	newType.IsComponent = true

	return newType, nil
}

func flattenSecurityFields(fields ast.Fields, opts *Opts) ast.Fields {
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
			newField.Type.Fields = flattenSecuritySchemes(newField.Type.Fields, opts)
			containsOptions = true
			flattenedFields = flattenedFields.MustAddField(&newField, opts.Subsystem.Config.MaintainOpenAPIOrder())
		} else {
			flattenedFields = flattenedFields.MustAddField(&newField, opts.Subsystem.Config.MaintainOpenAPIOrder())
		}
	}

	if !containsOptions {
		flattenedFields = flattenSecuritySchemes(flattenedFields, opts)
	}

	return flattenedFields
}

func flattenSecuritySchemes(fields ast.Fields, opts *Opts) ast.Fields {
	flattenedFields := ast.Fields{}

	for _, scheme := range fields {
		if len(fields) > 1 && len(scheme.Type.Fields) > 1 {
			flattenedFields = flattenedFields.MustAddField(scheme, opts.Subsystem.Config.MaintainOpenAPIOrder())
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

				// Propagate deprecation from the scheme wrapper TypeDef to the flattened field.
				// Without this, deprecation info is lost when a scheme type is flattened
				// into individual fields on the Security struct.
				if scheme.Type.Comments != nil && scheme.Type.Comments.Deprecated {
					if newSchemeField.Comments == nil {
						newSchemeField.Comments = &ast.Comment{}
					}
					if newSchemeField.Comments.Description == "" {
						newSchemeField.Comments.Description = scheme.Type.Comments.Description
					}
					newSchemeField.Comments.Deprecated = true
					newSchemeField.Comments.DeprecationMessage = scheme.Type.Comments.DeprecationMessage
					newSchemeField.Comments.DeprecationReplacement = scheme.Type.Comments.DeprecationReplacement
				}

				if scheme.Optional {
					newSchemeField.Optional = true
				}

				flattenedFields = flattenedFields.MustAddField(&newSchemeField, opts.Subsystem.Config.MaintainOpenAPIOrder())
			}
		}
	}

	return flattenedFields
}
