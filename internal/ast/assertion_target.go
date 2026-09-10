package ast

import (
	"fmt"

	"github.com/speakeasy-api/openapi/expression"
)

// Describes the target of an assertion, such as status code or response body.
type AssertionTarget string

const (
	AssertionTargetResponseBody AssertionTarget = "responseBody"
	AssertionTargetStatusCode   AssertionTarget = "statusCode"
)

// Converts an Expression to an AssertionTarget.
func AssertionTargetFromExpression(expr expression.Expression) (AssertionTarget, error) {
	switch expr.GetType() {
	case expression.ExpressionTypeResponse:
		_, responseTarget, _, _ := expr.GetParts()

		switch responseTarget {
		case "body":
			return AssertionTargetResponseBody, nil
		default:
			return "", fmt.Errorf("unsupported response target for assertion: %s", responseTarget)
		}
	case expression.ExpressionTypeStatusCode:
		return AssertionTargetStatusCode, nil
	default:
		return "", fmt.Errorf("unsupported target for assertion: %s", expr)
	}
}
