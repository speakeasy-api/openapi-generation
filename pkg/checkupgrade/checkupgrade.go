// Package checkupgrade provides functionality to compare a gen.yaml configuration
// against the defaults for new SDKs, identifying which settings differ from
// the recommended defaults.
package checkupgrade

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/templates"
	config "github.com/speakeasy-api/sdk-gen-config"
	"gopkg.in/yaml.v3"
)

// ConfigDifference represents a single configuration value that differs from the newSDK default
type ConfigDifference struct {
	Key          string
	CurrentValue any
	NewSDKValue  any
	Description  string
}

// SectionResult contains the comparison results for a configuration section
type SectionResult struct {
	Name        string
	Differences []ConfigDifference
	Matches     []ConfigDifference // Only populated when IncludeMatches is true
}

// Result contains the full comparison results
type Result struct {
	GenYamlPath string
	Generation  *SectionResult
	Languages   map[string]*SectionResult
}

// Options configures the check-upgrade behavior
type Options struct {
	IncludeMatches bool // Include matching values in results
}

// Check analyzes a gen.yaml file and compares it against newSDK defaults
func Check(repoPath string, opts Options) (*Result, error) {
	genYamlPath := FindGenYaml(repoPath)
	if genYamlPath == "" {
		return nil, fmt.Errorf("could not find gen.yaml in %s (looked in: .speakeasy/gen.yaml, gen.yaml)", repoPath)
	}

	data, err := os.ReadFile(genYamlPath)
	if err != nil {
		return nil, fmt.Errorf("reading gen.yaml: %w", err)
	}

	var genConfig map[string]any
	if err := yaml.Unmarshal(data, &genConfig); err != nil {
		return nil, fmt.Errorf("parsing gen.yaml: %w", err)
	}

	result := &Result{
		GenYamlPath: genYamlPath,
		Languages:   make(map[string]*SectionResult),
	}

	// Check generation section
	if generation, ok := genConfig["generation"].(map[string]any); ok {
		generationDefaults := generate.GetGenerationConfigFields(true)
		result.Generation = checkGenerationSection(generation, generationDefaults, opts.IncludeMatches)
	}

	// Find and check language targets
	targets := FindTargets(genConfig)
	for _, target := range targets {
		langConfig, ok := genConfig[target].(map[string]any)
		if !ok {
			continue
		}

		langDefaults, err := generate.GetLanguageConfigDefaults(target, true)
		if err != nil {
			continue
		}

		// Get field definitions for descriptions
		t, err := templates.GetTargetFromTargetString(target)
		var langFields []config.SDKGenConfigField
		if err == nil {
			langFields, _ = templates.GetLanguageConfigFields(t, true)
		}

		result.Languages[target] = checkLanguageSection(target, langConfig, langDefaults, langFields, opts.IncludeMatches)
	}

	return result, nil
}

