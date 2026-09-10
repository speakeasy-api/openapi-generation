package validation

import (
	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	baseLinter "github.com/speakeasy-api/openapi/linter"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/openapi/linter"
	"github.com/speakeasy-api/openapi/openapi/linter/rules"
	"github.com/speakeasy-api/openapi/validation"
	config "github.com/speakeasy-api/sdk-gen-config"
)

const (
	RulesetSpeakeasyRecommended = "speakeasy-recommended"
	RulesetSpeakeasyGeneration  = "speakeasy-generation"
	RulesetSpeakeasyUnsupported = "speakeasy-unsupported-constructs"
	RulesetSpeakeasyOpenAPI     = "speakeasy-openapi"
	RulesetOpenAPIStandard      = "openapi-standard"
	RulesetOWASP                = "owasp"
)

type configAwareRule interface {
	SetConfig(cfg *config.Configuration)
}

type docInfoAwareRule interface {
	SetDocInfo(docInfo *document.DocumentInfo)
}

type targetAwareRule interface {
	SetTarget(target types.Target)
}

type Rule = baseLinter.RuleRunner[*openapi.OpenAPI]

// GetAllRulesetNames returns all available ruleset names
func GetAllRulesetNames() []string {
	return []string{
		RulesetSpeakeasyRecommended,
		RulesetSpeakeasyGeneration,
		RulesetSpeakeasyUnsupported,
		RulesetSpeakeasyOpenAPI,
		RulesetOpenAPIStandard,
		RulesetOWASP,
	}
}

// NewRuleset returns a fresh set of rule instances for the given ruleset name.
// This ensures each validator gets its own rule instances, preventing state pollution.
func NewRuleset(name string) []Rule {
	switch name {
	case RulesetSpeakeasyRecommended:
		return newSpeakeasyRecommendedRules()
	case RulesetSpeakeasyGeneration:
		return newSpeakeasyGenerationRules()
	case RulesetSpeakeasyUnsupported:
		return newSpeakeasyUnsupportedRules()
	case RulesetSpeakeasyOpenAPI:
		return newSpeakeasyOpenAPIRules()
	case RulesetOpenAPIStandard:
		return newOpenAPIStandardRules()
	case RulesetOWASP:
		return newOWASPRules()
	default:
		return nil
	}
}

func newSpeakeasyRecommendedRules() []Rule {
	return []Rule{
		// Custom generator rules - collision detection
		&DuplicateSchemaName{},
		&DuplicateOperationName{},
		&DuplicateProperties{},
		&DuplicatePathParams{},
		&DuplicateTag{},
		&DuplicateErrors{},
		&DuplicateInlineSchemas{},
		&DuplicateModelNamespace{},
		&DuplicateComponentSchemas{}, // In Recommended + OpenAPI only (NOT in Generation)
		&ValidateCasing{},            // In Recommended + OpenAPI only (NOT in Generation)

		// Custom generator rules - type/schema validation
		&ValidateTypes{},
		&ValidateCompositeSchemas{}, // In Recommended + Generation + OpenAPI
		&ValidateConstsDefaults{},   // In Recommended + Generation + OpenAPI
		&ValidateContentType{},      // In Recommended + Generation + OpenAPI
		&ValidateEnums{},            // In Recommended + Generation
		&ValidateBase64InputMode{},  // In Recommended + Generation
		&ValidateExtensions{},       // In Recommended + Generation
		&ValidateDocument{},         // In Recommended + Generation + OpenAPI
		&ValidateParameters{},       // In Recommended + Generation + OpenAPI
		&ValidatePaths{},            // In Recommended + Generation + OpenAPI
		&ValidateRequests{},         // In Recommended + Generation + OpenAPI
		&ValidateResponses{},        // In Recommended + Generation + OpenAPI
		&ValidateSecurity{},         // In Recommended + Generation + OpenAPI
		&ValidateServers{},          // In Recommended + Generation + OpenAPI (SeverityHint)
		&ValidateDeprecation{},      // In Recommended + Generation + OpenAPI
		&SuggestDiscriminator{},     // In Recommended + OpenAPI only (NOT in Generation)
		&PathParams{},               // In Recommended + Generation + OpenAPI (custom path param validation)
		&MissingErrorResponse{},
		&MissingExamples{},               // In Recommended + OpenAPI only (NOT in Generation)
		&Pagination{},                    // In Recommended + Generation + OpenAPI
		&Retries{},                       // In Recommended + Generation + OpenAPI
		&PaginationNullableRequestBody{}, // In Recommended + Generation + OpenAPI

		&CsharpOptionalNullableDecimal{}, // C#-only warning: presence-aware optional+nullable scalar decimal

		// Built-in semantic rules (in Recommended + Generation + OpenAPI)
		&rules.PathDeclarationsRule{},         // semantic-path-declarations
		&rules.PathQueryRule{},                // semantic-path-query (was PathNotIncludeQuery)
		&rules.TypedEnumRule{},                // semantic-typed-enum
		&rules.NoEvalInMarkdownRule{},         // semantic-no-eval-in-markdown
		&rules.NoScriptTagsInMarkdownRule{},   // semantic-no-script-tags-in-markdown
		&rules.OperationSuccessResponseRule{}, // style-operation-success-response

		// Built-in rules (in Recommended + OpenAPI only, NOT in Generation)
		&rules.UnusedComponentRule{},     // semantic-unused-component
		&rules.OAS3HostNotExampleRule{},  // style-oas3-host-not-example
		&rules.OperationIdRule{},         // semantic-operation-operation-id (SeverityWarning override)
		&rules.DuplicatedEnumRule{},      // semantic-duplicated-enum (SeverityWarning override)
		&rules.OperationTagDefinedRule{}, // style-operation-tag-defined (SeverityWarning override)
	}
}

