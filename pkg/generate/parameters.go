package generate

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/contenttypes"
	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/internal/handlers"
	"github.com/speakeasy-api/openapi-generation/v2/internal/namer"
	"github.com/speakeasy-api/openapi-generation/v2/internal/openapi"
	"github.com/speakeasy-api/openapi-generation/v2/internal/resolution"
	"github.com/speakeasy-api/openapi-generation/v2/internal/schemas"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/speakeasy-api/openapi/pointer"

	"github.com/speakeasy-api/openapi/jsonschema/oas3"
	oas "github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/sequencedmap"
)

type paramOAS struct {
	Param *oas.ReferencedParameter
	Idx   int
}

func paramAllowsReserved(ctx context.Context, rp *oas.ReferencedParameter, ext *extensions.Extensions) bool {
	param := rp.MustGetObject()
	allowed, unsupported := ext.DoesParamAllowReserved(param)
	if unsupported {
		if param.GetIn() == oas.ParameterInPath {
			logging.LogWarning(ctx, "allowReserved keyword ignored", errors.NewUnsupportedError(fmt.Sprintf("path parameter %q has `allowReserved: true`, which only applies to query parameters; use `x-speakeasy-param-encoding-override: allowReserved` on the parameter instead", param.GetName()), nil))
		} else {
			logging.LogWarning(ctx, "allowReserved extension ignored", errors.NewUnsupportedError(fmt.Sprintf("%s parameter %q has `x-speakeasy-param-encoding-override: allowReserved`, which only applies to path parameters; use `allowReserved: true` on the parameter instead", param.GetIn(), param.GetName()), nil))
		}
	}
	return allowed
}

type param struct {
	Field    *ast.FieldDef
	Schema   *oas3.JSONSchema[oas3.Referenceable]
	Examples []*ast.Example
}

type handleParametersOptions struct {
	contextStack        ast.ContextStack
	itemParams          []*oas.ReferencedParameter
	paramsOAS           []*oas.ReferencedParameter
	globalNameOverrides []*extensions.NameOverride
	scope               ast.Scope
	globals             *ast.TypeDef
	opID                string
	docInfo             *document.DocumentInfo
}

