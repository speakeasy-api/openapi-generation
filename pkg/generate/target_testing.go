package generate

import (
	"bufio"
	"bytes"
	"context"
	goerrors "errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/speakeasy-api/openapi-generation/v2/internal/env"
	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/internal/instrument"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/testcontainers/testcontainers-go"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Runs target testing only. This assumes that the target has been previously
// generated into the outDir.
func (g *Generator) RunTargetTesting(ctx context.Context, target string, outDir string) error {
	g.warningLogger = logging.NewWarningLogger(g.validationOnly, func(err error) {
		var vErr *errors.ValidationError
		if errors.As(err, &vErr) {
			g.result.Insert(vErr.GetResult())
		}
	})

	ctx = logging.With(ctx, g.log)
	ctx = logging.WithWarningLogger(ctx, g.warningLogger)

	var span trace.Span

	if g.tracer != nil && g.verboseOutput {
		shutdown, err := instrument.SetupOTelSDK(ctx, "stdout")
		if err == nil {
			defer shutdown(ctx) //nolint:errcheck
		}
		ctx, span = g.tracer.Start(ctx, "Generator.RunTargetTesting")
		defer span.End()
	}

	if g.targetConfig == nil {
		if err := g.Init(ctx, target, outDir); err != nil {
			return err
		}
	}

	return logging.WithStepCtx(ctx, "Run Target Testing", func(ctx context.Context) error {
		return g.runTargetTesting(ctx)
	})
}

// Runs target-defined testing processes from targetconfig.TestingConfiguration.
func (g *Generator) runTargetTesting(ctx context.Context) error {
	if !g.shouldRunTargetTesting(ctx) {
		return nil
	}

	var err error
	var outText string

	ctx, span := g.tracer.Start(ctx, "Generator.testTarget")
	defer func() {
		span.RecordError(err)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	// Detect if the mock server exists and start it if it does
	mockServerDirectory, err := g.getMockServerDirectory(ctx, g.target)
	if err != nil {
		logging.From(ctx).Error(fmt.Sprintf("Could not get mock server directory: %s", err))
	}

	if mockServerDirectory != "" {
		mockServerDirectory = filepath.Join(g.outDir, mockServerDirectory)
	}

	mockServerEnabled := true
	if g.subsystem.Config.Generation.MockServer != nil {
		mockServerEnabled = !g.subsystem.Config.Generation.MockServer.Disabled
	}

	mockServerExists := false
	if mockServerDirectory != "" && mockServerEnabled {
		if _, err := os.Stat(filepath.Join(mockServerDirectory, "main.go")); err == nil {
			mockServerExists = true
		}
	}

	if mockServerExists && !g.disableMockServer {
		logging.From(ctx).Info("Starting mock server")

		mockServerProcess, err := newMockServerProcess(ctx, mockServerDirectory)
		if err != nil {
			logging.From(ctx).Error(fmt.Sprintf("Could not start mock server: %s", err))
			//nolint:gocritic
			os.Exit(1)
		}
		defer func() {
			logging.From(ctx).Info("Shutting down mock server")
			if err := mockServerProcess.Close(ctx); err != nil {
				logging.From(ctx).Error(fmt.Sprintf("Could not shut down mock server cleanly: %s", err))
				os.Exit(1)
			}
		}()

		url, err := mockServerProcess.URL(ctx)
		if err != nil {
			logging.From(ctx).Error(fmt.Sprintf("Could not get mock server URL: %s", err))
			os.Exit(1)
		}

		_ = os.Setenv("TEST_SERVER_URL", url)
		// Special case for the review SDKs TODO: probably want to do this conditionally
		_ = os.Setenv("TEST_URL", url)
	}

	argVars := map[string]string{}
	workingDir := g.outDir
	outText, err = g.targetConfig.Testing.Runner.Run(ctx, workingDir, argVars)
	if err != nil {
		logging.From(ctx).Warn("cannot test target - " + outText)
		return err
	}

	return nil
}

// Returns true if the target has the tests feature enabled and test running
// configuration.
func (g *Generator) shouldRunTargetTesting(ctx context.Context) bool {
	// Ensure internal tests always run for all targets, regardless of tests feature.
	if !env.IsDebug() && !g.subsystem.Features.IsFeatureSupported(ctx, features.FeatureTests) {
		return false
	}

	if g.targetConfig.Testing == nil || g.targetConfig.Testing.Runner == nil {
		return false
	}

	return true
}

func newMockServerProcess(ctx context.Context, mockServerDirectory string) (*mockServerProcess, error) {
	p, err := newMockServerDockerProcess(ctx, mockServerDirectory)
	if err != nil {
		logging.From(ctx).Warn(err.Error() + ", falling back to go run")

		p, err = newMockServerGoProcess(ctx, mockServerDirectory)
		if err != nil {
			return nil, err
		}
	}

	return p, nil
}

func newMockServerGoProcess(ctx context.Context, mockServerDirectory string) (*mockServerProcess, error) {
	if _, err := exec.LookPath("go"); err != nil {
		return nil, errors.New("go is not installed")
	}

	cmd := exec.Command("go", "run", "main.go")
	cmd.Dir = mockServerDirectory

	var outb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = cmd.Stdout
	assignProcessGroup(cmd)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sigChan
		_ = cmd.Process.Signal(syscall.SIGTERM)
	}()

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start mockserver: %w", err)
	}

	// read output until "starting server" is printed
	reader := bufio.NewReader(&outb)
	timeout := time.After(5 * time.Minute)

	for {
		lineChannel := make(chan string)
		errorChannel := make(chan error)

		// Read line in a goroutine so we can timeout
		go func() {
			line, err := reader.ReadString('\n')
			if err != nil {
				errorChannel <- err
				return
			}
			lineChannel <- line
		}()

		// Wait for either a line to be read, an error, or timeout
		select {
		case line := <-lineChannel:
			logging.From(ctx).Info(line)
			if strings.Contains(line, "starting server") {
				return &mockServerProcess{
					command: cmd,
				}, nil
			}
		case err := <-errorChannel:
			if goerrors.Is(err, io.EOF) {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			killProcess(cmd)
			return nil, fmt.Errorf("failed to read mockserver stdout: %w", err)
		case <-timeout:
			killProcess(cmd)
			return nil, errors.New("timeout waiting for server to start")
		}
	}
}

type mockServerProcess struct {
	container Container
	command   *exec.Cmd
}

// Container defines the interface needed by mockServerProcess
// Implementations are in the platform-specific files
type Container interface {
	Endpoint(ctx context.Context, s string) (string, error)
	Terminate(ctx context.Context, opts ...testcontainers.TerminateOption) error
}

func (m *mockServerProcess) URL(ctx context.Context) (string, error) {
	endpoint := ""

	if m.container != nil {
		var err error
		endpoint, err = m.container.Endpoint(ctx, "")
		if err != nil {
			return "", err
		}
	} else if m.command != nil {
		endpoint = "localhost:18080"
	}

	return "http://" + endpoint, nil
}

func (m *mockServerProcess) Close(ctx context.Context) error {
	if m.container != nil {
		return m.container.Terminate(ctx)
	} else if m.command != nil {
		killProcess(m.command)
	}

	return nil
}
