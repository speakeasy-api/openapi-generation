package ast_post_processing

import (
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
)

// MarkOpenEnums marks enums as open (tolerating unknown values) based on configuration and usage.
// openSingleValue extends the pass to single-value enums (fixes.openSingleValueEnumsOct2026);
// enums that act as union discriminator tags stay closed either way.
func MarkOpenEnums(types []*ast.TypeDef, cfg map[string]any, openSingleValue bool) {
	forwardCompatibleEnumsByDefault := false
	if forwardCompatibleEnumsByDefaultConfig, ok := cfg["forwardCompatibleEnumsByDefault"].(bool); ok {
		forwardCompatibleEnumsByDefault = forwardCompatibleEnumsByDefaultConfig
	}

	if !forwardCompatibleEnumsByDefault {
		return
	}

	var tagEnums map[*ast.TypeDef]bool
	if openSingleValue {
		tagEnums = collectDiscriminatorTagEnums(types)
	}

	for _, typeDef := range types {
		markEnumAsOpenIfNeeded(typeDef, openSingleValue, tagEnums)
	}
}

func markEnumAsOpenIfNeeded(typeDef *ast.TypeDef, openSingleValue bool, tagEnums map[*ast.TypeDef]bool) {
	if typeDef.Type != ast.DataTypeEnum || typeDef.Enum == nil {
		return
	}

	if !typeDef.UsedInResponse {
		return
	}

	if len(typeDef.Enum.Values) == 0 {
		return
	}

	if len(typeDef.Enum.Values) == 1 {
		if !openSingleValue || tagEnums[typeDef] {
			return
		}
	}

	if typeDef.Extensions != nil && typeDef.Extensions.All != nil {
		if _, hasExplicitSetting := typeDef.Extensions.All["x-speakeasy-unknown-values"]; hasExplicitSetting {
			// Respect the explicit setting - don't override
			return
		}
	}

	if matchesSpecialCasePattern(typeDef.Enum.Values) {
		return
	}

	typeDef.Enum.Open = true
}

// collectDiscriminatorTagEnums returns the enum types that identify a union
// member through the union's discriminator property. Those stay closed: the
// union is dispatched on their value and open unions cover unknown tags.
func collectDiscriminatorTagEnums(types []*ast.TypeDef) map[*ast.TypeDef]bool {
	tagEnums := map[*ast.TypeDef]bool{}

	for _, typeDef := range types {
		if typeDef == nil || typeDef.Discriminator == nil {
			continue
		}

		for _, mapping := range typeDef.Discriminator.Mapping {
			if mapping == nil || mapping.Type == nil {
				continue
			}

			collectDiscriminatorTagEnumsFromType(
				mapping.Type,
				typeDef.Discriminator.TypePropertyName,
				tagEnums,
				map[*ast.TypeDef]bool{},
			)
		}
	}

	return tagEnums
}

func collectDiscriminatorTagEnumsFromType(member *ast.TypeDef, propertyName string, tagEnums, visited map[*ast.TypeDef]bool) {
	if member == nil || visited[member] {
		return
	}
	visited[member] = true

	field := findDiscriminatorField(member, propertyName)
	if field != nil && field.Type != nil && field.Type.Type == ast.DataTypeEnum {
		tagEnums[field.Type] = true
	}

	if member.Type != ast.DataTypeUnion {
		return
	}
	for _, associatedType := range member.AssociatedTypes {
		collectDiscriminatorTagEnumsFromType(associatedType, propertyName, tagEnums, visited)
	}
}

func findDiscriminatorField(member *ast.TypeDef, propertyName string) *ast.FieldDef {
	for _, field := range member.Fields {
		if field.OriginalName == propertyName {
			return field
		}
	}

	return member.Fields.GetField(propertyName)
}

// Special case patterns that should automatically be marked as closed
var specialCasePatterns = []map[string]bool{
	{
		"monday":    true,
		"tuesday":   true,
		"wednesday": true,
		"thursday":  true,
		"friday":    true,
		"saturday":  true,
		"sunday":    true,
	},
	{
		"mon": true,
		"tue": true,
		"wed": true,
		"thu": true,
		"fri": true,
		"sat": true,
		"sun": true,
	},
	{
		"january":   true,
		"february":  true,
		"march":     true,
		"april":     true,
		"may":       true,
		"june":      true,
		"july":      true,
		"august":    true,
		"september": true,
		"october":   true,
		"november":  true,
		"december":  true,
	},
	{
		"jan": true,
		"feb": true,
		"mar": true,
		"apr": true,
		"may": true,
		"jun": true,
		"jul": true,
		"aug": true,
		"sep": true,
		"oct": true,
		"nov": true,
		"dec": true,
	},
}

func matchesSpecialCasePattern(enumValues []string) bool {
	if len(enumValues) == 0 {
		return false
	}

	for _, pattern := range specialCasePatterns {
		if len(enumValues) != len(pattern) {
			continue
		}

		allMatch := true
		for _, enumValue := range enumValues {
			normalizedValue := strings.ToLower(enumValue)

			if !pattern[normalizedValue] {
				allMatch = false
				break
			}
		}

		if allMatch {
			return true
		}
	}

	return false
}