func (g *Generator) handleParameters(ctx context.Context, opts handleParametersOptions) (*ast.RequestParams, ast.Fields, bool, error) {
	if len(opts.itemParams) == 0 && len(opts.paramsOAS) == 0 {
		return nil, nil, false, nil
	}

	params := []paramOAS{}

	paramIdx := 0

	// Resolve item parameters
	if err := resolveParameters(ctx, opts.itemParams, opts.docInfo); err != nil {
		return nil, nil, false, err
	}
	if err := resolveParameters(ctx, opts.paramsOAS, opts.docInfo); err != nil {
		return nil, nil, false, err
	}

	// Deduplicate parameters if multiple parameters are present.
	uniquePathLevelParams := getUniqueParamNames(opts.itemParams)
	for _, ip := range uniquePathLevelParams {
		itemParam := ip.MustGetObject()
		overridden := false
		for _, p := range opts.paramsOAS {
			param := p.MustGetObject()
			if itemParam.GetName() == param.GetName() && itemParam.GetIn() == param.GetIn() {
				overridden = true
				break
			}
		}
		if !overridden {
			params = append(params, paramOAS{Param: ip, Idx: paramIdx})
			paramIdx++
		}
	}

	uniqueOpLevelParams := getUniqueParamNames(opts.paramsOAS)
	for _, p := range uniqueOpLevelParams {
		params = append(params, paramOAS{Param: p, Idx: paramIdx})
		paramIdx++
	}

	pathParamsOAS := []paramOAS{}
	queryParamsOAS := []paramOAS{}
	headerParamsOAS := []paramOAS{}
	cookieParamsOAS := []paramOAS{}

	for _, param := range params {
		p := param.Param.MustGetObject()

		ignore, err := g.subsystem.Extensions.Ignore(p.GetExtensions())
		if err != nil {
			return nil, nil, false, err
		}
		if ignore {
			g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureIgnores)
			continue
		}

		switch p.GetIn() {
		case oas.ParameterInPath:
			pathParamsOAS = append(pathParamsOAS, param)
		case oas.ParameterInQuery:
			queryParamsOAS = append(queryParamsOAS, param)
		case oas.ParameterInHeader:
			headerParamsOAS = append(headerParamsOAS, param)
		case oas.ParameterInCookie:
			cookieParamsOAS = append(cookieParamsOAS, param)
		}
	}

	globalParams := ast.Fields{}

	var err error
	var pathParams []*ast.Param
	pathParams, globalParams, err = g.handlePathParameters(ctx, handleSpecificParametersOptions{
		contextStack:        opts.contextStack,
		params:              pathParamsOAS,
		globalParams:        globalParams,
		globalNameOverrides: opts.globalNameOverrides,
		scope:               opts.scope,
		globals:             opts.globals,
		opID:                opts.opID,
		docInfo:             opts.docInfo,
	})
	if err != nil {
		return nil, nil, false, err
	}

	var queryParams []*ast.Param
	queryParams, globalParams, err = g.handleQueryParameters(ctx, handleSpecificParametersOptions{
		contextStack:        opts.contextStack,
		params:              queryParamsOAS,
		globalParams:        globalParams,
		globalNameOverrides: opts.globalNameOverrides,
		scope:               opts.scope,
		globals:             opts.globals,
		opID:                opts.opID,
		docInfo:             opts.docInfo,
	})
	if err != nil {
		return nil, nil, false, err
	}

	var headerParams []*ast.Param
	var usesUserAgentHeader bool
	headerParams, globalParams, usesUserAgentHeader, err = g.handleHeaderParameters(ctx, handleSpecificParametersOptions{
		contextStack:        opts.contextStack,
		params:              headerParamsOAS,
		globalParams:        globalParams,
		globalNameOverrides: opts.globalNameOverrides,
		scope:               opts.scope,
		globals:             opts.globals,
		opID:                opts.opID,
		docInfo:             opts.docInfo,
	})
	if err != nil {
		return nil, nil, false, err
	}

	if err := handleCookieParameters(ctx, opts.contextStack, cookieParamsOAS); err != nil {
		return nil, nil, false, err
	}

	if pathParams == nil && queryParams == nil && headerParams == nil {
		return nil, globalParams, false, nil
	}

	return &ast.RequestParams{
		PathParams:   pathParams,
		QueryParams:  queryParams,
		HeaderParams: headerParams,
	}, globalParams, usesUserAgentHeader, nil
}

type handleSpecificParametersOptions struct {
	contextStack        ast.ContextStack
	params              []paramOAS
	globalParams        ast.Fields
	globalNameOverrides []*extensions.NameOverride
	scope               ast.Scope
	globals             *ast.TypeDef
	opID                string
	docInfo             *document.DocumentInfo
}

