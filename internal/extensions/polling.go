package extensions

import (
	"errors"
	"fmt"

	"github.com/speakeasy-api/openapi/arazzo/criterion"
	"github.com/speakeasy-api/openapi/expression"
	"gopkg.in/yaml.v3"
)

// Describes parsed and normalized x-speakeasy-polling extension configuration.
type Polling struct {
	// Collection of polling options.
	Options PollingOptions `json:"options" yaml:"options"`
}

// Clone creates a deep copy of the Polling
func (p *Polling) Clone() *Polling {
	if p == nil {
		return nil
	}

	cloned := &Polling{
		Options: p.Options.Clone(),
	}

	return cloned
}

// Collection of polling criterion.
type PollingCriteria []*PollingCriterion

// Clone creates a deep copy of the PollingCriteria.
func (c PollingCriteria) Clone() PollingCriteria {
	if c == nil {
		return nil
	}

	cloned := make(PollingCriteria, 0, len(c))

	for _, criterion := range c {
		cloned = append(cloned, criterion.Clone())
	}

	return cloned
}

// Describes a single polling criterion, such as a target condition.
type PollingCriterion struct {
	// Condition for the polling criterion. For simple type criterion, this is
	// typically a full expression such as `$statusCode == 200`. For regex type
	// criterion, this is the regular expression pattern.
	Condition *criterion.Condition `json:"condition,omitempty" yaml:"condition,omitempty"`

	// Context is the expression to the value to be evaluated. Required for
	// regex type criterion.
	Context *expression.Expression `json:"context,omitempty" yaml:"context,omitempty"`

	// Type is the type of criterion. Defaults to CriterionTypeSimple.
	Type criterion.CriterionType `json:"type,omitempty" yaml:"type,omitempty"`
}

