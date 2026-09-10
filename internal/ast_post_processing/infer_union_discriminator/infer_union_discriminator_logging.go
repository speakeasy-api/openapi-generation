package infer_union_discriminator

import (
	"fmt"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
)

// getLocationInfo returns the file:line:col information for a type
// schemaPath is the full path to the schema file for clickable links
func getLocationInfo(typeDef *ast.TypeDef, schemaPath string) string {
	// Use the union's own location first, fall back to first member if not available
	loc := typeDef.Location
	if loc == nil && len(typeDef.AssociatedTypes) > 0 {
		loc = typeDef.AssociatedTypes[0].Location
	}
	if loc == nil || loc.Node == nil {
		return "?:?:?"
	}
	if schemaPath == "" {
		return fmt.Sprintf("%d:%d", loc.Node.Line, loc.Node.Column)
	}
	return fmt.Sprintf("%s:%d:%d", schemaPath, loc.Node.Line, loc.Node.Column)
}

// summarizeType returns a short description of the type similar to collectTypeNameInfo
func summarizeType(typeDef *ast.TypeDef) string {
	switch typeDef.Type {
	case ast.DataTypeEnum:
		if typeDef.Enum == nil || len(typeDef.Enum.Values) == 0 {
			return "enum"
		}
		values := typeDef.Enum.Values
		snapshot := "enum: " + strings.Join(values[:min(3, len(values))], ", ")
		if len(values) > 3 {
			snapshot += " ..."
		}
		return snapshot
	case ast.DataTypeClass, ast.DataTypeError:
		fields := typeDef.Fields
		if len(fields) == 0 {
			return string(typeDef.Type) + ": empty"
		}
		maxFields := 10
		descParts := make([]string, 0, maxFields)
		for i := 0; i < min(maxFields, len(fields)); i++ {
			f := fields[i]
			typ := string(f.Type.Type)
			if f.Const != nil {
				typ = fmt.Sprint(f.Const.Value)
			}
			if f.Type.IsComponent {
				typ = f.Type.Name
			}
			if f.Type.Enum != nil && len(f.Type.Enum.Values) == 1 {
				typ = f.Type.Enum.Values[0]
			}

			name := f.OriginalName
			if name == "" {
				name = f.Name
			}
			if f.Optional {
				name += "?"
			}
			descParts = append(descParts, fmt.Sprintf("%s: %s", name, typ))
		}
		snapshot := strings.Join(descParts, ", ")
		if len(fields) > maxFields {
			snapshot += " ..."
		}
		return string(typeDef.Type) + ": " + snapshot
	case ast.DataTypeUnion:
		return fmt.Sprintf("union: %d members", len(typeDef.AssociatedTypes))
	default:
		return string(typeDef.Type)
	}
}

func summarizeUnion(typeDef *ast.TypeDef) string {
	// One line per member with discriminator mapping if available
	members := make([]string, len(typeDef.AssociatedTypes))
	for i, member := range typeDef.AssociatedTypes {
		summary := summarizeType(member)

		// Add discriminator mapping if available
		if typeDef.Discriminator != nil && i < len(typeDef.Discriminator.Mapping) {
			mapping := typeDef.Discriminator.Mapping[i]
			if mapping != nil {
				summary = fmt.Sprintf("%s [%s=%s]", summary, typeDef.Discriminator.TypePropertyName, mapping.Name)
			}
		}

		members[i] = summary
	}
	return "  - " + strings.Join(members, "\n  - ")
}

// logDiscriminatorResult logs the result of discriminator inference for a union type
func logDiscriminatorResult(logger logging.Logger, typeDef *ast.TypeDef, err error, schemaPath string) {
	typeName := typeDef.Name
	if typeName == "" {
		typeName = "<unnamed>"
	}
	location := getLocationInfo(typeDef, schemaPath)
	description := summarizeUnion(typeDef)

	if typeDef.Discriminator == nil {
		logger.Debug(fmt.Sprintf("❌ no_discriminator - %s for union %q %s \n%s", err.Error(), typeName, location, description))
	} else {
		msg := "✓ explicit_discriminator"
		if typeDef.Discriminator.Inferred {
			msg = "✓ inferred_discriminator"
		}
		logger.Debug(fmt.Sprintf("%s: %q for union %q %s \n%s", msg, typeDef.Discriminator.TypePropertyName, typeName, location, description))
	}
}