func (g *Generator) handlePathParameters(ctx context.Context, opts handleSpecificParametersOptions) ([]*ast.Param, ast.Fields, error) {
	if len(opts.params) == 0 {
		return nil, opts.globalParams, nil
	}

	globalParams := opts.globalParams

	pathParams := []*ast.Param{}

	childContextStack := append(opts.contextStack, ast.ContextFrame{
		Type:       ast.ContextTypeParameter,
		Identifier: "pathParam",
	})

	for _, prm := range opts.params {
		p := prm.Param.MustGetObject()

		var err error
		var param *param

		requested := paramAllowsReserved(ctx, prm.Param, g.subsystem.Extensions)
		supported := g.subsystem.Features.IsFeatureSupported(ctx, features.FeatureAllowReserved)
		if requested && !supported {
			logging.LogWarning(ctx, "allowReserved not supported", errors.NewUnsupportedError(fmt.Sprintf("path parameter %q has `x-speakeasy-param-encoding-override: allowReserved`, which this language does not support; the parameter will use standard percent encoding", p.GetName()), nil))
		}

		allowReserved := requested && supported

		if p.GetContent().Len() > 0 {
			param, err = g.handleContentParameter(ctx, ast.ParamTypePathParam, childContextStack, prm.Param, opts.globalNameOverrides, opts.scope, opts.docInfo)
			if err != nil {
				return nil, nil, err
			}
		} else {
			if p.GetStyle() != oas.SerializationStyleSimple && p.GetStyle() != "" {
				return nil, nil, errors.NewUnsupportedError(fmt.Sprintf("path parameters with style %s not currently supported", p.GetStyle()), p.GetCore().Style.GetKeyNodeOrRoot(p.GetRootNode()))
			}

			param, err = g.handleNormalParam(ctx, childContextStack, prm.Param, schemas.ParamTypePath, opts.globalNameOverrides, opts.scope, opts.docInfo)
			if err != nil {
				return nil, nil, err
			}

			if !param.Field.Type.IsPrimitive() {
				node := param.Schema.GetResolvedSchema().GetRootNode()
				if param.Schema.GetResolvedSchema().IsSchema() {
					node = param.Schema.GetResolvedSchema().GetPropertyNode("Type")
				}

				if !param.Field.Type.IsObjectType() && !param.Field.Type.IsContainer() {
					return nil, nil, errors.NewUnsupportedError(fmt.Sprintf("path parameters with type %s not currently supported", param.Field.Type.Type), node)
				}

				if !param.Field.Type.IsSimpleObjectOrContainerType() {
					return nil, nil, errors.NewUnsupportedError("path parameters of array or object type must be of simple array, map, or object", node)
				}
			}

			style := oas.SerializationStyleSimple
			if p.GetStyle() != "" {
				style = p.GetStyle()
			}

			paramAnnotation := &ast.ParamAnnotation{
				AllowReserved:        allowReserved,
				Explode:              openapi.GetLibOpenAPIExplodeDefault(p),
				FieldType:            param.Field.Type,
				Hidden:               g.subsystem.Extensions.IsGlobalHidden(ctx, p),
				Name:                 p.GetName(),
				ParamType:            ast.ParamTypePathParam,
				RequiredForOperation: !param.Field.Optional,
				Style:                string(style),
			}

			param.Field.Annotations = append(param.Field.Annotations, paramAnnotation)
		}

		description := prm.Param.GetDescription()
		if description == "" {
			description = p.GetDescription()
		}

		if description != "" {
			param.Field.Comments = &ast.Comment{
				Description: description,
			}
		} else if param.Field.Type.Comments != nil && param.Field.Type.Comments.Description != "" {
			param.Field.Comments = &ast.Comment{
				Description: param.Field.Type.Comments.Description,
			}
		}

		param.Field.Comments, err = handlers.HandleDeprecated(ctx, param.Field.Comments, p.GetDeprecated(), p.GetExtensions(), g.subsystem)
		if err != nil {
			return nil, nil, err
		}

		param.Field.ParameterIndex = pointer.From(prm.Idx)

		globalHidden, globalParam := g.handleGlobalParameter(ctx, opts.globals, param.Field, opts.opID)
		if globalParam != nil {
			globalParams, err = globalParams.AddField(globalParam, g.subsystem.Config.MaintainOpenAPIOrder())
			if err != nil {
				return nil, nil, errors.NewValidationError("duplicated field generated for global path parameter", p.GetCore().Name.GetKeyNodeOrRoot(p.GetRootNode()), fmt.Errorf("path parameter %s duplicated: %w", p.GetName(), err))
			}
		}

		pathParams = append(pathParams, &ast.Param{
			Field:    param.Field,
			Hidden:   globalHidden,
			Examples: param.Examples,
		})
	}

	if len(pathParams) == 0 {
		return nil, globalParams, nil
	}

	return pathParams, globalParams, nil
}

func getUniqueParamNames(params []*oas.ReferencedParameter) []*oas.ReferencedParameter {
	uniqueParams := sequencedmap.New[string, *oas.ReferencedParameter]()
	for _, param := range params {
		paramObj := param.MustGetObject()
		paramId := fmt.Sprintf("%s_%s", paramObj.GetName(), paramObj.GetIn())
		_, ok := uniqueParams.Get(paramId)
		if !ok {
			uniqueParams.Set(paramId, param)
		}
	}
	uniqueParamsCollection := make([]*oas.ReferencedParameter, 0, uniqueParams.Len())
	for _, param := range uniqueParams.All() {
		uniqueParamsCollection = append(uniqueParamsCollection, param)
	}
	return uniqueParamsCollection
}

