package ast

import (
	"errors"
	"fmt"

	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/speakeasy-api/openapi/arazzo/criterion"
)

// Creates a new Polling from the given extension configuration.
func (o *BaseOperation) NewPollingFromExtensionsPolling(extensionConfiguration *extensions.Polling) (*Polling, error) {
	if o == nil {
		return nil, errors.New("operation is nil for polling conversion")
	}

	if extensionConfiguration == nil {
		return nil, nil
	}

	result := &Polling{}

	options, err := o.pollingOptionsFromExtensionsPollingOptions(extensionConfiguration.Options)

	if err != nil {
		return nil, fmt.Errorf("failed to convert polling options: %w", err)
	}

	result.Options = options

	return result, nil
}

// Creates a new PollingOptions from the given extension configuration.
func (o *BaseOperation) pollingOptionsFromExtensionsPollingOptions(extensionConfiguration extensions.PollingOptions) (PollingOptions, error) {
	if o == nil {
		return nil, errors.New("operation is nil for polling options conversion")
	}

	if extensionConfiguration == nil {
		return nil, nil
	}

	result := make(PollingOptions, 0, len(extensionConfiguration))

	for _, opt := range extensionConfiguration {
		option, err := o.pollingOptionFromExtensionsPollingOption(opt)

		if err != nil {
			return nil, err
		}

		result = append(result, option)
	}

	return result, nil
}

// Creates a new PollingOption from the given extension configuration.
func (o *BaseOperation) pollingOptionFromExtensionsPollingOption(extensionConfiguration *extensions.PollingOption) (*PollingOption, error) {
	if o == nil {
		return nil, errors.New("operation is nil for polling option conversion")
	}

	if extensionConfiguration == nil {
		return nil, nil
	}

	result := &PollingOption{
		DelaySeconds:    extensionConfiguration.DelaySeconds,
		IntervalSeconds: extensionConfiguration.IntervalSeconds,
		LimitCount:      extensionConfiguration.LimitCount,
		Name:            extensionConfiguration.Name,
	}

	failureCriteria, err := o.pollingAssertionsFromExtensionsPollingCriteria(extensionConfiguration.FailureCriteria)

	if err != nil {
		return nil, fmt.Errorf("failed to convert failure criteria: %w", err)
	}

	result.FailureCriteria = failureCriteria

	successCriteria, err := o.pollingAssertionsFromExtensionsPollingCriteria(extensionConfiguration.SuccessCriteria)

	if err != nil {
		return nil, fmt.Errorf("failed to convert success criteria: %w", err)
	}

	result.SuccessCriteria = successCriteria

	return result, nil
}

// Creates a new polling-based Assertions from the given extension
// configuration.
func (o *BaseOperation) pollingAssertionsFromExtensionsPollingCriteria(extensionCriteria extensions.PollingCriteria) (Assertions, error) {
	if o == nil {
		return nil, errors.New("operation is nil for polling assertions conversion")
	}

	if extensionCriteria == nil {
		return nil, nil
	}

	result := make(Assertions, 0, len(extensionCriteria))

	for _, pollingCriterion := range extensionCriteria {
		if pollingCriterion == nil || pollingCriterion.Condition == nil {
			continue
		}

		switch pollingCriterion.Type {
		case criterion.CriterionTypeRegex:
			assertion, err := o.pollingAssertionFromRegexCriterion(result, pollingCriterion)

			if err != nil {
				return nil, fmt.Errorf("failed to convert polling regex criterion: %w", err)
			}

			result = append(result, assertion)
		case criterion.CriterionTypeSimple:
			assertion, err := o.pollingAssertionFromArazzoCondition(result, pollingCriterion.Condition)

			if err != nil {
				return nil, fmt.Errorf("failed to convert polling criterion: %w", err)
			}

			result = append(result, assertion)
		default:
			return nil, fmt.Errorf("unsupported polling criterion type: %s", pollingCriterion.Type)
		}
	}

	return result, nil
}

