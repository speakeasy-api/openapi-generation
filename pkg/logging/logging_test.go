package logging_test

import (
	"context"
	"errors"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestLogWarning(t *testing.T) {
	testErr := errors.New("test error")
	var loggedErr error

	logger := logging.NewWarningLogger(false, func(err error) {
		loggedErr = err
	})

	ctx := logging.WithWarningLogger(t.Context(), logger)

	fromCtx, ok := logging.WarningLoggerFrom(ctx)
	require.True(t, ok)
	require.Equal(t, logger, fromCtx)

	logging.LogWarning(ctx, "test message", testErr)
	require.Equal(t, testErr, loggedErr)
}

// mockStep implements Step for testing step tracking behavior.
type mockStep struct {
	name   string
	status string
}

func (m *mockStep) Succeed() { m.status = "success" }
func (m *mockStep) Fail()    { m.status = "failed" }
func (m *mockStep) Skip()    { m.status = "skipped" }

// mockLogger implements Logger for testing.
type mockLogger struct {
	steps []*mockStep // Track all steps created, in order
}

func (m *mockLogger) Debug(msg string, fields ...zapcore.Field)   {}
func (m *mockLogger) Info(msg string, fields ...zapcore.Field)    {}
func (m *mockLogger) Warn(msg string, fields ...zapcore.Field)    {}
func (m *mockLogger) Error(msg string, fields ...zapcore.Field)   {}
func (m *mockLogger) Github(msg string)                           {}
func (m *mockLogger) With(fields ...zapcore.Field) logging.Logger { return m }
func (m *mockLogger) Scope(name string) logging.Logger            { return m }
func (m *mockLogger) StartStep(msg string) logging.Step {
	step := &mockStep{name: msg}
	m.steps = append(m.steps, step)
	return step
}

func TestWithStepCtx_AllSubstepsSkip_ParentSkips(t *testing.T) {
	mock := &mockLogger{}
	ctx := logging.With(t.Context(), mock)

	err := logging.WithStepCtx(ctx, "Parent", func(ctx context.Context) error {
		// Create a substep that skips
		step := logging.StartStepCtx(ctx, "Child")
		step.Skip()
		return nil
	})

	require.NoError(t, err)
	require.Len(t, mock.steps, 2, "Should have created 2 steps (parent + child)")
	require.Equal(t, "Parent", mock.steps[0].name)
	require.Equal(t, "Child", mock.steps[1].name)
	require.Equal(t, "skipped", mock.steps[0].status, "Parent step should skip when all substeps skip")
	require.Equal(t, "skipped", mock.steps[1].status, "Child step should be skipped")
}

func TestWithStepCtx_SubstepSucceeds_ParentSucceeds(t *testing.T) {
	mock := &mockLogger{}
	ctx := logging.With(t.Context(), mock)

	err := logging.WithStepCtx(ctx, "Parent", func(ctx context.Context) error {
		// Create a substep that succeeds
		step := logging.StartStepCtx(ctx, "Child")
		step.Succeed()
		return nil
	})

	require.NoError(t, err)
	require.Len(t, mock.steps, 2, "Should have created 2 steps (parent + child)")
	require.Equal(t, "success", mock.steps[0].status, "Parent step should succeed when a substep succeeds")
	require.Equal(t, "success", mock.steps[1].status, "Child step should be success")
}

func TestWithStepCtx_NoSubsteps_ParentSucceeds(t *testing.T) {
	mock := &mockLogger{}
	ctx := logging.With(t.Context(), mock)

	err := logging.WithStepCtx(ctx, "Parent", func(ctx context.Context) error {
		// No substeps created
		return nil
	})

	require.NoError(t, err)
	require.Len(t, mock.steps, 1, "Should have created 1 step (parent only)")
	require.Equal(t, "success", mock.steps[0].status, "Parent step should succeed when no substeps")
}
