package validation

import (
	"context"
	"errors"
	"fmt"

	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	baseLinter "github.com/speakeasy-api/openapi/linter"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/openapi/linter"
	_ "github.com/speakeasy-api/openapi/openapi/linter/customrules"
	"github.com/speakeasy-api/openapi/validation"
	config "github.com/speakeasy-api/sdk-gen-config"
)

// LintResult contains the raw output from the linter plus any parsing errors
type LintResult struct {
	Output          *baseLinter.Output
	DocInfo         *baseLinter.DocumentInfo[*openapi.OpenAPI]
	ParseErrors     []error
	LintError       error
	CliVersion      string
	ParseValidOps   bool
	SpecBytes       []byte
	RegisteredRules []Rule // Rules that were registered for this validation
}

type Option func(*options)

func WithFilteredRules(rules []string) Option {
	return func(o *options) {
		o.filteredRules = rules
	}
}

func WithParseValidOperations() Option {
	return func(o *options) {
		o.parseValidOperations = true
	}
}

func WithDisallowConfiguration() Option {
	return func(o *options) {
		o.dissallowConfiguration = true
	}
}

func WithWorkingDir(workingDir string) Option {
	return func(o *options) {
		o.workingDir = workingDir
	}
}

func WithRuleset(ruleset string) Option {
	return func(o *options) {
		o.ruleset = ruleset
	}
}

func WithCliVersion(cliVersion string) Option {
	return func(o *options) {
		o.cliVersion = cliVersion
	}
}

func WithUniformSeverity() Option {
	return func(o *options) {
		o.uniformSeverity = true
	}
}

type options struct {
	filteredRules          []string
	dissallowConfiguration bool
	workingDir             string
	ruleset                string
	cliVersion             string
	parseValidOperations   bool
	uniformSeverity        bool
}

type Validator struct {
	// Linter components
	linter    *linter.Linter
	config    *baseLinter.Config
	genConfig *config.Configuration

	// Validation state
	defaultRuleset   string
	usedRulesets     []string
	usedRules        []Rule
	registeredRules  []Rule // Track actually registered rule instances
	opts             *options
	configLoaded     bool   // Track if lint.yaml has been loaded
	loadedSchemaPath string // Track the schema path we loaded config for
}

func NewValidator(cfg *config.Configuration, defaultRuleset string, opts ...Option) (*Validator, error) {
	o := options{}

	for _, opt := range opts {
		opt(&o)
	}

	// Create linter config based on ruleset (with default filters)
	linterConfig := createLinterConfig(defaultRuleset, &o)

	// Create new linter instance without default rules (we register our own)
	lntr, err := linter.NewLinter(linterConfig, linter.WithoutDefaultRules())
	if err != nil {
		return nil, fmt.Errorf("failed to create linter: %w", err)
	}

	// Register rules based on whether filtering is enabled
	var registeredRules []Rule
	if len(o.filteredRules) > 0 {
		// When filtered rules are specified, ONLY register those specific rules
		// Get all available rules from all rulesets
		allRules := make(map[string]Rule)
		for _, rulesetName := range GetAllRulesetNames() {
			for _, rule := range NewRuleset(rulesetName) {
				allRules[rule.ID()] = rule
			}
		}

		for _, ruleID := range o.filteredRules {
			translatedID := TranslateRuleName(ruleID)
			if rule, ok := allRules[translatedID]; ok {
				// Set config if the rule needs it
				if configRule, isConfigAware := rule.(configAwareRule); isConfigAware {
					configRule.SetConfig(cfg)
				}
				lntr.Registry().Register(rule)
				registeredRules = append(registeredRules, rule)
			}
		}
	} else {
		// No filtering - register all rules from the ruleset
		registeredRules = NewRuleset(defaultRuleset)
		for _, rule := range registeredRules {
			if configRule, isConfigAware := rule.(configAwareRule); isConfigAware {
				configRule.SetConfig(cfg)
			}
			lntr.Registry().Register(rule)
		}
	}

	return &Validator{
		linter:          lntr,
		config:          linterConfig,
		genConfig:       cfg,
		defaultRuleset:  defaultRuleset,
		opts:            &o,
		configLoaded:    false,
		registeredRules: registeredRules,
	}, nil
}

