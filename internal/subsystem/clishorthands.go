package subsystem

import "sync"

// CLIShorthands collects the auto-assigned CLI flag shorthands that templates
// resolve for each operation, for persistence in the gen.lock cliShorthands
// section. Recording is ignored until Start, so template runs that must not
// replace the lock section (snippet-only generation) leave it untouched.
type CLIShorthands struct {
	mu         sync.Mutex
	operations map[string]any
}

func (c *CLIShorthands) Start() {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.operations = map[string]any{}
}

// Record stores one operation's resolution, omitting empty flags and retired
// maps so operations without shorthands do not add lock entries.
func (c *CLIShorthands) Record(operationID string, resolution map[string]any) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.operations == nil {
		return
	}
	entry := map[string]any{}
	for _, key := range []string{"flags", "retired"} {
		if value, ok := resolution[key].(map[string]any); ok && len(value) > 0 {
			entry[key] = value
		}
	}
	if len(entry) == 0 {
		delete(c.operations, operationID)
		return
	}
	c.operations[operationID] = entry
}

// Result returns the recorded section and whether recording was started.
func (c *CLIShorthands) Result() (map[string]any, bool) {
	if c == nil {
		return nil, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.operations, c.operations != nil
}
