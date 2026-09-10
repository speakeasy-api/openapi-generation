package ast

// Collection of Assertion.
type Assertions []*Assertion

// Creates a deep copy of the Assertions.
func (a Assertions) Clone() Assertions {
	if a == nil {
		return nil
	}

	cloned := make(Assertions, 0, len(a))

	for _, assertion := range a {
		cloned = append(cloned, assertion.Clone())
	}

	return cloned
}

// If found, returns the Assertion with given AssertionTarget.
func (a Assertions) FindAssertionByTarget(target AssertionTarget) *Assertion {
	for _, assertion := range a {
		if assertion.TargetType == target {
			return assertion
		}
	}

	return nil
}