func newSpeakeasyGenerationRules() []Rule {
	return []Rule{
		// Critical custom generator rules for SDK generation
		// Collision detection
		&DuplicateSchemaName{},
		&DuplicateOperationName{},
		&DuplicateProperties{},
		&DuplicatePathParams{},
		&DuplicateTag{},
		&DuplicateErrors{},
		&DuplicateInlineSchemas{},
		&DuplicateModelNamespace{},

		// Type/schema validation
		&ValidateTypes{},
		&ValidateCompositeSchemas{}, // In Recommended + Generation + OpenAPI
		&ValidateConstsDefaults{},   // In Recommended + Generation + OpenAPI
		&ValidateContentType{},      // In Recommended + Generation + OpenAPI
		&ValidateEnums{},            // In Recommended + Generation
		&ValidateBase64InputMode{},  // In Recommended + Generation
		&ValidateExtensions{},       // In Recommended + Generation
		&ValidateDocument{},         // In Recommended + Generation + OpenAPI
		&ValidateParameters{},       // In Recommended + Generation + OpenAPI
		&ValidatePaths{},            // In Recommended + Generation + OpenAPI
		&ValidateRequests{},         // In Recommended + Generation + OpenAPI
		&ValidateResponses{},        // In Recommended + Generation + OpenAPI
		&ValidateSecurity{},         // In Recommended + Generation + OpenAPI
		&ValidateServers{},          // In Recommended + Generation + OpenAPI (SeverityHint)
		&ValidateDeprecation{},      // In Recommended + Generation + OpenAPI
		&PathParams{},               // In Recommended + Generation + OpenAPI
		&MissingErrorResponse{},
		&Pagination{},                    // In Recommended + Generation + OpenAPI
		&Retries{},                       // In Recommended + Generation + OpenAPI
		&PaginationNullableRequestBody{}, // In Recommended + Generation + OpenAPI

		&CsharpOptionalNullableDecimal{}, // C#-only warning: presence-aware optional+nullable scalar decimal

		// Critical built-in semantic rules (in Recommended + Generation + OpenAPI)
		&rules.PathDeclarationsRule{},         // semantic-path-declarations
		&rules.PathQueryRule{},                // semantic-path-query
		&rules.TypedEnumRule{},                // semantic-typed-enum
		&rules.NoEvalInMarkdownRule{},         // semantic-no-eval-in-markdown
		&rules.NoScriptTagsInMarkdownRule{},   // semantic-no-script-tags-in-markdown
		&rules.OperationSuccessResponseRule{}, // style-operation-success-response
	}
}

