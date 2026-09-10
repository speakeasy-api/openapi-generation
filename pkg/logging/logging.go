package logging

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/speakeasy-api/openapi-generation/v2/internal/env"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextKey string

func (c contextKey) String() string {
	return "logger-" + string(c)
}

const (
	loggerContextKey        = contextKey("context")
	warningLoggerContextKey = contextKey("warning")
	stepTrackerContextKey   = contextKey("stepTracker")
)

// StepStatus represents the outcome of a progress step.
type StepStatus string

const (
	StepStatusPending StepStatus = "pending"
	StepStatusSuccess StepStatus = "success"
	StepStatusFailed  StepStatus = "failed"
	StepStatusSkipped StepStatus = "skipped"
)

// Step is a handle for tracking progress of an operation.
// Call Succeed(), Fail(), or Skip() when the operation completes.
type Step interface {
	Succeed()
	Fail()
	Skip()
}

// stepTracker tracks the outcomes of substeps within a WithStep block.
// It is stored in context and used to determine if a parent step should skip
// when all its substeps have skipped.
type stepTracker struct {
	hasSubsteps bool // Were any substeps created?
	allSkipped  bool // Are ALL substeps skipped? (initialized true)
}

// trackedStep wraps a Step and updates the tracker when the step completes.
type trackedStep struct {
	Step
	tracker *stepTracker
}

func (ts *trackedStep) Succeed() {
	ts.tracker.hasSubsteps = true
	ts.tracker.allSkipped = false
	ts.Step.Succeed()
}

func (ts *trackedStep) Fail() {
	ts.tracker.hasSubsteps = true
	ts.tracker.allSkipped = false
	ts.Step.Fail()
}

func (ts *trackedStep) Skip() {
	ts.tracker.hasSubsteps = true
	// Do not change allSkipped - it remains true if all substeps skip
	ts.Step.Skip()
}

// withStepTracker returns a new context with a step tracker attached.
func withStepTracker(ctx context.Context, tracker *stepTracker) context.Context {
	return context.WithValue(ctx, stepTrackerContextKey, tracker)
}

// stepTrackerFrom returns the step tracker from context, if any.
func stepTrackerFrom(ctx context.Context) *stepTracker {
	if tracker, ok := ctx.Value(stepTrackerContextKey).(*stepTracker); ok {
		return tracker
	}
	return nil
}

type Logger interface {
	Debug(msg string, fields ...zapcore.Field)
	Info(msg string, fields ...zapcore.Field)
	Warn(msg string, fields ...zapcore.Field)
	Error(msg string, fields ...zapcore.Field)
	Github(msg string) // Prints only when running in GitHub Actions
	With(fields ...zapcore.Field) Logger

	// Scope returns a Logger with a path prefix for StartStep calls.
	// Scopes can be nested: logger.Scope("A").Scope("B").StartStep("C")
	// produces path ["A", "B", "C"] in the workflow tracking UI.
	Scope(name string) Logger

	// StartStep starts a trackable progress step.
	// The step appears as "pending" until Succeed/Fail/Skip is called.
	// Path is formed by: [scope prefix...] + msg
	StartStep(msg string) Step
}

func LogWarning(ctx context.Context, msg string, err error) {
	if l, ok := WarningLoggerFrom(ctx); ok {
		l.LogWarning(ctx, msg, err)
	}
}

func WithWarningLogger(ctx context.Context, l *WarningLogger) context.Context {
	return context.WithValue(ctx, warningLoggerContextKey, l)
}

func WarningLoggerFrom(ctx context.Context) (*WarningLogger, bool) {
	l, ok := ctx.Value(warningLoggerContextKey).(*WarningLogger)
	if !ok {
		return nil, false
	}
	if l == nil {
		return nil, false
	}

	return l, true
}

// With returns a new context with the given logger added to the context.
func With(ctx context.Context, l Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey, l)
}

// From returns the logger associated with the given context.
func From(ctx context.Context) Logger {
	if l, ok := ctx.Value(loggerContextKey).(Logger); ok {
		return l
	}
	return NewLogger(LevelFromEnv())
}

// WithFields returns a new context with the given fields added to the logger.
func WithFields(ctx context.Context, fields ...zap.Field) context.Context {
	if len(fields) == 0 {
		return ctx
	}

	return With(ctx, From(ctx).With(fields...))
}

// ZapLogger wraps zap.Logger and implements the Logger interface.
type ZapLogger struct {
	*zap.Logger
	scopePath []string // Track nested scopes for GitHub Actions group naming
}

