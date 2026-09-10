package processrunner

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
)

// A set of commands to be ran by the generator, such as target-specific
// compilation and testing commands.
type Runner struct {
	// The commands to run, if dependencies are met.
	Commands []Command

	// Command dependency verification, before running the commands. Repeated
	// commands only require a single dependency.
	Dependencies Dependencies
}

// Verifies the Runner is prepared to run the commands by checking the dependencies.
func (r Runner) CheckDependencies(ctx context.Context, dir string, argVars map[string]string) (string, error) {
	return r.Dependencies.check(ctx, dir, argVars)
}

// Checks dependencies and runs all processes. Use other methods, such as
// CheckDependencies and RunCommands, for additional control if necessary.
func (r Runner) Run(ctx context.Context, dir string, argVars map[string]string) (string, error) {
	if output, err := r.CheckDependencies(ctx, dir, argVars); err != nil {
		return output, err
	}

	stdout, stderr, err := r.RunCommands(ctx, dir, argVars)

	if err != nil {
		return fmt.Sprintf("failed running commands: \n%s\n%s", stdout, stderr), err
	}

	return "", nil
}

// Runs the commands. Consecutive commands with Parallel set to true are ran
// concurrently as a group. All commands in a parallel group must complete
// before the next non-parallel command runs. Errors from parallel commands
// are collected and joined.
func (r Runner) RunCommands(ctx context.Context, dir string, argVars map[string]string) (string, string, error) {
	var stdoutBuilder, stderrBuilder strings.Builder

	i := 0
	for i < len(r.Commands) {
		command := r.Commands[i]

		if !command.Parallel {
			stdout, stderr, err := command.run(ctx, dir, argVars)
			if err != nil {
				return stdout, stderr, err
			}

			_, _ = stdoutBuilder.WriteString(stdout)
			_, _ = stderrBuilder.WriteString(stderr)
			i++

			continue
		}

		// Collect consecutive parallel commands into a group.
		groupStart := i
		for i < len(r.Commands) && r.Commands[i].Parallel {
			i++
		}

		group := r.Commands[groupStart:i]

		stdout, stderr, err := runParallelGroup(ctx, group, dir, argVars)

		_, _ = stdoutBuilder.WriteString(stdout)
		_, _ = stderrBuilder.WriteString(stderr)

		if err != nil {
			return stdoutBuilder.String(), stderrBuilder.String(), err
		}
	}

	return stdoutBuilder.String(), stderrBuilder.String(), nil
}

// commandResult holds the output from a single parallel command execution.
type commandResult struct {
	stdout string
	stderr string
	err    error
}

// runParallelGroup runs all commands concurrently, waits for all to complete,
// then returns combined output and any joined errors in original command order.
func runParallelGroup(ctx context.Context, commands []Command, dir string, argVars map[string]string) (string, string, error) {
	results := make([]commandResult, len(commands))

	var wg sync.WaitGroup
	wg.Add(len(commands))

	for idx, cmd := range commands {
		go func(i int, c Command) {
			defer wg.Done()

			stdout, stderr, err := c.run(ctx, dir, argVars)
			results[i] = commandResult{stdout: stdout, stderr: stderr, err: err}
		}(idx, cmd)
	}

	wg.Wait()

	var stdoutBuilder, stderrBuilder strings.Builder
	var errs error

	for _, r := range results {
		_, _ = stdoutBuilder.WriteString(r.stdout)
		_, _ = stderrBuilder.WriteString(r.stderr)

		if r.err != nil {
			errs = errors.Join(errs, r.err)
		}
	}

	return stdoutBuilder.String(), stderrBuilder.String(), errs
}
