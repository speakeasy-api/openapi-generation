// Copyright 2022 Dave Shanley / Quobix
// SPDX-License-Identifier: MIT
package report

import (
	"bytes"
	_ "embed"
	"fmt"
	"html"
	"html/template"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/speakeasy-api/openapi/validation"
	"gopkg.in/yaml.v3"
)

//go:embed templates/report-template.gohtml
var reportTemplate string

//go:embed templates/header.gohtml
var header string

//go:embed templates/footer.gohtml
var footer string

//go:embed ui/build/static/js/vacuumReport.js
var bundledJS string

//go:embed ui/build/static/js/hydrate.js
var hydrateJS string

//go:embed ui/src/css/report.css
var reportCSS string

type HTMLReport interface {
	GenerateReport(testMode bool, version string) []byte
}

// MaxViolations the maximum number of violations the report will render per broken rule.
// TODO: make this configurable
const MaxViolations = 100

// RuleCategory represents a category of validation rules
type RuleCategory struct {
	ID          string
	Name        string
	Description string
}

// ReportStatistics contains validation statistics
type ReportStatistics struct {
	TotalErrors   int
	TotalWarnings int
	TotalHints    int
}

// ValidationReportData contains all data needed for report generation using the new linter
type ValidationReportData struct {
	Results      []*validation.Error      // Validation errors from the linter
	Categories   []*RuleCategory          // Rule categories
	Statistics   *ReportStatistics        // Statistics summary
	SpecBytes    []byte                   // Original spec bytes for source display
	Generated    time.Time                // When the report was generated
	RuleMetadata map[string]*RuleMetadata // Rule ID to metadata mapping
}

// RuleMetadata contains metadata about a rule
type RuleMetadata struct {
	Summary     string
	Description string
	HowToFix    string
}

// ReportData is the internal struct used by the HTML template
type ReportData struct {
	BundledJS      template.JS              `json:"bundledJS"`
	HydrateJS      template.JS              `json:"hydrateJS"`
	ShoelaceJS     string                   `json:"shoelaceJS"`
	ReportCSS      template.CSS             `json:"reportCSS"`
	Statistics     *ReportStatistics        `json:"reportStatistics"`
	TestMode       bool                     `json:"test"`
	RuleCategories []*RuleCategory          `json:"ruleCategories"`
	RuleResults    []*validation.Error      `json:"ruleResults"`
	MaxViolations  int                      `json:"maxViolations"`
	Generated      time.Time                `json:"generated"`
	Version        string                   `json:"version"`
	SpecString     []string                 `json:"-"`
	RuleMetadata   map[string]*RuleMetadata `json:"-"` // Map of rule ID to metadata
}

func NewHTMLReport(data *ValidationReportData) HTMLReport {
	return &htmlReport{data: data}
}

type htmlReport struct {
	data *ValidationReportData
}