// newUnsupportedOnlyRules returns the rules that report OpenAPI constructs the
// generator cannot represent. They belong to no other ruleset, so they only run
// when a lint.yaml opts into speakeasy-unsupported-constructs.
func newUnsupportedOnlyRules() []Rule {
	return []Rule{
		&ValidateNotSchemas{},
	}
}

// newSpeakeasyUnsupportedRules is a strict superset of the generation ruleset, so
// generation's own "must validate against the generation ruleset" check still
// holds when a spec selects it.
func newSpeakeasyUnsupportedRules() []Rule {
	return append(newSpeakeasyGenerationRules(), newUnsupportedOnlyRules()...)
}

func newSpeakeasyOpenAPIRules() []Rule {
	return []Rule{
		// All speakeasy-recommended rules plus additional ones
		// Collision detection
		&DuplicateSchemaName{},
		&DuplicateOperationName{},
		&DuplicateProperties{},
		&DuplicatePathParams{},
		&DuplicateTag{},
		&DuplicateErrors{},
		&DuplicateInlineSchemas{},
		&DuplicateModelNamespace{},

		// Type/schema validation
		&ValidateTypes{},
		&ValidateCompositeSchemas{},  // In Recommended + Generation + OpenAPI
		&ValidateConstsDefaults{},    // In Recommended + Generation + OpenAPI
		&ValidateContentType{},       // In Recommended + Generation + OpenAPI
		&ValidateParameters{},        // In Recommended + Generation + OpenAPI
		&ValidatePaths{},             // In Recommended + Generation + OpenAPI
		&ValidateRequests{},          // In Recommended + Generation + OpenAPI
		&ValidateResponses{},         // In Recommended + Generation + OpenAPI
		&ValidateSecurity{},          // In Recommended + Generation + OpenAPI
		&ValidateServers{},           // In Recommended + Generation + OpenAPI (SeverityHint)
		&ValidateDocument{},          // In Recommended + Generation + OpenAPI
		&ValidateDeprecation{},       // In Recommended + Generation + OpenAPI
		&SuggestDiscriminator{},      // In Recommended + OpenAPI only (NOT in Generation)
		&ValidateCasing{},            // In Recommended + OpenAPI only (NOT in Generation)
		&DuplicateComponentSchemas{}, // In Recommended + OpenAPI only (NOT in Generation)
		&PathParams{},                // In Recommended + Generation + OpenAPI
		&MissingErrorResponse{},
		&MissingExamples{},               // In Recommended + OpenAPI only (NOT in Generation)
		&Pagination{},                    // In Recommended + Generation + OpenAPI
		&Retries{},                       // In Recommended + Generation + OpenAPI
		&PaginationNullableRequestBody{}, // In Recommended + Generation + OpenAPI

		// Built-in semantic rules (in Recommended + Generation + OpenAPI)
		&rules.PathDeclarationsRule{},         // semantic-path-declarations
		&rules.PathQueryRule{},                // semantic-path-query
		&rules.TypedEnumRule{},                // semantic-typed-enum
		&rules.NoEvalInMarkdownRule{},         // semantic-no-eval-in-markdown
		&rules.NoScriptTagsInMarkdownRule{},   // semantic-no-script-tags-in-markdown
		&rules.OperationSuccessResponseRule{}, // style-operation-success-response

		// Built-in rules (in Recommended + OpenAPI only, NOT in Generation)
		&rules.UnusedComponentRule{},     // semantic-unused-component
		&rules.OAS3HostNotExampleRule{},  // style-oas3-host-not-example
		&rules.OperationIdRule{},         // semantic-operation-operation-id (SeverityWarning override)
		&rules.DuplicatedEnumRule{},      // semantic-duplicated-enum (SeverityWarning override)
		&rules.OperationTagDefinedRule{}, // style-operation-tag-defined (SeverityWarning override)
		&rules.OAS3ExampleMissingRule{},  // oas3-example-missing
		&rules.LinkOperationRule{},       // validation-link-operation (validates link.operationId/operationRef exists)
	}
}

