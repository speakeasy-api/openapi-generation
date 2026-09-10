package namer

import "github.com/speakeasy-api/openapi-generation/v2/internal/ast"

// enumNamer computes the lowercased enum constant names for a TypeDef, used
// during name resolution to detect conflicts between enum constants and type names.
type enumNamer interface {
	EnumNames(t *ast.TypeDef) []string
}

// newEnumNamer returns the enumNamer for the given target, or nil if the
// target does not need enum name resolution.
func newEnumNamer(target string) enumNamer {
	switch target {
	case "go", "mockserver", "terraform":
		return NewGoEnumNamer()
	default:
		return nil
	}
}
