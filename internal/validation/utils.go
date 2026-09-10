package validation

import (
	"github.com/speakeasy-api/openapi-generation/v2/internal/validation/sanitization"
)

func findOperationNameConflict(currentResult sanitization.Result, uniques []sanitizedOperationNameResult) (string, string, *sanitizedOperationNameResult) {
	for _, unique := range uniques {
		sanitized, original := currentResult.GetConflicting(unique.Result)
		if sanitized != "" && original != "" {
			return sanitized, original, &unique
		}
	}

	return "", "", nil
}

func findNameConflict(currentResult sanitization.Result, uniques []nameReference) (string, string, *nameReference) {
	for _, unique := range uniques {
		sanitized, original := currentResult.GetConflicting(unique.Result)
		if sanitized != "" && original != "" {
			return sanitized, original, &unique
		}
	}

	return "", "", nil
}

func findEnumConflict(currentResult sanitization.Result, uniques []enumNameReference) (string, string, *enumNameReference) {
	for _, unique := range uniques {
		sanitized, original := currentResult.GetConflicting(unique.Result)
		if sanitized != "" && original != "" {
			return sanitized, original, &unique
		}
	}

	return "", "", nil
}
