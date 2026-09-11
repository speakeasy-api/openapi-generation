package main

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func loadFixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return b
}

func newAcc(threshold int, ignore []IgnoreEntry) *npmAuditAccumulator {
	return &npmAuditAccumulator{
		findings:   make(map[string]*Finding),
		threshold:  threshold,
		ignoreList: ignore,
	}
}

func TestParseNPMAudit_FixtureFields(t *testing.T) {
	acc := newAcc(severityRanks["low"], nil)
	if err := parseNPMAudit(loadFixture(t, "npm-audit-mcp-typescript.json"), "mcp-typescript", acc); err != nil {
		t.Fatalf("parse: %v", err)
	}

	fastURI, ok := acc.findings["GHSA-q3j6-qgpj-74h6|fast-uri"]
	if !ok {
		t.Fatalf("missing fast-uri finding; got keys %v", keysOf(acc.findings))
	}
	if fastURI.Severity != "HIGH" {
		t.Errorf("fast-uri severity: got %q want HIGH", fastURI.Severity)
	}
	if fastURI.Ecosystem != "npm" || fastURI.Source != "npm-audit" {
		t.Errorf("fast-uri ecosystem/source: %+v", fastURI)
	}
	if fastURI.Range != "<=3.1.0" {
		t.Errorf("fast-uri range: got %q want <=3.1.0", fastURI.Range)
	}
	if fastURI.FixAvailable != "true" {
		t.Errorf("fast-uri fixAvailable: got %q want true", fastURI.FixAvailable)
	}
	if !slices.Equal(fastURI.Paths, []string{"node_modules/fast-uri"}) {
		t.Errorf("fast-uri paths: %v", fastURI.Paths)
	}
	if !slices.Equal(fastURI.Templates, []string{"mcp-typescript"}) {
		t.Errorf("fast-uri templates: %v", fastURI.Templates)
	}
}

func TestParseNPMAudit_ViaStringSkipped(t *testing.T) {
	// express-rate-limit has via: ["ip-address"] (string ref), should not surface
	// a finding on its own — only the actual ip-address advisory should.
	acc := newAcc(severityRanks["low"], nil)
	if err := parseNPMAudit(loadFixture(t, "npm-audit-mcp-typescript.json"), "mcp-typescript", acc); err != nil {
		t.Fatalf("parse: %v", err)
	}
	for k := range acc.findings {
		if k == "|express-rate-limit" || k == "npm-0|express-rate-limit" {
			t.Errorf("string via produced finding for express-rate-limit: key=%q", k)
		}
	}
}

func TestParseNPMAudit_ThresholdFilter(t *testing.T) {
	// hono fixture has 6 advisory objects, 1 of severity "low".
	// Threshold = moderate must filter out the low one but keep the moderates.
	acc := newAcc(severityRanks["moderate"], nil)
	if err := parseNPMAudit(loadFixture(t, "npm-audit-mcp-typescript.json"), "mcp-typescript", acc); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if _, present := acc.findings["GHSA-hm8q-7f3q-5f36|hono"]; present {
		t.Errorf("low-severity advisory should be filtered at moderate threshold")
	}
	// Positive assertions: moderate + high advisories must still be present.
	for _, key := range []string{
		"GHSA-458j-xx4x-4375|hono",       // moderate
		"GHSA-9vqf-7f2p-gf9v|hono",       // moderate
		"GHSA-q3j6-qgpj-74h6|fast-uri",   // high
		"GHSA-v2v4-37r5-5v8g|ip-address", // moderate
	} {
		if _, present := acc.findings[key]; !present {
			t.Errorf("expected finding %q at moderate threshold, missing", key)
		}
	}
}

func TestParseNPMAudit_CrossTemplateDedup(t *testing.T) {
	acc := newAcc(severityRanks["low"], nil)
	if err := parseNPMAudit(loadFixture(t, "npm-audit-mcp-typescript.json"), "mcp-typescript", acc); err != nil {
		t.Fatalf("parse 1: %v", err)
	}
	if err := parseNPMAudit(loadFixture(t, "npm-audit-synthetic-other.json"), "sdk-typescriptv2", acc); err != nil {
		t.Fatalf("parse 2: %v", err)
	}

	// Overlap: fast-uri present in both. Templates merged, paths merged + deduped.
	fastURI := acc.findings["GHSA-q3j6-qgpj-74h6|fast-uri"]
	if fastURI == nil {
		t.Fatal("fast-uri missing")
		return
	}
	if !slices.Equal(fastURI.Templates, []string{"mcp-typescript", "sdk-typescriptv2"}) {
		t.Errorf("templates not merged: %v", fastURI.Templates)
	}
	wantPaths := []string{"node_modules/fast-uri", "node_modules/some-dep/node_modules/fast-uri"}
	if !slices.Equal(fastURI.Paths, wantPaths) {
		t.Errorf("paths merge: got %v want %v", fastURI.Paths, wantPaths)
	}

	// Unique to mcp-typescript only.
	hono := acc.findings["GHSA-458j-xx4x-4375|hono"]
	if hono == nil || !slices.Equal(hono.Templates, []string{"mcp-typescript"}) {
		t.Errorf("hono templates: %+v", hono)
	}

	// Unique to sdk-typescriptv2 only.
	lodash := acc.findings["GHSA-jf85-cpcp-j695|lodash"]
	if lodash == nil || !slices.Equal(lodash.Templates, []string{"sdk-typescriptv2"}) {
		t.Errorf("lodash templates: %+v", lodash)
	}
}

