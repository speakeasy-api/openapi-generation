package validation

import (
	stderrors "errors"
	"strings"
	"time"

	"github.com/speakeasy-api/openapi-generation/v2/internal/validation/report"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	baseLinter "github.com/speakeasy-api/openapi/linter"
	"github.com/speakeasy-api/openapi/openapi"
	openapiValidation "github.com/speakeasy-api/openapi/validation"
)

// transformLintResult converts the LintResult to the validation.Result format
// The linter output is already filtered by the library based on the config
func transformLintResult(lintResult *LintResult) *Result {
	// Handle lint error (fatal error that prevented linting)
	if lintResult.LintError != nil {
		return NewResult(
			nil,
			[]error{errors.NewValidationError("failed to run linter", nil, lintResult.LintError)},
			lintResult.CliVersion,
			nil,
		)
	}

	// Transform linter output (already filtered by library)
	result := transformOutput(lintResult.Output, lintResult.DocInfo, lintResult.SpecBytes, lintResult.RegisteredRules, lintResult.CliVersion, lintResult.ParseValidOps)

	// Note: Parse errors are already included in the linter output via preExistingErrors parameter,
	// so we don't need to add them separately here

	return result
}

// transformOutput converts the linter Output to the validation.Result format
// All match filters have already been applied by the library
func transformOutput(output *baseLinter.Output, docInfo *baseLinter.DocumentInfo[*openapi.OpenAPI], specBytes []byte, registeredRules []Rule, cliVersion string, parseValidOperations bool) *Result {
	// Build rule metadata map from registered rules
	ruleMetadata := buildRuleMetadata(registeredRules)

	// Build internal report data
	reportData := &report.ValidationReportData{
		Results:      []*openapiValidation.Error{},
		Categories:   []*report.RuleCategory{},
		Statistics:   &report.ReportStatistics{},
		SpecBytes:    specBytes,
		Generated:    time.Now(),
		RuleMetadata: ruleMetadata,
	}

	var validationErrors []error

	// Transform each error from the linter output (already filtered by library)
	for _, err := range output.Results {
		// Check if it's a validation error from the linter
		var vErr *openapiValidation.Error
		if stderrors.As(err, &vErr) {
			// Add to report data
			reportData.Results = append(reportData.Results, vErr)

			// Update statistics
			switch vErr.Severity {
			case openapiValidation.SeverityError:
				reportData.Statistics.TotalErrors++
			case openapiValidation.SeverityWarning:
				reportData.Statistics.TotalWarnings++
			case openapiValidation.SeverityHint:
				reportData.Statistics.TotalHints++
			}

			// Add metadata for this rule if not already present
			if _, exists := ruleMetadata[vErr.Rule]; !exists {
				ruleMetadata[vErr.Rule] = getRuleMetadataForID(vErr.Rule)
			}

			// Convert to old error format for GetValidationErrors()
			oldErr := convertValidationError(vErr)
			validationErrors = append(validationErrors, oldErr)
		} else {
			// For non-validation errors, wrap them as validation errors
			validationErrors = append(validationErrors, errors.NewValidationError(err.Error(), nil, err))
		}
	}

	// Build operation validity map if requested
	var operationToFailingLineNumber map[string]int
	if parseValidOperations && docInfo != nil {
		operationToFailingLineNumber = buildOperationValidity(output, docInfo)
	}

	// Create Result with internal report data
	return NewResult(reportData, validationErrors, cliVersion, operationToFailingLineNumber)
}

// convertValidationError converts a new validation.Error to an old errors.ValidationError
func convertValidationError(newErr *openapiValidation.Error) *errors.ValidationError {
	// Map severity
	severity := mapSeverity(newErr.Severity)

	// Translate rule name back to old name for backward compatibility (optional)
	// We can keep the new name or translate it back
	ruleName := newErr.Rule

	// Create the appropriate error type with node for line/column tracking
	var err *errors.ValidationError
	switch severity {
	case errors.SeverityError:
		err = errors.NewValidationError(newErr.UnderlyingError.Error(), newErr.Node, nil)
	case errors.SeverityWarn:
		err = errors.NewValidationWarning(newErr.UnderlyingError.Error(), newErr.Node, nil)
	case errors.SeverityHint:
		err = errors.NewValidationHint(newErr.UnderlyingError.Error(), newErr.Node, nil)
	default:
		err = errors.NewValidationError(newErr.UnderlyingError.Error(), newErr.Node, nil)
	}

	err.Rule = ruleName

	return err
}

// mapSeverity maps new validation.Severity to old errors.Severity
func mapSeverity(newSev openapiValidation.Severity) errors.Severity {
	switch newSev {
	case openapiValidation.SeverityError:
		return errors.SeverityError
	case openapiValidation.SeverityWarning:
		return errors.SeverityWarn // Old package uses "Warn" not "Warning"
	case openapiValidation.SeverityHint:
		return errors.SeverityHint
	default:
		return errors.SeverityError // Default to error if unknown
	}
}

