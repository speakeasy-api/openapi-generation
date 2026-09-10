package snaptest

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"
	generationaccess "github.com/speakeasy-api/generation-context/access"
	"github.com/speakeasy-api/openapi-generation/v2/internal/env"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"
)

// Options holds the configuration for a snapshot test
type Options struct {
	Spec         string
	GenYaml      string
	IncludeGlobs []string
	// ExcludeGlobs lists path patterns that must NOT exist after generation.
	// A trailing "/**" matches everything beneath a directory (e.g. "docs/**");
	// otherwise filepath.Match semantics apply to the slash-relative path. Any
	// match fails the test. Useful for asserting a feature suppressed output —
	// e.g. `generation.documentation: none` skipping `docs/**`.
	ExcludeGlobs  []string
	Expected      string
	GenOpts       []generate.GeneratorOptions
	AfterGenerate func(t *testing.T, tempDir string)
	ExpectErrors  []string
	ShouldCompile *bool
}

type GenerateReport struct {
	MissingExpected   []string
	GenerationErrors  []error
	CompilationErrors []error
	LintErrors        []error
	IgnoredCompile    []error
	IgnoredLint       []error
}

func (r GenerateReport) HasFailures() bool {
	return len(r.MissingExpected) > 0 ||
		len(r.GenerationErrors) > 0 ||
		len(r.CompilationErrors) > 0 ||
		len(r.LintErrors) > 0
}

