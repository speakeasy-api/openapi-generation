package namer

import (
	"strings"

	"github.com/ettle/strcase"
	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/casing"
	"github.com/speakeasy-api/openapi-generation/v2/internal/sanitization"
)

// enumCasing* are the disambiguation suffixes appended when two enum values
// produce the same base name. They are used as-is (already PascalCase).
const (
	enumCasingUpper = "Upper"
	enumCasingLower = "Lower"
	enumCasingMixed = "Mixed"
)

// goEnumNamer computes enum constant names for Go-like targets (go, mockserver,
// terraform). Each constant name is the PascalCase type name followed by a
// PascalCase value name, lowercased for conflict detection.
type goEnumNamer struct {
	caser *casing.Casing
}

// NewGoEnumNamer returns a goEnumNamer configured for Go-like targets.
func NewGoEnumNamer() goEnumNamer {
	return goEnumNamer{caser: casing.New()}
}

// EnumNames returns the lowercased enum constant names for t. Each name is the
// PascalCase type name followed by a PascalCase value name.
func (n goEnumNamer) EnumNames(t *ast.TypeDef) []string {
	if t.Enum == nil {
		return nil
	}

	typeName := n.caser.ToGoPascal(sanitization.SanitizeName(t.Name))

	if len(t.Enum.Names) > 0 {
		result := make([]string, len(t.Enum.Names))
		for i, name := range t.Enum.Names {
			result[i] = strings.ToLower(typeName + strcase.ToPascal(sanitization.SanitizeName(name)))
		}
		return result
	}

	return n.namesFromValues(typeName, t.Enum.Values)
}

// baseName converts a raw enum value to a PascalCase base name, treating an
// empty value as "Unknown".
func (n goEnumNamer) baseName(value string) string {
	name := strings.TrimSpace(value)
	if name == "" {
		name = "Unknown"
	}
	return strcase.ToPascal(sanitization.SanitizeName(name))
}

// casingSuffix classifies a value as "Upper", "Lower", or "Mixed", used as a
// disambiguation suffix when multiple values share the same base name.
func (n goEnumNamer) casingSuffix(v string) string {
	if strings.ToUpper(v) == v {
		return enumCasingUpper
	}
	if strings.ToLower(v) == v {
		return enumCasingLower
	}
	return enumCasingMixed
}

// namesFromValues derives lowercased enum constant names from raw values. When
// two values produce the same base name, a casing-style suffix ("Upper",
// "Lower", or "Mixed") is appended to disambiguate.
func (n goEnumNamer) namesFromValues(typeName string, values []string) []string {
	baseNames := make([]string, len(values))
	nameCounts := make(map[string]int, len(values))
	for i, v := range values {
		base := n.baseName(v)
		baseNames[i] = base
		nameCounts[base]++
	}

	result := make([]string, len(values))
	for i, v := range values {
		name := baseNames[i]
		if nameCounts[name] > 1 {
			name += n.casingSuffix(v)
		}
		result[i] = strings.ToLower(typeName + n.caser.ToGoPascal(sanitization.SanitizeName(name)))
	}
	return result
}
