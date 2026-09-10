package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/templates"
)

func main() {
	format := flag.String("format", "json", "Output format: json, summary, markdown")
	target := flag.String("target", "", "Specific target (empty = all targets)")
	scan := flag.Bool("scan", false, "Scan for vulnerabilities via OSV and optionally npm audit")
	osvSeverity := flag.String("osv-severity", "low", "Minimum severity for OSV findings (low|moderate|high|critical)")
	npmAudit := flag.String("npm-audit", "moderate", "Minimum severity for npm audit findings (low|moderate|high|critical), or \"skip\" to disable the npm audit scan")
	flag.Parse()

	if *scan {
		osvThreshold, ok := severityThreshold(*osvSeverity)
		if !ok {
			fmt.Fprintf(os.Stderr, "Error: invalid -osv-severity value %q (expected one of: low, moderate, high, critical)\n", *osvSeverity)
			os.Exit(1)
		}

		// Resolve -npm-audit to a threshold. 0 means skip the npm audit scan.
		npmThreshold := 0
		if npmAuditEnabled(*npmAudit) {
			r, ok := severityThreshold(*npmAudit)
			if !ok {
				fmt.Fprintf(os.Stderr, "Error: invalid -npm-audit value %q (expected one of: low, moderate, high, critical, or \"skip\" to disable)\n", *npmAudit)
				os.Exit(1)
			}
			npmThreshold = r
		}
		if err := runSecurityScan(*format, osvThreshold, npmThreshold); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if err := listDependencies(*format, *target); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func listDependencies(format, target string) error {
	var allDeps map[string][]templates.TemplateDependency
	var err error

	if target != "" {
		t := types.NewTargetFromTemplate(target)
		deps, err := templates.GetTemplateDependencies(t)
		if err != nil {
			return fmt.Errorf("failed to get dependencies for target %s: %w", target, err)
		}
		allDeps = map[string][]templates.TemplateDependency{target: deps}
	} else {
		allDeps, err = templates.GetAllTemplateDependencies()
		if err != nil {
			return fmt.Errorf("failed to get all dependencies: %w", err)
		}
	}

	switch format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(allDeps)

	case "summary":
		for targetName, deps := range allDeps {
			fmt.Printf("\n=== %s ===\n", targetName)
			for _, dep := range deps {
				fmt.Printf("  %s@%s (%s)\n", dep.Name, dep.Version, dep.Category)
			}
		}
		return nil

	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

// OSV API types
type OSVBatchQuery struct {
	Queries []OSVQuery `json:"queries"`
}

type OSVQuery struct {
	Package OSVPackage `json:"package"`
	Version string     `json:"version,omitempty"`
}

type OSVPackage struct {
	Name      string `json:"name"`
	Ecosystem string `json:"ecosystem"`
}

type OSVBatchResponse struct {
	Results []OSVResult `json:"results"`
}

type OSVResult struct {
	Vulns []OSVVuln `json:"vulns,omitempty"`
}

type OSVVuln struct {
	ID               string         `json:"id"`
	Summary          string         `json:"summary"`
	Details          string         `json:"details"`
	Severity         []OSVSeverity  `json:"severity,omitempty"`
	Modified         string         `json:"modified"`
	DatabaseSpecific map[string]any `json:"database_specific,omitempty"`
	References       []OSVReference `json:"references,omitempty"`
}

type OSVSeverity struct {
	Type  string `json:"type"`
	Score string `json:"score"`
}

type OSVReference struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

// Ignore file types
type IgnoreFile struct {
	Ignore []IgnoreEntry `json:"ignore"`
}

type IgnoreEntry struct {
	ID      string `json:"id"`
	Package string `json:"package,omitempty"`
	Reason  string `json:"reason"`
	Expires string `json:"expires,omitempty"`
}

// Finding represents a vulnerability finding
type Finding struct {
	ID               string   `json:"id"`
	Severity         string   `json:"severity"`
	Description      string   `json:"description"`
	URL              string   `json:"url"`
	Package          string   `json:"package"`
	Ecosystem        string   `json:"ecosystem"`
	Version          string   `json:"version"`
	Templates        []string `json:"templates"`
	Source           string   `json:"source,omitempty"` // "osv" or "npm-audit"
	Range            string   `json:"range,omitempty"`
	InstalledVersion string   `json:"installedVersion,omitempty"`
	FixAvailable     string   `json:"fixAvailable,omitempty"`
	Paths            []string `json:"paths,omitempty"`
}

func runSecurityScan(format string, osvThreshold, npmAuditThreshold int) error {
	slackWebhook := os.Getenv("SECURITY_NOTIFICATIONS_SLACK_WEBHOOK")

	// Load ignore file if present
	ignoreList := loadIgnoreFile(".security-ignore.json")

	// Get all dependencies grouped by template
	allDeps, err := templates.GetAllTemplateDependencies()
	if err != nil {
		return fmt.Errorf("failed to get dependencies: %w", err)
	}

	// Build unique packages to scan (dedupe by ecosystem+name+version)
	type pkgKey struct {
		ecosystem string
		name      string
		version   string
	}
	uniquePkgs := make(map[pkgKey][]string) // maps to template names
	var queries []OSVQuery

	for templateName, deps := range allDeps {
		for _, dep := range deps {
			minVersion := extractMinVersion(dep.Version)
			key := pkgKey{ecosystem: dep.Ecosystem, name: dep.Name, version: minVersion}

			if _, exists := uniquePkgs[key]; !exists {
				query := OSVQuery{
					Package: OSVPackage{
						Name:      dep.Name,
						Ecosystem: dep.Ecosystem,
					},
				}
				if minVersion != "" {
					query.Version = minVersion
				}
				queries = append(queries, query)
			}
			uniquePkgs[key] = append(uniquePkgs[key], templateName)
		}
	}

	fmt.Printf("Scanning %d unique package versions via OSV...\n", len(queries))

	// Query OSV batch API
	results, err := queryOSVBatch(queries)
	if err != nil {
		return fmt.Errorf("failed to query OSV: %w", err)
	}

	// Collect unique vulnerability IDs and their affected packages
	type vulnPackage struct {
		vulnID    string
		query     OSVQuery
		templates []string
	}
	var vulnPackages []vulnPackage
	var ignoredCount int
	seenVulns := make(map[string]bool)

	for i, result := range results {
		if i >= len(queries) {
			break
		}
		query := queries[i]

		for _, vuln := range result.Vulns {
			// Check if ignored
			if isIgnored(vuln.ID, query.Package.Name, ignoreList) {
				ignoredCount++
				continue
			}

			// Find templates affected
			key := pkgKey{
				ecosystem: query.Package.Ecosystem,
				name:      query.Package.Name,
				version:   query.Version,
			}

			vulnPackages = append(vulnPackages, vulnPackage{
				vulnID:    vuln.ID,
				query:     query,
				templates: uniquePkgs[key],
			})

			if !seenVulns[vuln.ID] {
				seenVulns[vuln.ID] = true
			}
		}
	}

	// Fetch full details for each unique vulnerability
	vulnDetails := make(map[string]*OSVVuln)
	for vulnID := range seenVulns {
		details, err := fetchVulnDetails(vulnID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to fetch details for %s: %v\n", vulnID, err)
			continue
		}
		vulnDetails[vulnID] = details
	}

	// Process results into findings
	findings := make([]Finding, 0, len(vulnPackages))

	for _, vp := range vulnPackages {
		vuln := vulnDetails[vp.vulnID]

		finding := Finding{
			ID:        vp.vulnID,
			Source:    "osv",
			Package:   vp.query.Package.Name,
			Ecosystem: vp.query.Package.Ecosystem,
			Version:   vp.query.Version,
			Templates: vp.templates,
		}

		// Get details from fetched vulnerability info
		if vuln != nil {
			if vuln.Summary != "" {
				finding.Description = vuln.Summary
			} else if vuln.Details != "" {
				finding.Description = vuln.Details
			}

			// Get severity from database_specific
			if vuln.DatabaseSpecific != nil {
				if sev, ok := vuln.DatabaseSpecific["severity"].(string); ok {
					finding.Severity = strings.ToUpper(sev)
				}
			}
		}

		if finding.Severity == "" {
			finding.Severity = "UNKNOWN"
		}

		// Filter OSV findings by the -osv-severity threshold.
		// UNKNOWN severities are never silently dropped.
		if !meetsSeverityThreshold(finding.Severity, osvThreshold) {
			continue
		}

		// Get URL
		switch {
		case strings.HasPrefix(vp.vulnID, "CVE-"):
			finding.URL = "https://nvd.nist.gov/vuln/detail/" + vp.vulnID
		case strings.HasPrefix(vp.vulnID, "GHSA-"):
			finding.URL = "https://github.com/advisories/" + vp.vulnID
		case vuln != nil && len(vuln.References) > 0:
			finding.URL = vuln.References[0].URL
		}

		findings = append(findings, finding)
	}

	// Run npm audit scan over generated TS SDKs and merge findings (unless disabled)
	if npmAuditThreshold > 0 {
		npmFindings, npmIgnored, err := runNPMAuditScan(npmAuditPaths, npmAuditThreshold, ignoreList)
		if err != nil {
			// Fail loudly: when npm audit is explicitly enabled, a failure must not
			// silently degrade the scan to OSV-only.
			return fmt.Errorf("npm audit scan failed: %w", err)
		}
		findings = append(findings, npmFindings...)
		ignoredCount += npmIgnored
	} else {
		fmt.Println("Skipping npm audit scan (disabled via -npm-audit flag)")
	}

	// Sort findings by severity
	severityOrder := map[string]int{"CRITICAL": 0, "HIGH": 1, "MODERATE": 2, "MEDIUM": 3, "LOW": 4, "UNKNOWN": 5}
	sort.Slice(findings, func(i, j int) bool {
		return severityOrder[findings[i].Severity] < severityOrder[findings[j].Severity]
	})

	// Send Slack notification if webhook is configured
	if slackWebhook != "" && len(findings) > 0 {
		if err := sendSlackNotification(findings, slackWebhook); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to send Slack notification: %v\n", err)
		}
	}

	// Output results based on format
	switch format {
	case "json":
		return outputJSON(findings, ignoredCount)
	case "markdown":
		outputMarkdown(findings, ignoredCount)
	default:
		outputText(findings, ignoredCount)
	}

	return nil
}

// ScanResult is the JSON output structure for security scan results
type ScanResult struct {
	Findings     []Finding `json:"findings"`
	IgnoredCount int       `json:"ignoredCount"`
	TotalCount   int       `json:"totalCount"`
}

func outputJSON(findings []Finding, ignoredCount int) error {
	result := ScanResult{
		Findings:     findings,
		IgnoredCount: ignoredCount,
		TotalCount:   len(findings),
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func outputMarkdown(findings []Finding, ignoredCount int) {
	if len(findings) == 0 {
		fmt.Println("## ✅ Security Scan Results")
		fmt.Println()
		fmt.Println("No vulnerabilities found affecting current dependency versions.")
		if ignoredCount > 0 {
			fmt.Printf("\n*(%d finding(s) ignored via `.security-ignore.json`)*\n", ignoredCount)
		}
		return
	}

	fmt.Println("## ⚠️ Security Scan Results")
	fmt.Println()
	fmt.Printf("Found **%d vulnerability(s)** affecting current versions", len(findings))
	if ignoredCount > 0 {
		fmt.Printf(" (%d ignored)", ignoredCount)
	}
	fmt.Println()
	fmt.Println()

	// Summary by severity
	bySeverity := make(map[string]int)
	for _, f := range findings {
		bySeverity[f.Severity]++
	}
	fmt.Print("**Summary:** ")
	for _, sev := range []string{"CRITICAL", "HIGH", "MODERATE", "MEDIUM", "LOW", "UNKNOWN"} {
		if count, ok := bySeverity[sev]; ok {
			fmt.Printf("%s: %d  ", sev, count)
		}
	}
	fmt.Println()
	fmt.Println()

	// Table header
	fmt.Println("| Severity | Advisory | Package | Version | Templates | Description |")
	fmt.Println("|----------|----------|---------|---------|-----------|-------------|")

	for _, f := range findings {
		desc := f.Description
		if len(desc) > 60 {
			desc = desc[:57] + "..."
		}
		// Escape pipe characters in description
		desc = strings.ReplaceAll(desc, "|", "\\|")

		fmt.Printf("| %s | [%s](%s) | `%s` | %s | %s | %s |\n",
			f.Severity,
			f.ID,
			f.URL,
			f.Package,
			f.Version,
			strings.Join(f.Templates, ", "),
			desc,
		)
	}

	fmt.Println()
	fmt.Println("<details>")
	fmt.Println("<summary>How to ignore findings</summary>")
	fmt.Println()
	fmt.Println("Add entries to `.security-ignore.json`:")
	fmt.Println("```json")
	fmt.Println(`{"ignore": [{"id": "GHSA-xxxx", "reason": "Not applicable because...", "expires": "2025-12-31"}]}`)
	fmt.Println("```")
	fmt.Println("</details>")
}

func outputText(findings []Finding, ignoredCount int) {
	if len(findings) == 0 {
		fmt.Println("No vulnerabilities found affecting current dependency versions.")
		if ignoredCount > 0 {
			fmt.Printf("(%d ignored via .security-ignore.json)\n", ignoredCount)
		}
		return
	}

	fmt.Printf("\nFound %d vulnerability(s) affecting current versions:\n", len(findings))
	if ignoredCount > 0 {
		fmt.Printf("(%d ignored via .security-ignore.json)\n", ignoredCount)
	}
	fmt.Println()

	for _, f := range findings {
		fmt.Printf("  %s [%s]\n", f.ID, f.Severity)
		fmt.Printf("    Package: %s@%s (%s)\n", f.Package, f.Version, f.Ecosystem)
		fmt.Printf("    Templates: %s\n", strings.Join(f.Templates, ", "))
		fmt.Printf("    More info: %s\n", f.URL)
		switch {
		case len(f.Description) > 100:
			fmt.Printf("    %s...\n\n", f.Description[:100])
		case f.Description != "":
			fmt.Printf("    %s\n\n", f.Description)
		default:
			fmt.Println()
		}
	}

	fmt.Println("To ignore a finding, add it to .security-ignore.json:")
	fmt.Println(`  {"ignore": [{"id": "GHSA-xxxx", "reason": "Not applicable because...", "expires": "2025-12-31"}]}`)
}

// extractMinVersion extracts the minimum version from a version constraint
// Examples: "^3.25.0" -> "3.25.0", ">=8.2" -> "8.2", "~> 1.73.2" -> "1.73.2"
func extractMinVersion(constraint string) string {
	constraint = strings.TrimSpace(constraint)

	// Handle empty
	if constraint == "" {
		return ""
	}

	// Patterns to extract version number
	patterns := []string{
		`^\^(.+)$`,         // npm caret: ^3.25.0
		`^~>\s*(.+)$`,      // ruby pessimistic: ~> 1.73.2 (must precede npm tilde)
		`^~(.+)$`,          // npm tilde: ~3.25.0
		`^>=\s*(.+)$`,      // gte: >=8.2
		`^>\s*(.+)$`,       // gt: >8.2
		`^v?(\d+\.\d+.*)$`, // plain version: 3.25.0 or v3.25.0
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(constraint); len(matches) > 1 {
			version := strings.TrimSpace(matches[1])
			// Clean up any trailing constraints (e.g., "^3.25.0 || ^4.0.0" -> "3.25.0")
			if idx := strings.Index(version, " "); idx > 0 {
				version = version[:idx]
			}
			// Remove any remaining special chars
			version = strings.TrimPrefix(version, "v")
			return version
		}
	}

	// If no pattern matches, try to extract any semver-like string
	semverRe := regexp.MustCompile(`(\d+\.\d+(?:\.\d+)?)`)
	if matches := semverRe.FindStringSubmatch(constraint); len(matches) > 1 {
		return matches[1]
	}

	return ""
}

func queryOSVBatch(queries []OSVQuery) ([]OSVResult, error) {
	batchQuery := OSVBatchQuery{Queries: queries}

	jsonBody, err := json.Marshal(batchQuery)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post("https://api.osv.dev/v1/querybatch", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OSV API returned status %d: %s", resp.StatusCode, string(body))
	}

	var batchResp OSVBatchResponse
	if err := json.NewDecoder(resp.Body).Decode(&batchResp); err != nil {
		return nil, err
	}

	return batchResp.Results, nil
}

func fetchVulnDetails(vulnID string) (*OSVVuln, error) {
	url := "https://api.osv.dev/v1/vulns/" + vulnID
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OSV API returned status %d: %s", resp.StatusCode, string(body))
	}

	var vuln OSVVuln
	if err := json.NewDecoder(resp.Body).Decode(&vuln); err != nil {
		return nil, err
	}

	return &vuln, nil
}

func loadIgnoreFile(path string) []IgnoreEntry {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil // File doesn't exist, that's fine
	}

	var ignoreFile IgnoreFile
	if err := json.Unmarshal(data, &ignoreFile); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to parse %s: %v\n", path, err)
		return nil
	}

	return ignoreFile.Ignore
}

func isIgnored(vulnID, packageName string, ignoreList []IgnoreEntry) bool {
	for _, entry := range ignoreList {
		if entry.ID != vulnID {
			continue
		}
		// If package is specified, it must match
		if entry.Package != "" && entry.Package != packageName {
			continue
		}
		// TODO: Check expiry date
		return true
	}
	return false
}

func severityEmoji(severity string) string {
	switch severity {
	case "CRITICAL":
		return ":rotating_light:"
	case "HIGH":
		return ":red_circle:"
	case "MODERATE", "MEDIUM":
		return ":large_orange_circle:"
	case "LOW":
		return ":large_yellow_circle:"
	default:
		return ":white_circle:"
	}
}

func sendSlackNotification(findings []Finding, webhookURL string) error {
	var sb strings.Builder
	sb.WriteString("<!subteam^S07FJCX22CA> :warning: *Security Scan: Vulnerabilities Found in SDK Dependencies*\n\n")

	// Group by severity
	bySeverity := make(map[string]int)
	for _, f := range findings {
		bySeverity[f.Severity]++
	}

	// Summary with emojis
	sb.WriteString("*Summary:* ")
	for _, sev := range []string{"CRITICAL", "HIGH", "MODERATE", "MEDIUM", "LOW", "UNKNOWN"} {
		if count, ok := bySeverity[sev]; ok {
			fmt.Fprintf(&sb, "%s %s: %d  ", severityEmoji(sev), sev, count)
		}
	}
	sb.WriteString("\n\n")

	// Show findings with details
	shown := 0
	for _, f := range findings {
		if shown >= 5 {
			fmt.Fprintf(&sb, "\n_... and %d more vulnerabilities_\n", len(findings)-5)
			break
		}

		fmt.Fprintf(&sb, "%s *<%s|%s>* `[%s]`\n", severityEmoji(f.Severity), f.URL, f.ID, f.Severity)
		fmt.Fprintf(&sb, "    Package: `%s@%s` (%s)\n", f.Package, f.Version, f.Ecosystem)
		fmt.Fprintf(&sb, "    Affects: %s\n", strings.Join(f.Templates, ", "))
		if f.Description != "" {
			desc := f.Description
			if len(desc) > 80 {
				desc = desc[:77] + "..."
			}
			fmt.Fprintf(&sb, "    _%s_\n", desc)
		}
		sb.WriteString("\n")
		shown++
	}

	sb.WriteString("\n<https://github.com/speakeasy-api/openapi-generation|View Repository>")

	payload := map[string]string{"text": sb.String()}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Slack returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