// DoTestSnapshot runs a snapshot test with the given options
func DoTestSnapshot(t *testing.T, opts Options) {
	t.Helper()
	var err error

	// Create a stable directory for generation using test name in a dedicated snapshot tests location
	snapshotTestsDir := filepath.Join(getTempDir(), "speakeasy-snapshot-tests")
	tempDir := filepath.Join(snapshotTestsDir, t.Name())

	t.Logf("tempDir: %s", tempDir)

	// Clear the directory contents to ensure clean state while preserving the directory structure
	if !env.SnapshotsSkipCleanDir() {
		err = clearDirectoryContents(tempDir)
		require.NoError(t, err)
	}
	err = os.MkdirAll(tempDir, 0o755)
	require.NoError(t, err)

	// Always keep the temp directory for debugging - no cleanup
	t.Logf("Generated SDK preserved at: %s", tempDir)

	// Parse user-provided genYaml to determine target and overrides
	var userConfig map[string]any
	err = yaml.Unmarshal([]byte(opts.GenYaml), &userConfig)
	require.NoError(t, err)

	// Determine target from config (look for language keys like 'go', 'typescript', 'python', etc.)
	target := "python" // default
	for key := range userConfig {
		if key != "configVersion" && key != "generation" {
			target = key
			break
		}
	}

	// Get default configuration for the target
	defaultConfig, err := config.GetDefaultConfig(true, generate.GetLanguageConfigDefaults, map[string]bool{target: true})
	require.NoError(t, err)

	// Convert default config to map for merging
	defaultBytes, err := yaml.Marshal(defaultConfig)
	require.NoError(t, err)

	var defaultMap map[string]any
	err = yaml.Unmarshal(defaultBytes, &defaultMap)
	require.NoError(t, err)

	// Deep merge user config into default config
	deepMerge(defaultMap, userConfig)

	// Write the merged configuration to gen.yaml
	genYamlPath := filepath.Join(tempDir, ".speakeasy", "gen.yaml")
	err = os.MkdirAll(filepath.Dir(genYamlPath), 0o755)
	require.NoError(t, err)

	mergedConfigBytes, err := yaml.Marshal(defaultMap)
	require.NoError(t, err)

	err = os.WriteFile(genYamlPath, mergedConfigBytes, 0o644)
	require.NoError(t, err)

	// Write the OpenAPI spec to .speakeasy/out.openapi.yaml for debugging
	specPath := filepath.Join(tempDir, ".speakeasy", "out.openapi.yaml")
	err = os.WriteFile(specPath, []byte(opts.Spec), 0o644)
	require.NoError(t, err)

	// Create a new generator with debug logging enabled and use the public Generate method for full generation with compilation
	debugLogger := logging.NewLogger(zapcore.DebugLevel)
	genOpts := []generate.GeneratorOptions{
		generate.WithDebuggingEnabled(),
		generate.WithLogger(debugLogger),
	}

	// Append any extra generator options from the test (e.g. split thresholds).
	genOpts = append(genOpts, opts.GenOpts...)

	// Isolate per-test gem/bundle state so parallel snapshot tests don't
	// race on shared system directories. For example, multiple Ruby tests
	// running "bundle install" concurrently can corrupt the shared Bundler
	// installation. The compile-env directory is a sibling of the SDK
	// output directory (not underneath it) because tools like Sorbet use
	// "--dir ." to type-check everything under the output directory —
	// placing gem sources there causes duplicate definition errors against
	// the pre-generated RBI stubs.
	compileEnvDir := tempDir + "-compile-env"
	genOpts = append(genOpts, generate.WithCompileEnv(map[string]string{
		"GEM_HOME":    filepath.Join(compileEnvDir, "gem"),
		"BUNDLE_PATH": filepath.Join(compileEnvDir, "bundle"),
	}))

	// Support WATCH_TEMPLATES=1 + WATCH_TEMPLATES_LOCATION environment variables (unified with CLI)
	if env.WatchTemplatesEnabled() {
		watchTemplatesLocation := env.WatchTemplatesLocation()
		if watchTemplatesLocation != "" {
			genOpts = append(genOpts, generate.WithWatchTemplatesLocation(watchTemplatesLocation))
		}
	}

	shouldCompile := !env.SnapshotSkipCompile()
	if opts.ShouldCompile != nil {
		shouldCompile = *opts.ShouldCompile
	}
	if !shouldCompile {
		t.Logf("Skipping compilation")
	}
	ctx := generationaccess.WithDirect(context.Background(), generationaccess.ElectAGPL())

	runGenerate := func() []error {
		generator, err := generate.New(genOpts...)
		require.NoError(t, err)
		return generator.Generate(ctx, []byte(opts.Spec), "test-schema.yaml", target, tempDir, false, shouldCompile)
	}

	firstRunExpectErrors := opts.ExpectErrors
	if opts.AfterGenerate != nil {
		// opts.ExpectErrors will be checked against the post-AfterGenerate re-run.
		firstRunExpectErrors = nil
	}
	report := analyzeGenerateErrors(runGenerate(), shouldCompile, firstRunExpectErrors)
	emitGenerateReport(t, report)

	if opts.AfterGenerate != nil {
		opts.AfterGenerate(t, tempDir)
		report = analyzeGenerateErrors(runGenerate(), shouldCompile, opts.ExpectErrors)
		emitGenerateReport(t, report)
	}

	// Read gen.yaml content to display at end of test
	var _ []byte
	if content, err := os.ReadFile(genYamlPath); err == nil {
		_ = content
	}

	// Assert that no excluded files were generated. Independent of the inline
	// snapshot comparison so a test can assert pure absence without an
	// expected-content block.
	excludeViolation := !assertNoExcludedFiles(t, tempDir, opts.ExcludeGlobs)

	// Compare with expected snapshot or update it if empty/update requested.
	// Skipped entirely for assertion-only tests (no include globs), e.g. tests
	// that only assert absence via ExcludeGlobs.
	var snapshotMismatch bool
	if len(opts.IncludeGlobs) > 0 {
		// Collect files matching the include globs
		snapshotContent, err := collectSnapshotFiles(tempDir, opts.IncludeGlobs)
		require.NoError(t, err)

		// Check if we should update snapshots
		updateSnapshots := env.UpdateSnapshots()

		if opts.Expected == "" || updateSnapshots {
			// Generate the snapshot content and update the test file
			err = updateTestSnapshot(t, snapshotContent)
			require.NoError(t, err)

			if updateSnapshots && opts.Expected != "" {
				t.Logf("Snapshot updated for test %s", t.Name())
			}
		} else {
			snapshotMismatch = !matchInlineSnapshot(t, tempDir, snapshotContent, opts.Expected)
		}
	}

	// Fail the test if there were generation errors, compilation errors, a
	// snapshot mismatch, or an excluded file was generated.
	if report.HasFailures() || snapshotMismatch || excludeViolation {
		t.FailNow()
	}
}

