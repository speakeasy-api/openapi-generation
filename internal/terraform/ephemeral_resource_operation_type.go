package terraform

// Describes an individual Terraform ephemeral resource operation type.
type EphemeralResourceOperationType string

const (
	// Represents a close operation for a Terraform ephemeral resource.
	EphemeralResourceOperationTypeClose EphemeralResourceOperationType = "close"

	// Represents an open operation for a Terraform ephemeral resource.
	EphemeralResourceOperationTypeOpen EphemeralResourceOperationType = "open"
)

var (
	// Collection of valid Terraform ephemeral resource operation types.
	EphemeralResourceOperationTypes = []EphemeralResourceOperationType{
		EphemeralResourceOperationTypeClose,
		EphemeralResourceOperationTypeOpen,
	}
)
