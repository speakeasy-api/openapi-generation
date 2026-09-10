package targetconfig

import (
	"context"

	"github.com/speakeasy-api/openapi-generation/v2/internal/processrunner"
)

// Represents a [processrunner.Command] configuration in primitive Go types for
// simplifying the conversion from goja. Use the Convert method to finish
// the conversion for generator usage.
type GojaRunnerCommand struct {
	Command               string
	Args                  []string
	EnvironmentVariables  map[string]string
	IgnoreCommandNotFound bool
	StdoutDiscard         bool
	Parallel              bool
}

// Converts the GojaRunnerCommand into its [processrunner.Command] form for
// the generator.
func (g GojaRunnerCommand) Convert(ctx context.Context) (processrunner.Command, error) {
	result := processrunner.Command{
		Command:               g.Command,
		Args:                  g.Args,
		EnvironmentVariables:  g.EnvironmentVariables,
		IgnoreCommandNotFound: g.IgnoreCommandNotFound,
		StdoutDiscard:         g.StdoutDiscard,
		Parallel:              g.Parallel,
	}

	return result, nil
}
