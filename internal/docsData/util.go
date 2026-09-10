package docsData

import (
	"encoding/json"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
)

func GetFormattedExamples(examples ast.Examples) []string {
	formattedExamples := make([]string, 0)
	for _, example := range examples {
		// Sometimes examples ends up being an empty string here if the examples in
		// the spec isn't formatted correctly, which we want to filter out
		if example.Value == nil || example.Value.Value == "" {
			continue
		}
		formattedExamples = append(formattedExamples, example.Value.Value)
	}
	return formattedExamples
}

func GetSerializedDefault(field *ast.FieldDef) *string {
	var defaultValueStr *string
	if field.Default != nil && field.Default.Value != nil {
		bytes, err := json.Marshal(field.Default.Value)
		if err == nil {
			str := string(bytes)
			defaultValueStr = &str
		}
	}
	return defaultValueStr
}
