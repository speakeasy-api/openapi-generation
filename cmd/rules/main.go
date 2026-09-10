package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/validation"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/markdown"
	oaValidation "github.com/speakeasy-api/openapi/validation"
	config "github.com/speakeasy-api/sdk-gen-config"
)

func main() {
	v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyRecommended)
	if err != nil {
		panic(err)
	}

	rules := v.GetRules()

	content := make([][]string, 0, 1+len(rules))
	content = append(content, []string{"Rule ID", "Default Severity", "Description"})

	// Get sorted rule IDs
	ruleIDs := make([]string, 0, len(rules))
	for id := range rules {
		ruleIDs = append(ruleIDs, id)
	}
	slices.Sort(ruleIDs)

	for _, id := range ruleIDs {
		rule := rules[id]
		content = append(content, []string{id, rule.DefaultSeverity().String(), rule.Description()})
	}

	fmt.Println("# All Rules")
	fmt.Println()
	fmt.Println(markdown.CreateMarkdownTable(content))

	rulesets := v.GetRulesets()

	keys := make([]string, 0, len(rulesets))
	for k := range rulesets {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	var rulesetMarkdown strings.Builder
	rulesetMarkdown.WriteString("\n\n# Rulesets")

	for _, name := range keys {
		ruleset := rulesets[name]
		fmt.Fprintf(&rulesetMarkdown, "\n\n## %s\n\n", name)

		content := make([][]string, 0, 1+len(ruleset))
		content = append(content, []string{"Rule ID", "Severity"})

		for _, rule := range ruleset {
			content = append(content, []string{rule.ID(), rule.DefaultSeverity().String()})
		}

		rulesetMarkdown.WriteString(markdown.CreateMarkdownTable(content))
	}

	fmt.Println(rulesetMarkdown.String())

	// Print validation errors from the openapi library
	// These are non-rule-based errors from document parsing/validation
	validationErrorIDs := []string{
		oaValidation.RuleValidationRequiredField,
		oaValidation.RuleValidationTypeMismatch,
		oaValidation.RuleValidationDuplicateKey,
		oaValidation.RuleValidationInvalidFormat,
		oaValidation.RuleValidationEmptyValue,
		oaValidation.RuleValidationInvalidReference,
		oaValidation.RuleValidationInvalidSyntax,
		oaValidation.RuleValidationInvalidSchema,
		oaValidation.RuleValidationInvalidTarget,
		oaValidation.RuleValidationAllowedValues,
		oaValidation.RuleValidationMutuallyExclusiveFields,
		oaValidation.RuleValidationOperationNotFound,
		oaValidation.RuleValidationOperationIdUnique,
		oaValidation.RuleValidationOperationParameters,
		oaValidation.RuleValidationSchemeNotFound,
		oaValidation.RuleValidationTagNotFound,
		oaValidation.RuleValidationSupportedVersion,
		oaValidation.RuleValidationCircularReference,
	}

	fmt.Print("\n\n# Validation Errors\n\n")

	validationContent := make([][]string, 0, 1+len(validationErrorIDs))
	validationContent = append(validationContent, []string{"Rule ID", "Default Severity", "Description"})

	for _, id := range validationErrorIDs {
		description := ""
		if info, ok := oaValidation.RuleInfoForID(id); ok {
			description = info.Description
		}
		validationContent = append(validationContent, []string{id, oaValidation.SeverityError.String(), description})
	}

	fmt.Println(markdown.CreateMarkdownTable(validationContent))
}
