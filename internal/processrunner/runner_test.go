package processrunner

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"go.uber.org/zap"
)

func testContext() context.Context {
	return logging.With(context.Background(), logging.NewLoggerFromZap(zap.NewNop()))
}

func echoCmd(msg string) Command {
	return Command{
		Command: "echo",
		Args:    []string{msg},
	}
}

func parallelEchoCmd(msg string) Command {
	return Command{
		Command:  "echo",
		Args:     []string{msg},
		Parallel: true,
	}
}

func failCmd() Command {
	return Command{
		Command: "false",
	}
}

func parallelFailCmd() Command {
	return Command{
		Command:  "false",
		Parallel: true,
	}
}

func TestRunCommands_SequentialOnly(t *testing.T) {
	runner := Runner{
		Commands: []Command{
			echoCmd("hello"),
			echoCmd("world"),
		},
	}

	stdout, _, err := runner.RunCommands(testContext(), t.TempDir(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "hello") || !strings.Contains(stdout, "world") {
		t.Fatalf("expected stdout to contain hello and world, got: %q", stdout)
	}
}

func TestRunCommands_SequentialStopsOnError(t *testing.T) {
	runner := Runner{
		Commands: []Command{
			echoCmd("before"),
			failCmd(),
			echoCmd("after"),
		},
	}

	_, _, err := runner.RunCommands(testContext(), t.TempDir(), nil)
	if err == nil {
		t.Fatal("expected error from failing command")
	}
}

func TestRunCommands_AllParallel(t *testing.T) {
	runner := Runner{
		Commands: []Command{
			parallelEchoCmd("alpha"),
			parallelEchoCmd("beta"),
			parallelEchoCmd("gamma"),
		},
	}

	stdout, _, err := runner.RunCommands(testContext(), t.TempDir(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "alpha") || !strings.Contains(stdout, "beta") || !strings.Contains(stdout, "gamma") {
		t.Fatalf("expected all parallel outputs, got: %q", stdout)
	}
}

func TestRunCommands_MixedSequentialAndParallel(t *testing.T) {
	runner := Runner{
		Commands: []Command{
			echoCmd("first"),
			parallelEchoCmd("p1"),
			parallelEchoCmd("p2"),
			echoCmd("last"),
		},
	}

	stdout, _, err := runner.RunCommands(testContext(), t.TempDir(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, expected := range []string{"first", "p1", "p2", "last"} {
		if !strings.Contains(stdout, expected) {
			t.Fatalf("expected stdout to contain %q, got: %q", expected, stdout)
		}
	}
}

func TestRunCommands_ParallelErrorAggregation(t *testing.T) {
	runner := Runner{
		Commands: []Command{
			parallelEchoCmd("ok"),
			parallelFailCmd(),
			parallelFailCmd(),
		},
	}

	_, _, err := runner.RunCommands(testContext(), t.TempDir(), nil)
	if err == nil {
		t.Fatal("expected error from failing parallel commands")
	}

	// errors.Join separates errors with newlines — two failures should
	// produce at least two error segments.
	errStr := err.Error()
	if strings.Count(errStr, "exit status") < 2 {
		t.Fatalf("expected at least 2 joined errors, got: %q", errStr)
	}
}

func TestRunCommands_ParallelActuallyConcurrent(t *testing.T) {
	// Each parallel command sleeps for 200ms. If run sequentially, 3 commands
	// would take >= 600ms. If parallel, they should finish in ~200ms.
	runner := Runner{
		Commands: []Command{
			{Command: "sleep", Args: []string{"0.2"}, Parallel: true},
			{Command: "sleep", Args: []string{"0.2"}, Parallel: true},
			{Command: "sleep", Args: []string{"0.2"}, Parallel: true},
		},
	}

	start := time.Now()
	_, _, err := runner.RunCommands(testContext(), t.TempDir(), nil)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// With true parallelism, 3x200ms should complete well under 500ms.
	// Use 500ms as threshold to allow for CI overhead.
	if elapsed >= 500*time.Millisecond {
		t.Fatalf("parallel commands took %v, expected < 500ms — commands may not be running concurrently", elapsed)
	}
}

func TestRunCommands_SequentialAfterParallelError(t *testing.T) {
	runner := Runner{
		Commands: []Command{
			parallelEchoCmd("ok"),
			parallelFailCmd(),
			echoCmd("should-not-run"),
		},
	}

	_, _, err := runner.RunCommands(testContext(), t.TempDir(), nil)
	if err == nil {
		t.Fatal("expected error from parallel group")
	}
}
