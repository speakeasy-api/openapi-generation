package snapshotdispatch

import (
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"
)

func TestDispatcherNormalizesEstablishedTargetLabels(t *testing.T) {
	data, err := os.ReadFile("../../.github/workflows/snapshot-dispatch.yml")
	if err != nil {
		t.Fatalf("read snapshot dispatcher: %v", err)
	}
	workflow := string(data)

	expected := map[string]string{
		"snapshot":                "snapshot",
		"snapshot-all":            "snapshot-all",
		"cli":                     "snapshot-cli",
		"csharp":                  "snapshot-csharp",
		"go":                      "snapshot-go",
		"java":                    "snapshot-java",
		"mcp-typescript":          "snapshot-mcp-typescript",
		"php":                     "snapshot-php",
		"postman":                 "snapshot-postman",
		"python":                  "snapshot-python",
		"ruby":                    "snapshot-ruby",
		"terraform":               "snapshot-terraform",
		"typescript":              "snapshot-typescript",
		"snapshot-cli":            "snapshot-cli",
		"snapshot-csharp":         "snapshot-csharp",
		"snapshot-go":             "snapshot-go",
		"snapshot-java":           "snapshot-java",
		"snapshot-mcp-typescript": "snapshot-mcp-typescript",
		"snapshot-php":            "snapshot-php",
		"snapshot-postman":        "snapshot-postman",
		"snapshot-python":         "snapshot-python",
		"snapshot-ruby":           "snapshot-ruby",
		"snapshot-terraform":      "snapshot-terraform",
		"snapshot-typescript":     "snapshot-typescript",
	}

	mappingPattern := regexp.MustCompile(`\['([^']+)', '([^']+)'\]`)
	actual := map[string]string{}
	for _, match := range mappingPattern.FindAllStringSubmatch(workflow, -1) {
		if previous, exists := actual[match[1]]; exists {
			t.Fatalf("dispatcher contains duplicate alias %q with values %q and %q", match[1], previous, match[2])
		}
		actual[match[1]] = match[2]
	}
	if len(actual) != len(expected) {
		t.Fatalf("dispatcher contains %d aliases, want %d: %#v", len(actual), len(expected), actual)
	}
	for alias, selector := range expected {
		if actual[alias] != selector {
			t.Errorf("dispatcher alias %q = %q, want %q", alias, actual[alias], selector)
		}
	}

	selectorSet := map[string]struct{}{}
	for _, selector := range actual {
		selectorSet[selector] = struct{}{}
	}
	selectors := make([]string, 0, len(selectorSet))
	for selector := range selectorSet {
		selectors = append(selectors, selector)
	}
	slices.Sort(selectors)
	expectedSelectors := []string{
		"snapshot",
		"snapshot-all",
		"snapshot-cli",
		"snapshot-csharp",
		"snapshot-go",
		"snapshot-java",
		"snapshot-mcp-typescript",
		"snapshot-php",
		"snapshot-postman",
		"snapshot-python",
		"snapshot-ruby",
		"snapshot-terraform",
		"snapshot-typescript",
	}
	if !slices.Equal(selectors, expectedSelectors) {
		t.Errorf("dispatcher selector codomain = %v, want %v", selectors, expectedSelectors)
	}

	required := []string{
		"pull_request_target:",
		"issue_comment:",
		"!selectorAliases.has(context.payload.label.name)",
		".map((label) => selectorAliases.get(label.name))",
		".filter(Boolean)",
		")].sort();",
		"Customer/repository",
	}
	for _, value := range required {
		if !strings.Contains(workflow, value) {
			t.Errorf("dispatcher is missing normalized selector contract %q", value)
		}
	}

	for _, value := range []string{"allowedSelectors", "snapshot-unity"} {
		if strings.Contains(workflow, value) {
			t.Errorf("dispatcher retains obsolete selector contract %q", value)
		}
	}
}
