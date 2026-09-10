package targetconfig

import (
	"context"
	"regexp"

	"github.com/speakeasy-api/openapi-generation/v2/internal/processrunner"
)

// Represents a [processrunner.Dependencies] configuration in primitive Go types
// for simplifying the conversion from goja. Use the Convert method to finish
// the conversion for generator usage.
type GojaRunnerDependencies struct {
	ArgVariables         map[string]string
	Commands             []GojaRunnerCommandDependency
	EnvironmentVariables map[string]string
}

// Converts the GojaRunnerDependencies into its [processrunner.Dependencies]
// form for the generator.
func (g GojaRunnerDependencies) Convert(ctx context.Context) (processrunner.Dependencies, error) {
	result := processrunner.Dependencies{
		ArgVariables:         make(map[string]*regexp.Regexp, len(g.ArgVariables)),
		Commands:             make([]processrunner.CommandDependency, 0, len(g.Commands)),
		EnvironmentVariables: make(map[string]*regexp.Regexp, len(g.EnvironmentVariables)),
	}

	for argVarName, argVarString := range g.ArgVariables {
		argVarRegexp, err := regexp.Compile(argVarString)

		if err != nil {
			return result, err
		}

		result.ArgVariables[argVarName] = argVarRegexp
	}

	for _, gojaCommandDependency := range g.Commands {
		command, err := gojaCommandDependency.Convert(ctx)

		if err != nil {
			return result, err
		}

		result.Commands = append(result.Commands, command)
	}

	for envVarName, envVarString := range g.EnvironmentVariables {
		envVarRegexp, err := regexp.Compile(envVarString)

		if err != nil {
			return result, err
		}

		result.EnvironmentVariables[envVarName] = envVarRegexp
	}

	return result, nil
}