func analyzeGenerateErrors(errs []error, shouldCompile bool, expectedErrors []string) GenerateReport {
	report := GenerateReport{}
	// Track which errors match an expected error string
	matchedExpected := make(map[int]bool)
	if len(expectedErrors) > 0 {
		for _, expected := range expectedErrors {
			found := false
			for i, genErr := range errs {
				if matchedExpected[i] {
					continue
				}
				if strings.Contains(genErr.Error(), expected) {
					found = true
					matchedExpected[i] = true
					break
				}
			}
			if !found {
				report.MissingExpected = append(report.MissingExpected, expected)
			}
		}
	}

	for i, genErr := range errs {
		if matchedExpected[i] {
			continue
		}
		errLower := strings.ToLower(genErr.Error())
		switch {
		case strings.Contains(errLower, "lint"):
			if shouldCompile {
				report.LintErrors = append(report.LintErrors, genErr)
			} else {
				report.IgnoredLint = append(report.IgnoredLint, genErr)
			}
		case strings.Contains(errLower, "compil"):
			if shouldCompile {
				report.CompilationErrors = append(report.CompilationErrors, genErr)
			} else {
				report.IgnoredCompile = append(report.IgnoredCompile, genErr)
			}
		default:
			report.GenerationErrors = append(report.GenerationErrors, genErr)
		}
	}

	return report
}

func emitGenerateReport(t *testing.T, report GenerateReport) {
	t.Helper()

	for _, expected := range report.MissingExpected {
		t.Errorf("expected generation error containing %q", expected)
	}
	for _, genErr := range report.GenerationErrors {
		t.Errorf("Generation error: %v", genErr)
	}
	for _, genErr := range report.LintErrors {
		t.Errorf("SDK lint failed: %v", genErr)
	}
	for _, genErr := range report.CompilationErrors {
		t.Errorf("SDK compilation failed: %v", genErr)
	}
	for _, genErr := range report.IgnoredLint {
		t.Logf("Lint error ignored (SNAPSHOTS_SKIP_COMPILE=1): %v", genErr)
	}
	for _, genErr := range report.IgnoredCompile {
		t.Logf("Compilation error ignored (SNAPSHOTS_SKIP_COMPILE=1): %v", genErr)
	}
}

// assertNoExcludedFiles walks the generated output and fails the test if any
// file matches one of the exclude globs. Returns true when nothing matched
// (the success case). See Options.ExcludeGlobs for the pattern semantics.
func assertNoExcludedFiles(t *testing.T, dir string, globs []string) bool {
	t.Helper()
	if len(globs) == 0 {
		return true
	}

	ok := true
	if err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(dir, p)
		if relErr != nil {
			return relErr
		}
		rel = filepath.ToSlash(rel)
		for _, g := range globs {
			matched, matchErr := matchExcludeGlob(g, rel)
			if matchErr != nil {
				return fmt.Errorf("malformed exclude glob %q: %w", g, matchErr)
			}
			if matched {
				ok = false
				t.Errorf("expected no file matching exclude glob %q, but found: %s", g, rel)
			}
		}
		return nil
	}); err != nil {
		ok = false
		t.Errorf("walking generated output for exclude-glob assertion: %v", err)
	}
	return ok
}

// matchExcludeGlob reports whether the slash-relative path rel matches the
// exclude glob. A trailing "/**" matches the directory itself and everything
// beneath it; otherwise path.Match semantics apply, including its
// ErrBadPattern for malformed globs.
func matchExcludeGlob(glob, rel string) (bool, error) {
	if prefix, found := strings.CutSuffix(glob, "/**"); found {
		return rel == prefix || strings.HasPrefix(rel, prefix+"/"), nil
	}
	return path.Match(glob, rel)
}

