package terraform

// Returns true if the given schema version is valid.
func IsResourceSchemaVersionValid(version int64) bool {
	return version >= 0
}
