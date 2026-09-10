package validation

import (
	stderrors "errors"
	"sort"
	"strings"
	"time"

	"github.com/speakeasy-api/openapi-generation/v2/internal/validation/report"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi/validation"
)

type Result struct {
	reportData                   *report.ValidationReportData // Internal report format (replacement for Vacuum motor.RuleSetExecutionResult)
	errs                         []error
	cliVersion                   string
	operationToFailingLineNumber map[string]int
}

func NewResult(reportData *report.ValidationReportData, errs []error, cliVersion string, operationToFailingLineNumber map[string]int) *Result {
	// If no report data provided, initialize empty
	if reportData == nil {
		reportData = &report.ValidationReportData{
			Results:    []*validation.Error{},
			Categories: []*report.RuleCategory{},
			Statistics: &report.ReportStatistics{},
			Generated:  time.Now(),
		}
	}

	return &Result{
		reportData:                   reportData,
		errs:                         errs,
		cliVersion:                   cliVersion,
		operationToFailingLineNumber: operationToFailingLineNumber,
	}
}

func (r *Result) Insert(res *validation.Error) {
	if filterJSONSchemaErrors(res.Rule, res.UnderlyingError.Error()) {
		return
	}

	r.reportData.Results = append(r.reportData.Results, res)

	// Update statistics
	switch res.Severity {
	case validation.SeverityError:
		r.reportData.Statistics.TotalErrors++
	case validation.SeverityWarning:
		r.reportData.Statistics.TotalWarnings++
	case validation.SeverityHint:
		r.reportData.Statistics.TotalHints++
	}
}

func filterJSONSchemaErrors(ruleID string, message string) bool {
	if ruleID != "validate-json-schema" {
		return false
	}

	messagesToFilter := []string{
		"missing properties: 'description'",           // Ignore missing description property as we consider that completely optional
		"/response/required",                          // Ignore response needing to be marked as required
		"minimum 1 items required, but found 0 items", // Ignore empty `required` arrays as we just consider all properties optional in this case
	}

	for _, msg := range messagesToFilter {
		if strings.Contains(strings.ToLower(message), msg) {
			return true
		}
	}

	return false
}

func (r *Result) AddError(err error) {
	r.errs = append(r.errs, err)
}

func (r *Result) GetValidationErrors() []error {
	// Sort errors by severity (hint, warning, error) then by line number
	sorted := make([]error, len(r.errs))
	copy(sorted, r.errs)

	// Define severity order: hint (1) < warn (2) < error (3)
	severityOrder := func(sev errors.Severity) int {
		switch sev {
		case errors.SeverityHint:
			return 1
		case errors.SeverityWarn:
			return 2
		case errors.SeverityError:
			return 3
		default:
			return 4 // Unknown severity comes last
		}
	}

	sort.SliceStable(sorted, func(i, j int) bool {
		var errI, errJ *errors.ValidationError
		okI := stderrors.As(sorted[i], &errI)
		okJ := stderrors.As(sorted[j], &errJ)

		// If either error can't be cast, maintain original order
		if !okI || !okJ {
			return false
		}

		// First, compare by severity (hint < warning < error)
		orderI := severityOrder(errI.Severity)
		orderJ := severityOrder(errJ.Severity)
		if orderI != orderJ {
			return orderI < orderJ
		}

		// Within same severity, sort by line number (ascending)
		// Handle nil nodes
		lineI := 0
		if errI.Node != nil {
			lineI = errI.Node.Line
		}
		lineJ := 0
		if errJ.Node != nil {
			lineJ = errJ.Node.Line
		}

		return lineI < lineJ
	})

	return sorted
}

func (r *Result) HasFatalErrors() bool {
	hasErrors := false
	vErrs := r.GetValidationErrors()
	for _, e := range vErrs {
		var vErr *errors.ValidationError
		if errors.As(e, &vErr) {
			if vErr.Severity == errors.SeverityError {
				hasErrors = true
			}
		} else {
			hasErrors = true
		}
	}
	return hasErrors
}

func (r *Result) GenerateReport() []byte {
	// Use new report generation with internal data
	// This will be properly implemented in Phase 5
	r.reportData.Generated = time.Now()
	rep := report.NewHTMLReport(r.reportData)
	return rep.GenerateReport(false, r.cliVersion)
}

func (r *Result) GetValidOperations() []string {
	var validOperations []string
	for operationID, line := range r.operationToFailingLineNumber {
		if line == -1 {
			validOperations = append(validOperations, operationID)
		}
	}
	return validOperations
}

func (r *Result) GetInvalidOperations() []string {
	var invalidOperations []string
	for operationID, line := range r.operationToFailingLineNumber {
		if line != -1 {
			invalidOperations = append(invalidOperations, operationID)
		}
	}
	return invalidOperations
}
