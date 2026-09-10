package targetconfig

import (
	"context"
)

// Represents a LintConfiguration in primitive Go types for simplifying the
// conversion from goja. Use the Convert method to finish the conversion for
// generator usage.
type GojaLintConfiguration struct {
	Runner *GojaRunnerConfiguration
}

// Converts the GojaLintConfiguration into its LintConfiguration form for
// the generator.
func (c *GojaLintConfiguration) Convert(ctx context.Context) (*LintConfiguration, error) {
	if c == nil {
		return nil, nil
	}

	runner, err := c.Runner.Convert(ctx)

	if err != nil {
		return nil, err
	}

	result := &LintConfiguration{
		Runner: runner,
	}

	return result, nil
}
