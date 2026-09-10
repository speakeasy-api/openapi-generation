package changelogs_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/templates"
	"github.com/stretchr/testify/require"

	"github.com/speakeasy-api/openapi-generation/v2/changelogs"
	"github.com/stretchr/testify/assert"
)

func TestCompileChangelog(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		lang        string
		expectError bool
	}{
		{
			name: "csharp",
			lang: "csharp",
		},
		{
			name: "go",
			lang: "go",
		},
		{
			name: "java", // only javav2/ exists
			lang: "java",
		},
		{
			name: "javav2",
			lang: "javav2",
		},
		{
			name: "mcp-typescript",
			lang: "mcp-typescript",
		},
		{
			name: "mockserver",
			lang: "mockserver",
		},
		{
			name: "php",
			lang: "php",
		},
		{
			name: "postman",
			lang: "postman",
		},
		{
			name: "python", // only pythonv2/ exists
			lang: "python",
		},
		{
			name: "pythonv2",
			lang: "pythonv2",
		},
		{
			name: "ruby",
			lang: "ruby",
		},
		{
			name: "terraform",
			lang: "terraform",
		},
		{
			name: "typescript", // both typescript/ and typescriptv2/ exist
			lang: "typescript",
		},
		{
			name: "typescriptv2",
			lang: "typescriptv2",
		},
		{
			name: "unity",
			lang: "unity",
		},
		{
			name:        "nonexistent language returns error",
			lang:        "nonexistent",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			changelog, err := changelogs.CompileChangelog(tc.lang)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err, "CompileChangelog should not error for %s", tc.lang)
			require.NotEmpty(t, changelog, "changelog should not be empty for %s", tc.lang)

			// Verify changelog contains properly formatted entries
			sections := changelogs.SplitRegex.Split(changelog, -1)
			entryCount := 0

			for _, section := range sections {
				if len(section) == 0 {
					continue
				}

				section = "## " + section
				entryCount++

				assert.True(t, changelogs.HeaderRegex.MatchString(section),
					"section should match header format: %s", section[:min(100, len(section))])

				assert.True(t, changelogs.VersionMatchRegex.MatchString(section),
					"section should match version format: %s", section[:min(100, len(section))])
			}

			assert.GreaterOrEqual(t, entryCount, 1,
				"expected at least 1 entry for %s, got %d", tc.lang, entryCount)

			assert.Contains(t, changelog, "core: ",
				"changelog should contain core feature entries")
		})
	}
}

func TestCompileChangelog_AllAvailableTemplates(t *testing.T) {
	t.Parallel()

	allLanguages := templates.GetAvailableTemplates()

	for _, templateVersion := range allLanguages {
		t.Run(templateVersion, func(t *testing.T) {
			t.Parallel()

			_, err := changelogs.CompileChangelog(templateVersion)
			assert.NoError(t, err)
		})
	}
}

// The way this currently works, is the first non-hidden template determines
// the "versions" of the features. I.e.
//
//	"typescript" refers to "typescriptv2", and the changelog is created by comparing the SDK features against the
//	"python" refers to "python".
func TestGetChangelog_templates(t *testing.T) {
	t.Parallel()

	allLanguages := templates.GetAvailableTemplates()

	for _, language := range allLanguages {
		t.Run(language, func(t *testing.T) {
			t.Parallel()

			versions, err := changelogs.GetLatestVersions(language)
			require.NoError(t, err)
			changelog, err := changelogs.GetChangeLog(language, versions, map[string]string{})
			require.NoError(t, err)
			require.NotEmpty(t, changelog)
		})
	}
}

func TestGetChangelog_featureVersionsMatch(t *testing.T) {
	t.Parallel()

	allLanguages := templates.GetAvailableTemplates()

	for _, language := range allLanguages {
		t.Run(language, func(t *testing.T) {
			t.Parallel()

			changelog, err := changelogs.CompileChangelog(language)
			require.NoError(t, err)

			sections := changelogs.SplitRegex.Split(changelog, -1)
			for _, section := range sections {
				if len(section) == 0 {
					continue
				}
				section = "## " + section

				t.Run(section, func(t *testing.T) {
					require.True(t, changelogs.VersionMatchRegex.MatchString(section), "%s should match %s", section, changelogs.VersionMatchRegex)
				})
			}
		})
	}
}