func newOpenAPIStandardRules() []Rule {
	return []Rule{
		// OpenAPI Standard ruleset = all custom rules + all built-in rules from the library
		// This is the most comprehensive ruleset with all validation rules enabled

		// Custom generator rules - collision detection
		&DuplicateSchemaName{},
		&DuplicateOperationName{},
		&DuplicateProperties{},
		&DuplicatePathParams{},
		&DuplicateTag{},
		&DuplicateErrors{},
		&DuplicateInlineSchemas{},
		&DuplicateModelNamespace{},

		// Custom generator rules - type/schema validation
		&ValidateTypes{},
		&ValidateCompositeSchemas{},
		&ValidateConstsDefaults{},
		&ValidateContentType{},
		&ValidateParameters{},
		&ValidatePaths{},
		&ValidateRequests{},
		&ValidateResponses{},
		&ValidateSecurity{},
		&ValidateServers{},
		&ValidateDocument{},
		&ValidateDeprecation{},
		&SuggestDiscriminator{},
		&ValidateCasing{},
		&DuplicateComponentSchemas{},
		&PathParams{},
		&MissingErrorResponse{},
		&MissingExamples{},
		&Pagination{},
		&Retries{},

		// Built-in semantic rules - all rules from openapi library
		// Component and schema rules
		&rules.ComponentDescriptionRule{},
		&rules.DuplicatedEnumRule{},
		&rules.NoRefSiblingsRule{},
		&rules.OASSchemaCheckRule{},
		&rules.TypedEnumRule{},
		&rules.UnusedComponentRule{},

		// Info and metadata rules
		&rules.ContactPropertiesRule{},
		&rules.DescriptionDuplicationRule{},
		&rules.InfoContactRule{},
		&rules.InfoDescriptionRule{},
		&rules.InfoLicenseRule{},
		&rules.LicenseURLRule{},

		// Path rules
		&rules.NoAmbiguousPathsRule{},
		&rules.NoVerbsInPathRule{},
		&rules.PathDeclarationsRule{},
		&rules.PathParamsRule{},
		&rules.PathQueryRule{},
		&rules.PathTrailingSlashRule{},
		&rules.PathsKebabCaseRule{},

		// Operation rules
		&rules.LinkOperationRule{},
		&rules.OperationDescriptionRule{},
		&rules.OperationErrorResponseRule{},
		&rules.OperationIDValidInURLRule{},
		&rules.OperationIdRule{},
		&rules.OperationSingularTagRule{},
		&rules.OperationSuccessResponseRule{},
		&rules.OperationTagDefinedRule{},
		&rules.OperationTagsRule{},

		// Tag rules
		&rules.OpenAPITagsRule{},
		&rules.TagDescriptionRule{},
		&rules.TagsAlphabeticalRule{},

		// Markdown rules
		&rules.NoEvalInMarkdownRule{},
		&rules.NoScriptTagsInMarkdownRule{},

		// OAS3 specific rules
		&rules.OAS3APIServersRule{},
		&rules.OAS3ExampleMissingRule{},
		&rules.OAS3HostNotExampleRule{},
		&rules.OAS3HostTrailingSlashRule{},
		&rules.OAS3NoNullableRule{},
		&rules.OAS3ParameterDescriptionRule{},
	}
}

