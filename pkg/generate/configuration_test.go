package generate_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/speakeasy-api/sdk-gen-config/workflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckMCPTargetNameSupported_Success(t *testing.T) {
	assert.True(t, generate.CheckMCPTargetNameSupported("mcp-typescript"))
	assert.False(t, generate.CheckMCPTargetNameSupported("common"))
	assert.False(t, generate.CheckMCPTargetNameSupported("go"))
	assert.False(t, generate.CheckMCPTargetNameSupported("java"))
	assert.False(t, generate.CheckMCPTargetNameSupported("mockserver"))
	assert.False(t, generate.CheckMCPTargetNameSupported("python"))
	assert.False(t, generate.CheckMCPTargetNameSupported("terraform"))
	assert.False(t, generate.CheckMCPTargetNameSupported("typescript"))
	assert.False(t, generate.CheckMCPTargetNameSupported("typescriptv2"))
}

func TestCheckSDKTargetNameSupported_Success(t *testing.T) {
	assert.True(t, generate.CheckSDKTargetNameSupported("cli"))
	assert.True(t, generate.CheckSDKTargetNameSupported("csharp"))
	assert.True(t, generate.CheckSDKTargetNameSupported("go"))
	assert.True(t, generate.CheckSDKTargetNameSupported("java"))
	assert.True(t, generate.CheckSDKTargetNameSupported("php"))
	assert.True(t, generate.CheckSDKTargetNameSupported("python"))
	assert.True(t, generate.CheckSDKTargetNameSupported("ruby"))
	assert.True(t, generate.CheckSDKTargetNameSupported("typescript"))
	assert.False(t, generate.CheckSDKTargetNameSupported("common"))
	assert.False(t, generate.CheckSDKTargetNameSupported("mcp-typescript"))
	assert.False(t, generate.CheckSDKTargetNameSupported("mockserver"))
	assert.False(t, generate.CheckSDKTargetNameSupported("terraform"))
	assert.False(t, generate.CheckSDKTargetNameSupported("typescriptv2"))
}

func TestGetSupportedMCPTargetNames_Success(t *testing.T) {
	langs := generate.GetSupportedMCPTargetNames()

	assert.Equal(t, []string{"mcp-typescript"}, langs)
}

func TestGetSupportedSDKTargetNames_Success(t *testing.T) {
	langs := generate.GetSupportedSDKTargetNames()

	assert.Equal(t, []string{"cli", "csharp", "go", "java", "php", "postman", "python", "ruby", "typescript", "unity"}, langs)
}

func TestGetSupportedMCPTargets_DEBUG_Success(t *testing.T) {
	t.Setenv("SPEAKEASY_DEBUG", "true")
	targets := generate.GetSupportedMCPTargets()

	assert.Equal(t, []types.Target{
		{Target: "mcp-typescript", Template: "mcp-typescript", Maturity: "Beta"},
	}, targets)
}

func TestGetSupportedMCPTargets_Success(t *testing.T) {
	t.Setenv("SPEAKEASY_DEBUG", "false")
	targets := generate.GetSupportedMCPTargets()

	assert.Equal(t, []types.Target{
		{Target: "mcp-typescript", Template: "mcp-typescript", Maturity: "Beta"},
	}, targets)
}

func TestGetSupportedSDKTargets_DEBUG_Success(t *testing.T) {
	t.Setenv("SPEAKEASY_DEBUG", "true")
	targets := generate.GetSupportedSDKTargets()

	assert.Equal(t, []types.Target{
		{Target: "cli", Template: "cli", Maturity: "Beta"},
		{Target: "csharp", Template: "csharp", Maturity: "GA"},
		{Target: "go", Template: "go", Maturity: "GA"},
		{Target: "java", Template: "javav2", Maturity: "GA"},
		{Target: "mockserver", Template: "mockserver", Maturity: "Alpha"},
		{Target: "php", Template: "php", Maturity: "GA"},
		{Target: "postman", Template: "postman", Maturity: "Alpha"},
		{Target: "python", Template: "pythonv2", Maturity: "GA"},
		{Target: "ruby", Template: "ruby", Maturity: "GA"},
		{Target: "typescript", Template: "typescriptv2", Maturity: "GA"},
		{Target: "unity", Template: "unity", Maturity: "Beta"},
	}, targets)
}

func TestGetSupportedSDKTargets_Success(t *testing.T) {
	t.Setenv("SPEAKEASY_DEBUG", "false")
	targets := generate.GetSupportedSDKTargets()

	assert.Equal(t, []types.Target{
		{Target: "cli", Template: "cli", Maturity: "Beta"},
		{Target: "csharp", Template: "csharp", Maturity: "GA"},
		{Target: "go", Template: "go", Maturity: "GA"},
		{Target: "java", Template: "java", Maturity: "GA"},
		{Target: "php", Template: "php", Maturity: "GA"},
		{Target: "postman", Template: "postman", Maturity: "Alpha"},
		{Target: "python", Template: "python", Maturity: "GA"},
		{Target: "ruby", Template: "ruby", Maturity: "GA"},
		{Target: "typescript", Template: "typescript", Maturity: "GA"},
		{Target: "unity", Template: "unity", Maturity: "Beta"},
	}, targets)
}

