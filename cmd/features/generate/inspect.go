package main

import (
	"go/ast"
	"go/token"
	"log"
)

func inspect(rootNode ast.Node, typ string, onFindConst func(string, *ast.ValueSpec)) {
	ast.Inspect(rootNode, func(n ast.Node) bool {
		decl, ok := n.(*ast.GenDecl)
		// Not a const continue
		if !ok || decl.Tok != token.CONST {
			return true
		}

		currentType := ""

		for _, spec := range decl.Specs {
			vs := spec.(*ast.ValueSpec)
			if vs.Type == nil && len(vs.Values) > 0 {
				currentType = ""

				ce, ok := vs.Values[0].(*ast.CallExpr)
				if !ok {
					continue
				}
				id, ok := ce.Fun.(*ast.Ident)
				if !ok {
					continue
				}

				currentType = id.Name
			}
			if vs.Type != nil {
				id, ok := vs.Type.(*ast.Ident)
				if !ok {
					continue
				}

				currentType = id.Name
			}
			if currentType != typ {
				continue
			}

			if len(vs.Names) != 1 {
				log.Fatalf("expected only one name in const declaration, got %d", len(vs.Names))
			}

			name := vs.Names[0].Name
			if name == "_" {
				continue
			}

			onFindConst(name, vs)
		}

		return false
	})
}