// Checks the given parameter field against all global fields by comparing name,
// type, and annotations to find a global parameter match. When matched, returns
// whether the parameter should be hidden and the matched global parameter.
//
// The parameter being hidden is determined by:
//   - A matching global parameter.
//   - Either the parameter field or global parameter field has the hidden
//     parameter annotation.
//   - The hidden globals feature is supported by the template.
//
// When the parameter matches a global parameter, but is not hidden, the
// parameter is marked as optional and its type is set to the global parameter
// type.
func (g *Generator) handleGlobalParameter(ctx context.Context, globals *ast.TypeDef, param *ast.FieldDef, opID string) (bool, *ast.FieldDef) {
	hidden := false
	var globalParam *ast.FieldDef

	if globals == nil {
		return hidden, globalParam
	}

	paramAnno := param.Annotations.Get(ast.AnnotationTypeParam).(*ast.ParamAnnotation)

	for _, globalField := range globals.Fields {
		if param.Name != globalField.Name {
			continue
		}

		_, anno := globalField.Annotations.Find(paramAnno)
		if anno == nil {
			continue
		}
		globalAnno := anno.(*ast.ParamAnnotation)

		if paramAnno.ParamType != globalAnno.ParamType {
			continue
		}

		typeMatches := (globalField.Type.IsCustomClass() && param.Type.IsEqual(globalField.Type) == nil) || (!globalField.Type.IsCustomClass() && param.Type.Type == globalField.Type.Type)

		if !typeMatches {
			continue
		}

		if param.Type.Scope != ast.ScopeShared {
			g.subsystem.Register.UnregisterType(param.Type)
		}

		paramAnno.HasGlobal = true
		if !slices.Contains(globalAnno.OperationsForGlobal, opID) {
			globalAnno.OperationsForGlobal = append(globalAnno.OperationsForGlobal, opID)
		}

		hidden = globalAnno.Hidden || paramAnno.Hidden

		if hidden {
			g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureHiddenGlobals)
		}
		if !g.subsystem.Features.IsFeatureSupported(ctx, features.FeatureHiddenGlobals) {
			hidden = false
		}

		if hidden {
			// Ensure the hidden determination is propagated to the parameter annotation in addition to the ast.Param
			// to simplify templating logic only receiving the FieldDef.
			paramAnno.Hidden = true
		} else {
			// The parameter becomes optional if there is a matched global parameter with the same name and type
			param.Optional = true
			param.Type = globalField.Type
		}

		globalParam = globalField
		break
	}

	return hidden, globalParam
}

