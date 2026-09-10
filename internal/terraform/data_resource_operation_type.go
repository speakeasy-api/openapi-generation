package terraform

// Describes an individual Terraform data resource operation type.
type DataResourceOperationType string

const (
	// Represents a read operation for a Terraform data resource.
	DataResourceOperationTypeRead DataResourceOperationType = "read"
)

var (
	// Collection of valid Terraform data resource operation types.
	DataResourceOperationTypes = []DataResourceOperationType{
		DataResourceOperationTypeRead,
	}
)
