package ast

import (
	"fmt"

	"github.com/speakeasy-api/openapi/arazzo/criterion"
)

// Describes the assertion type, such as the operator in a condition.
type AssertionType string

const (
	AssertionTypeEqual    AssertionType = "equal"
	AssertionTypeNotEqual AssertionType = "notEqual"
	AssertionTypeRegex    AssertionType = "regex"
)

// Converts an Arazzo criterion operator to an AssertionType.
func AssertionTypeFromArazzoOperator(operator criterion.Operator) (AssertionType, error) {
	switch operator {
	case criterion.OperatorEQ:
		return AssertionTypeEqual, nil
	case criterion.OperatorNE:
		return AssertionTypeNotEqual, nil
	default:
		return "", fmt.Errorf("unsupported operator for assertion type: %s", operator)
	}
}
