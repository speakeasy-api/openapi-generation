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

// prettierPrintWidth mirrors the default Prettier line width used by the docs
// site, so regenerating the page does not fight the site's formatter.
const prettierPrintWidth = 80

const (
	cellImplemented          = "✅"
	cellPartiallyImplemented = "⚠️"
	cellNotImplemented       = "⛔"
	cellIgnored              = "➖"
)

var (
	parensRegex     = regexp.MustCompile(`\s*\([^)]*\)`)
	jsIdentifier    = regexp.MustCompile(`^[A-Za-z_$][A-Za-z0-9_$]*$`)
	jsStringEscaper = strings.NewReplacer(`\`, `\\`, `"`, `\"`)
	defaultOutput   = "maturity.mdx"
	defaultCSV      = "reports/implemented-features.csv"

	// targetOrder is the row order of the feature support level table.
	targetOrder = []string{
		"typescript", "python", "go", "java", "csharp",
		"php", "ruby", "terraform", "postman", "mcp-typescript", "cli",
	}
	// matrixColumns is the column order of every per-category feature table.
	matrixColumns = []string{
		"typescript", "python", "go", "java", "csharp",
		"php", "ruby", "terraform", "mcp-typescript", "postman", "cli",
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
	maturityLevels = [][2]string{
		{"Alpha", "An early preview of upcoming features intended for gathering feedback. Alpha versions are less complete and likely to be unstable, with frequent updates and significant changes."},
		{"Beta", "A stable release version that includes many features of GA but is still ongoing development and customer feedback. Core interfaces are stable and ready for production use."},
		{"General availability (GA)", "A fully supported release that includes all functionalities altering the type interface from OpenAPI Specification keywords (for example, [oneOf](/docs/sdks/customize/data-model/oneof-schemas))."},
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

// property is a single key/value pair of a JavaScript object literal embedded
// in the generated MDX.
type property struct {
	key   string
	value string
}

// propOrder selects which of the <Table /> element's two attributes is written
// first, since the existing page does not use the same order everywhere.
type propOrder int

const (
	columnsFirst propOrder = iota
	dataFirst
)

func main() {
	outputPath := flag.String("out", defaultOutput, "Path to output MDX file")
	flag.Parse()

	csvPath := filepath.FromSlash(defaultCSV)
	outPath := filepath.FromSlash(*outputPath)

	writeFile(outPath, render(readCSV(csvPath)))
}

// render turns the implemented features report into the maturity page, with
// one table per feature category.
func render(records [][]string) string {
	headers, rows := records[0][1:], records[1:]

	// Column headings carry the template name, which for some targets is
	// suffixed with a generation version (for example "pythonv2").
	for i, h := range headers {
		headers[i] = strings.TrimSuffix(strings.TrimSpace(h), "v2")
	}

	grouped := initCategoryBuckets(categoryOrder)
	for _, row := range rows {
		feature := cleanFeatureName(row[0])

		category := categoryMap[feature]
		if category == "" {
			continue
		}

		cells := make([]property, 0, 1+len(matrixColumns))
		cells = append(cells, property{key: "feature", value: "`" + feature + "`"})
		for _, lang := range matrixColumns {
			cell := cellNotImplemented
			if idx := indexOf(headers, lang); idx >= 0 {
				cell = normalizeCell(row[idx+1])
			}
			cells = append(cells, property{key: lang, value: cell})
		}

		grouped[category] = append(grouped[category], cells)
	}

	return buildMarkdown(grouped)
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
	case ":white_check_mark:":
		return cellImplemented
	case ":warning:":
		return cellPartiallyImplemented
	case ":heavy_minus_sign:":
		return cellIgnored
	case ":no_entry:", "", "-", "—":
		return cellNotImplemented
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

func initCategoryBuckets(order []string) map[string][][]property {
	buckets := make(map[string][][]property, len(order))
	for _, cat := range order {
		buckets[cat] = [][]property{}
	}
	return buckets
}

func buildMarkdown(grouped map[string][][]property) string {
	var b strings.Builder

	b.WriteString(`---
title: "SDK Feature Matrix by Category"
description: "Explore Speakeasy's support for generated targets."
sidebar_position: 4
sidebar_label: Language Maturity
---

import { Table } from "@/mdx/components";

## Maturity levels

Maturity levels indicate the extent of development on a generation target.

`)

	maturityRows := make([][]property, 0, len(maturityLevels))
	for _, level := range maturityLevels {
		maturityRows = append(maturityRows, []property{
			{key: "level", value: level[0]},
			{key: "description", value: level[1]},
		})
	}
	writeTable(&b, []property{
		{key: "level", value: "Maturity level"},
		{key: "description", value: "Description"},
	}, maturityRows, dataFirst)

	b.WriteString(`
## Feature support levels

Feature support levels indicate the extent of additional functionalities provided.

`)

	targetRows := make([][]property, 0, len(targetOrder))
	for _, target := range targetOrder {
		name := languageDisplay[target]
		if methodologyPath, ok := types.MethodologyPaths[target]; ok {
			name = fmt.Sprintf("[%s](%s)", name, methodologyPath)
		}

		targetRows = append(targetRows, []property{
			{key: "target", value: name},
			{key: "maturity", value: string(types.TargetMaturity[target])},
			{key: "support", value: string(types.SupportLevels[target])},
		})
	}
	writeTable(&b, []property{
		{key: "target", value: "Target"},
		{key: "maturity", value: "Maturity level"},
		{key: "support", value: "Feature support level"},
	}, targetRows, columnsFirst)

	b.WriteString(`
## Deprecated generation targets

- TypeScript Beta (v1)
- Java Beta (v1)

## SDK Feature Matrix by Category

This document outlines the OpenAPI and SDK features supported by Speakeasy. Features are grouped by category to help quickly locate what's available per SDK.

**Legend**:

- ` + cellImplemented + ` Implemented
- ` + cellPartiallyImplemented + ` Partially Implemented (missing Readme sections or tests)
- ` + cellNotImplemented + ` Not Implemented
- ` + cellIgnored + ` Ignored

_Note: This is not a complete list. Some SDK features are language-specific or not yet documented here._
`)

	columns := make([]property, 0, 1+len(matrixColumns))
	columns = append(columns, property{key: "feature", value: "Feature"})
	for _, lang := range matrixColumns {
		columns = append(columns, property{key: lang, value: languageDisplay[lang]})
	}

	for _, category := range categoryOrder {
		rows := grouped[category]
		if len(rows) == 0 {
			continue
		}

		fmt.Fprintf(&b, "\n## %s\n\n", category)
		writeTable(&b, columns, rows, columnsFirst)
	}

	return b.String()
}

// writeTable renders a <Table /> element in the shape Prettier would produce
// for it, so the docs site's formatter leaves the generated page alone.
func writeTable(b *strings.Builder, columns []property, rows [][]property, order propOrder) {
	writeColumns := func() {
		b.WriteString("  columns={[\n")
		for _, column := range columns {
			fmt.Fprintf(b, "    { key: %s, header: %s },\n", jsString(column.key), jsString(column.value))
		}
		b.WriteString("  ]}\n")
	}

	writeData := func() {
		b.WriteString("  data={[\n")
		for _, row := range rows {
			b.WriteString("    {\n")
			for _, prop := range row {
				writeProperty(b, "      ", prop)
			}
			b.WriteString("    },\n")
		}
		b.WriteString("  ]}\n")
	}

	b.WriteString("<Table\n")
	if order == dataFirst {
		writeData()
		writeColumns()
	} else {
		writeColumns()
		writeData()
	}
	b.WriteString("/>\n")
}

// writeProperty breaks a property onto its own line when the single-line form
// would overflow Prettier's print width, matching how Prettier wraps it.
func writeProperty(b *strings.Builder, indent string, prop property) {
	key := jsKey(prop.key)
	value := jsString(prop.value)

	if len(indent)+len(key)+len(": ")+len(value)+len(",") > prettierPrintWidth {
		fmt.Fprintf(b, "%s%s:\n%s  %s,\n", indent, key, indent, value)
		return
	}

	fmt.Fprintf(b, "%s%s: %s,\n", indent, key, value)
}

// jsKey quotes an object key only when it is not a bare JavaScript identifier,
// which is what Prettier does.
func jsKey(key string) string {
	if jsIdentifier.MatchString(key) {
		return key
	}
	return jsString(key)
}

func jsString(value string) string {
	return `"` + jsStringEscaper.Replace(value) + `"`
}

func writeFile(path, content string) {
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		panic(fmt.Errorf("failed to write markdown: %w", err))
	}
}