func (g *Generator) handleQueryParameters(ctx context.Context, opts handleSpecificParametersOptions) ([]*ast.Param, ast.Fields, error) {
	if len(opts.params) == 0 {
		return nil, opts.globalParams, nil
	}

	globalParams := opts.globalParams
	queryParams := []*ast.Param{}

	childContextStack := append(opts.contextStack, ast.ContextFrame{
		Type:       ast.ContextTypeParameter,
		Identifier: "queryParam",
	})

	for _, prm := range opts.params {
		p := prm.Param.MustGetObject()

		if paramAllowsReserved(ctx, prm.Param, g.subsystem.Extensions) && !g.subsystem.Features.IsFeatureSupported(ctx, features.FeatureAllowReserved) {
			logging.LogWarning(ctx, "allowReserved not supported", errors.NewUnsupportedError(fmt.Sprintf("query parameter %q has `allowReserved: true`, which this language does not support; the parameter value will use standard percent encoding", p.GetName()), nil))
		}

		var err error
		var param *param

		if p.GetContent().Len() > 0 {
			param, err = g.handleContentParameter(ctx, ast.ParamTypeQueryParam, childContextStack, prm.Param, opts.globalNameOverrides, opts.scope, opts.docInfo)
			if err != nil {
				return nil, nil, err
			}
		} else {
			param, err = g.handleNormalQueryParameter(ctx, childContextStack, prm.Param, opts.globalNameOverrides, opts.scope, opts.docInfo)
			if err != nil {
				return nil, nil, err
			}
		}

		description := prm.Param.GetDescription()
		if description == "" {
			description = p.GetDescription()
		}

		if description != "" {
			param.Field.Comments = &ast.Comment{
				Description: description,
			}
		} else if param.Field.Type.Comments != nil && param.Field.Type.Comments.Description != "" {
			param.Field.Comments = &ast.Comment{
				Description: param.Field.Type.Comments.Description,
			}
		}

		param.Field.Comments, err = handlers.HandleDeprecated(ctx, param.Field.Comments, p.GetDeprecated(), p.GetExtensions(), g.subsystem)
		if err != nil {
			return nil, nil, err
		}

		param.Field.ParameterIndex = pointer.From(prm.Idx)

		globalHidden, globalParam := g.handleGlobalParameter(ctx, opts.globals, param.Field, opts.opID)
		if globalParam != nil {
			globalParams, err = globalParams.AddField(globalParam, g.subsystem.Config.MaintainOpenAPIOrder())
			if err != nil {
				return nil, nil, errors.NewValidationError("duplicated field generated for global query parameter", p.GetCore().Name.GetKeyNodeOrRoot(p.GetRootNode()), fmt.Errorf("query parameter %s duplicated: %w", p.GetName(), err))
			}
		}

		allowEmptyValue, err := g.subsystem.Extensions.HandleAllowEmptyQueryParameterValueExtension(p)
		if err != nil {
			return nil, nil, err
		}

		queryParams = append(queryParams, &ast.Param{
			Field:           param.Field,
			Hidden:          globalHidden,
			Examples:        param.Examples,
			AllowEmptyValue: allowEmptyValue,
		})
	}

	if len(queryParams) == 0 {
		return nil, globalParams, nil
	}

	return queryParams, globalParams, nil
}

func (g *Generator) handleContentParameter(ctx context.Context, paramType string, contextStack ast.ContextStack, rp *oas.ReferencedParameter, globalNameOverrides []*extensions.NameOverride, scope ast.Scope, docInfo *document.DocumentInfo) (*param, error) {
	p := rp.MustGetObject()

	examples, err := getParamExamples(ctx, p, docInfo)
	if err != nil {
		return nil, err
	}

	for contentType, content := range p.GetContent().AllOrdered(g.schemas.Config.GetSequencedMapIterationOrder()) {
		switch {
		case contenttypes.IsJSON(contentType):
			s := scope

			schema := content.GetSchema()

			resolvedSchema, err := resolution.Resolve(ctx, schema, docInfo)
			if err != nil {
				return nil, err
			}

			if rp.IsReference() {
				s = ast.ScopeShared
				// We are setting the schema ref to the param ref, so that we can ensure it is associated as a dependency correctly
				if !schema.IsReference() {
					schema = oas3.NewReferencedScheme(ctx, rp.GetReference(), resolvedSchema)
				}
			}

			paramField, err := g.schemas.HandleSchema(ctx, schemas.Params{
				ContextStack:        addPropScope(contextStack, rp),
				SerializationMethod: ast.SerializationMethodJSON,
				Schema:              schema,
				Scope:               s,
				IsRequest:           true,
				Depth:               1,
				DocInfo:             docInfo,
			})
			if err != nil {
				return nil, err
			}
			paramField.Name = p.GetName()
			paramField.OriginalName = p.GetName()
			paramField, err = g.overrideParameterName(ctx, paramField, p, globalNameOverrides)
			if err != nil {
				return nil, err
			}

			required := false
			if p.Required != nil && p.GetRequired() {
				required = true
			}

			paramField.Optional = !required

			paramAnnotation := &ast.ParamAnnotation{
				AllowReserved:        false,
				FieldType:            paramField.Type,
				Hidden:               g.subsystem.Extensions.IsGlobalHidden(ctx, p),
				Name:                 p.GetName(),
				ParamType:            paramType,
				RequiredForOperation: required,
				Serialization:        string(ast.SerializationMethodJSON),
			}

			paramField.Annotations = append(paramField.Annotations, paramAnnotation)

			if len(examples) == 0 {
				if p.GetExamples().Len() > 0 {
					examples = []*ast.Example{}
					for name, e := range content.GetExamples().All() {
						example, err := resolution.Resolve(ctx, e, docInfo)
						if err != nil {
							return nil, err
						}

						examples = append(examples, ast.NewExample(name, example.GetDescription(), resolveExampleValue(example)))
					}
				} else if content.Example != nil {
					examples = []*ast.Example{
						ast.NewExample("", "", content.Example),
					}
				}
			}

			return &param{
				Field:    paramField,
				Examples: examples,
			}, nil
		default:
			return nil, errors.NewUnsupportedError(fmt.Sprintf("%s with content type %s not currently supported", paramType, contentType), p.GetCore().Content.GetMapKeyNodeOrRoot(contentType, p.GetRootNode()))
		}
	}

	return nil, fmt.Errorf("failed to generate %s", paramType)
}