func newOWASPRules() []Rule {
	return []Rule{
		// All OWASP security rules
		&rules.OwaspNoNumericIDsRule{},
		&rules.OwaspNoHttpBasicRule{},
		&rules.OwaspNoAPIKeysInURLRule{},
		&rules.OwaspNoCredentialsInURLRule{},
		&rules.OwaspAuthInsecureSchemesRule{},
		&rules.OwaspJWTBestPracticesRule{},
		&rules.OwaspProtectionGlobalUnsafeRule{},
		&rules.OwaspProtectionGlobalUnsafeStrictRule{},
		&rules.OwaspProtectionGlobalSafeRule{},
		&rules.OwaspDefineErrorValidationRule{},
		&rules.OwaspDefineErrorResponses401Rule{},
		&rules.OwaspDefineErrorResponses500Rule{},
		&rules.OwaspDefineErrorResponses429Rule{},
		&rules.OwaspRateLimitRule{},
		&rules.OwaspRateLimitRetryAfterRule{},
		&rules.OwaspArrayLimitRule{},
		&rules.OwaspStringLimitRule{},
		&rules.OwaspStringRestrictedRule{},
		&rules.OwaspIntegerFormatRule{},
		&rules.OwaspIntegerLimitRule{},
		&rules.OwaspNoAdditionalPropertiesRule{},
		&rules.OwaspAdditionalPropertiesConstrainedRule{},
		&rules.OwaspSecurityHostsHttpsOAS3Rule{},
	}
}

// SeverityOverrides allows overriding default severities for specific rules in specific rulesets
var SeverityOverrides = map[string]map[string]validation.Severity{
	RulesetSpeakeasyRecommended: {
		(&rules.OperationIdRule{}).ID():         validation.SeverityWarning,
		(&rules.DuplicatedEnumRule{}).ID():      validation.SeverityWarning,
		(&rules.OperationTagDefinedRule{}).ID(): validation.SeverityWarning,
		(&rules.NoAmbiguousPathsRule{}).ID():    validation.SeverityWarning,
	},
	RulesetSpeakeasyOpenAPI: {
		(&rules.OperationIdRule{}).ID():         validation.SeverityWarning,
		(&rules.DuplicatedEnumRule{}).ID():      validation.SeverityWarning,
		(&rules.OperationTagDefinedRule{}).ID(): validation.SeverityWarning,
		(&rules.NoAmbiguousPathsRule{}).ID():    validation.SeverityWarning,
	},
}

// NewLinterWithRuleset creates a new linter with the specified ruleset registered
func NewLinterWithRuleset(rulesetName string) (*linter.Linter, error) {
	// Create registry
	registry := baseLinter.NewRegistry[*openapi.OpenAPI]()

	// Register rules for this ruleset
	rules := NewRuleset(rulesetName)
	for _, rule := range rules {
		registry.Register(rule)
	}

	// Create config for this ruleset
	cfg := GetRulesetConfig(rulesetName)

	return linter.NewLinter(cfg, linter.WithoutDefaultRules())
}

// RegisterRuleset registers all rules for a ruleset, applying config-dependent rules.
func RegisterRuleset(registry *baseLinter.Registry[*openapi.OpenAPI], rulesetName string, cfg *config.Configuration) {
	if registry == nil {
		return
	}

	rules := NewRuleset(rulesetName)
	if rules == nil {
		return
	}

	for _, rule := range rules {
		if configRule, ok := rule.(configAwareRule); ok {
			configRule.SetConfig(cfg)
		}

		registry.Register(rule)
	}
}

// GetRulesetConfig returns a linter configuration for the specified ruleset name
func GetRulesetConfig(rulesetName string) *baseLinter.Config {
	// Handle aliases - "vacuum" is the old name for "openapi-standard"
	if rulesetName == "vacuum" {
		rulesetName = RulesetOpenAPIStandard
	}

	// Get rules for this ruleset
	rules := NewRuleset(rulesetName)
	if rules == nil {
		return baseLinter.NewConfig()
	}

	cfg := &baseLinter.Config{
		Extends: []string{}, // No extends - we explicitly list all rules
		Rules:   []baseLinter.RuleEntry{},
	}

	// Enable all rules in the ruleset
	for _, rule := range rules {
		ruleID := rule.ID()

		entry := baseLinter.RuleEntry{
			ID: ruleID,
		}

		// Check if there's a severity override
		if overrides, hasOverrides := SeverityOverrides[rulesetName]; hasOverrides {
			if severity, hasSeverity := overrides[ruleID]; hasSeverity {
				entry.Severity = &severity
			}
		}

		cfg.Rules = append(cfg.Rules, entry)
	}

	return cfg
}
