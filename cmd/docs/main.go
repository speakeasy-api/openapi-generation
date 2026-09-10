package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
)

var (
	parensRegex       = regexp.MustCompile(`\s*\([^)]*\)`)
	notSupportedEmoji = "⛔"
	defaultOutput     = "features-table.md"
	defaultCSV        = "reports/implemented-features.csv"

	languageOrder = []string{
		"typescript", "python", "go", "java", "csharp",
		"php", "ruby", "terraform", "postman", "mcp-typescript", "cli",
	}
	languageDisplay = map[string]string{
		"typescript":     "TypeScript",
		"python":         "Python",
		"go":             "Go",
		"java":           "Java",
		"csharp":         "C#",
		"php":            "PHP",
		"ruby":           "Ruby",
		"terraform":      "Terraform",
		"postman":        "Postman",
		"mcp-typescript": "MCP Typescript",
		"cli":            "CLI",
	}
	categoryOrder = []string{
		"Customization Basics", "Structure", "Data Model", "Customize Methods",
		"Responses & Error Handling", "Global Parameters", "Configure Servers",
		"Security & Authentication", "SDK Behavior", "Add Webhooks",
		"Add Custom Code", "Environment", "Documentation & Dev Experience",
	}
	categoryMap = map[string]string{
		// Customization Basics
		"core":      "Customization Basics",
		"callbacks": "Customization Basics",

		// Structure
		"nameOverrides":     "Structure",
		"multiLevelTagging": "Structure",
		"groups":            "Structure",
		"includes":          "Structure",
		"ignores":           "Structure",

		// Data Model
		"enums":                "Data Model",
		"nullables":            "Data Model",
		"bigint":               "Data Model",
		"decimal":              "Data Model",
		"openEnums":            "Data Model",
		"anyOf":                "Data Model",
		"oneOf":                "Data Model",
		"allOf":                "Data Model",
		"unions":               "Data Model",
		"sliceUnions":          "Data Model",
		"sets":                 "Data Model",
		"typeOverrides":        "Data Model",
		"constsAndDefaults":    "Data Model",
		"additionalProperties": "Data Model",
		"disallowCircularRefs": "Data Model",

		// Customize Methods
		"flattening":            "Customize Methods",
		"methodArguments":       "Customize Methods",
		"methodSecurity":        "Customize Methods",
		"methodServerURLs":      "Customize Methods",
		"tagBasedOrdering":      "Customize Methods",
		"pagination":            "Customize Methods",
		"retries":               "Customize Methods",
		"urlBasedPagination":    "Customize Methods",
		"defaultEnabledRetries": "Customize Methods",

		// Responses & Error Handling
		"inputOutputModels": "Responses & Error Handling",
		"jsonlResponses":    "Responses & Error Handling",
		"uploadStreams":     "Responses & Error Handling",
		"getRequestBodies":  "Responses & Error Handling",
		"responseFormat":    "Responses & Error Handling",
		"errors":            "Responses & Error Handling",
		"errorUnions":       "Responses & Error Handling",
		"enumUnions":        "Responses & Error Handling",

		// Global Parameters
		"deepObjectParams":       "Global Parameters",
		"acceptHeaders":          "Global Parameters",
		"allowReserved":          "Global Parameters",
		"additionalDependencies": "Global Parameters",

		// Configure Servers
		"globalServerURLs":         "Configure Servers",
		"serverIDs":                "Configure Servers",
		"globals":                  "Configure Servers",
		"serverEvents":             "Configure Servers",
		"serverEventsSentinels":    "Configure Servers",
		"globalSecurityFlattening": "Configure Servers",

		// Security & Authentication
		"globalSecurity":          "Security & Authentication",
		"globalSecurityCallbacks": "Security & Authentication",
		"oauth2ClientCredentials": "Security & Authentication",
		"oauth2Password":          "Security & Authentication",
		"customSecuritySchemes":   "Security & Authentication",
		"deprecations":            "Security & Authentication",

		// SDK Behavior
		"operationTimeout":    "SDK Behavior",
		"flatRequests":        "SDK Behavior",
		"stringNumberFormats": "SDK Behavior",
		"downloadStreams":     "SDK Behavior",

		// Add Webhooks
		"webhooks":        "Add Webhooks",
		"webhookHandlers": "Add Webhooks",

		// Add Custom Code
		"customCodeRegions": "Add Custom Code",
		"sdkHooks":          "Add Custom Code",
		"mockServer":        "Add Custom Code",
		"mockserver":        "Add Custom Code",

		// Environment
		"envVarGlobals":          "Environment",
		"envVarSecurityUsage":    "Environment",
		"configurableModuleName": "Environment",
		"hiddenGlobals":          "Environment",

		// Documentation & Dev Experience
		"docs":                    "Documentation & Dev Experience",
		"snippetGeneration":       "Documentation & Dev Experience",
		"readmeGeneration":        "Documentation & Dev Experience",
		"documentationGeneration": "Documentation & Dev Experience",
		"examples":                "Documentation & Dev Experience",
		"tests":                   "Documentation & Dev Experience",
	}
)

