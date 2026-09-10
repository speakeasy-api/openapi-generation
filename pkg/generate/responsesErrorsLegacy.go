package generate

import (
	"context"
	"fmt"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"gopkg.in/yaml.v3"
)

// Deprecated: use g.subsystem.Config.Generation.Fixes.SharedErrorComponentsApr2025
func (g *Generator) cloneTypeToErrorsScopeLegacy(ctx context.Context, typ *ast.TypeDef, options errorTypeUpdateOptions, visited map[*ast.TypeDef]bool) (*ast.TypeDef, error) {
	if visited[typ] {
		return typ, nil
	}

	visited[typ] = true

	switch typ.Type {
	case ast.DataTypeClass:
		fallthrough
	case ast.DataTypeError:
		return g.updateErrorClassScopeLegacy(ctx, typ, options, visited)
	case ast.DataTypeUnion:
		if !g.subsystem.Features.IsFeatureSupported(ctx, features.FeatureErrorUnions) {
			return nil, errors.NewUnsupportedError("union types are not currently supported as error types", nil)
		}
		g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureErrorUnions)

		return g.updateErrorUnionScopeLegacy(ctx, typ, options, visited)
	default:
		if options.isRootType {
			return nil, errors.NewUnsupportedError("only class and union types can be root error types", nil)
		}
		typ = g.subsystem.Register.UpdateRegisteredTypeToErrorsScopeLegacy(ctx, typ, false)
		return typ, nil
	}
}

// Deprecated: use g.subsystem.Config.Generation.Fixes.SharedErrorComponentsApr2025
func (g *Generator) updateErrorClassScopeLegacy(ctx context.Context, typ *ast.TypeDef, options errorTypeUpdateOptions, visited map[*ast.TypeDef]bool) (*ast.TypeDef, error) {
	g.addErrorResponseMetadata(ctx, typ, options)

	for i := range typ.Fields {
		childType, err := g.updateTypeToErrorsScopeLegacy(ctx, typ.Fields[i].Type, visited)
		if err != nil {
			return nil, err
		}

		typ.Fields[i].Type = childType
	}

	typ = g.subsystem.Register.UpdateRegisteredTypeToErrorsScopeLegacy(ctx, typ, options.isRootType)
	return typ, nil
}

// Deprecated: use g.subsystem.Config.Generation.Fixes.SharedErrorComponentsApr2025
func (g *Generator) updateErrorUnionScopeLegacy(ctx context.Context, typ *ast.TypeDef, options errorTypeUpdateOptions, visited map[*ast.TypeDef]bool) (*ast.TypeDef, error) {
	// Build mapping from old child types to new child types for updating Discriminator.Mapping
	childTypeMap := make(map[*ast.TypeDef]*ast.TypeDef)

	if g.subsystem.Features.IsUnionWrapper(ctx) {
		g.addErrorResponseMetadata(ctx, typ, options)

		for i, child := range typ.AssociatedTypes {
			originalChild := child
			t, err := g.updateTypeToErrorsScopeLegacy(ctx, child, visited)
			if err != nil {
				return nil, err
			}
			childTypeMap[originalChild] = t
			typ.AssociatedTypes[i] = t
		}

		typ = g.subsystem.Register.UpdateRegisteredTypeToErrorsScopeLegacy(ctx, typ, false)
	} else {
		for i, child := range typ.AssociatedTypes {
			originalChild := child
			t, err := g.cloneTypeToErrorsScopeLegacy(ctx, child, options, visited)
			if err != nil {
				return nil, err
			}
			childTypeMap[originalChild] = t
			typ.AssociatedTypes[i] = t
		}

		typ = g.subsystem.Register.UpdateRegisteredTypeToErrorsScopeLegacy(ctx, typ, false)
	}

	// Also update Discriminator.Mapping to point to the new types
	if typ.Discriminator != nil {
		for _, mapping := range typ.Discriminator.Mapping {
			if mapping.Type != nil {
				if newType, ok := childTypeMap[mapping.Type]; ok {
					mapping.Type = newType
				}
			}
		}
	}

	return typ, nil
}

// Deprecated: use g.subsystem.Config.Generation.Fixes.SharedErrorComponentsApr2025
func (g *Generator) updateTypeToErrorsScopeLegacy(ctx context.Context, typ *ast.TypeDef, visited map[*ast.TypeDef]bool) (*ast.TypeDef, error) {
	switch typ.Type {
	case ast.DataTypeClass:
		// We only move inline types to the errors scope
		if !typ.IsComponent {
			return g.cloneTypeToErrorsScopeLegacy(ctx, typ, errorTypeUpdateOptions{}, visited)
		}
	case ast.DataTypeMap:
		fallthrough
	case ast.DataTypeArray:
		itemType, err := g.updateTypeToErrorsScopeLegacy(ctx, typ.ItemType, visited)
		if err != nil {
			return nil, err
		}
		typ.ItemType = itemType
	case ast.DataTypeEnum:
		if !typ.IsComponent {
			typ = g.subsystem.Register.UpdateRegisteredTypeToErrorsScopeLegacy(ctx, typ, false)
		}
	case ast.DataTypeUnion:
		if !typ.IsComponent {
			for i, child := range typ.AssociatedTypes {
				t, err := g.updateTypeToErrorsScopeLegacy(ctx, child, visited)
				if err != nil {
					return nil, err
				}
				typ.AssociatedTypes[i] = t
			}

			typ = g.subsystem.Register.UpdateRegisteredTypeToErrorsScopeLegacy(ctx, typ, false)
		}
	}

	return typ, nil
}

type duplicateResponseFieldError struct {
	inner     error
	fieldName string
}

func (e *duplicateResponseFieldError) Error() string {
	return fmt.Sprintf("%s duplicated with different types: %s", e.fieldName, e.inner)
}

func (e *duplicateResponseFieldError) Unwrap() error {
	return e.inner
}

func (e *duplicateResponseFieldError) IntoValidationError(node *yaml.Node) *errors.ValidationError {
	return errors.NewValidationError("duplicate response fields with different types", node, e)
}
