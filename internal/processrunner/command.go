package processrunner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
	"text/template"
	"time"

	"github.com/speakeasy-api/openapi-generation/v2/internal/env"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"go.uber.org/zap"
)

const CanceledCommandWaitDelay = 100 * time.Millisecond

// An individual command to be ran.
type Command struct {
	// Command name.
	Command string

	// Arguments to pass to the command.
	//
	// If runtime argument variables are given, each argument is ran through
	// [text/template] to enable variable substition. For example, templating
	// an Example variable via {{ .Example }} anywhere in the string.
	//
	// Available variables are:
	//
	//  - PackageVersion: Generator selected new version for target.
	//
	// It is highly recommended add any templating variables in the runner
	// dependencies to prevent confusing errors.
	Args []string

	// Mapping of environment variable names to values for overriding.
	EnvironmentVariables map[string]string

	// When enabled, return without error if the command is not found. This
	// allows optional tooling commands to be ran, but only if present.
	IgnoreCommandNotFound bool

	// When enabled, discards stdout output.
	StdoutDiscard bool

	// When enabled, consecutive commands with Parallel set to true will be
	// ran concurrently. All parallel commands in a group must complete before
	// the next non-parallel command runs.
	Parallel bool
}

// Runs the command, returning the stdout, stderr, and any run errors.
func (c Command) run(ctx context.Context, dir string, argVars map[string]string) (string, string, error) {
	args, err := c.templateArgVariables(argVars)
	if err != nil {
		return "", "", fmt.Errorf("error templating argument variables: %w", err)
	}

	cmdStr := fmt.Sprintf("%s %s", c.Command, strings.Join(args, " "))
	step := logging.From(ctx).StartStep(cmdStr)

	environ := os.Environ()
	if len(c.EnvironmentVariables) > 0 {
		environ = slices.Concat(environ, environSlice(c.EnvironmentVariables))
	}
	environ = childEnvironment(environ)

	var stdout, stderr bytes.Buffer

	cmd := exec.CommandContext(ctx, c.Command, args...)
	cmd.Dir = dir
	cmd.Env = environ

	if !c.StdoutDiscard {
		cmd.Stdout = &stdout
	}

	cmd.Stderr = &stderr

	if ctx.Done() != nil {
		// If the context is cancelled while cmd is running, exec will
		// automatically send a shutdown signal (c.Process.Kill()).
		// Setting a non-zero WaitDelay prevents indefinitely waiting for:
		// - a child process that fails to listen to the shutdown signal
		// - I/O pipes left unclosed or still being read until EOF
		cmd.WaitDelay = CanceledCommandWaitDelay
	}

	logging.From(ctx).Info(fmt.Sprintf("Running command: %s %s", c.Command, strings.Join(args, " ")), zap.String("path", dir))

	start := time.Now()
	if err := cmd.Run(); err != nil {
		if env.IsDebug() {
			logging.From(ctx).Debug(fmt.Sprintf("Command failed after %s: %s", time.Since(start), cmdStr))
		}

		if c.IgnoreCommandNotFound && errors.Is(err, exec.ErrNotFound) {
			logging.From(ctx).Debug(fmt.Sprintf("optional command %s not found, skipping", c.Command))
			step.Skip()
			return "", "", nil
		}

		step.Fail()
		return stdout.String(), stderr.String(), err
	}

	if env.IsDebug() {
		logging.From(ctx).Debug(stdout.String())
		logging.From(ctx).Debug(fmt.Sprintf("Command completed in %s: %s", time.Since(start), cmdStr))
	}

	step.Succeed()
	return stdout.String(), stderr.String(), nil
}

// If runtime argument variables are given, each argument is ran through
// [text/template] to enable variable substition.
func (c Command) templateArgVariables(argVars map[string]string) ([]string, error) {
	if len(argVars) == 0 {
		return c.Args, nil
	}

	var errs error

	result := slices.Clone(c.Args)

	for index, arg := range c.Args {
		if !strings.Contains(arg, "{{") {
			continue
		}

		tmplName := fmt.Sprintf("arg[%d]", index)
		tmpl, err := template.New(tmplName).Parse(arg)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		var templatedArg strings.Builder

		if err := tmpl.Execute(&templatedArg, argVars); err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		result[index] = templatedArg.String()
	}

	result = slices.DeleteFunc(result, func(v string) bool { return v == "" })

	return result, errs
}
