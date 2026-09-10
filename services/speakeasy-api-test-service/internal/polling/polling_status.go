package polling

// Describes the polling status, such as "completed".
type PollingStatus string

const (
	// PollingStatusCompleted indicates the polling operation has completed successfully.
	PollingStatusCompleted PollingStatus = "completed"

	// PollingStatusFailed indicates the polling operation has failed.
	PollingStatusFailed PollingStatus = "failed"

	// PollingStatusPending indicates the polling operation is still pending.
	PollingStatusPending PollingStatus = "pending"

	// PollingStatusRunning indicates the polling operation is currently running.
	PollingStatusRunning PollingStatus = "running"
)