func main() {
	outputPath := flag.String("out", defaultOutput, "Path to output markdown file")
	flag.Parse()

	csvPath := filepath.FromSlash(defaultCSV)
	outPath := filepath.FromSlash(*outputPath)

	records := readCSV(csvPath)
	headersRaw, rows := records[0][1:], records[1:]

	// Clean up header names (trim spaces and strip any "v2" suffix)
	for i, h := range headersRaw {
		hdr := strings.TrimSpace(h)
		hdr = strings.TrimSuffix(hdr, "v2")
		headersRaw[i] = hdr
	}

	// Map from header index to column language
	headerMap := map[int]string{}
	for i, h := range headersRaw {
		headerMap[i] = h
	}

	grouped := initCategoryBuckets(categoryOrder)
	for _, row := range rows {
		feature := cleanFeatureName(row[0])
		if feature == "reactQueryHooks" {
			continue
		}
		if category := categoryMap[feature]; category != "" {
			// Reorder and normalize row to match languageOrder
			cells := make([]string, len(languageOrder))
			for i, lang := range languageOrder {
				idx := indexOf(headersRaw, lang)
				if idx >= 0 {
					cells[i] = normalizeCell(row[idx+1])
				} else {
					cells[i] = notSupportedEmoji
				}
			}
			grouped[category] = append(grouped[category], append([]string{"`" + feature + "`"}, cells...))
		}
	}

	markdown := buildMarkdown(languageOrder, grouped)
	writeFile(outPath, markdown)
}

func indexOf(slice []string, target string) int {
	for i, s := range slice {
		if s == target {
			return i
		}
	}
	return -1
}

func normalizeCell(cell string) string {
	switch strings.TrimSpace(cell) {
	case "", "-", "—":
		return notSupportedEmoji
	case ":white_check_mark:":
		return "✅"
	case ":warning:":
		return "⚠️"
	case ":no_entry:":
		return "⛔"
	case ":heavy_minus_sign:":
		return "⛔"
	default:
		return cell
	}
}

func readCSV(path string) [][]string {
	file, err := os.Open(path)
	if err != nil {
		panic(fmt.Errorf("failed to open CSV: %w", err))
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		panic(fmt.Errorf("failed to read CSV: %w", err))
	}
	if len(records) < 2 {
		panic("CSV must have at least one header and one data row")
	}
	return records
}

func cleanFeatureName(raw string) string {
	return strings.TrimSpace(parensRegex.ReplaceAllString(raw, ""))
}

func initCategoryBuckets(order []string) map[string][][]string {
	buckets := make(map[string][][]string, len(order))
	for _, cat := range order {
		buckets[cat] = [][]string{}
	}
	return buckets
}

func buildMarkdown(langKeys []string, grouped map[string][][]string) string {
	var b strings.Builder

	b.WriteString(`---
description: "Explore Speakeasy's support for generated targets."
sidebar_position: 4
slug: /code-generation-concepts/
sidebar_label: Language Maturity
---

import { Table } from "@/mdx/components";


## Maturity levels

<Table
  data={[
    { level: "Alpha", description: "An early preview of upcoming features intended for gathering feedback. Alpha versions are less complete and likely to be unstable, with frequent updates and significant changes." },
    { level: "Beta", description: "A pre-release version that includes many features of GA but is still subject to significant modifications based on user feedback. The interface is considered stable enough for testing." },
    { level: "General availability (GA)", description: "A fully supported release that includes all functionalities altering the type interface from OpenAPI Specification keywords (for example, [oneOf](/docs/customize-sdks/oneof))." }
  ]}
  columns={[
    { key: "level", header: "Maturity level" },
    { key: "description", header: "Description" }
  ]}
/>

## Feature support levels

Feature support levels indicate the extent of additional functionalities provided.`)

	b.WriteString("\n")
	b.WriteString("<Table\n columns={[{\"key\": \"target\", header: \"Target\"}, {key: \"maturity\", header: \"Maturity level\"}, {key: \"support\", header: \"Feature support level\"}]}\n\n data={[")
	for _, target := range languageOrder {
		maturity := types.TargetMaturity[target]
		methodologyPath, methodologyOk := types.MethodologyPaths[target]
		supportLevel := types.SupportLevels[target]
		prettyTarget := languageDisplay[target]
		b.WriteString("    { ")
		if methodologyOk {
			fmt.Fprintf(&b, "target: \"[%s](%s)\", ", prettyTarget, methodologyPath)
		} else {
			fmt.Fprintf(&b, "target: \"%s\", ", prettyTarget)
		}
		fmt.Fprintf(&b, "maturity: \"%s\", ", maturity)
		fmt.Fprintf(&b, "support: \"%s\"", supportLevel)
		b.WriteString("    },\n")
	}
	b.WriteString("  ]}\n/>\n\n")

	b.WriteString(`## Deprecated generation targets

* TypeScript Beta (v1)
* Java Beta (v1)

# SDK Feature Matrix by Category

This document outlines the OpenAPI and SDK features supported by Speakeasy. Features are grouped by category to help you quickly locate what's available per SDK.

**Legend**: ✅ Implemented, ⚠️ Partially Implemented (missing Readme sections or tests), ⛔ Not Implemented, ➖ Ignored**

_Note: This is not a complete list. Some SDK features are language-specific or not yet documented here._

`)
	for _, category := range categoryOrder {
		rows := grouped[category]
		if len(rows) == 0 {
			continue
		}

		fmt.Fprintf(&b, "## %s\n\n", category)
		b.WriteString("<Table\n  columns={[{ key: \"feature\", header: \"Feature\" },\n")
		for _, lang := range langKeys {
			fmt.Fprintf(&b, "    { key: \"%s\", header: \"%s\" },\n", lang, languageDisplay[lang])
		}
		b.WriteString("  ]}\n  data={[\n")

		for _, row := range rows {
			b.WriteString("    { ")
			fmt.Fprintf(&b, "feature: \"%s\"", row[0])
			for i, val := range row[1:] {
				fmt.Fprintf(&b, ", \"%s\": \"%s\"", langKeys[i], val)
			}
			b.WriteString(" },\n")
		}
		b.WriteString("  ]}\n/>\n\n")
	}
	return b.String()
}

func writeFile(path, content string) {
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		panic(fmt.Errorf("failed to write markdown: %w", err))
	}
}
