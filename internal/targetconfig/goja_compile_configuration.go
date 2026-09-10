package targetconfig

import (
	"context"
)

// Represents a CompileConfiguration in primitive Go types for simplifying the
// conversion from goja. Use the Convert method to finish the conversion for
// generator usage.
type GojaCompileConfiguration struct {
	Runner *GojaRunnerConfiguration
}

// Converts the GojaCompileConfiguration into its CompileConfiguration form for
// the generator.
func (c *GojaCompileConfiguration) Convert(ctx context.Context) (*CompileConfiguration, error) {
	if c == nil {
		return nil, nil
	}

	runner, err := c.Runner.Convert(ctx)

	if err != nil {
		return nil, err
	}

	result := &CompileConfiguration{
		Runner: runner,
	}

	return result, nil
}