// Clone creates a deep copy of the PollingCriterion.
func (c *PollingCriterion) Clone() *PollingCriterion {
	if c == nil {
		return nil
	}

	cloned := &PollingCriterion{
		Type: c.Type,
	}

	if c.Condition != nil {
		// Deep copy the Condition
		cloned.Condition = &criterion.Condition{
			Expression: c.Condition.Expression,
			Operator:   c.Condition.Operator,
			Value:      c.Condition.Value,
		}
	}

	if c.Context != nil {
		ctx := *c.Context
		cloned.Context = &ctx
	}

	return cloned
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for PollingCriterion.
func (c *PollingCriterion) UnmarshalYAML(node *yaml.Node) error {
	var raw struct {
		Condition *string `json:"condition,omitempty" yaml:"condition,omitempty"`
		Context   *string `json:"context,omitempty" yaml:"context,omitempty"`
		Type      *string `json:"type,omitempty" yaml:"type,omitempty"`
	}

	if err := node.Decode(&raw); err != nil {
		return err
	}

	if raw.Condition == nil {
		return errors.New("condition is required for polling criterion")
	}

	// Parse and validate Type, defaulting to simple
	criterionType := criterion.CriterionTypeSimple

	if raw.Type != nil {
		switch criterion.CriterionType(*raw.Type) {
		case criterion.CriterionTypeSimple:
			criterionType = criterion.CriterionTypeSimple
		case criterion.CriterionTypeRegex:
			criterionType = criterion.CriterionTypeRegex
		default:
			return fmt.Errorf("unsupported criterion type '%s': only 'simple' and 'regex' are supported", *raw.Type)
		}
	}

	result := PollingCriterion{
		Type: criterionType,
	}

	// Parse Context if provided
	if raw.Context != nil {
		ctx := expression.Expression(*raw.Context)
		result.Context = &ctx
	}

	// Handle based on criterion type
	switch criterionType {
	case criterion.CriterionTypeRegex:
		if result.Context == nil {
			return errors.New("context is required for regex criterion type")
		}

		// For regex type, store the pattern in a synthetic Condition
		result.Condition = &criterion.Condition{
			Value: *raw.Condition,
		}
	default:
		// For simple type, parse using the Arazzo criterion parser
		arazzoCriterion := &criterion.Criterion{
			Condition: *raw.Condition,
		}

		parsedCondition, err := arazzoCriterion.GetCondition()
		if err != nil {
			return fmt.Errorf("failed to parse condition '%s': %w", *raw.Condition, err)
		}

		if parsedCondition == nil {
			return fmt.Errorf("condition '%s' could not be parsed", *raw.Condition)
		}

		result.Condition = parsedCondition
	}

	*c = result

	return nil
}

// Describes a single polling option.
type PollingOption struct {
	// Delay in seconds before polling calls begin. Defaults to 1.
	DelaySeconds *int64 `json:"delaySeconds,omitempty" yaml:"delaySeconds,omitempty"`

	// Descibes immediate failure criteria for the polling option. When all
	// matching criteria are met (AND boolean), the operation will immediately
	// return an error.
	FailureCriteria PollingCriteria `json:"failureCriteria,omitempty" yaml:"failureCriteria,omitempty"`

	// Interval between polling calls in seconds. Defaults to 1.
	IntervalSeconds *int64 `json:"intervalSeconds,omitempty" yaml:"intervalSeconds,omitempty"`

	// Name of the polling option.
	Name string `json:"name" yaml:"name"`

	// Number of polling calls not matching the FailureCriteria or
	// SuccessCriteria before returning a timeout error. Defaults to 60.
	LimitCount *int64 `json:"limitCount,omitempty" yaml:"limitCount,omitempty"`

	// Descibes success criteria for the polling option. When all matching
	// criteria are met (AND boolean), the operation will return successfully.
	SuccessCriteria PollingCriteria `json:"successCriteria,omitempty" yaml:"successCriteria,omitempty"`
}

// Clone creates a deep copy of the PollingOption.
func (o *PollingOption) Clone() *PollingOption {
	if o == nil {
		return nil
	}

	cloned := &PollingOption{
		DelaySeconds:    clonePtr(o.DelaySeconds),
		FailureCriteria: o.FailureCriteria.Clone(),
		IntervalSeconds: clonePtr(o.IntervalSeconds),
		LimitCount:      clonePtr(o.LimitCount),
		SuccessCriteria: o.SuccessCriteria.Clone(),
		Name:            o.Name,
	}

	return cloned
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for PollingOption.
func (o *PollingOption) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("expected mapping node, got %v", node.Kind)
	}

	// Initialize with default values
	result := PollingOption{
		DelaySeconds:    ptr(int64(1)),
		IntervalSeconds: ptr(int64(1)),
		LimitCount:      ptr(int64(60)),
	}

	// Parse the mapping node content directly
	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]

		if keyNode.Kind != yaml.ScalarNode {
			continue
		}

		switch keyNode.Value {
		case "name":
			if err := valueNode.Decode(&result.Name); err != nil {
				return fmt.Errorf("failed to decode name: %w", err)
			}
		case "delaySeconds":
			var delay int64
			if err := valueNode.Decode(&delay); err != nil {
				return fmt.Errorf("failed to decode delaySeconds: %w", err)
			}
			if delay < 0 {
				return errors.New("delaySeconds must be greater than or equal to zero")
			}
			result.DelaySeconds = &delay
		case "intervalSeconds":
			var intervalSeconds int64
			if err := valueNode.Decode(&intervalSeconds); err != nil {
				return fmt.Errorf("failed to decode intervalSeconds: %w", err)
			}
			if intervalSeconds < 0 {
				return errors.New("intervalSeconds must be greater than or equal to zero")
			}
			result.IntervalSeconds = &intervalSeconds
		case "limitCount":
			var limitCount int64
			if err := valueNode.Decode(&limitCount); err != nil {
				return fmt.Errorf("failed to decode limitCount: %w", err)
			}
			if limitCount <= 0 {
				return errors.New("limitCount must be greater than zero")
			}
			result.LimitCount = &limitCount
		case "successCriteria":
			if err := valueNode.Decode(&result.SuccessCriteria); err != nil {
				return fmt.Errorf("failed to decode successCriteria: %w", err)
			}

		case "failureCriteria":
			if err := valueNode.Decode(&result.FailureCriteria); err != nil {
				return fmt.Errorf("failed to decode failureCriteria: %w", err)
			}
		}
	}

	if result.Name == "" {
		return errors.New("name must be a non-empty string")
	}

	if len(result.SuccessCriteria) == 0 {
		return errors.New("at least one successCriteria must be defined")
	}

	*o = result

	return nil
}

// Collection of PollingOption.
type PollingOptions []*PollingOption

// Clone creates a deep copy of the PollingOptions.
func (o PollingOptions) Clone() PollingOptions {
	if o == nil {
		return nil
	}

	cloned := make(PollingOptions, 0, len(o))

	for _, option := range o {
		cloned = append(cloned, option.Clone())
	}

	return cloned
}

// Handles parsing of the x-speakeasy-polling extension from the given OpenAPI
// extensions map.
func (e *Extensions) HandlePollingExtension(extensions OAExtensions) (*Polling, error) {
	if extensions.Len() == 0 {
		return nil, nil
	}

	yamlNode, ok := e.findExtension(extensions, ExtPolling)

	if !ok {
		return nil, nil
	}

	result, err := e.parsePolling(yamlNode)
	if err != nil {
		return nil, fmt.Errorf("failed to parse extension value: %w", err)
	}

	return result, nil
}

// Parses the given YAML node into a Polling structure.
func (e *Extensions) parsePolling(node *yaml.Node) (*Polling, error) {
	switch node.Kind {
	case yaml.SequenceNode:
		result := &Polling{
			Options: make(PollingOptions, 0, len(node.Content)),
		}

		for index, content := range node.Content {
			pollingOption := new(PollingOption)

			if err := pollingOption.UnmarshalYAML(content); err != nil {
				return nil, fmt.Errorf("failed to parse polling option %d: %w", index, err)
			}

			result.Options = append(result.Options, pollingOption)
		}

		return result, nil
	default:
		return nil, fmt.Errorf("unsupported YAML node kind: %v", node.Kind)
	}
}
