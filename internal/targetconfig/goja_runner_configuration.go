package targetconfig

import (
	"context"

	"github.com/speakeasy-api/openapi-generation/v2/internal/processrunner"
)

// Represents a [processrunner.Runner] configuration in primitive Go types for
// simplifying the conversion from goja. Use the Convert method to finish
// the conversion for generator usage.
type GojaRunnerConfiguration struct {
	Commands     []GojaRunnerCommand
	Dependencies *GojaRunnerDependencies
}

// Converts the GojaRunnerConfiguration into its [processrunner.Runner] form for
// the generator.
func (g *GojaRunnerConfiguration) Convert(ctx context.Context) (*processrunner.Runner, error) {
	if g == nil {
		return nil, nil
	}

	result := &processrunner.Runner{
		Commands: make([]processrunner.Command, 0, len(g.Commands)),
	}

	for _, gojaCommand := range g.Commands {
		command, err := gojaCommand.Convert(ctx)

		if err != nil {
			return nil, err
		}

		result.Commands = append(result.Commands, command)
	}

	dependencies, err := g.Dependencies.Convert(ctx)

	if err != nil {
		return nil, err
	}

	result.Dependencies = dependencies

	return result, nil
}
