package validation

import (
	"errors"
	"fmt"

	"github.com/speakeasy-api/openapi-generation/v2/internal/validation/sanitization"
	"github.com/speakeasy-api/openapi/validation"
)

type operationFilenameConflictChecker struct {
	uniqueFileNames []sanitizedOperationNameResult
}

func newOperationFilenameConflictChecker() *operationFilenameConflictChecker {
	return &operationFilenameConflictChecker{
		uniqueFileNames: []sanitizedOperationNameResult{},
	}
}

func (f *operationFilenameConflictChecker) Check(operationInfo operationNameInfo, rule Rule) *validation.Error {
	opID := operationInfo.OperationID
	node := operationInfo.OperationIDNode
	hasOpId := operationInfo.HasOperationID

	sanitizedFileNameResult := sanitization.GetSanitizedFileNameResult(opID)

	if sanitizedFileName, otherFileName, other := findOperationNameConflict(sanitizedFileNameResult, f.uniqueFileNames); other != nil {
		var problemStatement, matching string

		if hasOpId {
			problemStatement = fmt.Sprintf("operationId `%s` (`%s`)", sanitizedFileNameResult.Original, sanitizedFileName)
			matching = fmt.Sprintf("operationId `%s` (`%s`)", other.Result.Original, otherFileName)
		} else {
			problemStatement = fmt.Sprintf("`%s` (`%s`) without `operationId`", sanitizedFileNameResult.Original, sanitizedFileName)
			matching = fmt.Sprintf("operationId `%s` (`%s`)", other.Result.Original, otherFileName)
		}

		message := fmt.Sprintf("`%s` in `%s` - `%s` will collide with `%s` [line `%d`] when converted to file name", problemStatement, operationInfo.HTTPPath, operationInfo.HTTPMethod, matching, other.Line)

		return &validation.Error{
			Rule:            rule.ID(),
			Severity:        rule.DefaultSeverity(),
			Node:            node,
			UnderlyingError: errors.New(message),
		}
	} else {
		f.uniqueFileNames = append(f.uniqueFileNames, sanitizedOperationNameResult{
			Result: sanitizedFileNameResult,
			Line:   node.Line,
		})
	}

	return nil
}
