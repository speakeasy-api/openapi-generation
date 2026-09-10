package validation

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"

	baseLinter "github.com/speakeasy-api/openapi/linter"
	"github.com/speakeasy-api/openapi/pointer"
	"github.com/speakeasy-api/openapi/validation"
	"github.com/speakeasy-api/sdk-gen-config/lint"
)

// loadLinterConfig loads and converts lint.yaml to linter.Config
// It handles:
// - Loading lint.yaml via sdk-gen-config/lint
// - Converting to linter.Config format
// - Adding default match filters
// - Merging extended rulesets
// Returns the config and a list of rulesets that should be registered
func loadLinterConfig(ctx context.Context, schemaPath string, defaultRuleset string, opts *options) (*baseLinter.Config, []string, error) {
	// Start with the base ruleset configuration
	rulesetName := defaultRuleset
	if opts.ruleset != "" {
		rulesetName = opts.ruleset
	}
	config := GetRulesetConfig(rulesetName)
	rulesets := []string{rulesetName} // Track rulesets to register

	// Add default match filters as RuleEntry items
	config.Rules = append(config.Rules, getDefaultRuleEntries()...)

	// Don't load lint.yaml if configuration is disallowed
	if opts.dissallowConfiguration {
		return config, rulesets, nil
	}

	// Load lint.yaml if present
	lintCfg, configPath, err := loadLintYAML(ctx, schemaPath, opts)
	if err != nil {
		return nil, nil, err
	}

	// If no lint.yaml, return base config
	if lintCfg == nil {
		return config, rulesets, nil
	}

	// Wire custom rules configuration from lint.yaml to linter config
	var hasCustomRules bool
	if lintCfg.CustomRules != nil && len(lintCfg.CustomRules.Paths) > 0 {
		hasCustomRules = true
		config.CustomRules = &baseLinter.CustomRulesConfig{
			Paths:   lintCfg.CustomRules.Paths,
			Timeout: lintCfg.CustomRules.Timeout,
		}
	}

	// Determine which ruleset to use
	rulesetName = lintCfg.DefaultRuleset
	if opts.ruleset != "" {
		rulesetName = opts.ruleset
	}

	// If no ruleset specified, return base config
	if rulesetName == "" {
		return config, rulesets, nil
	}

	// Get the ruleset from lint.yaml
	ruleset, ok := lintCfg.Rulesets[rulesetName]
	if !ok {
		// Check if it's a built-in ruleset
		if NewRuleset(rulesetName) != nil {
			// Use built-in ruleset config
			return config, rulesets, nil
		}
		return nil, nil, fmt.Errorf("ruleset `%s` not found in `%s` or built-in rulesets", rulesetName, configPath)
	}

	// Merge extended rulesets from lint.yaml
	if len(ruleset.Rulesets) > 0 {
		// Start fresh with merged configs from all extended rulesets
		mergedConfig := &baseLinter.Config{
			Extends:     []string{},
			Rules:       []baseLinter.RuleEntry{},
			CustomRules: config.CustomRules, // Preserve custom rules from lint.yaml
		}

		// Track extended rulesets for registration
		rulesets = ruleset.Rulesets

		for _, extendedRulesetName := range ruleset.Rulesets {
			extendedConfig := GetRulesetConfig(extendedRulesetName)

			// Merge extends
			mergedConfig.Extends = append(mergedConfig.Extends, extendedConfig.Extends...)

			// Merge rules (later rulesets append to the list)
			mergedConfig.Rules = append(mergedConfig.Rules, extendedConfig.Rules...)
		}

		// Add default filters
		mergedConfig.Rules = append(mergedConfig.Rules, getDefaultRuleEntries()...)

		config = mergedConfig
	}

	// Apply rule overrides from lint.yaml
	for _, rule := range ruleset.Rules {
		if rule.ID == "" {
			continue
		}

		// Translate old rule names to new names
		translatedID := TranslateRuleName(rule.ID)

		// Create RuleEntry
		entry := baseLinter.RuleEntry{
			ID: translatedID,
		}

		// If rule has a match field, it's a match filter
		if rule.Match != nil {
			entry.Match = rule.Match
		}

		// Apply severity override
		if rule.Severity != "" {
			if rule.Severity == "off" {
				// "off" is treated as disabled
				entry.Disabled = pointer.From(true)
			} else {
				severity := mapStringToSeverity(rule.Severity)
				entry.Severity = &severity
			}
		}

		// Apply disabled flag
		if rule.Disabled {
			entry.Disabled = pointer.From(true)
		}

		// Add to rules list
		config.Rules = append(config.Rules, entry)
	}

	// Use "all" extends so custom rules (loaded by NewLinter into the registry)
	// are enabled alongside the explicitly registered generator rules.
	if hasCustomRules {
		config.Extends = append(config.Extends, "all")
	}

	return config, rulesets, nil
}

// loadLintYAML loads lint.yaml using sdk-gen-config/lint
func loadLintYAML(_ context.Context, schemaPath string, opts *options) (*lint.Lint, string, error) {
	searchPaths := []string{}

	if opts.workingDir != "" {
		searchPaths = append(searchPaths, opts.workingDir)
	}
	if schemaPath != "" {
		configSearchDir := filepath.Dir(schemaPath)
		searchPaths = append(searchPaths, configSearchDir)
	}

	if len(searchPaths) == 0 {
		return nil, "", nil
	}

	config, configPath, err := lint.Load(searchPaths)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// No lint.yaml found - not an error, just use defaults
			return nil, "", nil
		}
		return nil, "", fmt.Errorf("failed to load `lint.yaml`: %w", err)
	}

	return config, configPath, nil
}