func TestGetTemplateChangelog(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		templateVersion   string
		previousVersions  map[string]string
		targetVersions    map[string]string
		expectedChangelog string
	}{
		"both nil": {
			templateVersion:   "typescriptv2",
			previousVersions:  nil,
			targetVersions:    nil,
			expectedChangelog: ``,
		},
		"both empty": {
			templateVersion:   "typescriptv2",
			previousVersions:  map[string]string{},
			targetVersions:    map[string]string{},
			expectedChangelog: ``,
		},
		"invalid version bump - empty previous version": {
			templateVersion: "typescriptv2",
			previousVersions: map[string]string{
				"sdkHooks": "",
			},
			targetVersions: map[string]string{
				"sdkHooks": "0.3.0",
			},
			expectedChangelog: testMustReadFile(t, filepath.Join("typescriptv2", "sdkHooks-0.3.0.md")) +
				"\n",
		},
		"invalid version bump - 0.0.0 previous version": {
			templateVersion: "typescriptv2",
			previousVersions: map[string]string{
				"sdkHooks": "0.0.0",
			},
			targetVersions: map[string]string{
				"sdkHooks": "0.3.0",
			},
			expectedChangelog: testMustReadFile(t, filepath.Join("typescriptv2", "sdkHooks-0.3.0.md")) +
				"\n\n\n" +
				testMustReadFile(t, filepath.Join("typescriptv2", "sdkHooks-0.2.0.md")) +
				"\n",
		},
		"invalid version bump - empty target version": {
			templateVersion: "typescriptv2",
			previousVersions: map[string]string{
				"core": "3.21.10",
			},
			targetVersions: map[string]string{
				"core": "",
			},
			expectedChangelog: ``,
		},
		"invalid version bump - 0.0.0 target version": {
			templateVersion: "typescriptv2",
			previousVersions: map[string]string{
				"core": "3.21.10",
			},
			targetVersions: map[string]string{
				"core": "0.0.0",
			},
			expectedChangelog: ``,
		},
		"invalid version bump - backwards version": {
			templateVersion: "typescriptv2",
			previousVersions: map[string]string{
				"core": "3.21.10",
			},
			targetVersions: map[string]string{
				"core": "3.21.9",
			},
			expectedChangelog: ``,
		},
		"single feature no bump": {
			templateVersion: "typescriptv2",
			previousVersions: map[string]string{
				"core": "3.21.9",
			},
			targetVersions: map[string]string{
				"core": "3.21.9",
			},
			expectedChangelog: ``,
		},
		"single feature bump": {
			templateVersion: "typescriptv2",
			previousVersions: map[string]string{
				"core": "3.21.9",
			},
			targetVersions: map[string]string{
				"core": "3.21.10",
			},
			expectedChangelog: testMustReadFile(t, filepath.Join("typescriptv2", "core-3.21.10.md")) +
				"\n",
		},
		"multiple feature no bump": {
			templateVersion: "typescriptv2",
			previousVersions: map[string]string{
				"core":     "3.21.9",
				"sdkHooks": "0.2.0",
			},
			targetVersions: map[string]string{
				"core":     "3.21.9",
				"sdkHooks": "0.2.0",
			},
			expectedChangelog: ``,
		},
		"multiple feature bump": {
			templateVersion: "typescriptv2",
			previousVersions: map[string]string{
				"core":     "3.21.9",
				"sdkHooks": "0.2.0",
			},
			targetVersions: map[string]string{
				"core":     "3.21.10",
				"sdkHooks": "0.3.0",
			},
			expectedChangelog: testMustReadFile(t, filepath.Join("typescriptv2", "core-3.21.10.md")) +
				"\n\n\n" +
				testMustReadFile(t, filepath.Join("typescriptv2", "sdkHooks-0.3.0.md")) +
				"\n",
		},
		"new feature - previous versions nil": {
			templateVersion:  "typescriptv2",
			previousVersions: nil,
			targetVersions: map[string]string{
				"sdkHooks": "0.3.0",
			},
			expectedChangelog: testMustReadFile(t, filepath.Join("typescriptv2", "sdkHooks-0.3.0.md")) +
				"\n",
		},
		"new feature - previous versions empty": {
			templateVersion:  "typescriptv2",
			previousVersions: map[string]string{},
			targetVersions: map[string]string{
				"sdkHooks": "0.3.0",
			},
			expectedChangelog: testMustReadFile(t, filepath.Join("typescriptv2", "sdkHooks-0.3.0.md")) +
				"\n",
		},
		"new features - previous versions nil": {
			templateVersion:  "typescriptv2",
			previousVersions: nil,
			targetVersions: map[string]string{
				"deepObjectParams": "0.1.0",
				"sdkHooks":         "0.3.0",
			},
			expectedChangelog: testMustReadFile(t, filepath.Join("typescriptv2", "deepObjectParams-0.1.0.md")) +
				"\n\n\n" +
				testMustReadFile(t, filepath.Join("typescriptv2", "sdkHooks-0.3.0.md")) +
				"\n",
		},
		"new features - previous versions empty": {
			templateVersion:  "typescriptv2",
			previousVersions: map[string]string{},
			targetVersions: map[string]string{
				"deepObjectParams": "0.1.0",
				"sdkHooks":         "0.3.0",
			},
			expectedChangelog: testMustReadFile(t, filepath.Join("typescriptv2", "deepObjectParams-0.1.0.md")) +
				"\n\n\n" +
				testMustReadFile(t, filepath.Join("typescriptv2", "sdkHooks-0.3.0.md")) +
				"\n",
		},
		"removed feature - target versions nil": {
			templateVersion: "typescriptv2",
			previousVersions: map[string]string{
				"sdkHooks": "0.3.0",
			},
			targetVersions:    nil,
			expectedChangelog: ``,
		},
		"removed feature - target versions empty": {
			templateVersion: "typescriptv2",
			previousVersions: map[string]string{
				"sdkHooks": "0.3.0",
			},
			targetVersions:    map[string]string{},
			expectedChangelog: ``,
		},
		"removed features - target versions empty": {
			templateVersion: "typescriptv2",
			previousVersions: map[string]string{
				"deepObjectParams": "0.1.0",
				"sdkHooks":         "0.3.0",
			},
			targetVersions:    map[string]string{},
			expectedChangelog: ``,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := changelogs.GetTemplateChangeLog(testCase.templateVersion, testCase.targetVersions, testCase.previousVersions)
			require.NoError(t, err)
			assert.Equal(t, testCase.expectedChangelog, got)
		})
	}
}