// FindGenYaml locates the gen.yaml file in a repository
func FindGenYaml(repoPath string) string {
	candidates := []string{
		filepath.Join(repoPath, ".speakeasy", "gen.yaml"),
		filepath.Join(repoPath, "gen.yaml"),
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// FindTargets returns the language targets configured in a gen.yaml
func FindTargets(genConfig map[string]any) []string {
	supportedTargets := templates.AllSupportedTemplateNames()
	var targets []string

	for key := range genConfig {
		for _, supported := range supportedTargets {
			if key == supported {
				targets = append(targets, key)
				break
			}
		}
	}

	sort.Strings(targets)
	return targets
}

// shouldExcludeKey returns true if the key should be excluded from comparison
func shouldExcludeKey(key string) bool {
	return strings.HasSuffix(key, "Name") || key == "version"
}

func checkGenerationSection(actual map[string]any, defaults []config.SDKGenConfigField, includeMatches bool) *SectionResult {
	defaultsMap := make(map[string]any)
	descriptionMap := make(map[string]string)
	for _, field := range defaults {
		if field.DefaultValue != nil {
			defaultsMap[field.Name] = *field.DefaultValue
		}
		if field.Description != nil {
			descriptionMap[field.Name] = *field.Description
		}
	}

	flatActual := flattenMap(actual, "")

	result := &SectionResult{Name: "generation"}

	for key, defaultVal := range defaultsMap {
		if shouldExcludeKey(key) {
			continue
		}

		actualVal, exists := flatActual[key]
		if !exists {
			if includeMatches {
				result.Matches = append(result.Matches, ConfigDifference{
					Key:         key,
					NewSDKValue: defaultVal,
					Description: descriptionMap[key],
				})
			}
			continue
		}

		if !valuesEqual(actualVal, defaultVal) {
			result.Differences = append(result.Differences, ConfigDifference{
				Key:          key,
				CurrentValue: actualVal,
				NewSDKValue:  defaultVal,
				Description:  descriptionMap[key],
			})
		} else if includeMatches {
			result.Matches = append(result.Matches, ConfigDifference{
				Key:          key,
				CurrentValue: actualVal,
				NewSDKValue:  defaultVal,
				Description:  descriptionMap[key],
			})
		}
	}

	sortDifferences(result.Differences)
	sortDifferences(result.Matches)

	return result
}

func checkLanguageSection(target string, actual map[string]any, defaults *config.LanguageConfig, fields []config.SDKGenConfigField, includeMatches bool) *SectionResult {
	defaultsMap := make(map[string]any)
	defaultsMap["version"] = defaults.Version
	for k, v := range defaults.Cfg {
		defaultsMap[k] = v
	}

	// Build description map from fields
	descriptionMap := make(map[string]string)
	for _, field := range fields {
		if field.Description != nil {
			descriptionMap[field.Name] = *field.Description
		}
	}

	flatActual := flattenMap(actual, "")

	result := &SectionResult{Name: target}

	for key, defaultVal := range defaultsMap {
		if shouldExcludeKey(key) {
			continue
		}

		actualVal, exists := flatActual[key]
		if !exists {
			if includeMatches {
				result.Matches = append(result.Matches, ConfigDifference{
					Key:         key,
					NewSDKValue: defaultVal,
					Description: descriptionMap[key],
				})
			}
			continue
		}

		if !valuesEqual(actualVal, defaultVal) {
			result.Differences = append(result.Differences, ConfigDifference{
				Key:          key,
				CurrentValue: actualVal,
				NewSDKValue:  defaultVal,
				Description:  descriptionMap[key],
			})
		} else if includeMatches {
			result.Matches = append(result.Matches, ConfigDifference{
				Key:          key,
				CurrentValue: actualVal,
				NewSDKValue:  defaultVal,
				Description:  descriptionMap[key],
			})
		}
	}

	sortDifferences(result.Differences)
	sortDifferences(result.Matches)

	return result
}

func sortDifferences(diffs []ConfigDifference) {
	sort.Slice(diffs, func(i, j int) bool {
		return diffs[i].Key < diffs[j].Key
	})
}

func flattenMap(m map[string]any, prefix string) map[string]any {
	result := make(map[string]any)

	for key, val := range m {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := val.(type) {
		case map[string]any:
			result[fullKey] = v
			for k, fv := range flattenMap(v, fullKey) {
				result[k] = fv
			}
		default:
			result[fullKey] = val
		}
	}

	return result
}

func valuesEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	aMap, aIsMap := a.(map[string]any)
	bMap, bIsMap := b.(map[string]any)
	if aIsMap && bIsMap {
		return mapsEqual(aMap, bMap)
	}

	aSlice, aIsSlice := toSlice(a)
	bSlice, bIsSlice := toSlice(b)
	if aIsSlice && bIsSlice {
		return slicesEqual(aSlice, bSlice)
	}

	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)

	return aStr == bStr
}

func toSlice(v any) ([]any, bool) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil, false
	}

	result := make([]any, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		result[i] = rv.Index(i).Interface()
	}
	return result, true
}

func slicesEqual(a, b []any) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !valuesEqual(a[i], b[i]) {
			return false
		}
	}
	return true
}

func mapsEqual(a, b map[string]any) bool {
	if len(a) != len(b) {
		return false
	}
	for k, av := range a {
		bv, exists := b[k]
		if !exists {
			return false
		}
		if !valuesEqual(av, bv) {
			return false
		}
	}
	return true
}

// FormatValue formats a value for display
func FormatValue(v any) string {
	switch val := v.(type) {
	case map[string]any:
		parts := make([]string, 0, len(val))
		for k, v := range val {
			parts = append(parts, fmt.Sprintf("%s:%v", k, FormatValue(v)))
		}
		sort.Strings(parts)
		return "{" + strings.Join(parts, ", ") + "}"
	case []any:
		parts := make([]string, len(val))
		for i, v := range val {
			parts[i] = FormatValue(v)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	default:
		return fmt.Sprintf("%v", v)
	}
}