// getDefaultRuleEntries converts default match filters to RuleEntry items
// These are the built-in match filters from default_filters.go
func getDefaultRuleEntries() []baseLinter.RuleEntry {
	warnSeverity := validation.SeverityWarning

	bt := "`?"

	return []baseLinter.RuleEntry{
		// Downgrade required field validation errors to warnings for common lenient cases
		// Note: upstream library wraps field names in backticks (e.g., `info.title` is required)
		{
			ID:       "validation-required-field",
			Match:    regexp.MustCompile(`.*` + bt + `info\.title` + bt + ` is required.*`),
			Severity: &warnSeverity,
		},
		{
			ID:       "validation-required-field",
			Match:    regexp.MustCompile(`.*` + bt + `info\.version` + bt + ` is required.*`),
			Severity: &warnSeverity,
		},
		{
			ID:       "validation-required-field",
			Match:    regexp.MustCompile(`.*` + bt + `openapi\.info` + bt + ` is required.*`),
			Severity: &warnSeverity,
		},
		{
			ID:       "validation-required-field",
			Match:    regexp.MustCompile(`.*` + bt + `response\.description` + bt + ` is required.*`),
			Severity: &warnSeverity,
		},
		{
			ID:       "validation-required-field",
			Match:    regexp.MustCompile(`.*` + bt + `serverVariable\.default` + bt + ` is required.*`),
			Severity: &warnSeverity,
		},
		{
			ID:       "validation-required-field",
			Match:    regexp.MustCompile(`.*` + bt + `parameter\.in=path` + bt + ` requires ` + bt + `required=true` + bt + `.*`),
			Severity: &warnSeverity,
		},
		{
			ID:       "validation-required-field",
			Match:    regexp.MustCompile(`.*oAuthFlow\.tokenUrl is required for type=.*`),
			Severity: &warnSeverity,
		},

		// Downgrade invalid format errors to warnings
		{
			ID:       "validation-invalid-format",
			Match:    regexp.MustCompile(`.*contact\.email is not a valid email address.*`),
			Severity: &warnSeverity,
		},
		{
			ID:       "validation-invalid-format",
			Match:    regexp.MustCompile(`.*callback expression is invalid.*`),
			Severity: &warnSeverity,
		},

		// Downgrade type mismatch errors for boolean exclusiveMin/Max (OpenAPI 3.0 vs 3.1)
		{
			ID:       "validation-type-mismatch",
			Match:    regexp.MustCompile(`.*schema\..*?exclusiveMinimum expected ` + bt + `number` + bt + `, got ` + bt + `boolean` + bt + `.*`),
			Severity: &warnSeverity,
		},
		{
			ID:       "validation-type-mismatch",
			Match:    regexp.MustCompile(`.*schema\..*?exclusiveMaximum expected ` + bt + `number` + bt + `, got ` + bt + `boolean` + bt + `.*`),
			Severity: &warnSeverity,
		},
		{
			ID:       "validation-type-mismatch",
			Match:    regexp.MustCompile(`.*examples expected ` + bt + `sequence` + bt + `, got ` + bt + `object` + bt + `.*`),
			Severity: &warnSeverity,
		},

		// Downgrade invalid syntax warnings
		{
			ID:       "validation-invalid-syntax",
			Match:    regexp.MustCompile(`.*server variable .*? has no default value.*`),
			Severity: &warnSeverity,
		},

		// Downgrade allowed-values errors to warnings
		{
			ID:       "validation-allowed-values",
			Match:    regexp.MustCompile(`.*` + bt + `parameter\.allowEmptyValue` + bt + ` is only valid for ` + bt + `in=query` + bt + `.*`),
			Severity: &warnSeverity,
		},

		// Downgrade invalid schema warnings for duplicate required items
		{
			ID:       "validation-invalid-schema",
			Match:    regexp.MustCompile(`.*schema\..*?required items at \d+ and \d+ are equal.*`),
			Severity: &warnSeverity,
		},

		// Downgrade invalid reference warnings
		{
			ID:       "validation-invalid-reference",
			Match:    regexp.MustCompile(`.*link\.operationId value.*`),
			Severity: &warnSeverity,
		},

		// Downgrade scheme not found errors
		{
			ID:       "validation-scheme-not-found",
			Match:    regexp.MustCompile(`.*securityRequirement scheme .+ is not defined in components\.securitySchemes.*`),
			Severity: &warnSeverity,
		},

		// Downgrade all operation parameter errors to warnings
		{
			ID:       "validation-operation-parameters",
			Match:    regexp.MustCompile(`.*`),
			Severity: &warnSeverity,
		},
	}
}

// mapStringToSeverity converts a severity string to a validation.Severity
func mapStringToSeverity(s string) validation.Severity {
	switch s {
	case "error":
		return validation.SeverityError
	case "warn", "warning":
		return validation.SeverityWarning
	case "hint", "info":
		return validation.SeverityHint
	default:
		return validation.SeverityError
	}
}