// buildOperationValidity creates a map of operation IDs to their failing line numbers
// Operations with no errors are marked with -1, operations with errors have the error line number
func buildOperationValidity(output *baseLinter.Output, docInfo *baseLinter.DocumentInfo[*openapi.OpenAPI]) map[string]int {
	operationToFailingLineNumber := make(map[string]int)

	// Use the index to iterate through operations
	if docInfo.Index == nil {
		return operationToFailingLineNumber
	}

	// Initialize all operations as valid (-1 means no errors)
	for _, opIndexNode := range docInfo.Index.Operations {
		operation := opIndexNode.Node
		if operation == nil {
			continue
		}

		operationID := operation.GetOperationID()
		if operationID == "" {
			continue
		}

		operationToFailingLineNumber[operationID] = -1
	}

	// Process validation errors and mark affected operations as invalid
	// Only consider actual errors (severity=Error), not warnings or hints
	for _, err := range output.Results {
		var vErr *openapiValidation.Error
		if !stderrors.As(err, &vErr) || vErr.Node == nil || vErr.Severity != openapiValidation.SeverityError {
			continue
		}

		errorNode := vErr.Node
		errorLine := errorNode.Line

		// Use the index's node-to-operation mapping to find which operations are affected
		affectedOps := docInfo.Index.GetNodeOperations(errorNode)

		for _, opIndexNode := range affectedOps {
			operation := opIndexNode.Node
			if operation == nil {
				continue
			}

			operationID := operation.GetOperationID()
			if operationID == "" {
				continue
			}

			// Mark this operation as invalid with the error line number
			// Only update if this is the first error for this operation (keep the earliest error line)
			if currentLine, exists := operationToFailingLineNumber[operationID]; !exists || currentLine == -1 {
				operationToFailingLineNumber[operationID] = errorLine
			}
		}
	}

	return operationToFailingLineNumber
}

// getRuleMetadataForID gets metadata for a single rule ID
func getRuleMetadataForID(ruleID string) *report.RuleMetadata {
	metadata := &report.RuleMetadata{}

	// For built-in validation rules (from openapi library), look up metadata
	if isBuiltInValidationRule(ruleID) {
		if ruleInfo, ok := openapiValidation.RuleInfoForID(ruleID); ok {
			metadata.Summary = ruleInfo.Summary
			metadata.Description = ruleInfo.Description
			metadata.HowToFix = ruleInfo.HowToFix
		}
	}

	// Generate summary if we don't have one
	if metadata.Summary == "" {
		metadata.Summary = buildRuleSummary(metadata.Description, ruleID)
	}

	return metadata
}

// buildRuleMetadata builds a map of rule ID to metadata (description, how-to-fix) from registered rules
func buildRuleMetadata(registeredRules []Rule) map[string]*report.RuleMetadata {
	metadata := make(map[string]*report.RuleMetadata)

	for _, rule := range registeredRules {
		ruleID := rule.ID()

		// Get summary (short description) from the rule
		summary := ""
		if summarizer, ok := rule.(interface{ Summary() string }); ok {
			summary = summarizer.Summary()
		}

		// Get description from the rule
		description := ""
		if describer, ok := rule.(interface{ Description() string }); ok {
			description = describer.Description()
		}

		// Get how-to-fix if available
		howToFix := ""
		if fixer, ok := rule.(interface{ HowToFix() string }); ok {
			howToFix = fixer.HowToFix()
		}

		// For built-in validation rules (from openapi library), look up metadata
		if isBuiltInValidationRule(ruleID) {
			if ruleInfo, ok := openapiValidation.RuleInfoForID(ruleID); ok {
				// Use the library's metadata if our custom rule didn't provide it
				if summary == "" {
					summary = ruleInfo.Summary
				}
				if description == "" {
					description = ruleInfo.Description
				}
				if howToFix == "" {
					howToFix = ruleInfo.HowToFix
				}
			}
		}

		if summary == "" {
			summary = buildRuleSummary(description, ruleID)
		}

		metadata[ruleID] = &report.RuleMetadata{
			Summary:     summary,
			Description: description,
			HowToFix:    howToFix,
		}
	}

	return metadata
}

func buildRuleSummary(description string, ruleID string) string {
	summary := strings.TrimSpace(description)
	if summary != "" {
		summary = strings.SplitN(summary, "\n", 2)[0]
		if idx := strings.Index(summary, ". "); idx != -1 {
			summary = summary[:idx+1]
		}
		summary = strings.TrimSpace(summary)
		if len(summary) > 120 {
			summary = strings.TrimSpace(summary[:117]) + "..."
		}
		return summary
	}

	if ruleID == "" {
		return ""
	}

	shortID := ruleID
	if idx := strings.Index(shortID, "-"); idx != -1 {
		shortID = shortID[idx+1:]
	}
	shortID = strings.TrimSpace(strings.ReplaceAll(shortID, "-", " "))
	if shortID == "" {
		return ""
	}

	return strings.ToUpper(shortID[:1]) + shortID[1:]
}

// isBuiltInValidationRule checks if a rule ID is from the built-in validation category
func isBuiltInValidationRule(ruleID string) bool {
	return len(ruleID) > len("validation-") && ruleID[:len("validation-")] == "validation-"
}