func TestParseNPMAudit_IgnoreList(t *testing.T) {
	acc := newAcc(severityRanks["low"], []IgnoreEntry{{ID: "GHSA-q3j6-qgpj-74h6"}})
	if err := parseNPMAudit(loadFixture(t, "npm-audit-mcp-typescript.json"), "mcp-typescript", acc); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if _, present := acc.findings["GHSA-q3j6-qgpj-74h6|fast-uri"]; present {
		t.Errorf("ignored finding still present")
	}
	if acc.ignored == 0 {
		t.Errorf("ignored counter not incremented")
	}
}

func TestAdvisoryIDFromURL(t *testing.T) {
	cases := map[string]string{
		"https://github.com/advisories/GHSA-q3j6-qgpj-74h6":  "GHSA-q3j6-qgpj-74h6",
		"https://github.com/advisories/GHSA-q3j6-qgpj-74h6/": "GHSA-q3j6-qgpj-74h6",
		"https://nvd.nist.gov/vuln/detail/CVE-2023-12345":    "CVE-2023-12345",
		"":          "",
		"not-an-id": "",
	}
	for in, want := range cases {
		if got := advisoryIDFromURL(in); got != want {
			t.Errorf("advisoryIDFromURL(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFormatFixAvailable(t *testing.T) {
	cases := map[string]string{
		`true`:                            "true",
		`false`:                           "false",
		`{"name":"hono","version":"4.0"}`: "hono@4.0",
		`{"name":"hono"}`:                 "hono",
		``:                                "",
		`null`:                            "",
	}
	for in, want := range cases {
		got := formatFixAvailable([]byte(in))
		if got != want {
			t.Errorf("formatFixAvailable(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestParseNPMAudit_MalformedJSON(t *testing.T) {
	acc := newAcc(severityRanks["low"], nil)
	err := parseNPMAudit([]byte("{not-json"), "x", acc)
	if err == nil {
		t.Fatal("expected error on malformed JSON")
	}
	if len(acc.findings) != 0 {
		t.Errorf("findings populated despite parse error: %d", len(acc.findings))
	}
}

func TestParseNPMAudit_UnknownSeverityFiltered(t *testing.T) {
	// Advisory with severity "info" -> rank 0. Filtered when threshold >= 1.
	raw := []byte(`{
		"vulnerabilities": {
			"foo": {
				"name": "foo",
				"severity": "info",
				"via": [{
					"source": 1,
					"name": "foo",
					"title": "info-level advisory",
					"url": "https://github.com/advisories/GHSA-aaaa-bbbb-cccc",
					"severity": "info",
					"range": "<1"
				}],
				"nodes": ["node_modules/foo"],
				"fixAvailable": false
			}
		}
	}`)

	acc := newAcc(severityRanks["low"], nil)
	if err := parseNPMAudit(raw, "x", acc); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(acc.findings) != 0 {
		t.Errorf("info severity should be filtered at threshold=low (rank 1), got %d findings", len(acc.findings))
	}

	// With threshold 0 (no filter), should surface.
	acc = newAcc(0, nil)
	if err := parseNPMAudit(raw, "x", acc); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(acc.findings) != 1 {
		t.Errorf("info severity should pass threshold=0, got %d findings", len(acc.findings))
	}
}

func TestParseNPMAudit_DescriptionFromTitle(t *testing.T) {
	acc := newAcc(severityRanks["low"], nil)
	if err := parseNPMAudit(loadFixture(t, "npm-audit-mcp-typescript.json"), "mcp-typescript", acc); err != nil {
		t.Fatalf("parse: %v", err)
	}
	f := acc.findings["GHSA-q3j6-qgpj-74h6|fast-uri"]
	want := "fast-uri vulnerable to path traversal via percent-encoded dot segments"
	if f.Description != want {
		t.Errorf("Description: got %q want %q", f.Description, want)
	}
	if f.URL != "https://github.com/advisories/GHSA-q3j6-qgpj-74h6" {
		t.Errorf("URL: %q", f.URL)
	}
	if f.Package != "fast-uri" {
		t.Errorf("Package: %q", f.Package)
	}
}

func keysOf(m map[string]*Finding) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