func testMustReadFile(t *testing.T, filePath string) string {
	t.Helper()

	content, err := os.ReadFile(filePath)
	require.NoError(t, err, "failed to read file %s", filePath)

	return string(content)
}

func TestChangelogEntriesBetweenVersions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		lang        string
		feature     string
		newVersion  string
		oldVersion  string
		expected    []string
		expectError bool
	}{
		{
			name:       "go core single version",
			lang:       "go",
			feature:    "core",
			newVersion: "3.1.0",
			oldVersion: "3.1.0",
			expected:   []string{"core-3.1.0.md"},
		},
		{
			name:       "go core patch version range",
			lang:       "go",
			feature:    "core",
			newVersion: "3.1.2",
			oldVersion: "3.1.0",
			expected:   []string{"core-3.1.0.md", "core-3.1.1.md", "core-3.1.2.md"},
		},
		{
			name:       "go core minor version range",
			lang:       "go",
			feature:    "core",
			newVersion: "3.2.0",
			oldVersion: "3.1.5",
			expected:   []string{"core-3.1.5.md", "core-3.1.6.md", "core-3.2.0.md"},
		},
		{
			name:       "go core no matching versions",
			lang:       "go",
			feature:    "core",
			newVersion: "1.0.0",
			oldVersion: "0.9.0",
			expected:   []string{},
		},
		{
			name:       "go nonexistent feature",
			lang:       "go",
			feature:    "nonexistent",
			newVersion: "1.0.0",
			oldVersion: "0.9.0",
			expected:   []string{},
		},
		{
			name:        "nonexistent language directory returns error",
			lang:        "nonexistent",
			feature:     "core",
			newVersion:  "1.0.0",
			oldVersion:  "0.9.0",
			expectError: true,
		},
		{
			name:       "python uses pythonv2 directory",
			lang:       "python",
			feature:    "core",
			newVersion: "4.8.1",
			oldVersion: "4.8.0",
			expected:   []string{"core-4.8.0.md", "core-4.8.1.md"},
		},
		{
			name:       "java uses javav2 directory",
			lang:       "java",
			feature:    "core",
			newVersion: "3.26.2",
			oldVersion: "3.26.1",
			expected:   []string{"core-3.26.1.md", "core-3.26.2.md"},
		},
		{
			name:       "pythonv2 directory exists",
			lang:       "pythonv2",
			feature:    "core",
			newVersion: "4.8.1",
			oldVersion: "4.8.0",
			expected:   []string{"core-4.8.0.md", "core-4.8.1.md"},
		},
		{
			name:       "javav2 directory exists",
			lang:       "javav2",
			feature:    "core",
			newVersion: "3.26.2",
			oldVersion: "3.26.1",
			expected:   []string{"core-3.26.1.md", "core-3.26.2.md"},
		},
		{
			name:        "invalid new version",
			lang:        "go",
			feature:     "core",
			newVersion:  "invalid",
			oldVersion:  "3.1.0",
			expectError: true,
		},
		{
			name:        "invalid old version",
			lang:        "go",
			feature:     "core",
			newVersion:  "3.1.0",
			oldVersion:  "invalid",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := changelogs.ChangelogEntriesBetweenVersions(tc.lang, tc.feature, tc.newVersion, tc.oldVersion)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.ElementsMatch(t, tc.expected, result)
		})
	}
}
