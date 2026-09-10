package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/changeset"
	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
)

func setupTestDir(t *testing.T, featureVersions map[string]string) {
	t.Helper()
	setupTestDirForTemplate(t, "go", featureVersions)
}

func setupTestDirForTemplate(t *testing.T, template string, featureVersions map[string]string) {
	t.Helper()

	dir := t.TempDir()
	t.Chdir(dir)

	// Create .changesets directory
	if err := os.MkdirAll(changeset.Dir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a minimal features.ts
	templateDir := filepath.Join(dir, "templates", "templates", template)
	if err := os.MkdirAll(templateDir, 0o755); err != nil {
		t.Fatal(err)
	}

	var b strings.Builder
	b.WriteString("const supportedFeatures = {\n")
	for name, ver := range featureVersions {
		fmt.Fprintf(&b, "  %s: \"%s\",\n", name, ver)
	}
	b.WriteString("};\n")

	if err := os.WriteFile(filepath.Join(templateDir, "features.ts"), []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create changelogs directory for target
	if err := os.MkdirAll(filepath.Join(dir, "changelogs", template), 0o755); err != nil {
		t.Fatal(err)
	}

	target := types.NewTargetFromTemplate(template)
	writeReviewGenLock(t, template, target.Target, featureVersions)
}

func writeReviewGenLock(t *testing.T, template string, target string, featureVersions map[string]string) {
	t.Helper()

	genLockPath, ok := reviewGenLockPath(template)
	if !ok {
		t.Fatalf("expected review gen.lock path for template %s", template)
	}

	if err := os.MkdirAll(filepath.Dir(genLockPath), 0o755); err != nil {
		t.Fatal(err)
	}

	var b strings.Builder
	b.WriteString("lockVersion: 2.0.0\n")
	b.WriteString("features:\n")
	fmt.Fprintf(&b, "  %s:\n", target)
	for name, ver := range featureVersions {
		fmt.Fprintf(&b, "    %s: %s\n", name, ver)
	}

	if err := os.WriteFile(genLockPath, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestApplyChangesets_SingleChangeset(t *testing.T) {
	setupTestDir(t, map[string]string{
		"core": "3.13.21",
	})

	cs := &changeset.Changeset{
		ID:          "test-001",
		Features:    []string{"core"},
		Targets:     []string{"go"},
		Type:        "fix",
		Bump:        "patch",
		Description: "fixed JSON serialization",
		Author:      "testuser",
		Date:        "2026-03-10",
	}

	if _, err := changeset.Write(cs); err != nil {
		t.Fatal(err)
	}

	if err := applyChangesets(false); err != nil {
		t.Fatal(err)
	}

	// Verify features.ts was updated
	ver, err := readFeatureVersion(features.FeatureCore, types.NewTargetFromTemplate("go"))
	if err != nil {
		t.Fatal(err)
	}
	if ver != "3.13.22" {
		t.Errorf("expected version 3.13.22, got %s", ver)
	}

	reviewGenLockPath, ok := reviewGenLockPath("go")
	if !ok {
		t.Fatal("expected review gen.lock path for go")
	}

	reviewGenLock, err := os.ReadFile(reviewGenLockPath)
	if err != nil {
		t.Fatal(err)
	}
	if expected := "    core: 3.13.22\n"; !strings.Contains(string(reviewGenLock), expected) {
		t.Errorf("expected review gen.lock to contain %q, got: %s", expected, string(reviewGenLock))
	}

	// Verify changelog file was created
	changelogFile := filepath.Join("changelogs", "go", "core-3.13.22.md")
	data, err := os.ReadFile(changelogFile)
	if err != nil {
		t.Fatalf("changelog file not created: %v", err)
	}
	content := string(data)
	if expected := "## core: 3.13.22 - 2026-03-10\n"; !strings.Contains(content, expected) {
		t.Errorf("changelog missing header, got: %s", content)
	}
	if expected := "### :bug: Bug Fixes\n"; !strings.Contains(content, expected) {
		t.Errorf("changelog missing category, got: %s", content)
	}
	if expected := "fixed JSON serialization"; !strings.Contains(content, expected) {
		t.Errorf("changelog missing description, got: %s", content)
	}

	// Verify changesets were deleted
	remaining, err := changeset.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(remaining) != 0 {
		t.Errorf("expected 0 remaining changesets, got %d", len(remaining))
	}
}

func TestApplyChangesets_MultipleChangesetsSameFeature(t *testing.T) {
	setupTestDir(t, map[string]string{
		"core": "3.13.21",
	})

	// Two changesets both targeting core on go
	cs1 := &changeset.Changeset{
		ID:          "aaa-001",
		Features:    []string{"core"},
		Targets:     []string{"go"},
		Type:        "fix",
		Bump:        "patch",
		Description: "first fix",
		Author:      "user1",
		Date:        "2026-03-09",
	}
	cs2 := &changeset.Changeset{
		ID:          "bbb-002",
		Features:    []string{"core"},
		Targets:     []string{"go"},
		Type:        "feat",
		Bump:        "patch",
		Description: "second fix",
		Author:      "user2",
		Date:        "2026-03-10",
	}

	if _, err := changeset.Write(cs1); err != nil {
		t.Fatal(err)
	}
	if _, err := changeset.Write(cs2); err != nil {
		t.Fatal(err)
	}

	if err := applyChangesets(false); err != nil {
		t.Fatal(err)
	}

	// Should have bumped twice: 3.13.21 -> 3.13.22 -> 3.13.23
	ver, err := readFeatureVersion(features.FeatureCore, types.NewTargetFromTemplate("go"))
	if err != nil {
		t.Fatal(err)
	}
	if ver != "3.13.23" {
		t.Errorf("expected version 3.13.23, got %s", ver)
	}

	// Both changelog files should exist
	if _, err := os.Stat(filepath.Join("changelogs", "go", "core-3.13.22.md")); err != nil {
		t.Errorf("expected core-3.13.22.md to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join("changelogs", "go", "core-3.13.23.md")); err != nil {
		t.Errorf("expected core-3.13.23.md to exist: %v", err)
	}
}

func TestApplyChangesets_MajorBump(t *testing.T) {
	setupTestDir(t, map[string]string{
		"core": "3.13.21",
	})

	cs := &changeset.Changeset{
		ID:          "test-001",
		Features:    []string{"core"},
		Targets:     []string{"go"},
		Type:        "feat",
		Bump:        "major",
		Description: "breaking change",
		Author:      "testuser",
		Date:        "2026-03-10",
	}

	if _, err := changeset.Write(cs); err != nil {
		t.Fatal(err)
	}

	if err := applyChangesets(false); err != nil {
		t.Fatal(err)
	}

	ver, err := readFeatureVersion(features.FeatureCore, types.NewTargetFromTemplate("go"))
	if err != nil {
		t.Fatal(err)
	}
	if ver != "4.0.0" {
		t.Errorf("expected version 4.0.0, got %s", ver)
	}
}

func TestApplyChangesets_MinorBump(t *testing.T) {
	setupTestDir(t, map[string]string{
		"core": "3.13.21",
	})

	cs := &changeset.Changeset{
		ID:          "test-001",
		Features:    []string{"core"},
		Targets:     []string{"go"},
		Type:        "feat",
		Bump:        "minor",
		Description: "new feature",
		Author:      "testuser",
		Date:        "2026-03-10",
	}

	if _, err := changeset.Write(cs); err != nil {
		t.Fatal(err)
	}

	if err := applyChangesets(false); err != nil {
		t.Fatal(err)
	}

	ver, err := readFeatureVersion(features.FeatureCore, types.NewTargetFromTemplate("go"))
	if err != nil {
		t.Fatal(err)
	}
	if ver != "3.14.0" {
		t.Errorf("expected version 3.14.0, got %s", ver)
	}
}

func TestReviewGenLockPath(t *testing.T) {
	tests := []struct {
		template string
		want     string
		ok       bool
	}{
		{template: "go", want: "zSDKs/sdk-go/.speakeasy/gen.lock", ok: true},
		{template: "mcp-typescript", want: "zSDKs/mcp-typescript/.speakeasy/gen.lock", ok: true},
		{template: "terraform", want: "zSDKs/terraform-provider-testing/.speakeasy/gen.lock", ok: true},
		{template: "mockserver", want: "", ok: false},
	}

	for _, tt := range tests {
		got, ok := reviewGenLockPath(tt.template)
		if got != tt.want || ok != tt.ok {
			t.Errorf("reviewGenLockPath(%q) = (%q, %t), want (%q, %t)", tt.template, got, ok, tt.want, tt.ok)
		}
	}
}

func TestBumpReviewSDKFeatureVersion_UsesBaseTargetForVersionedTemplates(t *testing.T) {
	tests := []struct {
		template string
		target   string
	}{
		{template: "pythonv2", target: "python"},
		{template: "javav2", target: "java"},
		{template: "typescriptv2", target: "typescript"},
	}

	for _, tt := range tests {
		t.Run(tt.template, func(t *testing.T) {
			setupTestDirForTemplate(t, tt.template, map[string]string{
				"core": "1.2.3",
			})

			target := types.NewTargetFromTemplate(tt.template)
			if err := bumpReviewSDKFeatureVersion(features.FeatureCore, target, "1.2.4", false); err != nil {
				t.Fatal(err)
			}

			genLockPath, ok := reviewGenLockPath(tt.template)
			if !ok {
				t.Fatalf("expected review gen.lock path for template %s", tt.template)
			}

			reviewGenLock, err := os.ReadFile(genLockPath)
			if err != nil {
				t.Fatal(err)
			}

			if expected := fmt.Sprintf("  %s:\n    core: 1.2.4\n", tt.target); !strings.Contains(string(reviewGenLock), expected) {
				t.Fatalf("expected review gen.lock to contain %q, got: %s", expected, string(reviewGenLock))
			}
		})
	}
}

func TestApplyChangesets_InvalidFeatureFails(t *testing.T) {
	setupTestDir(t, map[string]string{
		"core": "3.13.21",
	})

	cs := &changeset.Changeset{
		ID:          "test-invalid-feature",
		Features:    []string{"notAFeature"},
		Targets:     []string{"go"},
		Type:        "fix",
		Bump:        "patch",
		Description: "bad feature",
		Author:      "testuser",
		Date:        "2026-03-10",
	}

	if _, err := changeset.Write(cs); err != nil {
		t.Fatal(err)
	}

	err := applyChangesets(false)
	if err == nil || !strings.Contains(err.Error(), `invalid feature "notAFeature"`) {
		t.Fatalf("expected invalid feature error, got %v", err)
	}

	remaining, err := changeset.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(remaining) != 1 {
		t.Fatalf("expected changeset to remain after failure, got %d", len(remaining))
	}
}

func TestApplyChangesets_InvalidTargetFails(t *testing.T) {
	setupTestDir(t, map[string]string{
		"core": "3.13.21",
	})

	cs := &changeset.Changeset{
		ID:          "test-invalid-target",
		Features:    []string{"core"},
		Targets:     []string{"notATarget"},
		Type:        "fix",
		Bump:        "patch",
		Description: "bad target",
		Author:      "testuser",
		Date:        "2026-03-10",
	}

	if _, err := changeset.Write(cs); err != nil {
		t.Fatal(err)
	}

	err := applyChangesets(false)
	if err == nil || !strings.Contains(err.Error(), `invalid target "notATarget"`) {
		t.Fatalf("expected invalid target error, got %v", err)
	}

	remaining, err := changeset.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(remaining) != 1 {
		t.Fatalf("expected changeset to remain after failure, got %d", len(remaining))
	}
}

func TestApplyChangesets_UnsupportedFeatureSkipped(t *testing.T) {
	setupTestDir(t, map[string]string{
		"core": "3.13.21",
		// "pagination" is NOT in the features.ts, so it should be skipped
	})

	cs := &changeset.Changeset{
		ID:          "test-001",
		Features:    []string{"core", "pagination"},
		Targets:     []string{"go"},
		Type:        "fix",
		Bump:        "patch",
		Description: "fix something",
		Author:      "testuser",
		Date:        "2026-03-10",
	}

	if _, err := changeset.Write(cs); err != nil {
		t.Fatal(err)
	}

	if err := applyChangesets(false); err != nil {
		t.Fatal(err)
	}

	// Core should be bumped
	ver, err := readFeatureVersion(features.FeatureCore, types.NewTargetFromTemplate("go"))
	if err != nil {
		t.Fatal(err)
	}
	if ver != "3.13.22" {
		t.Errorf("expected version 3.13.22, got %s", ver)
	}

	// Pagination changelog should NOT exist
	if _, err := os.Stat(filepath.Join("changelogs", "go", "pagination-0.0.1.md")); !os.IsNotExist(err) {
		t.Error("expected pagination changelog to NOT exist")
	}
}

func TestApplyChangesets_DryRun(t *testing.T) {
	setupTestDir(t, map[string]string{
		"core": "3.13.21",
	})

	cs := &changeset.Changeset{
		ID:          "test-001",
		Features:    []string{"core"},
		Targets:     []string{"go"},
		Type:        "fix",
		Bump:        "patch",
		Description: "fixed something",
		Author:      "testuser",
		Date:        "2026-03-10",
	}

	if _, err := changeset.Write(cs); err != nil {
		t.Fatal(err)
	}

	if err := applyChangesets(true); err != nil {
		t.Fatal(err)
	}

	// Version should NOT have changed
	ver, err := readFeatureVersion(features.FeatureCore, types.NewTargetFromTemplate("go"))
	if err != nil {
		t.Fatal(err)
	}
	if ver != "3.13.21" {
		t.Errorf("expected version 3.13.21 (unchanged), got %s", ver)
	}

	// Changeset should NOT have been deleted
	remaining, err := changeset.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(remaining) != 1 {
		t.Errorf("expected 1 remaining changeset, got %d", len(remaining))
	}
}

func TestApplyChangesets_DryRun_MissingReviewGenLockFails(t *testing.T) {
	setupTestDir(t, map[string]string{
		"core": "3.13.21",
	})

	genLockPath, ok := reviewGenLockPath("go")
	if !ok {
		t.Fatal("expected review gen.lock path for go")
	}
	if err := os.Remove(genLockPath); err != nil {
		t.Fatal(err)
	}

	cs := &changeset.Changeset{
		ID:          "test-001",
		Features:    []string{"core"},
		Targets:     []string{"go"},
		Type:        "fix",
		Bump:        "patch",
		Description: "fixed something",
		Author:      "testuser",
		Date:        "2026-03-10",
	}

	if _, err := changeset.Write(cs); err != nil {
		t.Fatal(err)
	}

	err := applyChangesets(true)
	if err == nil || !strings.Contains(err.Error(), "review gen.lock") {
		t.Fatalf("expected review gen.lock error, got %v", err)
	}
}

func TestApplyChangesets_DryRun_MissingChangelogDirFails(t *testing.T) {
	setupTestDir(t, map[string]string{
		"core": "3.13.21",
	})

	if err := os.RemoveAll(filepath.Join("changelogs", "go")); err != nil {
		t.Fatal(err)
	}

	cs := &changeset.Changeset{
		ID:          "test-001",
		Features:    []string{"core"},
		Targets:     []string{"go"},
		Type:        "fix",
		Bump:        "patch",
		Description: "fixed something",
		Author:      "testuser",
		Date:        "2026-03-10",
	}

	if _, err := changeset.Write(cs); err != nil {
		t.Fatal(err)
	}

	err := applyChangesets(true)
	if err == nil || !strings.Contains(err.Error(), "changelog directory missing") {
		t.Fatalf("expected changelog directory error, got %v", err)
	}
}

func TestApplyChangesets_DryRun_LeavesReviewGenLockUnchanged(t *testing.T) {
	setupTestDir(t, map[string]string{
		"core": "3.13.21",
	})

	genLockPath, ok := reviewGenLockPath("go")
	if !ok {
		t.Fatal("expected review gen.lock path for go")
	}
	before, err := os.ReadFile(genLockPath)
	if err != nil {
		t.Fatal(err)
	}

	cs := &changeset.Changeset{
		ID:          "test-001",
		Features:    []string{"core"},
		Targets:     []string{"go"},
		Type:        "fix",
		Bump:        "patch",
		Description: "fixed something",
		Author:      "testuser",
		Date:        "2026-03-10",
	}

	if _, err := changeset.Write(cs); err != nil {
		t.Fatal(err)
	}

	if err := applyChangesets(true); err != nil {
		t.Fatal(err)
	}

	after, err := os.ReadFile(genLockPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Errorf("expected review gen.lock unchanged in dry-run")
	}

	if _, err := os.Stat(filepath.Join("changelogs", "go", "core-3.13.22.md")); !os.IsNotExist(err) {
		t.Error("expected changelog to NOT be written in dry-run")
	}
}

func TestApplyChangesets_Empty(t *testing.T) {
	setupTestDir(t, map[string]string{
		"core": "3.13.21",
	})

	// No changesets to apply
	if err := applyChangesets(false); err != nil {
		t.Fatal(err)
	}

	// Version should not have changed
	ver, err := readFeatureVersion(features.FeatureCore, types.NewTargetFromTemplate("go"))
	if err != nil {
		t.Fatal(err)
	}
	if ver != "3.13.21" {
		t.Errorf("expected version 3.13.21, got %s", ver)
	}
}

func TestCreateChangelogEntry(t *testing.T) {
	cs := &changeset.Changeset{
		Type:        "feat",
		Description: "added new pagination support",
		Author:      "testuser",
		Date:        "2026-03-10",
	}

	entry := createChangelogEntry(cs, features.FeatureCore, "3.13.22")

	expected := "## core: 3.13.22 - 2026-03-10\n### :bee: New Features\n- added new pagination support *(commit by [@testuser](https://github.com/testuser))*\n"
	if entry != expected {
		t.Errorf("unexpected changelog entry.\nExpected:\n%s\nGot:\n%s", expected, entry)
	}
}

func TestCreateChangelogEntry_UnknownType(t *testing.T) {
	cs := &changeset.Changeset{
		Type:        "unknown",
		Description: "some change",
		Author:      "testuser",
		Date:        "2026-03-10",
	}

	entry := createChangelogEntry(cs, features.FeatureCore, "3.13.22")

	if !strings.Contains(entry, ":flying_saucer: Other Changes") {
		t.Errorf("expected unknown type to fall back to 'Other Changes', got: %s", entry)
	}
}