// createLinterConfig creates a linter configuration based on the ruleset name
// with default filters included
func createLinterConfig(ruleset string, options *options) *baseLinter.Config {
	// Get the base configuration for the ruleset
	cfg := GetRulesetConfig(ruleset)

	// Add default filters (these can be overridden by lint.yaml later)
	if !options.dissallowConfiguration {
		cfg.Rules = append(cfg.Rules, getDefaultRuleEntries()...)
	}

	// Apply filtered rules if specified
	if len(options.filteredRules) > 0 {
		// When filtered rules are specified, we need to disable all other rules
		// and only enable the ones in the filter list
		cfg.Rules = []baseLinter.RuleEntry{}
		cfg.Extends = []string{} // Don't extend anything when filtering

		// Enable only the filtered rules (rules are enabled by default unless disabled: true)
		for _, ruleID := range options.filteredRules {
			// Translate old rule names to new names
			newRuleID := TranslateRuleName(ruleID)
			cfg.Rules = append(cfg.Rules, baseLinter.RuleEntry{
				ID: newRuleID,
			})
		}
	}

	return cfg
}

// ValidateDocument validates an already-parsed OpenAPI document from DocumentInfo
// This is the optimized API that eliminates duplicate parsing between validation and generation
// Optionally accepts parse errors from openapi.Unmarshal which will be included in the output
func (v *Validator) ValidateDocument(ctx context.Context, docInfo *document.DocumentInfo, target types.Target, parseErrors ...error) *Result {
	// Load lint.yaml configuration if not already loaded or schema path changed
	if !v.configLoaded || v.loadedSchemaPath != docInfo.SchemaPath {
		// Load lint.yaml and create merged config
		linterConfig, rulesets, err := loadLinterConfig(ctx, docInfo.SchemaPath, v.defaultRuleset, v.opts)
		if err != nil {
			return &Result{
				errs: []error{fmt.Errorf("failed to load lint configuration: %w", err)},
			}
		}

		// Update the config and recreate linter without default rules
		v.config = linterConfig
		var lntrErr error
		v.linter, lntrErr = linter.NewLinter(linterConfig, linter.WithoutDefaultRules())
		if lntrErr != nil {
			return &Result{
				errs: []error{fmt.Errorf("failed to create linter: %w", lntrErr)},
			}
		}

		// Re-register rules based on whether filtering is enabled
		v.registeredRules = nil // Clear previous registrations
		if len(v.opts.filteredRules) > 0 {
			// When filtered rules are specified, ONLY register those specific rules
			allRules := make(map[string]Rule)
			for _, rulesetName := range GetAllRulesetNames() {
				for _, rule := range NewRuleset(rulesetName) {
					allRules[rule.ID()] = rule
				}
			}

			for _, ruleID := range v.opts.filteredRules {
				translatedID := TranslateRuleName(ruleID)
				if rule, ok := allRules[translatedID]; ok {
					if configRule, isConfigAware := rule.(configAwareRule); isConfigAware {
						configRule.SetConfig(v.genConfig)
					}
					v.linter.Registry().Register(rule)
					v.registeredRules = append(v.registeredRules, rule)
				}
			}
		} else {
			// No filtering - register all rules from the rulesets returned by loadLinterConfig
			// This includes the default ruleset and any extended rulesets from lint.yaml
			for _, rulesetName := range rulesets {
				rules := NewRuleset(rulesetName)
				for _, rule := range rules {
					if configRule, isConfigAware := rule.(configAwareRule); isConfigAware {
						configRule.SetConfig(v.genConfig)
					}
					v.linter.Registry().Register(rule)
					v.registeredRules = append(v.registeredRules, rule)
				}
			}
			// Store the registered rulesets
			v.usedRulesets = rulesets
		}

		// Mark config as loaded
		v.configLoaded = true
		v.loadedSchemaPath = docInfo.SchemaPath
	}

	// Use a default location if none provided (for testing and in-memory specs)
	location := docInfo.SchemaPath
	if location == "" {
		location = "openapi.yaml"
	}

	// Build index for operation validity tracking and efficient rule evaluation
	// This is required for GetValidOperations/GetInvalidOperations to work
	resolveOpts := docInfo.GetResolutionOptions(ctx)
	resolveOpts.TargetDocument = docInfo.Doc // Required by BuildIndex
	resolveOpts.TargetLocation = location    // Use normalized location (defaults to "openapi.yaml" if empty)

	// Enable node-to-operation mapping if operation validity tracking is requested
	var indexOpts []openapi.IndexOption
	if v.opts.parseValidOperations {
		indexOpts = append(indexOpts, openapi.WithNodeOperationMap())
	}

	index := openapi.BuildIndex(ctx, docInfo.Doc, resolveOpts, indexOpts...)

	// Create document info from parsed document with precomputed index
	linterDocInfo := baseLinter.NewDocumentInfoWithIndex(docInfo.Doc, location, index)

	// Get validation errors from the index (includes validation-unknown-properties and other index-time validation)
	indexErrors := index.GetAllErrors()

	// Combine parse errors and index errors
	allPreExistingErrors := append(parseErrors, indexErrors...)

	// Set docInfo and target on all registered rules that need them
	// Use the tracked registered rule instances to ensure we're setting state on the right instances
	for _, rule := range v.registeredRules {
		if docInfoRule, isDocInfoAware := rule.(docInfoAwareRule); isDocInfoAware {
			docInfoRule.SetDocInfo(docInfo)
		}
		if targetRule, isTargetAware := rule.(targetAwareRule); isTargetAware {
			targetRule.SetTarget(target)
		}
	}

	// Set up lint options
	opts := &baseLinter.LintOptions{
		VersionFilter: &docInfo.Doc.OpenAPI, // Filter rules by OpenAPI version
	}

	// Run linter with parse errors and index errors as pre-existing errors
	// The linter now applies all match filters internally based on the config
	output, err := v.linter.Lint(ctx, linterDocInfo, allPreExistingErrors, opts)

	// Create LintResult
	lintResult := &LintResult{
		Output:          output,
		DocInfo:         linterDocInfo,
		ParseErrors:     parseErrors,
		LintError:       err,
		CliVersion:      v.opts.cliVersion,
		ParseValidOps:   v.opts.parseValidOperations,
		SpecBytes:       docInfo.Schema,
		RegisteredRules: v.registeredRules,
	}

	// Transform linter output to Result format (no manual filter application needed)
	result := transformLintResult(lintResult)
	v.trackUsedRules(lintResult.Output)
	return result
}

