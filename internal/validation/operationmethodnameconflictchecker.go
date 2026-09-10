package validation

import (
	"errors"
	"fmt"
	"slices"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/internal/validation/sanitization"
	"github.com/speakeasy-api/openapi/validation"
)

type methodNameConflictChecker struct {
	uniqueMethodNames map[string][]sanitizedOperationNameResult
	target            types.Target
}

func newMethodNameConflictChecker(target types.Target) *methodNameConflictChecker {
	return &methodNameConflictChecker{
		target:            target,
		uniqueMethodNames: make(map[string][]sanitizedOperationNameResult),
	}
}

func (m *methodNameConflictChecker) Check(operationInfo operationNameInfo, rule Rule) []*validation.Error {
	failures := []*validation.Error{}

	opID := operationInfo.OperationID
	methodNameOverride := operationInfo.MethodNameOverride
	methodGroupOverride := operationInfo.MethodGroupOverride
	tags := operationInfo.Tags
	hasOpId := operationInfo.HasOperationID
	node := operationInfo.OperationIDNode
	httpPath := operationInfo.HTTPPath
	httpMethod := operationInfo.HTTPMethod

	methodName := opID
	if methodNameOverride != "" {
		methodName = methodNameOverride
		if operationInfo.MethodNameOverrideNode != nil {
			node = operationInfo.MethodNameOverrideNode
		}
	}

	// Call sanitization.GetSanitizedMethodNameResult inside the Check method
	sanitizedMethodNameResult := sanitization.GetSanitizedMethodNameResult(methodName)

	if methodGroupOverride != "" {
		tags = []string{methodGroupOverride}
	}

	if len(tags) == 0 {
		tags = append(tags, "")
	}

	sanitizedGroupNames := []string{}
	for _, tag := range tags {
		if !slices.Contains(sanitizedGroupNames, tag) {
			sanitizedGroupNames = append(sanitizedGroupNames, tag)
		}
	}

	for _, group := range sanitizedGroupNames {
		if _, ok := m.uniqueMethodNames[group]; !ok {
			m.uniqueMethodNames[group] = []sanitizedOperationNameResult{}
		}

		if sanitizedMethodName, _, other := findOperationNameConflict(sanitizedMethodNameResult, m.uniqueMethodNames[group]); other != nil {

			targetSpecificMethodName, _ := sanitizedMethodNameResult.Results.Get(m.target.Target)
			if targetSpecificMethodName != "" {
				sanitizedMethodName = targetSpecificMethodName
			}
			fullMethodName := "sdk." + sanitizedMethodName + "()"

			if group != "" {
				fullMethodName = "sdk." + group + "." + sanitizedMethodName + "()"
			}

			message := fmt.Sprintf("method name `%s` will collide with operation `%s` [line `%d`]", fullMethodName, other.GetPath(), other.Line)
			switch {
			case !hasOpId:
				message += ", try adding `operationId`"
			case methodNameOverride == "":
				message += ", try using `x-speakeasy-name-override`"
			case group == "":
				message += ", try using `x-speakeasy-group`"
			default:
				message += ", try adjusting `x-speakeasy-name-override`"
			}

			failures = append(failures, &validation.Error{
				Rule:            rule.ID(),
				Severity:        rule.DefaultSeverity(),
				Node:            node,
				UnderlyingError: errors.New(message),
			})
		} else {
			m.uniqueMethodNames[group] = append(m.uniqueMethodNames[group], sanitizedOperationNameResult{
				Line:       node.Line,
				Result:     sanitizedMethodNameResult,
				HTTPPath:   httpPath,
				HTTPMethod: httpMethod,
			})
		}
	}

	return failures
}