func (h htmlReport) GenerateReport(test bool, version string) []byte {
	templateFuncs := template.FuncMap{
		"sortResults": func(results []*validation.Error) []*validation.Error {
			sort.Slice(results, func(i, j int) bool {
				if results[i].Node.Line < results[j].Node.Line {
					return true
				}
				if results[i].Node.Line > results[j].Node.Line {
					return false
				}
				// Compare by error message
				iMsg := results[i].UnderlyingError.Error()
				jMsg := results[j].UnderlyingError.Error()
				if iMsg != jMsg {
					return iMsg < jMsg
				}
				// Compare by rule ID
				if results[i].Rule != results[j].Rule {
					return results[i].Rule < results[j].Rule
				}
				return false
			})
			return results
		},
		"timeGenerated": func(t time.Time) string {
			return t.Format("02 Jan 2006 15:04:05 MST")
		},
		"extractResultsForCategory": func(cat string, allResults []*validation.Error) []*validation.Error {
			// Filter results by category based on rule ID prefix
			var categoryResults []*validation.Error

			if cat == "all" {
				categoryResults = allResults
			} else {
				// Filter by category prefix
				for _, result := range allResults {
					if strings.HasPrefix(result.Rule, cat+"-") {
						categoryResults = append(categoryResults, result)
					}
				}
			}

			// Sort by severity
			slices.SortFunc(categoryResults, func(a, b *validation.Error) int {
				return compareSeverityEnum(a.Severity, b.Severity)
			})

			return categoryResults
		},
		"limitResults": func(results []*validation.Error, maxResults int) []*validation.Error {
			if len(results) > maxResults {
				return results[:maxResults]
			}
			return results
		},
		"hasMore": func(total, maxResults int) bool {
			return total > maxResults
		},
		"moreCount": func(total, maxResults int) int {
			if total > maxResults {
				return total - maxResults
			}
			return 0
		},
		"groupByRule": func(results []*validation.Error) map[string][]*validation.Error {
			// Group results by rule ID
			grouped := make(map[string][]*validation.Error)
			for _, result := range results {
				grouped[result.Rule] = append(grouped[result.Rule], result)
			}
			return grouped
		},
		"getRuleKeys": func(grouped map[string][]*validation.Error) []string {
			// Get sorted list of rule IDs
			keys := make([]string, 0, len(grouped))
			for k := range grouped {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			return keys
		},
		"ruleSeverityIcon": func(sev validation.Severity) string {
			switch sev {
			case validation.SeverityError:
				return "❌"
			case validation.SeverityWarning:
				return "⚠️"
			case validation.SeverityHint:
				return "💠"
			}
			return ""
		},
		"renderSource": func(r *validation.Error, specData []string) template.HTML {
			builder := new(strings.Builder)
			reportLine := r.Node.Line
			startLine := reportLine - 5
			endLine := reportLine + 5

			if startLine < 0 {
				startLine = 0
			}

			if endLine > len(specData) {
				endLine = len(specData)
			}

			builder.WriteString("<pre tabindex=\"0\" class=\"chroma\"><code>")

			for index, line := range specData[startLine:endLine] {
				lineClass := "line"
				lineNumber := startLine + index + 1 // +1 for 0-based index

				if lineNumber == reportLine {
					lineClass = "line hl"
				}

				builder.WriteString("<span class=\"" + lineClass + "\">")
				builder.WriteString("<span class=\"ln\">" + strconv.Itoa(lineNumber) + "</span>")
				builder.WriteString("<span class=\"cl\">" + html.EscapeString(strings.TrimRight(line, "\n")) + "</span>")
				builder.WriteString("</span>\n")
			}

			builder.WriteString("</code></pre>")

			return template.HTML(builder.String())
		},
		"getNodeLine": func(node *yaml.Node) int {
			if node == nil {
				return -1
			}

			return node.Line
		},
		"getNodeColumn": func(node *yaml.Node) int {
			if node == nil {
				return -1
			}

			return node.Column
		},
		"formatErrorMessage": func(err error) string {
			if err == nil {
				return ""
			}
			return err.Error()
		},
	}
	tmpl := template.New("header")
	tmpl.Funcs(templateFuncs)
	t, err := tmpl.Parse(header)
	if err != nil {
		return []byte(fmt.Sprintf("failed to render: %v", err))
	}
	_, err = t.New("footer").Parse(footer)
	if err != nil {
		return []byte(fmt.Sprintf("failed to render: %v", err))
	}
	_, err = t.New("report").Parse(reportTemplate)
	if err != nil {
		return nil
	}

	var byteBuf bytes.Buffer

	// Build categories that have results
	allCategories := buildCategories()
	var catsFiltered []*RuleCategory

	// Count results per category
	categoryCounts := make(map[string]int)
	for _, result := range h.data.Results {
		// Extract category from rule ID (e.g., "validation-foo" -> "validation")
		parts := strings.SplitN(result.Rule, "-", 2)
		if len(parts) > 0 {
			categoryCounts[parts[0]]++
		}
	}

	// Add "all" category if we have any results
	if len(h.data.Results) > 0 {
		catsFiltered = append(catsFiltered, &RuleCategory{
			ID:          "all",
			Name:        "All",
			Description: "All validation results",
		})
	}

	// Add categories that have results
	for _, cat := range allCategories {
		if categoryCounts[cat.ID] > 0 {
			catsFiltered = append(catsFiltered, cat)
		}
	}

	var specStringData []string
	if len(h.data.SpecBytes) > 0 {
		specStringData = strings.Split(string(h.data.SpecBytes), "\n")
	}

	reportData := &ReportData{
		BundledJS:      template.JS(bundledJS),
		HydrateJS:      template.JS(hydrateJS),
		ReportCSS:      template.CSS(reportCSS),
		Statistics:     h.data.Statistics,
		RuleCategories: catsFiltered,
		TestMode:       test,
		RuleResults:    h.data.Results,
		MaxViolations:  MaxViolations,
		SpecString:     specStringData,
		Version:        version,
		Generated:      h.data.Generated,
		RuleMetadata:   h.data.RuleMetadata,
	}
	err = t.ExecuteTemplate(&byteBuf, "report", reportData)
	if err != nil {
		return []byte(fmt.Sprintf("failed to render: %v", err.Error()))
	}

	return byteBuf.Bytes()
}

func compareSeverityEnum(a, b validation.Severity) int {
	// Define priority: Error (1) > Warning (2) > Hint (3)
	severityPriority := map[validation.Severity]int{
		validation.SeverityError:   1,
		validation.SeverityWarning: 2,
		validation.SeverityHint:    3,
	}

	aPriority := severityPriority[a]
	bPriority := severityPriority[b]

	return aPriority - bPriority
}

// buildCategories returns the defined rule categories
func buildCategories() []*RuleCategory {
	return []*RuleCategory{
		{
			ID:          "validation",
			Name:        "Validation",
			Description: "OpenAPI schema validation rules",
		},
		{
			ID:          "generator",
			Name:        "Generator",
			Description: "SDK generator validation rules",
		},
		{
			ID:          "semantic",
			Name:        "Semantic",
			Description: "Semantic validation rules",
		},
		{
			ID:          "style",
			Name:        "Style",
			Description: "OpenAPI style guide rules",
		},
		{
			ID:          "collision",
			Name:        "Collision",
			Description: "Name collision detection rules",
		},
		{
			ID:          "completeness",
			Name:        "Completeness",
			Description: "API completeness checks",
		},
		{
			ID:          "deprecation",
			Name:        "Deprecation",
			Description: "Deprecation-related rules",
		},
		{
			ID:          "duplicate",
			Name:        "Duplicate",
			Description: "Duplicate detection rules",
		},
		{
			ID:          "missing",
			Name:        "Missing",
			Description: "Missing required elements",
		},
		{
			ID:          "path",
			Name:        "Path",
			Description: "Path parameter validation",
		},
		{
			ID:          "pagination",
			Name:        "Pagination",
			Description: "Pagination configuration rules",
		},
		{
			ID:          "retry",
			Name:        "Retry",
			Description: "Retry configuration rules",
		},
		{
			ID:          "content",
			Name:        "Content Type",
			Description: "Content type validation",
		},
		{
			ID:          "extension",
			Name:        "Extension",
			Description: "OpenAPI extension validation",
		},
		{
			ID:          "enum",
			Name:        "Enum",
			Description: "Enum validation rules",
		},
		{
			ID:          "security",
			Name:        "Security",
			Description: "Security scheme validation",
		},
		{
			ID:          "server",
			Name:        "Server",
			Description: "Server configuration validation",
		},
		{
			ID:          "type",
			Name:        "Type",
			Description: "Type validation rules",
		},
		{
			ID:          "request",
			Name:        "Request",
			Description: "Request validation rules",
		},
		{
			ID:          "response",
			Name:        "Response",
			Description: "Response validation rules",
		},
		{
			ID:          "parameter",
			Name:        "Parameter",
			Description: "Parameter validation rules",
		},
		{
			ID:          "schema",
			Name:        "Schema",
			Description: "Schema validation rules",
		},
	}
}