// Creates a new polling-based Assertion from the given regex criterion.
// Only response-based assertions are valid for polling criteria.
func (o *BaseOperation) pollingAssertionFromRegexCriterion(priorAssertions Assertions, pollingCriterion *extensions.PollingCriterion) (*Assertion, error) {
	if pollingCriterion == nil || pollingCriterion.Condition == nil {
		return nil, errors.New("missing condition for regex assertion")
	}

	if pollingCriterion.Context == nil {
		return nil, errors.New("missing context for regex assertion")
	}

	assertionTarget, err := AssertionTargetFromExpression(*pollingCriterion.Context)

	if err != nil {
		return nil, err
	}

	switch assertionTarget {
	case AssertionTargetResponseBody:
		statusCodeAssertion := priorAssertions.FindAssertionByTarget(AssertionTargetStatusCode)

		if statusCodeAssertion == nil {
			return nil, errors.New("polling response body assertions require a prior status code assertion, such as: $statusCode == 200")
		}

		statusCode := fmt.Sprint(statusCodeAssertion.Value)

		// Create a synthetic condition with the context expression and regex pattern
		condition := &criterion.Condition{
			Expression: *pollingCriterion.Context,
			Value:      pollingCriterion.Condition.Value,
		}

		return o.pollingResponseBodyAssertion(statusCode, "", AssertionTypeRegex, condition)
	case AssertionTargetStatusCode:
		result := &Assertion{
			Target:     o.Response,
			TargetType: AssertionTargetStatusCode,
			Type:       AssertionTypeRegex,
			Value:      pollingCriterion.Condition.Value,
		}

		return result, nil
	default:
		return nil, fmt.Errorf("unsupported assertion target for polling regex: %s", assertionTarget)
	}
}

// Creates a new polling-based Assertion from the given Arazzo criterion
// condition. Only response-based assertions are valid for polling criteria, so
// this will return an error if the condition expression does not target the
// response.
func (o *BaseOperation) pollingAssertionFromArazzoCondition(priorAssertions Assertions, condition *criterion.Condition) (*Assertion, error) {
	if condition == nil {
		return nil, errors.New("missing condition for assertion")
	}

	assertionType, err := AssertionTypeFromArazzoOperator(condition.Operator)

	if err != nil {
		return nil, err
	}

	switch assertionType {
	case AssertionTypeEqual, AssertionTypeNotEqual:
		break // Supported, do nothing
	default:
		// Other assertion types can be supported in the future.
		return nil, fmt.Errorf("unsupported assertion type for polling: %s", assertionType)
	}

	assertionTarget, err := AssertionTargetFromExpression(condition.Expression)

	if err != nil {
		return nil, err
	}

	switch assertionTarget {
	case AssertionTargetResponseBody:
		statusCodeAssertion := priorAssertions.FindAssertionByTarget(AssertionTargetStatusCode)

		if statusCodeAssertion == nil {
			return nil, errors.New("polling response body assertions require a prior status code assertion, such as: $statusCode == 200")
		}

		statusCode := fmt.Sprint(statusCodeAssertion.Value)

		return o.pollingResponseBodyAssertion(statusCode, "", assertionType, condition)
	case AssertionTargetStatusCode:
		result := &Assertion{
			Target:     o.Response,
			TargetType: AssertionTargetStatusCode,
			Type:       assertionType,
			Value:      condition.Value,
		}

		return result, nil
	default:
		return nil, fmt.Errorf("unsupported assertion target for polling: %s", assertionTarget)
	}
}

// Creates a new polling response body Assertion.
func (o *BaseOperation) pollingResponseBodyAssertion(statusCode string, contentType string, assertionType AssertionType, condition *criterion.Condition) (*Assertion, error) {
	if o == nil {
		return nil, errors.New("operation is nil for polling response body assertion")
	}

	if o.Response == nil {
		return nil, errors.New("operation has no response for polling response body assertion")
	}

	return o.Response.pollingResponseBodyAssertion(statusCode, contentType, assertionType, condition)
}