func (g *Generator) overrideParameterName(ctx context.Context, param *ast.FieldDef, p *oas.Parameter, globalNameOverrides []*extensions.NameOverride) (*ast.FieldDef, error) {
	for _, globalNameOverride := range globalNameOverrides {
		if globalNameOverride.ParameterName != "" {
			re := regexp.MustCompile(globalNameOverride.ParameterName)
			if re.MatchString(p.GetName()) {
				param.Name = globalNameOverride.GlobalParameterNameOverride
			}
		}
	}

	parameterOverride, overrideGlobalParameterName, err := g.subsystem.Extensions.HandleOperationParameterNameExtension(p)
	if err != nil {
		return nil, err
	}

	if overrideGlobalParameterName {
		param.Name = parameterOverride.Name
		g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureNameOverrides)
	}

	return param, nil
}

func (g *Generator) handleNormalQueryParameter(ctx context.Context, contextStack ast.ContextStack, rp *oas.ReferencedParameter, globalNameOverrides []*extensions.NameOverride, scope ast.Scope, docInfo *document.DocumentInfo) (*param, error) {
	p := rp.MustGetObject()

	param, err := g.handleNormalParam(ctx, contextStack, rp, schemas.ParamTypeQuery, globalNameOverrides, scope, docInfo)
	if err != nil {
		return nil, err
	}

	style := oas.SerializationStyleForm
	if p.GetStyle() != "" {
		style = p.GetStyle()
	}

	switch style {
	case "form":
	case "deepObject":
		if param.Field.Type.IsPrimitive() {
			return nil, errors.NewUnsupportedError(fmt.Sprintf("query parameters with type %s not supported for style deepObject", param.Field.Type.Type), param.Schema.MustGetResolvedSchema().GetSchema().GetCore().Type.GetKeyNodeOrRoot(param.Schema.MustGetResolvedSchema().GetSchema().GetRootNode()))
		}
	case "pipeDelimited":
		if !param.Field.Type.IsSimpleObjectOrContainerType() {
			return nil, errors.NewUnsupportedError(fmt.Sprintf("query parameters with type %s not supported for style pipeDelimited", param.Field.Type.Type), param.Schema.MustGetResolvedSchema().GetSchema().GetCore().Type.GetKeyNodeOrRoot(param.Schema.MustGetResolvedSchema().GetSchema().GetRootNode()))
		}
	default:
		return nil, errors.NewUnsupportedError(fmt.Sprintf("query parameters with style %s not currently supported", style), p.GetCore().Style.GetKeyNodeOrRoot(p.GetRootNode()))
	}

	paramAnnotation := &ast.ParamAnnotation{
		AllowReserved:        paramAllowsReserved(ctx, rp, g.subsystem.Extensions) && g.subsystem.Features.IsFeatureSupported(ctx, features.FeatureAllowReserved),
		Explode:              openapi.GetLibOpenAPIExplodeDefault(p),
		FieldType:            param.Field.Type,
		Hidden:               g.subsystem.Extensions.IsGlobalHidden(ctx, p),
		Name:                 p.GetName(),
		ParamType:            ast.ParamTypeQueryParam,
		RequiredForOperation: !param.Field.Optional,
		Style:                string(style),
	}

	param.Field.Annotations = append(param.Field.Annotations, paramAnnotation)

	return param, nil
}

