package generate

import (
	"context"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
)

func (g *Generator) generateHTTPMetadataType(_ context.Context) *ast.TypeDef {
	typeDef := ast.NewType(&ast.TypeDef{
		Name:        "HTTPMetadata",
		Type:        ast.DataTypeClass,
		Scope:       ast.ScopeShared,
		IsComponent: true,
		Output:      true,
		Fields: ast.Fields{
			&ast.FieldDef{
				Name: "Response",
				Type: ast.NewType(&ast.TypeDef{Type: ast.DataTypeResponse}, nil),
				Comments: &ast.Comment{
					Description: rawResponseDesc,
				},
				Annotations: ast.Annotations{
					&ast.JSONAnnotation{Ignore: true, FieldName: "-"},
					// Synthetic field — apply target casing even when the
					// language config says to preserve spec field names.
					// Without this, no-zod / preserveModelFieldNames variants
					// would carry the literal `Response` / `Request` casing
					// while the parent wrapper still uses `httpMeta` (which
					// is forced via NeedsCasing on the response side).
					&ast.NeedsCasingAnnotation{},
				},
			},
			&ast.FieldDef{
				Name: "Request",
				Type: ast.NewType(&ast.TypeDef{Type: ast.DataTypeRequest}, nil),
				Comments: &ast.Comment{
					Description: rawRequestDesc,
				},
				Annotations: ast.Annotations{
					&ast.JSONAnnotation{Ignore: true, FieldName: "-"},
					&ast.NeedsCasingAnnotation{},
				},
			},
		},
	}, ast.ContextStack{})

	return typeDef
}