func (l *ZapLogger) With(fields ...zapcore.Field) Logger {
	return &ZapLogger{
		Logger:    l.Logger.With(fields...),
		scopePath: l.scopePath,
	}
}

func (l *ZapLogger) Github(msg string) {
	if env.IsGithubAction() {
		fmt.Println(msg)
	}
}

func (l *ZapLogger) Scope(name string) Logger {
	newPath := make([]string, len(l.scopePath), len(l.scopePath)+1)
	copy(newPath, l.scopePath)
	newPath = append(newPath, name)
	return &ZapLogger{
		Logger:    l.Logger,
		scopePath: newPath,
	}
}

func (l *ZapLogger) StartStep(msg string) Step {
	return newGithubStep(l, msg)
}

type githubStep struct {
	logger     *ZapLogger
	name       string
	groupEnded bool
}

func newGithubStep(logger *ZapLogger, name string) *githubStep {
	s := &githubStep{
		logger: logger,
		name:   name,
	}
	s.start()
	return s
}

func (s *githubStep) fullName() string {
	if len(s.logger.scopePath) > 0 {
		return strings.Join(s.logger.scopePath, " > ") + " > " + s.name
	}
	return s.name
}

func (s *githubStep) start() {
	s.logger.Github("::group::" + s.fullName())
	// Note: Debug logging is intentionally NOT done here.
	// The caller (e.g., g.logStep) is responsible for debug logging.
	// This method only handles GitHub Actions group commands.
}

// closeGroup ensures ::endgroup:: is emitted exactly once, making teardown idempotent.
// This prevents duplicate endgroup markers when Skip() is called followed by Fail()/Succeed().
func (s *githubStep) closeGroup() {
	if s.groupEnded {
		return
	}
	s.groupEnded = true
	s.logger.Github("::endgroup::")
}

func (s *githubStep) Succeed() {
	s.closeGroup()
}

func (s *githubStep) Fail() {
	s.closeGroup()
}

func (s *githubStep) Skip() {
	s.closeGroup()
}

func NewLogger(lvl zapcore.Level) *ZapLogger {
	config, err := createConfig(lvl, "")
	if err != nil {
		fmt.Printf("Logger init failed with error: %s\n", err.Error())
		return &ZapLogger{Logger: zap.NewNop()}
	}

	l, err := config.Build()
	if err != nil {
		fmt.Printf("Logger init failed with error: %s\n", err.Error())
		l = zap.NewNop()
	}

	return &ZapLogger{Logger: l}
}

func NewLoggerFromZap(l *zap.Logger) *ZapLogger {
	return &ZapLogger{Logger: l}
}

func NewFileLogger(lvl zapcore.Level, outFile string) *ZapLogger {
	config, err := createConfig(lvl, outFile)
	if err != nil {
		fmt.Printf("Logger init failed with error: %s\n", err.Error())
		return &ZapLogger{Logger: zap.NewNop()}
	}

	l, err := config.Build()
	if err != nil {
		fmt.Printf("Logger init failed with error: %s\n", err.Error())
		l = zap.NewNop()
	}

	return &ZapLogger{Logger: l}
}

func createConfig(l zapcore.Level, outFile string) (*zap.Config, error) {
	level := zap.NewAtomicLevelAt(l)

	var err error

	outPaths := []string{"stdout"}
	errPaths := []string{"stderr"}
	if outFile != "" {
		dir := filepath.Dir(outFile)
		if err = os.MkdirAll(dir, 0o755); err == nil {
			if _, err := os.Stat(outFile); err == nil || os.IsNotExist(err) {
				f, err := os.Create(outFile)
				if err == nil {
					outPaths = []string{outFile}
					errPaths = []string{outFile}
					_ = f.Close()
				}
			}
		}
	}

	if err != nil {
		return nil, err
	}

	return &zap.Config{
		Level:    level,
		Encoding: "console",
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:     "message",
			LevelKey:       "level",
			TimeKey:        "time",
			NameKey:        "name",
			CallerKey:      "caller",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     timeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   callerEncoder,
		},
		OutputPaths:       outPaths,
		ErrorOutputPaths:  errPaths,
		DisableStacktrace: true,
	}, nil
}

func timeEncoder(time.Time, zapcore.PrimitiveArrayEncoder)                            {}
func callerEncoder(caller zapcore.EntryCaller, encoder zapcore.PrimitiveArrayEncoder) {}

func LevelFromEnv() zapcore.Level {
	level := os.Getenv("SPEAKEASY_LOG_LEVEL")
	if level == "" && env.IsDebug() {
		level = "debug"
	}

	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