func (g *Generator) handleHeaderParameters(ctx context.Context, opts handleSpecificParametersOptions) ([]*ast.Param, ast.Fields, bool, error) {
	if len(opts.params) == 0 {
		return nil, opts.globalParams, false, nil
	}

	globalParams := opts.globalParams
	headerParams := []*ast.Param{}

	usesUserAgentHeader := false

	for _, prm := range opts.params {
		p := prm.Param.MustGetObject()

		if p.GetContent().Len() > 0 {
			return nil, nil, false, errors.NewUnsupportedError("header parameters with content not currently supported", p.GetCore().Content.GetKeyNodeOrRoot(p.GetRootNode()))
		}

		if slices.Contains([]string{"accept", "content-type", "authorization"}, strings.ToLower(p.GetName())) {
			continue
		}

		if strings.ToLower(p.GetName()) == "user-agent" {
			usesUserAgentHeader = true
		}

		param, err := g.handleNormalParam(ctx, append(opts.contextStack, ast.ContextFrame{
			Type:       ast.ContextTypeParameter,
			Identifier: "header",
		}), prm.Param, schemas.ParamTypeHeader, opts.globalNameOverrides, opts.scope, opts.docInfo)
		if err != nil {
			return nil, nil, false, err
		}

		paramAnnotation := &ast.ParamAnnotation{
			AllowReserved:        false,
			Explode:              openapi.GetLibOpenAPIExplodeDefault(p),
			FieldType:            param.Field.Type,
			Hidden:               g.subsystem.Extensions.IsGlobalHidden(ctx, p),
			Name:                 p.GetName(),
			ParamType:            ast.ParamTypeHeader,
			RequiredForOperation: !param.Field.Optional,
			Style:                "simple",
		}

		param.Field.Annotations = append(param.Field.Annotations, paramAnnotation)

		description := prm.Param.GetDescription()
		if description == "" {
			description = p.GetDescription()
		}

		if description != "" {
			param.Field.Comments = &ast.Comment{
				Description: description,
			}
		} else if param.Field.Type.Comments != nil && param.Field.Type.Comments.Description != "" {
			param.Field.Comments = &ast.Comment{
				Description: param.Field.Type.Comments.Description,
			}
		}

		param.Field.Comments, err = handlers.HandleDeprecated(ctx, param.Field.Comments, p.GetDeprecated(), p.GetExtensions(), g.subsystem)
		if err != nil {
			return nil, nil, false, err
		}

		param.Field.ParameterIndex = pointer.From(prm.Idx)

		globalHidden, globalParam := g.handleGlobalParameter(ctx, opts.globals, param.Field, opts.opID)
		if globalParam != nil {
			globalParams, err = globalParams.AddField(globalParam, g.subsystem.Config.MaintainOpenAPIOrder())
			if err != nil {
				return nil, nil, false, errors.NewValidationError("duplicated field generated for global header parameter", p.GetCore().Name.GetKeyNodeOrRoot(p.GetRootNode()), fmt.Errorf("header parameter %s duplicated: %w", p.GetName(), err))
			}
		}

		headerParams = append(headerParams, &ast.Param{
			Field:    param.Field,
			Hidden:   globalHidden,
			Examples: param.Examples,
		})
	}

	if len(headerParams) == 0 {
		return nil, globalParams, usesUserAgentHeader, nil
	}

	return headerParams, globalParams, usesUserAgentHeader, nil
}

func handleCookieParameters(_ context.Context, _ ast.ContextStack, params []paramOAS) error {
	if len(params) == 0 {
		return nil
	}
	return errors.NewUnsupportedError("cookie parameters not currently supported", params[0].Param.MustGetObject().GetCore().In.GetKeyNodeOrRoot(params[0].Param.MustGetObject().GetRootNode()))
}

