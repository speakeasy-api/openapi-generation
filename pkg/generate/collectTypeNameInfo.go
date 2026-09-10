package generate

import (
	"fmt"
	"strings"

	"github.com/ettle/strcase"
	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
)

type typeNameInfo struct {
	Name string
	// A short summary of the type fields
	Description   string
	Depth         int
	ID            string
	NameWithDepth string
	Scope         string
	Type          *ast.TypeDef
}

// This is only used for unit testing the AST
func collectTypeNameInfo(a *ast.AST) []typeNameInfo {
	lines := make([]typeNameInfo, 0)
	seen := make(map[string]bool)

	getName := func(t *ast.TypeDef) string {
		if t.ItemType != nil {
			return "array"
		}

		parts := make([]string, 0)

		if t.Name != "" {
			pascal := strcase.ToPascal(t.Name)
			if strings.TrimSpace(pascal) != "" {
				parts = append(parts, pascal)
			}
		}
		return strings.Join(parts, ".")
	}

	getDescription := func(t *ast.TypeDef) string {
		var typeDescription string
		switch t.Type {
		case "enum":
			values := t.Enum.Values
			if len(values) == 0 {
				typeDescription = "enum"
			} else {
				typeDescription = "enum: " + strings.Join(values[:min(3, len(values))], ", ")
				if len(values) > 3 {
					typeDescription += " ..."
				}
			}
		case "class":
			fields := t.Fields
			descParts := make([]string, 0, 3)
			for i := 0; i < min(3, len(fields)); i++ {
				f := fields[i]
				name := f.Name
				typ := string(f.Type.Type)
				if f.Type.IsComponent {
					typ = f.Type.Name
				}
				descParts = append(descParts, fmt.Sprintf("%s: %s", name, typ))
			}
			typeDescription = strings.Join(descParts, ", ")
			if len(fields) > 3 {
				typeDescription += " ..."
			}
			if typeDescription == "" {
				typeDescription = "empty"
			}
			if t.Scope == ast.ScopeSDK {
				typeDescription = "SDK " + typeDescription
			}
		default:
			typeDescription = string(t.Type)
		}
		return typeDescription
	}

	visited := make(map[string]bool)

	visit := func(node ast.Node, parents []ast.Node, a *ast.AST) error {
		if ast.GetNodeType(node) != ast.NodeTypeTypeDef {
			return nil
		}

		t := node.(*ast.TypeDef)
		if !t.IsCustomType() {
			return nil
		}

		id := "scope:" + string(t.Scope) + " type:" + string(t.Type)
		if len(t.ContextStack) > 0 {
			id = t.GetRegistrationID()
		}

		if seen[id] {
			return nil
		}
		seen[id] = true
		numParentTypeDefs := 0
		for _, p := range parents {
			if ast.GetNodeType(p) == ast.NodeTypeTypeDef {
				numParentTypeDefs++
			}
		}
		description := getDescription(t)

		nameWithPath := strings.Repeat(" ", numParentTypeDefs) + getName(t)
		if description != "" {
			nameWithPath += " (" + description + ")"
		}

		lines = append(lines, typeNameInfo{
			Name:          t.Name,
			Depth:         numParentTypeDefs,
			ID:            id,
			NameWithDepth: nameWithPath,
			Description:   getDescription(t),
			Scope:         string(t.Scope),
			Type:          t,
		})
		return nil
	}

	_ = a.WalkSDK(a.MainSDK, []ast.Node{}, visit, visited)

	return lines
}