func TestGetConfigUsageFromType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		s        any
		parent   string
		expected map[string]any
	}{
		"nil": {
			s:        nil,
			expected: map[string]any{},
		},
		"empty": {
			s:        map[string]any{},
			expected: map[string]any{},
		},
		"parent": {
			s: map[string]any{
				"test": "test-value",
			},
			parent: "parent",
			expected: map[string]any{
				"parent_test": "test-value",
			},
		},
		"value": {
			s: map[string]any{
				"test": "test-value",
			},
			expected: map[string]any{
				"test": "test-value",
			},
		},
		"value-nested-array": {
			s: map[string]any{
				"testarray": []string{
					"test-value1",
					"test-value2",
				},
			},
			expected: map[string]any{
				"testarray_0": "test-value1",
				"testarray_1": "test-value2",
			},
		},
		"value-nested-map": {
			s: map[string]any{
				"testmap": map[string]any{
					"testkey1": "test-value1",
					"testkey2": "test-value2",
				},
			},
			expected: map[string]any{
				"testmap_testkey1": "test-value1",
				"testmap_testkey2": "test-value2",
			},
		},
		"value-nil": {
			s: map[string]any{
				"other": "other-value",
				"test":  nil, // intentionally nil value
			},
			parent: "parent",
			expected: map[string]any{
				"parent_other": "other-value",
			},
		},
	}
	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()
			got := generate.GetConfigUsageFromType(testCase.s, testCase.parent)
			assert.Equal(t, testCase.expected, got)
		})
	}
}

func TestConfigurationStability(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(rootDir, ".speakeasy"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(rootDir, ".speakeasy", "workflow.yaml"), []byte(`workflowVersion: 1.0.0
sources:
  widgets:
    inputs:
      - location: ./openapi.yaml
targets:
  python-widgets:
    target: python
    source: widgets
    output: dist/sp_python
    testing:
      enabled: false
  typescript-widgets:
    target: typescript
    source: widgets
    output: dist/sp_typescript
    testing:
      enabled: false
`), 0o644))

	wf, _, err := workflow.Load(rootDir)
	require.NoError(t, err)
	require.NotNil(t, wf.Targets["python-widgets"].Testing)
	require.NotNil(t, wf.Targets["typescript-widgets"].Testing)
	require.NotNil(t, wf.Targets["python-widgets"].Testing.Enabled)
	require.NotNil(t, wf.Targets["typescript-widgets"].Testing.Enabled)
	assert.False(t, *wf.Targets["python-widgets"].Testing.Enabled)
	assert.False(t, *wf.Targets["typescript-widgets"].Testing.Enabled)

	tests := []struct {
		name   string
		target string
		outDir string
	}{
		{
			name:   "python target workflow testing disabled",
			target: "python",
			outDir: "dist/sp_python",
		},
		{
			name:   "typescript target workflow testing disabled",
			target: "typescript",
			outDir: "dist/sp_typescript",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			outDir := filepath.Join(rootDir, tc.outDir)
			writeGenConfigWithTestsDisabled(t, outDir, tc.target)

			firstLoad := loadGenConfigForTarget(t, outDir, tc.target)
			assert.False(t, firstLoad.Config.Generation.Tests.GenerateTests)
			assert.False(t, firstLoad.Config.Generation.Tests.GenerateNewTests)
			assert.False(t, firstLoad.Config.Generation.Tests.SkipResponseBodyAssertions)

			firstCanonicalizedGenYAML := readGenYAML(t, outDir)
			require.NotContains(t, firstCanonicalizedGenYAML, "generateTests: true", "first pass must not rewrite explicit disabled tests to enabled")
			require.Contains(t, firstCanonicalizedGenYAML, "generateTests: false")

			secondLoad := loadGenConfigForTarget(t, outDir, tc.target)
			assert.False(t, secondLoad.Config.Generation.Tests.GenerateTests, "disabled generated tests config must remain stable across repeated loads")
			assert.False(t, secondLoad.Config.Generation.Tests.GenerateNewTests)
			assert.False(t, secondLoad.Config.Generation.Tests.SkipResponseBodyAssertions)

			secondCanonicalizedGenYAML := readGenYAML(t, outDir)
			assert.NotContains(t, secondCanonicalizedGenYAML, "generateTests: true", "second pass must not reintroduce generateTests: true")
		})
	}
}

func writeGenConfigWithTestsDisabled(t *testing.T, outDir, target string) {
	t.Helper()

	speakeasyDir := filepath.Join(outDir, ".speakeasy")
	require.NoError(t, os.MkdirAll(speakeasyDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(speakeasyDir, "gen.yaml"), []byte(fmt.Sprintf(`configVersion: %s
generation:
  mockServer:
    disabled: true
  tests:
    generateTests: false
    generateNewTests: false
    skipResponseBodyAssertions: false
%s:
  version: 0.0.1
`, config.Version, target)), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(speakeasyDir, "gen.lock"), []byte(fmt.Sprintf(`lockVersion: 2.0.0
id: gen-3005
management: {}
features:
  %s:
    core: 0.0.1
`, target)), 0o644))
}

func loadGenConfigForTarget(t *testing.T, outDir, target string) *config.Config {
	t.Helper()

	cfg, err := config.Load(
		outDir,
		config.WithLanguages(target),
		config.WithLanguageDefaultFunc(generate.GetLanguageConfigDefaults),
		config.WithUpgradeFunc(func(_, _, _, _ string, cfg map[string]any) (map[string]any, error) {
			return cfg, nil
		}),
	)
	require.NoError(t, err)
	return cfg
}

func readGenYAML(t *testing.T, outDir string) string {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(outDir, ".speakeasy", "gen.yaml"))
	require.NoError(t, err)
	return strings.ReplaceAll(string(data), "\r\n", "\n")
}
