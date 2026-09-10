package changes

import (
	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi/sequencedmap"
)

// diffSuccessResponses compares only the success responses of two operations
func diffSuccessResponses(a, b *ast.Operation) TypeDefDiffResult {
	return diffResponsesInner(a, b, false)
}

// diffErrorResponses compares only the error responses of two operations
func diffErrorResponses(a, b *ast.Operation) TypeDefDiffResult {
	return diffResponsesInner(a, b, true)
}

// diffResponsesInner compares responses filtered by error status
func diffResponsesInner(a, b *ast.Operation, errorOnly bool) TypeDefDiffResult {
	// Get status codes for both operations
	statusCodesA := getStatusCodes(a, errorOnly)
	statusCodesB := getStatusCodes(b, errorOnly)

	// If both operations have no responses, they're equal
	if statusCodesA.Len() == 0 && statusCodesB.Len() == 0 {
		return TypeDefDiffResult{Equal: true, Path: []PathSegment{}}
	}

	// If one has responses and the other doesn't, they're different
	if statusCodesA.Len() == 0 || statusCodesB.Len() == 0 {
		return TypeDefDiffResult{
			Equal:      false,
			Path:       []PathSegment{},
			Reason:     DiffReasonKind,
			IsBreaking: false,
		}
	}

	// Initialize result
	ret := TypeDefDiffResult{Equal: true, Path: []PathSegment{}}

	// Removed status codes
	for statusCode := range statusCodesA.Keys() {
		if _, exists := statusCodesB.Get(statusCode); !exists {
			// Status code removed
			diff := createStatusCodeDiff(statusCode, errorOnly, DiffReasonFieldRemoved)
			ret = mergeTypeDefDiffResult(ret, diff, ast.DataTypeAny)
		}
	}

	// Added status codes
	for statusCode := range statusCodesB.Keys() {
		if _, exists := statusCodesA.Get(statusCode); !exists {
			// Status code added
			diff := createStatusCodeDiff(statusCode, errorOnly, DiffReasonFieldAdded)
			ret = mergeTypeDefDiffResult(ret, diff, ast.DataTypeAny)
		}
	}

	// Status codes that exist in both
	for statusCode, respA := range statusCodesA.All() {
		respB, exists := statusCodesB.Get(statusCode)
		if !exists {
			continue
		}

		contentDiff := diffContentTypes(statusCode, respA, respB)
		ret = mergeTypeDefDiffResult(ret, contentDiff, ast.DataTypeAny)
	}

	return ret
}

// getStatusCodes extracts status codes from responses filtered by error type
func getStatusCodes(op *ast.Operation, errorOnly bool) *sequencedmap.Map[string, *ast.SubResponse] {
	statusMap := sequencedmap.New[string, *ast.SubResponse]()

	if op == nil || op.Response == nil {
		return statusMap
	}

	for _, resp := range op.Response.Responses {
		if resp.Error != errorOnly {
			continue
		}

		for _, code := range resp.Code {
			statusMap.Set(code, resp)
		}
	}

	return statusMap
}

// buildContentTypeMap creates a map of content type to ResponseBodyContent
func buildContentTypeMap(resp *ast.SubResponse) *sequencedmap.Map[string, *ast.ResponseBodyContent] {
	contentMap := sequencedmap.New[string, *ast.ResponseBodyContent]()
	for _, content := range resp.Content {
		contentMap.Set(content.ContentType, content)
	}
	return contentMap
}

// diffContentTypes compares content types between two responses with the same status code
func diffContentTypes(statusCode string, respA, respB *ast.SubResponse) TypeDefDiffResult {
	ret := TypeDefDiffResult{Equal: true, Path: []PathSegment{}}
	contentMapA := buildContentTypeMap(respA)
	contentMapB := buildContentTypeMap(respB)

	basePath := PathSegments{
		{Type: PathSegmentResponseStatus, Name: statusCode},
	}

	// Check for removed content types
	for contentType := range contentMapA.Keys() {
		if _, exists := contentMapB.Get(contentType); !exists {
			diff := TypeDefDiffResult{
				Equal:      false,
				Path:       basePath.Append(PathSegment{Type: PathSegmentResponseContentType, Name: contentType}),
				Reason:     DiffReasonFieldRemoved,
				IsBreaking: true,
			}
			ret = mergeTypeDefDiffResult(ret, diff, ast.DataTypeAny)
		}
	}

	// Check for added content types
	for contentType := range contentMapB.Keys() {
		if _, exists := contentMapA.Get(contentType); !exists {
			diff := TypeDefDiffResult{
				Equal:      false,
				Path:       basePath.Append(PathSegment{Type: PathSegmentResponseContentType, Name: contentType}),
				Reason:     DiffReasonFieldAdded,
				IsBreaking: false,
			}
			ret = mergeTypeDefDiffResult(ret, diff, ast.DataTypeAny)
		}
	}

	// Check content types that exist in both
	for contentType, contentA := range contentMapA.All() {
		contentB, exists := contentMapB.Get(contentType)
		if !exists {
			continue
		}

		// Compare the response content schemas
		if contentA.Content != nil && contentB.Content != nil &&
			contentA.Content.Type != nil && contentB.Content.Type != nil {
			params := DiffTypeDefsParams{
				A:         contentA.Content.Type,
				B:         contentB.Content.Type,
				Path:      basePath.Append(PathSegment{Type: PathSegmentResponseContentType, Name: contentType}),
				IsRequest: false,
			}
			diff := DiffTypeDefs(params)
			if !diff.Equal {
				ret = mergeTypeDefDiffResult(ret, diff, ast.DataTypeAny)
			}
		}
	}

	return ret
}

// createStatusCodeDiff creates a diff for an added or removed status code
func createStatusCodeDiff(statusCode string, isError bool, reason DiffReason) TypeDefDiffResult {
	return TypeDefDiffResult{
		Equal: false,
		Path: []PathSegment{
			{Type: PathSegmentResponseStatus, Name: statusCode},
		},
		Reason:     reason,
		IsBreaking: reason == DiffReasonFieldRemoved || (reason == DiffReasonFieldAdded && !isError),
	}
}
