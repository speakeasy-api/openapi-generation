package changes

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// testSpecPair helps create test configurations from inline spec strings
type testSpecPair struct {
	oldSpec string
	newSpec string
	lang    string
}

// createTestConfigs writes spec strings to temp files and returns configured GenerateOptions
func createTestConfigs(t *testing.T, specs testSpecPair) (oldConfig, newConfig GenerateOptions) {
	t.Helper()

	// Create temp directory for test files
	tmpDir := t.TempDir()

	oldSpecPath := filepath.Join(tmpDir, "old_spec.yaml")
	newSpecPath := filepath.Join(tmpDir, "new_spec.yaml")

	// Write specs to files
	if err := os.WriteFile(oldSpecPath, []byte(specs.oldSpec), 0644); err != nil {
		t.Fatalf("Failed to write old spec: %v", err)
	}
	if err := os.WriteFile(newSpecPath, []byte(specs.newSpec), 0644); err != nil {
		t.Fatalf("Failed to write new spec: %v", err)
	}

	oldConfig = GenerateOptions{
		Lang:       specs.lang,
		SchemaPath: oldSpecPath,
		OutDir:     filepath.Join(tmpDir, "old"),
	}
	newConfig = GenerateOptions{
		Lang:       specs.lang,
		SchemaPath: newSpecPath,
		OutDir:     filepath.Join(tmpDir, "new"),
	}

	return oldConfig, newConfig
}

// formatTestSnapshotWithDiff formats test output with both compact and full mode outputs
func formatTestSnapshotWithDiff(testName, oldSpec, newSpec string, diff SDKDiff) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "# Test: %s\n\n", testName)

	sb.WriteString("## Old Spec:\n```yaml\n")
	sb.WriteString(strings.TrimSpace(oldSpec))
	sb.WriteString("\n```\n\n")

	sb.WriteString("## New Spec:\n```yaml\n")
	sb.WriteString(strings.TrimSpace(newSpec))
	sb.WriteString("\n```\n\n")

	sb.WriteString("## Compact Mode:\n")
	sb.WriteString(ToMarkdown(diff, DetailLevelCompact))
	sb.WriteString("\n")

	sb.WriteString("## Full Mode:\n")
	sb.WriteString(ToMarkdown(diff, DetailLevelFull))

	return sb.String()
}
