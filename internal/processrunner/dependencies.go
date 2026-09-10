package processrunner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Dependencies for the runner commands.
type Dependencies struct {
	// Mapping of required argument variables names to value regular expressions
	// for argument templating.
	ArgVariables map[string]*regexp.Regexp

	// Required commands.
	Commands []CommandDependency

	// Mapping of required environment variable names to value regular
	// expressions.
	EnvironmentVariables map[string]*regexp.Regexp
}

// Verifies the argument variables, commands, and environment variables.
func (d Dependencies) check(ctx context.Context, dir string, argVars map[string]string) (string, error) {
	if err := d.checkArgVariables(argVars); err != nil {
		return "", err
	}

	if err := d.checkEnvironmentVariables(os.Environ()); err != nil {
		return "", err
	}

	return d.checkCommands(ctx, dir)
}

// Verifies the given argument variables against the dependencies.
func (d Dependencies) checkArgVariables(argVars map[string]string) error {
	if len(d.ArgVariables) == 0 {
		return nil
	}

	var errs error

	for name, pattern := range d.ArgVariables {
		got, ok := argVars[name]

		if !ok {
			errs = errors.Join(errs, fmt.Errorf("missing argument variable: %s", name))
			continue
		}

		if !pattern.MatchString(got) {
			errs = errors.Join(errs, fmt.Errorf("argument variable %s did not match expected pattern %s, got: %s", name, pattern, got))
		}
	}

	return errs
}

// Verifies the given commands against the dependencies.
func (d Dependencies) checkCommands(ctx context.Context, dir string) (string, error) {
	var errs error
	var outputBuilder strings.Builder

	for _, commandDependency := range d.Commands {
		output, err := commandDependency.check(ctx, dir)

		if err != nil {
			errs = errors.Join(errs, err)
			_, _ = outputBuilder.WriteString(output)
		}
	}

	if errs != nil {
		return outputBuilder.String(), errs
	}

	return "", nil
}

// Verifies the given environment variables against the dependencies.
func (d Dependencies) checkEnvironmentVariables(environ []string) error {
	if len(d.EnvironmentVariables) == 0 {
		return nil
	}

	envVars := environMap(environ)
	var errs error

	for name, pattern := range d.ArgVariables {
		got, ok := envVars[name]

		if !ok {
			errs = errors.Join(errs, fmt.Errorf("missing argument variable: %s", name))
			continue
		}

		if !pattern.MatchString(got) {
			errs = errors.Join(errs, fmt.Errorf("argument variable %s did not match expected pattern %s, got: %s", name, pattern, got))
		}
	}

	return errs
}
