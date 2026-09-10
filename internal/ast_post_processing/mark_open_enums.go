package ast_post_processing

import (
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
)

// MarkOpenEnums marks enums as open (tolerating unknown values) based on configuration and usage.
func MarkOpenEnums(types []*ast.TypeDef, cfg map[string]any) {
	forwardCompatibleEnumsByDefault := false
	if forwardCompatibleEnumsByDefaultConfig, ok := cfg["forwardCompatibleEnumsByDefault"].(bool); ok {
		forwardCompatibleEnumsByDefault = forwardCompatibleEnumsByDefaultConfig
	}

	if !forwardCompatibleEnumsByDefault {
		return
	}

	for _, typeDef := range types {
		markEnumAsOpenIfNeeded(typeDef)
	}
}

func markEnumAsOpenIfNeeded(typeDef *ast.TypeDef) {
	if typeDef.Type != ast.DataTypeEnum || typeDef.Enum == nil {
		return
	}

	if !typeDef.UsedInResponse {
		return
	}

	if len(typeDef.Enum.Values) <= 1 {
		// If there is only one value, it is almost always intended as a "const" rather than an enum
		return
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
