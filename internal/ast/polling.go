package ast

// Describes polling configuration derived from the x-speakeasy-polling
// extension configuration.
type Polling struct {
	// Collection of polling options.
	Options PollingOptions `json:"options" yaml:"options"`
}

// Creates a deep copy of the Polling
func (p *Polling) Clone() *Polling {
	if p == nil {
		return nil
	}

	cloned := &Polling{
		Options: p.Options.Clone(),
	}

	return cloned
}

// Describes a single polling option.
type PollingOption struct {
	// Delay in seconds before polling calls begin. Defaults to 1.
	DelaySeconds *int64 `json:"delaySeconds,omitempty" yaml:"delaySeconds,omitempty"`

	// Descibes immediate failure criteria for the polling option. When all
	// matching criteria are met (AND boolean), the operation will immediately
	// return an error.
	FailureCriteria Assertions `json:"failureCriteria,omitempty" yaml:"failureCriteria,omitempty"`

	// Interval between polling calls in seconds. Defaults to 1.
	IntervalSeconds *int64 `json:"intervalSeconds,omitempty" yaml:"intervalSeconds,omitempty"`

	// Name of the polling option.
	Name string `json:"name" yaml:"name"`

	// Number of polling calls not matching the FailureCriteria or
	// SuccessCriteria before returning a timeout error. Defaults to 60.
	LimitCount *int64 `json:"limitCount,omitempty" yaml:"limitCount,omitempty"`

	// Descibes success criteria for the polling option. When all matching
	// criteria are met (AND boolean), the operation will return successfully.
	SuccessCriteria Assertions `json:"successCriteria,omitempty" yaml:"successCriteria,omitempty"`
}

// Creates a deep copy of the PollingOption.
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

// Collection of PollingOption.
type PollingOptions []*PollingOption

// Creates a deep copy of the PollingOptions.
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
