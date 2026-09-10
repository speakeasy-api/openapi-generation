package changes

// IsOperationMCPEnabled returns true if the operation is not disabled for MCP.
// Operations without MCP extensions or with Disabled == false are considered enabled.
func IsOperationMCPEnabled(m MethodDiff) bool {
	if m.Operation == nil || m.Operation.Extensions == nil || m.Operation.Extensions.MCP == nil {
		return true
	}
	return !m.Operation.Extensions.MCP.Disabled
}