func (g *Generator) handleNormalParam(ctx context.Context, contextStack ast.ContextStack, rp *oas.ReferencedParameter, paramType schemas.ParamType, globalNameOverrides []*extensions.NameOverride, scope ast.Scope, docInfo *document.DocumentInfo) (*param, error) {
	p := rp.MustGetObject()
	schema := p.GetSchema()

	resolvedSchema, err := resolution.Resolve(ctx, schema, docInfo)
	if err != nil {
		return nil, err
	}

	if rp.IsReference() {
		scope = ast.ScopeShared
		// We are setting the schema ref to the param ref, so that we can ensure it is associated as a dependency correctly
		if !schema.IsReference() {
			schema = oas3.NewReferencedScheme(ctx, rp.GetReference(), resolvedSchema)
		}
	}

	paramField, err := g.schemas.HandleSchema(ctx, schemas.Params{
		ContextStack: addPropScope(contextStack, rp),
		ParamType:    paramType,
		Schema:       schema,
		Scope:        scope,
		IsRequest:    true,
		Depth:        1,
		MaxDepth:     1,
		DocInfo:      docInfo,
	})
	if err != nil {
		return nil, err
	}
	paramField.Name = p.GetName()
	paramField.OriginalName = p.GetName()
	if (paramField.Type.Comments == nil || paramField.Type.Comments.Description == "") && paramField.Type.Scope != ast.ScopeShared {
		description := rp.GetDescription()
		if description == "" {
			description = p.GetDescription()
		}

		paramField.Type.Comments = &ast.Comment{
			Description: description,
		}
	}

	if len(paramField.Type.Examples) == 0 && p.GetExample() != nil {
		exampleNode := ast.NewExample("", "", p.GetExample())
		paramField.Type.Examples = []*ast.Example{
			exampleNode,
		}
	}

	err = paramField.Type.Extensions.MergeWithoutOverwrite(g.subsystem.Extensions, paramField.Type, p.GetExtensions())
	if err != nil {
		return nil, err
	}

	paramField, err = g.overrideParameterName(ctx, paramField, p, globalNameOverrides)
	if err != nil {
		return nil, err
	}

	// OpenAPI requires path params to be required=true; we warn elsewhere but always force required here.
	switch paramType {
	case schemas.ParamTypePath:
		paramField.Optional = false
	default:
		paramField.Optional = p.Required == nil || !p.GetRequired()
	}

	examples, err := getParamExamples(ctx, p, docInfo)
	if err != nil {
		return nil, err
	}

	return &param{
		Field:    paramField,
		Schema:   schema,
		Examples: examples,
	}, nil
}

func addPropScope(contextStack ast.ContextStack, p *oas.ReferencedParameter) ast.ContextStack {
	if p.IsReference() {
		refName, refType := namer.GetRefName(p.GetReference())

		return ast.ContextStack{
			ast.ContextFrame{
				Type:       ast.ContextTypeRefType,
				Identifier: refType,
			},
			ast.ContextFrame{
				Type:       ast.ContextTypeRefName,
				Identifier: refName,
			},
		}
	} else {
		contextStack = append(contextStack, ast.ContextFrame{
			Type:       ast.ContextTypeProperty, // TODO this may need to be a different type like ContextTypeParameter
			Identifier: p.MustGetObject().GetName(),
		})
	}

	return contextStack
}

func getParamExamples(ctx context.Context, p *oas.Parameter, docInfo *document.DocumentInfo) ([]*ast.Example, error) {
	if p.GetExamples().Len() > 0 {
		examples := []*ast.Example{}
		for name, e := range p.GetExamples().All() {
			example, err := resolution.Resolve(ctx, e, docInfo)
			if err != nil {
				return nil, err
			}

			examples = append(examples, ast.NewExample(name, example.GetDescription(), resolveExampleValue(example)))
		}
		return examples, nil
	}

	if p.GetExample() != nil {
		return []*ast.Example{
			ast.NewExample("", "", p.GetExample()),
		}, nil
	}

	return nil, nil
}

func resolveParameters(ctx context.Context, params []*oas.ReferencedParameter, docInfo *document.DocumentInfo) error {
	for _, p := range params {
		_, err := resolution.Resolve(ctx, p, docInfo)
		if err != nil {
			return err
		}
	}

	return nil
}
