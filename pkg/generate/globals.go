package generate

import (
	"context"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/speakeasy-api/openapi/openapi"
)

func (g *Generator) handleGlobals(ctx context.Context, doc *openapi.OpenAPI, a *ast.AST, globalNameOverrides []*extensions.NameOverride, docInfo *document.DocumentInfo) (*ast.TypeDef, error) {
	globalVariables, err := g.subsystem.Extensions.HandleGlobalsExtension(ctx, doc)
	if err != nil {
		return nil, err
	}

	if globalVariables == nil || len(globalVariables.Parameters) == 0 {
		return nil, nil
	}

	// TODO: we probably need to improve our handling of this as certain types may be rewritten to not match the global anymore
	globalParams, _, _, err := g.handleParameters(ctx, handleParametersOptions{
		contextStack:        ast.ContextStack{},
		paramsOAS:           globalVariables.Parameters,
		globalNameOverrides: globalNameOverrides,
		scope:               ast.ScopeOperations,
		docInfo:             docInfo,
	})
	if err != nil {
		return nil, err
	}

	fields := ast.Fields{}

	if globalParams != nil {
		if globalParams.QueryParams != nil {
			fields = g.addGlobalParamFields(fields, globalParams.QueryParams)
		}
		if globalParams.PathParams != nil {
			fields = g.addGlobalParamFields(fields, globalParams.PathParams)
		}
		if globalParams.HeaderParams != nil {
			fields = g.addGlobalParamFields(fields, globalParams.HeaderParams)
		}
	}

	if len(fields) == 0 {
		return nil, nil
	}

	globals := g.subsystem.Register.RegisterType(ctx, ast.NewType(&ast.TypeDef{
		Name:   "Globals",
		Type:   ast.DataTypeClass,
		Fields: fields,
		Scope:  ast.ScopeGlobals,
	}, ast.ContextStack{}), true)

	if a != nil && a.MainSDK != nil {
		a.MainSDK.Globals = globals
	}

	return globals, nil
}

func (g *Generator) addGlobalParamFields(fields ast.Fields, paramFields []*ast.Param) ast.Fields {
	for _, param := range paramFields {
		param.Field.Optional = true // Forcing to true for now as we don't have a way to determine if a global parameter is required correctly

		param.Field.Annotations.Get(ast.AnnotationTypeParam).(*ast.ParamAnnotation).IsGlobal = true

		ff, err := fields.AddField(param.Field, g.subsystem.Config.MaintainOpenAPIOrder())
		// Error should never happen as we validate fields are unique
		if err != nil {
			continue
		}
		fields = ff
	}

	return fields
}
