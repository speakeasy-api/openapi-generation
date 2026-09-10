package generate

import (
	"context"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/internal/resolution"
	"github.com/speakeasy-api/openapi-generation/v2/internal/schemas"
	"github.com/speakeasy-api/openapi/jsonschema/oas3"
	"github.com/speakeasy-api/openapi/references"
)

func (g *Generator) handleComponents(ctx context.Context, docInfo *document.DocumentInfo, a *ast.AST) error {
	if docInfo.Doc.Components == nil {
		return nil
	}

	for name, js := range docInfo.Doc.Components.Schemas.AllOrdered(g.subsystem.Config.GetSequencedMapIterationOrder()) {
		schema, err := resolution.Resolve(ctx, js, docInfo)
		if err != nil {
			return err
		}

		include, err := g.subsystem.Extensions.IncludeSchema(schema)
		if err != nil {
			return err
		}

		if !include {
			continue
		}

		ref := references.Reference("#/components/schemas/" + name)
		js = oas3.NewReferencedScheme(ctx, ref, schema)

		g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureIncludes)

		typeName, refType, _, err := g.namer.GetTypeName(ctx, js, ast.ContextStack{}, "", false)
		if err != nil {
			return err
		}

		contextStack := ast.ContextStack{}
		contextStack.AppendRefType(refType)
		contextStack.Append(ast.ContextTypeRefName, typeName)

		f, err := g.schemas.HandleSchema(ctx, schemas.Params{
			Schema:              js,
			ContextStack:        contextStack,
			Scope:               ast.ScopeShared,
			SerializationMethod: ast.SerializationMethodJSON,
			DocInfo:             docInfo,
		})
		if err != nil {
			return err
		}

		if a != nil && a.MainSDK != nil {
			a.MainSDK.AdditionalTypes = append(a.MainSDK.AdditionalTypes, f.Type)
		}
	}

	return nil
}
