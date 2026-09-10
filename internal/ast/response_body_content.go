package ast

import (
	"errors"
	"fmt"

	"github.com/speakeasy-api/openapi/arazzo/criterion"
	"github.com/speakeasy-api/openapi/jsonpointer"
)

// ResponseBodyContent represents the type of content that is returned in a response
type ResponseBodyContent struct {
	SerializationMethod string    `yaml:",omitempty"` // The serialization method for the content
	ContentType         string    `yaml:",omitempty"` // The content type of the content
	Content             *FieldDef `yaml:",omitempty"` // The content field returned in the response type
	UsageExample        bool      `yaml:",omitempty"` // Whether or not the content should be used as a usage example
	Examples            Examples  `yaml:",omitempty"` // The examples of the content
	SSESentinel         string    `yaml:",omitempty"`
}

// Clone creates a deep copy of ResponseBodyContent
func (r *ResponseBodyContent) Clone() *ResponseBodyContent {
	if r == nil {
		return nil
	}

	return &ResponseBodyContent{
		Content:             r.Content.Clone(),
		ContentType:         r.ContentType,
		Examples:            r.Examples.Clone(),
		SerializationMethod: r.SerializationMethod,
		SSESentinel:         r.SSESentinel,
		UsageExample:        r.UsageExample,
	}
}

// Runs any ResponseBodyContent matcher against this ResponseBodyContent.
func (r *ResponseBodyContent) Match(matchers Matchers) error {
	if matchers.ResponseBodyContent != nil {
		return matchers.ResponseBodyContent(r)
	}

	return nil
}

// Returns true if ResponseBodyContent is equal to given ResponseBodyContent.
func (r *ResponseBodyContent) IsEqual(other *ResponseBodyContent) bool {
	if r == nil && other == nil || r == other {
		return true
	}
	if r == nil && other != nil || other == nil && r != nil {
		return false
	}
	if r.SerializationMethod != other.SerializationMethod {
		return false
	}
	if r.ContentType != other.ContentType {
		return false
	}
	if r.UsageExample != other.UsageExample {
		return false
	}
	return r.Content.IsEqual(other.Content)
}

// Returns a new polling response body Assertion based on ResponseBodyContent.
func (r *ResponseBodyContent) pollingResponseBodyAssertion(assertionType AssertionType, condition *criterion.Condition) (*Assertion, error) {
	if r == nil {
		return nil, errors.New("response body content is nil for polling response body assertion")
	}

	conditionValueStr := fmt.Sprint(condition.Value)
	var example *Example

	// Enable use case where the user is trying to check any value was received.
	// The null or empty string value might not actually represent the correct
	// type but when Example is not set, the templating logic will short circuit
	// to use existence checks.
	if assertionType != AssertionTypeNotEqual || (conditionValueStr != "null" && conditionValueStr != `""`) {
		var err error

		example, err = NewExampleFromString(condition.Expression.String(), "", conditionValueStr)
		if err != nil {
			return nil, fmt.Errorf("unable to create value example from %s for response body polling assertion: %w", condition.Value, err)
		}
	}

	jp := condition.Expression.GetJSONPointer()
	responseBodyAssertion := ResponseBodyAssertion{
		Content: r,
		Path:    string(jp),
		Value:   example,
	}
	result := &Assertion{
		Target:     r.Content,
		TargetType: AssertionTargetResponseBody,
		Type:       assertionType,
		Value:      responseBodyAssertion,
	}

	if jp == "" {
		return result, nil
	}

	jpTarget, err := jsonpointer.GetTarget(r.Content, jp)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve JSON Pointer %s for response body polling assertion: %w", jp, err)
	}

	if jpTargetTyp, ok := jpTarget.(*FieldDef); ok {
		if jpTargetTyp.IsAdditionalProperties {
			jpTarget = jpTargetTyp.Type.ItemType
		}
	}

	result.Target = jpTarget

	return result, nil
}
