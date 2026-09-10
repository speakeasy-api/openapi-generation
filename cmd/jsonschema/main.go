package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/templates"
)

type JSONSchema struct {
	Schema               string                    `json:"$schema"`
	Title                string                    `json:"title"`
	Type                 string                    `json:"type"`
	Description          string                    `json:"description,omitempty"`
	Properties           map[string]map[string]any `json:"properties"`
	Required             []string                  `json:"required,omitempty"`
	AdditionalProperties bool                      `json:"additionalProperties"`
}

func main() {
	outDir := flag.String("o", "", "path to the output directory")
	langArg := flag.String("l", "", "language to generate")
	flag.Parse()

	if langArg == nil {
		panic("language is required")
	}

	lang := *langArg
	template := lang

	// We are only building json schemas for latest template versions
	if _, ok := templates.GetMinimumTargetVersion()[lang]; ok {
		template += "v2"
	}

	configFields, err := templates.GetLanguageConfigFields(types.NewTargetFromTemplate(template), true)
	if err != nil {
		panic(err)
	}

	// Initialize JSON schema structure
	jsonSchema := JSONSchema{
		Schema:               "https://json-schema.org/draft/2020-12/schema",
		Title:                lang + " Configuration Schema",
		Type:                 "object",
		Description:          fmt.Sprintf("Schema for configuration specific to a %s SDK", lang),
		Properties:           make(map[string]map[string]any),
		Required:             []string{},
		AdditionalProperties: true,
	}
	for _, field := range configFields {
		property := map[string]any{}

		hasClearlyDiscernibleType := true
		typeIsBoolean := false

		// Determine the property type from DefaultValue if it's discernible
		if field.DefaultValue != nil {
			switch (*field.DefaultValue).(type) {
			case string:
				property["type"] = "string"
			case bool:
				property["type"] = "boolean"
				typeIsBoolean = true
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
				property["type"] = "number"
			default:
				hasClearlyDiscernibleType = false
			}
		}

		// Use ValidationRegex if available, boolean regex is irrelevant
		if field.ValidationRegex != nil && !typeIsBoolean {
			property["pattern"] = *field.ValidationRegex
			hasClearlyDiscernibleType = true
		}

		// Only add the specific json schema right now if we have a discernible primitive type or a validation regex
		if hasClearlyDiscernibleType {
			if field.Required {
				jsonSchema.Required = append(jsonSchema.Required, field.Name)
			}
			if field.Description != nil {
				property["description"] = *field.Description
			}
			jsonSchema.Properties[field.Name] = property
		}
	}

	schemaJSON, err := json.MarshalIndent(jsonSchema, "", "  ")
	if err != nil {
		panic("Error marshalling JSON schema" + err.Error())
	}

	if outDir != nil {
		if _, err := os.Stat(*outDir); os.IsNotExist(err) {
			if err := os.MkdirAll(*outDir, os.ModePerm); err != nil {
				panic("Failed to create output directory`: " + err.Error())
			}
		}

		filePath := fmt.Sprintf("%s/%s.schema.json", *outDir, lang)
		file, err := os.Create(filePath)
		if err != nil {
			panic("Failed to create file: " + err.Error())
		}
		defer file.Close()

		if _, err := file.Write(schemaJSON); err != nil {
			panic("Failed to write to file: " + err.Error())
		}
	} else {
		fmt.Println(string(schemaJSON))
	}
}
