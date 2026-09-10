package ast

// Describes an assertion.
type Assertion struct {
	// Type of assertion, such as the operator in a condition.
	Type AssertionType

	// Target type of the assertion, such as status code or response body.
	TargetType AssertionTarget

	// Target of the assertion, such as the computed FieldDef.
	Target any

	// Value to assert against.
	Value any
}

// Creates a deep copy of the Assertion.
func (a *Assertion) Clone() *Assertion {
	if a == nil {
		return nil
	}

	cloned := &Assertion{
		Type:       a.Type,
		TargetType: a.TargetType,
		Target:     a.Target, // Assuming Target is immutable or a pointer to an immutable type
		Value:      a.Value,  // Assuming Value is immutable or a pointer to an immutable type
	}

	return cloned
}
