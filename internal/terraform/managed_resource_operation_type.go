package terraform

// Describes an individual Terraform managed resource operation type.
type ManagedResourceOperationType string

const (
	// Represents a create operation for a Terraform managed resource.
	ManagedResourceOperationTypeCreate ManagedResourceOperationType = "create"

	// Represents a delete operation for a Terraform managed resource.
	ManagedResourceOperationTypeDelete ManagedResourceOperationType = "delete"

	// Represents a read operation for a Terraform managed resource.
	ManagedResourceOperationTypeRead ManagedResourceOperationType = "read"

	// Represents an update operation for a Terraform managed resource.
	ManagedResourceOperationTypeUpdate ManagedResourceOperationType = "update"
)

var (
	// Collection of valid Terraform managed resource operation types.
	ManagedResourceOperationTypes = []ManagedResourceOperationType{
		ManagedResourceOperationTypeCreate,
		ManagedResourceOperationTypeRead,
		ManagedResourceOperationTypeUpdate,
		ManagedResourceOperationTypeDelete,
	}
)
