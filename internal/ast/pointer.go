package ast

// Returns a pointer to a copy of the given pointer value.
func clonePtr[T any](v *T) *T {
	if v == nil {
		return nil
	}

	cloned := *v

	return &cloned
}

// Returns a pointer to the given value.
func ptr[T any](v T) *T {
	return &v
}
