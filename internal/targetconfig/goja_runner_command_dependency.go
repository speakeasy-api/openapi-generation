package targetconfig

import (
	"context"

	"github.com/speakeasy-api/openapi-generation/v2/internal/processrunner"
)

// Represents a [processrunner.CommandDependency] configuration in primitive Go
// types for simplifying the conversion from goja. Use the Convert method to
// finish the conversion for generator usage.
type GojaRunnerCommandDependency struct {
	Command              string
	InstallDocumentation string
	Version              *GojaRunnerCommandDependencyVersion
}

// Converts the GojaRunnerCommandDependency into its
// [processrunner.CommandDependency] form for the generator.
func (g GojaRunnerCommandDependency) Convert(ctx context.Context) (processrunner.CommandDependency, error) {
	result := processrunner.CommandDependency{
		Command:              g.Command,
		InstallDocumentation: g.InstallDocumentation,
	}

	commandVersionDependency, err := g.Version.Convert(ctx)

	if err != nil {
		return result, err
	}

	result.Version = commandVersionDependency

	return result, nil
}