// FilterErrors applies the linter's filters to errors and separates them by severity
// This is used for additional errors from resolution or other sources
func (v *Validator) FilterErrors(errs []error) ([]error, []error) {
	// Use the linter's FilterErrors to apply the same filters as Lint()
	filtered := v.linter.FilterErrors(errs)

	// Separate filtered errors by severity
	return separateErrorsBySeverity(filtered)
}

// separateErrorsBySeverity separates errors into warnings and errors based on severity
func separateErrorsBySeverity(errs []error) ([]error, []error) {
	var outWarns, outErrs []error
	for _, e := range errs {
		var vErr *validation.Error
		if errors.As(e, &vErr) && (vErr.Severity == validation.SeverityWarning || vErr.Severity == validation.SeverityHint) {
			outWarns = append(outWarns, e)
		} else {
			outErrs = append(outErrs, e)
		}
	}
	return outWarns, outErrs
}

// trackUsedRules updates the usedRules list with all rules that are registered and enabled in the linter
// This allows checking if a validator is using a specific ruleset (e.g., generation ruleset)
func (v *Validator) trackUsedRules(_ *baseLinter.Output) {
	// Track the rules from the rulesets we actually registered
	if len(v.usedRulesets) == 0 {
		return
	}

	var usedRules []Rule
	for _, rulesetName := range v.usedRulesets {
		ruleset := NewRuleset(rulesetName)
		if ruleset == nil {
			continue
		}
		usedRules = append(usedRules, ruleset...)
	}

	v.usedRules = usedRules
}

func (v *Validator) GetRules() map[string]Rule {
	// Flatten all rules from all rulesets into a single map
	rules := map[string]Rule{}

	for _, rulesetName := range GetAllRulesetNames() {
		for _, rule := range NewRuleset(rulesetName) {
			rules[rule.ID()] = rule
		}
	}

	return rules
}

func (v *Validator) GetRulesets() map[string][]Rule {
	// Build rulesets map from NewRuleset
	rulesets := map[string][]Rule{}
	for _, name := range GetAllRulesetNames() {
		rulesets[name] = NewRuleset(name)
	}
	return rulesets
}

func (v *Validator) GetUsedRules() []Rule {
	return v.usedRules
}

func (v *Validator) AreRulesUsed(rules []Rule) bool {
	// Check if all provided rules are in the usedRules list
	for _, rule := range rules {
		found := false
		for _, usedRule := range v.usedRules {
			// Compare by rule ID since rule instances may differ
			if rule.ID() == usedRule.ID() {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
