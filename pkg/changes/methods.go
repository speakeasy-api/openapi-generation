package changes

// diffMethodResult holds the result of comparing two methods
type diffMethodResult struct {
	HasChanges            bool
	ArgumentsResult       TypeDefDiffResult
	SuccessResponseResult TypeDefDiffResult
	ErrorResponseResult   TypeDefDiffResult
}

// diffMethod compares two operations and returns the method diff result
func diffMethod(a, b *methodInfo) diffMethodResult {
	// Compare arguments/request
	argResult := diffArguments(a.Operation, b.Operation)

	// Compare success responses
	successResult := diffSuccessResponses(a.Operation, b.Operation)

	// Compare error responses
	errorResult := diffErrorResponses(a.Operation, b.Operation)

	return diffMethodResult{
		HasChanges:            !argResult.Equal || !successResult.Equal || !errorResult.Equal,
		ArgumentsResult:       argResult,
		SuccessResponseResult: successResult,
		ErrorResponseResult:   errorResult,
	}
}