// collectSnapshotFiles walks the directory and collects files matching the include globs
func collectSnapshotFiles(dir string, includeGlobs []string) (string, error) {
	var result strings.Builder
	var allMatched []string

	// For each glob pattern, find matching files
	for _, glob := range includeGlobs {
		// Create full pattern path by joining with directory
		fullPattern := filepath.Join(dir, glob)

		// Use filepath.Glob to find matching files
		matches, err := filepath.Glob(fullPattern)
		if err != nil {
			return "", fmt.Errorf("invalid glob pattern %s: %w", glob, err)
		}

		// Convert back to relative paths
		for _, match := range matches {
			relPath, err := filepath.Rel(dir, match)
			if err != nil {
				return "", err
			}
			allMatched = append(allMatched, relPath)
		}
	}

	// Remove duplicates and sort
	fileSet := make(map[string]bool)
	for _, file := range allMatched {
		fileSet[file] = true
	}

	files := make([]string, 0, len(fileSet))
	for file := range fileSet {
		files = append(files, file)
	}

	// Sort files for consistent output
	sort.Strings(files)

	// Read and concatenate file contents
	for _, file := range files {
		fullPath := filepath.Join(dir, file)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return "", fmt.Errorf("failed to read file %s: %w", file, err)
		}

		fmt.Fprintf(&result, "--- %s ---\n", file)
		result.Write(content)
		result.WriteString("\n\n")
	}

	return result.String(), nil
}

// generateUnifiedDiff creates a unified diff between expected and actual content
func generateUnifiedDiff(expected, actual string) string {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(expected, actual, false)

	// Clean up the diffs for better readability
	diffs = dmp.DiffCleanupSemantic(diffs)

	var result strings.Builder
	result.WriteString("--- expected\n")
	result.WriteString("+++ actual\n")

	for _, diff := range diffs {
		switch diff.Type {
		case diffmatchpatch.DiffEqual:
			// Context lines - show some but not all
			lines := strings.Split(diff.Text, "\n")
			for i, line := range lines {
				if i < len(lines)-1 || line != "" { // Skip empty last line from split
					fmt.Fprintf(&result, " %s\n", line)
				}
			}
		case diffmatchpatch.DiffDelete:
			// Deleted lines
			lines := strings.Split(diff.Text, "\n")
			for i, line := range lines {
				if i < len(lines)-1 || line != "" { // Skip empty last line from split
					fmt.Fprintf(&result, "-%s\n", line)
				}
			}
		case diffmatchpatch.DiffInsert:
			// Added lines
			lines := strings.Split(diff.Text, "\n")
			for i, line := range lines {
				if i < len(lines)-1 || line != "" { // Skip empty last line from split
					fmt.Fprintf(&result, "+%s\n", line)
				}
			}
		}
	}

	return result.String()
}

// matchInlineSnapshot compares actual content with expected snapshot
func matchInlineSnapshot(t *testing.T, tempDir, actual, expected string) bool {
	t.Helper()
	if expected == "" {
		t.Logf("No expected snapshot provided. Actual content:\n%s", actual)
		return false
	}

	if actual != expected {
		// Generate diff
		diffContent := generateUnifiedDiff(expected, actual)

		// Create debug directory and save diff file
		debugDir := filepath.Join(tempDir, "__debug__")
		err := os.MkdirAll(debugDir, 0o755)
		if err == nil {
			diffPath := filepath.Join(debugDir, "snapshot.diff")
			err = os.WriteFile(diffPath, []byte(diffContent), 0o644)
			if err == nil {
				t.Logf("Diff saved to: %s", diffPath)
			}
		}

		t.Errorf("Snapshot mismatch for test %s.\n\nTo update snapshots, run:\n  UPDATE_SNAPS=1 go test -run %s\n\nDiff:\n%s",
			t.Name(), t.Name(), diffContent)
		return false
	}
	return true
}

