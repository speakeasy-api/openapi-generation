package terraform

// Describes an individual Terraform action operation type.
type ActionOperationType string

const (
	// Represents an invoke operation for a Terraform action.
	ActionOperationTypeInvoke ActionOperationType = "invoke"
)

var (
	// Collection of valid Terraform action operation types.
	ActionOperationTypes = []ActionOperationType{
		ActionOperationTypeInvoke,
	}
)
