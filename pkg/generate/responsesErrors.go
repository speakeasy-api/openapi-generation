package generate

import (
	"context"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
)

type errorTypeUpdateOptions struct {
	isRootType     bool
	addRawResponse bool
	httpMetaType   *ast.TypeDef
	resFormat      string
	depth          int
}

func (g *Generator) cloneTypeToErrorsScope(ctx context.Context, typ *ast.TypeDef, options errorTypeUpdateOptions) (*ast.TypeDef, error) {
	// Config.Generation.Fixes.SharedErrorComponentsApr2025 has been introduced as the legacy logic
	// could produce import cycles: `Shared -> Error -> Shared` due to flawed code when
	// cloning classes to the errors scope. Shallow cloning of classes resulted in updates to the
	// error classes mutating the shared classes.
	// With the introduction of this flag, the logic has been changed such that only the top level
	// response objects will get cloned or moved. Unlike the legacy logic, we don't recursively move
	// all child types to the errors scope.
	if g.subsystem.Config.Generation.Fixes.SharedErrorComponentsApr2025 {
		return g.cloneTypeToErrorsScopeInner(ctx, typ, options)
	}

	return g.cloneTypeToErrorsScopeLegacy(ctx, typ, options, map[*ast.TypeDef]bool{})
}

func (g *Generator) cloneTypeToErrorsScopeInner(ctx context.Context, t *ast.TypeDef, options errorTypeUpdateOptions) (*ast.TypeDef, error) {
	options.depth++

	switch t.Type {
	case ast.DataTypeError:
		return g.cloneErrorToErrorsScope(ctx, t, options)
	case ast.DataTypeClass:
		return g.cloneClassToErrorsScope(ctx, t, options)
	case ast.DataTypeUnion:
		return g.cloneUnionToErrorsScope(ctx, t, options)
	default:
		return nil, errors.NewUnsupportedError("only objects or union types can be used as error types", t.GetLocationNode())
	}
}

func (g *Generator) cloneErrorToErrorsScope(ctx context.Context, t *ast.TypeDef, _ errorTypeUpdateOptions) (*ast.TypeDef, error) {
	// Errors should already be in the errors scope, no moving or cloning is needed
	if t.Scope != ast.ScopeErrors {
		return nil, errors.NewUnsupportedError("unexpected error type not in the errors scope: "+t.Name, t.GetLocationNode())
	}
	if !g.subsystem.Register.IsTypeRegistered(t) {
		return g.subsystem.Register.RegisterType(ctx, t, false), nil
	}
	return t, nil
}

func (g *Generator) cloneClassToErrorsScope(ctx context.Context, t *ast.TypeDef, options errorTypeUpdateOptions) (*ast.TypeDef, error) {
	// Maybe we can cast this `class` into an `error` type or we have to clone it so we keep around the `class`
	t = g.subsystem.Register.UnregisterOrCloneTypeInPreparationForCreatingErrorType(ctx, t)

	// Cast this `class` into an `error` type
	t.Type = ast.DataTypeError
	// Move it to the errors scope
	t.Scope = ast.ScopeErrors

	// Drop the duplicate frames: they would create unnecessary "1", "2" suffixes
	// Note: this mutates both the original and the copy
	t.ContextStack.Filter(func(frame ast.ContextFrame) bool {
		return frame.Type != ast.ContextTypeRegisterDuplicate
	})

	// Add any fields like `err.HttpMeta` or `err.RawResponse`
	g.addErrorResponseMetadata(ctx, t, options)
	return g.subsystem.Register.RegisterType(ctx, t, false), nil
}

func (g *Generator) cloneUnionToErrorsScope(ctx context.Context, t *ast.TypeDef, options errorTypeUpdateOptions) (*ast.TypeDef, error) {
	if !g.subsystem.Features.IsFeatureSupported(ctx, features.FeatureErrorUnions) {
		return nil, errors.NewUnsupportedError("union types are not currently supported as error types", t.GetLocationNode())
	}
	g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureErrorUnions)

	// Only convert top level union options to error types
	if options.depth == 1 {
		// Build mapping from old child types to new child types
		childTypeMap := make(map[*ast.TypeDef]*ast.TypeDef)
		for i, child := range t.AssociatedTypes {
			originalChild := child
			child, err := g.cloneTypeToErrorsScopeInner(ctx, child, options)
			if err != nil {
				return nil, err
			}
			childTypeMap[originalChild] = child
			t.AssociatedTypes[i] = child
		}

		// Also update Discriminator.Mapping to point to the new types
		if t.Discriminator != nil {
			for _, mapping := range t.Discriminator.Mapping {
				if mapping.Type != nil {
					if newType, ok := childTypeMap[mapping.Type]; ok {
						mapping.Type = newType
					}
				}
			}
		}
	}

	t = g.subsystem.Register.UnregisterOrCloneTypeInPreparationForCreatingErrorType(ctx, t)
	t.Scope = ast.ScopeErrors
	if g.subsystem.Features.IsUnionWrapper(ctx) {
		g.addErrorResponseMetadata(ctx, t, options)
	}

	// Drop the duplicate frames: they would create unnecessary "1", "2" suffixes
	// Note: this mutates both the original and the copy
	t.ContextStack.Filter(func(frame ast.ContextFrame) bool {
		return frame.Type != ast.ContextTypeRegisterDuplicate
	})
	g.addErrorResponseMetadata(ctx, t, options)
	return g.subsystem.Register.RegisterType(ctx, t, false), nil
}

func (g *Generator) addErrorResponseMetadata(ctx context.Context, typ *ast.TypeDef, options errorTypeUpdateOptions) {
	if options.addRawResponse {
		switch options.resFormat {
		case responseFormatEnvelope:
			typ.Fields = typ.Fields.EnsureField(&ast.FieldDef{
				Name:     "RawResponse",
				Type:     ast.NewType(&ast.TypeDef{Type: ast.DataTypeResponse}, nil),
				Optional: true,
				Annotations: ast.Annotations{
					&ast.JSONAnnotation{Ignore: true, FieldName: "-"},
				},
				Comments: &ast.Comment{
					Description: rawResponseDesc,
				},
			}, g.subsystem.Config.MaintainOpenAPIOrder())
		case responseFormatEnvelopeHTTP:
			typ.Fields = typ.Fields.EnsureField(&ast.FieldDef{
				Name: "HttpMeta",
				Type: g.subsystem.Register.RegisterType(ctx, options.httpMetaType, false),
				Annotations: ast.Annotations{
					&ast.JSONAnnotation{Ignore: true, FieldName: "-"},
				},
			}, g.subsystem.Config.MaintainOpenAPIOrder())
		default:
			return
		}
	}
}