// updateTestSnapshot updates the test file with the generated snapshot content
func updateTestSnapshot(t *testing.T, snapshotContent string) error {
	t.Helper()

	// Find the test file by examining the call stack
	testFileName := getTestFileName(t)
	if testFileName == "" {
		return errors.New("could not determine test file name")
	}

	// Read the current test file
	content, err := os.ReadFile(testFileName)
	if err != nil {
		return fmt.Errorf("failed to read test file %s: %w", testFileName, err)
	}

	fileContent := string(content)

	// Find the expectedSnapshot variable assignment scoped to the current test function.
	// This ensures we update the right snapshot when multiple tests exist in one file.
	expectedSnapshotPattern := "expectedSnapshot := `"
	searchStart := 0
	// Use the top-level test function name (strip subtest suffixes like "/bar")
	testName := t.Name()
	if idx := strings.Index(testName, "/"); idx != -1 {
		testName = testName[:idx]
	}
	funcSignature := "func " + testName + "("
	funcIndex := strings.Index(fileContent, funcSignature)
	if funcIndex != -1 {
		searchStart = funcIndex
	}
	startIndex := strings.Index(fileContent[searchStart:], expectedSnapshotPattern)
	if startIndex == -1 {
		return fmt.Errorf("expectedSnapshot variable not found in test file %s", testFileName)
	}
	startIndex += searchStart

	// Find the start of the content (after the opening backtick)
	contentStart := startIndex + len(expectedSnapshotPattern)

	// Look for comment marker first (more reliable)
	commentMarker := "` // end of snapshot"
	markerIndex := strings.Index(fileContent[contentStart:], commentMarker)
	var contentEnd int

	if markerIndex != -1 {
		// Use comment marker approach for more reliable replacement
		contentEnd = contentStart + markerIndex
	} else {
		// Fallback to original approach for backwards compatibility
		backticksNewline := "`\n"
		backticksIndex := strings.Index(fileContent[contentStart:], backticksNewline)
		if backticksIndex == -1 {
			return fmt.Errorf("could not find end of expectedSnapshot variable in test file %s", testFileName)
		}
		contentEnd = contentStart + backticksIndex
	}

	// Escape any backticks in the content to avoid syntax errors
	escapedContent := strings.ReplaceAll(snapshotContent, "`", "` + \"`\" + `")

	// Replace the content between the backticks, adding comment marker
	var newContent string
	if markerIndex != -1 {
		// Replace content preserving comment marker
		newContent = fileContent[:contentStart] + escapedContent + fileContent[contentEnd:]
	} else {
		// Add comment marker for future reliability
		newContent = fileContent[:contentStart] + escapedContent + "` // end of snapshot" + fileContent[contentEnd+1:]
	}

	// Write the updated content back to the file
	err = os.WriteFile(testFileName, []byte(newContent), 0o644)
	if err != nil {
		return fmt.Errorf("failed to write updated test file %s: %w", testFileName, err)
	}

	return nil
}

// getTestFileName dynamically finds the test file that called DoTestSnapshot
func getTestFileName(t *testing.T) string {
	t.Helper()
	// Get the current working directory
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}

	// Get the test function name
	testName := t.Name()

	// Find all *_test.go files in the current directory
	testFiles, err := filepath.Glob(filepath.Join(wd, "*_test.go"))
	if err != nil {
		return ""
	}

	// Search each test file for the test function that contains DoTestSnapshot
	for _, testFile := range testFiles {
		content, err := os.ReadFile(testFile)
		if err != nil {
			continue
		}

		fileContent := string(content)

		// Check if this file contains the test function and a DoTestSnapshot/doTestSnapshot call
		hasTestFunc := strings.Contains(fileContent, "func "+testName+"(")
		hasSnapshotCall := strings.Contains(fileContent, "DoTestSnapshot(") || strings.Contains(fileContent, "doTestSnapshot(")
		if hasTestFunc && hasSnapshotCall {
			return testFile
		}
	}

	return ""
}

// deepMerge recursively merges src into dst
func deepMerge(dst, src map[string]any) {
	for key, srcVal := range src {
		if dstVal, exists := dst[key]; exists {
			// If both are maps, merge recursively
			if dstMap, dstOk := dstVal.(map[string]any); dstOk {
				if srcMap, srcOk := srcVal.(map[string]any); srcOk {
					deepMerge(dstMap, srcMap)
					continue
				}
			}
		}
		// Otherwise, overwrite the value
		dst[key] = srcVal
	}
}

// clearDirectoryContents removes all files and subdirectories within a directory
// while preserving the directory itself
func clearDirectoryContents(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		entryPath := filepath.Join(dir, entry.Name())
		if err := os.RemoveAll(entryPath); err != nil {
			return err
		}
	}

	return nil
}

// getTempDir returns the appropriate temporary directory for the current platform
func getTempDir() string {
	if runtime.GOOS == "windows" {
		// On Windows, use the TEMP or TMP environment variable, or fallback to C:\Temp
		if temp := os.Getenv("TEMP"); temp != "" {
			return temp
		}
		if temp := os.Getenv("TMP"); temp != "" {
			return temp
		}
		return `C:\Temp`
	}
	// On Unix-like systems, use /tmp
	return "/tmp"
}
