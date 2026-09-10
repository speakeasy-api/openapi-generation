package targetconfig

import (
	"context"
)

// Represents a TestingConfiguration in primitive Go types for simplifying the
// conversion from goja. Use the Convert method to finish the conversion for
// generator usage.
type GojaTestingConfiguration struct {
	Compile *GojaCompileConfiguration
	Runner  *GojaRunnerConfiguration
}

// Converts the GojaTestingConfiguration into its TestingConfiguration form for
// the generator.
func (c *GojaTestingConfiguration) Convert(ctx context.Context) (*TestingConfiguration, error) {
	if c == nil {
		return nil, nil
	}

	compile, err := c.Compile.Convert(ctx)

	if err != nil {
		return nil, err
	}

	runner, err := c.Runner.Convert(ctx)

	if err != nil {
		return nil, err
	}

	result := &TestingConfiguration{
		Compile: compile,
		Runner:  runner,
	}

	return result, nil
}
