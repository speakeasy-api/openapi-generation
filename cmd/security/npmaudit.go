package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// advisoryIDRe matches GHSA and CVE identifiers anywhere in a string.
var advisoryIDRe = regexp.MustCompile(`(GHSA-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{4}|CVE-\d{4}-\d+)`)

// Paths to run `npm audit` for (transitive dependency vulnerabilities)
var npmAuditPaths = []string{
	"zSDKs/sdk-typescriptv2",
	"zSDKs/mcp-typescript",
}

// severityRanks maps npm audit severity strings to a comparable integer.
// Higher = more severe. Unknown values resolve to 0 and are filtered out
// when any threshold > 0 is set.
var severityRanks = map[string]int{
	"critical": 4,
	"high":     3,
	"moderate": 2,
	"low":      1,
}

// severityThreshold resolves a severity string to its rank, accepting the
// "medium" alias for "moderate". The bool reports whether the value was valid.
func severityThreshold(severity string) (int, bool) {
	s := strings.ToLower(strings.TrimSpace(severity))
	if s == "medium" {
		s = "moderate"
	}
	rank, ok := severityRanks[s]
	return rank, ok
}

// meetsSeverityThreshold reports whether a severity string meets or exceeds the
// given threshold. A threshold <= 0 disables filtering (everything passes).
// Handles OSV severity casing/aliases ("MEDIUM" == moderate). Unknown/unscored
// severities always pass so they are never silently dropped.
func meetsSeverityThreshold(severity string, threshold int) bool {
	if threshold <= 0 {
		return true
	}
	rank, ok := severityThreshold(severity)
	if !ok {
		return true
	}
	return rank >= threshold
}

// npmAuditEnabled reports whether the npm audit scan should run given the
// `-npm-audit` flag value. "skip" (or empty) disables the scan.
func npmAuditEnabled(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "", "skip":
		return false
	}
	return true
}

// npm audit JSON types (auditReportVersion 2)
type npmAuditReport struct {
	Vulnerabilities map[string]npmAuditVuln `json:"vulnerabilities"`
}

type npmAuditVuln struct {
	Name     string            `json:"name"`
	Severity string            `json:"severity"`
	Via      []json.RawMessage `json:"via"`
	Range    string            `json:"range"`
	Nodes    []string          `json:"nodes"`
	// FixAvailable is either a bool or an object describing the fix.
	FixAvailable json.RawMessage `json:"fixAvailable"`
}

type npmAuditAdvisory struct {
	Source     int    `json:"source"`
	Name       string `json:"name"`
	Dependency string `json:"dependency"`
	Title      string `json:"title"`
	URL        string `json:"url"`
	Severity   string `json:"severity"`
	Range      string `json:"range"`
}

func runNPMAuditScan(paths []string, threshold int, ignoreList []IgnoreEntry) ([]Finding, int, error) {
	acc := &npmAuditAccumulator{
		findings:   make(map[string]*Finding),
		threshold:  threshold,
		ignoreList: ignoreList,
	}

	for _, dir := range paths {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Warning: npm audit path %s not found, skipping\n", dir)
			continue
		}

		args := []string{"audit", "--json", "--package-lock-only", "--omit=dev"}
		fmt.Printf("Running npm %s in %s...\n", strings.Join(args, " "), dir)

		cmd := exec.Command("npm", args...)
		cmd.Dir = dir
		out, err := cmd.Output()
		// npm audit exits non-zero when vulnerabilities are found; ignore exit
		// errors as long as we got JSON output to parse.
		if err != nil {
			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) || len(out) == 0 {
				return nil, acc.ignored, fmt.Errorf("npm audit failed in %s: %w", dir, err)
			}
		}

		if err := parseNPMAudit(out, filepath.Base(dir), loadInstalledVersions(dir), acc); err != nil {
			return nil, acc.ignored, fmt.Errorf("failed to parse npm audit output in %s: %w", dir, err)
		}
	}

	findings := make([]Finding, 0, len(acc.advisoryOrder))
	for _, k := range acc.advisoryOrder {
		findings = append(findings, *acc.findings[k])
	}
	return findings, acc.ignored, nil
}

