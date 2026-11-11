package subsystem

import "sync"

// CLIShorthands collects the auto-assigned CLI flag shorthands that templates
// resolve for each operation, for persistence in the gen.lock cliShorthands
// section. Recording is ignored until Start, so template runs that must not
// replace the lock section (snippet-only generation) leave it untouched.
type CLIShorthands struct {
	mu         sync.Mutex
	operations map[string]map[string]any
}

func (c *CLIShorthands) Start() {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.operations = map[string]map[string]any{}
}

// Record stores the resolution of one operation's request serialization
// method, omitting empty flags and retired maps so operations without
// shorthands do not add lock entries. A split operation keeps an empty method
// entry, which stops a later single command from inheriting a sibling's entry.
func (c *CLIShorthands) Record(operationID, serializationMethod string, resolution map[string]any, split bool) {
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
	methods := c.operations[operationID]
	if len(entry) == 0 && !split {
		delete(methods, serializationMethod)
		if len(methods) == 0 {
			delete(c.operations, operationID)
		}
		return
	}
	if methods == nil {
		methods = map[string]any{}
		c.operations[operationID] = methods
	}
	methods[serializationMethod] = entry
}

// Result returns the recorded section and whether recording was started.
func (c *CLIShorthands) Result() (map[string]any, bool) {
	if c == nil {
		return nil, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.operations == nil {
		return nil, false
	}
	result := make(map[string]any, len(c.operations))
	for operationID, methods := range c.operations {
		for _, entry := range methods {
			if len(entry.(map[string]any)) > 0 {
				result[operationID] = methods
				break
			}
		}
	}
	return result, true
}
