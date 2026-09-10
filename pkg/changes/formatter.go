package changes

import (
	"strings"

	"github.com/stoewer/go-strcase"
)

// FormatMethodName formats a method name according to language conventions
func FormatMethodName(methodParts []string, lang string) string {
	return FormatMethodNameWithSDK(methodParts, lang, "")
}

// FormatMethodNameWithSDK formats a method name according to language conventions with optional SDK name
// Returns in format: sdkName.method.name() or method.name() if sdkName is empty
func FormatMethodNameWithSDK(methodParts []string, lang, sdkName string) string {
	if len(methodParts) == 0 {
		return ""
	}

	// Format all parts including SDK names
	formattedParts := make([]string, 0, len(methodParts)+1)

	// Add SDK name if provided
	if sdkName != "" {
		// Format SDK name according to language conventions
		var formattedSDK string
		switch lang {
		case "python", "ruby":
			formattedSDK = strcase.SnakeCase(sdkName)
		case "typescript", "javascript", "java", "php":
			formattedSDK = strings.ToLower(sdkName)
		case "go", "csharp", "cli":
			formattedSDK = strcase.UpperCamelCase(sdkName)
		default:
			formattedSDK = sdkName
		}
		formattedParts = append(formattedParts, formattedSDK)
	}

	for i, part := range methodParts {
		// Format each part according to language conventions
		// Dashes in names like "business-classifications" are treated as word boundaries
		var formatted string
		switch lang {
		case "python", "ruby":
			formatted = strcase.SnakeCase(part)
		case "typescript", "javascript", "java":
			formatted = strcase.LowerCamelCase(part)
		case "php":
			if i == 0 {
				formatted = strings.ToLower(part)
			} else {
				formatted = strcase.LowerCamelCase(part)
			}
		case "go", "csharp", "cli":
			formatted = strcase.UpperCamelCase(part)
		default:
			formatted = part
		}
		formattedParts = append(formattedParts, formatted)
	}

	switch lang {
	case "php":
		return strings.Join(formattedParts, "->")
	default:
		return strings.Join(formattedParts, ".")
	}
}

// FormatPropertyName formats a property name according to language conventions
func FormatPropertyName(propertyName string, lang string) string {
	switch lang {
	case "python":
		// Python uses snake_case for properties
		return strcase.SnakeCase(propertyName)
	case "go", "cli":
		// Go uses PascalCase for exported properties
		return strcase.UpperCamelCase(propertyName)
	case "typescript", "javascript", "java", "php", "csharp", "ruby":
		// Most languages use camelCase for properties
		return strcase.LowerCamelCase(propertyName)
	default:
		// Return original if language not specified
		return propertyName
	}
}

// FormatFieldPath formats a field path according to language conventions
func FormatFieldPath(fieldParts []string, lang string) string {
	if len(fieldParts) == 0 {
		return ""
	}

	formattedParts := make([]string, 0, len(fieldParts))
	for i, part := range fieldParts {
		// Don't apply formatting to union options or Map<> - they contain type names that should preserve casing
		if strings.HasPrefix(part, "union(") || strings.HasPrefix(part, "Map<") {
			formattedParts = append(formattedParts, part)
			continue
		}

		// Dashes in names like "business-classifications" are treated as word boundaries
		var formatted string
		switch lang {
		case "python", "ruby":
			formatted = strcase.SnakeCase(part)
		case "typescript", "javascript", "java":
			formatted = strcase.LowerCamelCase(part)
		case "go", "cli":
			formatted = strcase.UpperCamelCase(part)
		case "csharp":
			if i == 0 {
				formatted = strcase.LowerCamelCase(part)
			} else {
				formatted = strcase.UpperCamelCase(part)
			}
		case "php":
			formatted = strcase.LowerCamelCase(part)
		default:
			formatted = part
		}
		formattedParts = append(formattedParts, formatted)
	}

	// Join parts intelligently - don't add separator before [] or {}
	var result strings.Builder
	sep := "."
	if lang == "php" {
		sep = "->"
	}
	for i, part := range formattedParts {
		if i > 0 && !strings.HasPrefix(part, "[") && !strings.HasPrefix(part, "{") {
			result.WriteString(sep)
		}
		result.WriteString(part)
	}
	return result.String()
}