// parseNPMAudit ingests one npm audit JSON report into the accumulator.
// Extracted from runNPMAuditScan to enable fixture-driven testing.
func parseNPMAudit(out []byte, templateName string, installed map[string]string, acc *npmAuditAccumulator) error {
	var report npmAuditReport
	if err := json.Unmarshal(out, &report); err != nil {
		return err
	}
	for _, vuln := range report.Vulnerabilities {
		for _, raw := range vuln.Via {
			acc.add(raw, vuln, templateName, installed)
		}
	}
	return nil
}

type npmLockfile struct {
	Packages map[string]struct {
		Version string `json:"version"`
	} `json:"packages"`
}

// loadInstalledVersions maps each node_modules path in the directory's
// package-lock.json to its installed version so npm audit findings carry the
// version the advisory was matched against.
func loadInstalledVersions(dir string) map[string]string {
	b, err := os.ReadFile(filepath.Join(dir, "package-lock.json"))
	if err != nil {
		return nil
	}
	var lock npmLockfile
	if err := json.Unmarshal(b, &lock); err != nil {
		return nil
	}
	versions := make(map[string]string, len(lock.Packages))
	for path, pkg := range lock.Packages {
		versions[path] = pkg.Version
	}
	return versions
}

func installedVersion(nodes []string, installed map[string]string) string {
	for _, node := range nodes {
		if v := installed[node]; v != "" {
			return v
		}
	}
	return ""
}

// npmAuditAccumulator collects findings from npm audit `via` entries across
// templates, deduplicating by (id, package).
type npmAuditAccumulator struct {
	findings      map[string]*Finding
	advisoryOrder []string
	ignored       int
	threshold     int
	ignoreList    []IgnoreEntry
}

// add processes one `via` entry from an npm audit report. `via` entries are
// either strings (pointers to other packages in the report, ignored here) or
// advisory objects carrying the metadata we surface.
func (a *npmAuditAccumulator) add(raw json.RawMessage, vuln npmAuditVuln, templateName string, installed map[string]string) {
	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		return
	}
	var adv npmAuditAdvisory
	if err := json.Unmarshal(raw, &adv); err != nil {
		return
	}
	if adv.URL == "" && adv.Title == "" {
		return
	}
	// Unknown advisory severities resolve to 0 and get filtered
	// when a non-zero threshold is set.
	if severityRanks[strings.ToLower(adv.Severity)] < a.threshold {
		return
	}

	id := advisoryIDFromURL(adv.URL)
	if id == "" {
		id = fmt.Sprintf("npm-%d", adv.Source)
	}

	version := installedVersion(vuln.Nodes, installed)
	if isIgnored(id, adv.Name, version, a.ignoreList) {
		a.ignored++
		return
	}

	key := fmt.Sprintf("%s|%s", id, adv.Name)

	if existing, ok := a.findings[key]; ok {
		if !slices.Contains(existing.Templates, templateName) {
			existing.Templates = append(existing.Templates, templateName)
		}
		for _, p := range vuln.Nodes {
			if !slices.Contains(existing.Paths, p) {
				existing.Paths = append(existing.Paths, p)
			}
		}
		return
	}

	f := &Finding{
		ID:           id,
		Severity:     strings.ToUpper(adv.Severity),
		Description:  adv.Title,
		URL:          adv.URL,
		Package:      adv.Name,
		Ecosystem:    "npm",
		Version:      version,
		Range:        adv.Range,
		Templates:    []string{templateName},
		Source:       "npm-audit",
		FixAvailable: formatFixAvailable(vuln.FixAvailable),
	}
	f.Paths = append(f.Paths, vuln.Nodes...)
	a.findings[key] = f
	a.advisoryOrder = append(a.advisoryOrder, key)
}

func advisoryIDFromURL(url string) string {
	return advisoryIDRe.FindString(url)
}

func formatFixAvailable(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var b bool
	if err := json.Unmarshal(raw, &b); err == nil {
		if b {
			return "true"
		}
		return "false"
	}
	var obj struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	if err := json.Unmarshal(raw, &obj); err == nil && obj.Name != "" {
		if obj.Version != "" {
			return obj.Name + "@" + obj.Version
		}
		return obj.Name
	}
	return ""
}
