package logging

import (
	"context"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"go.uber.org/zap"
)

type WarningLogger struct {
	validationOnly  bool            // whether or not this is a validation run only
	onWarningLogged OnWarningLogged // callback to be called when a warning is logged
	shouldLog       bool            // whether or not to log warnings

	warnings  []error             // warnings that have been logged
	logFilter map[string]struct{} // used to filter out duplicate warnings
}

type OnWarningLogged func(err error)

func NewWarningLogger(validationOnly bool, onWarningLogged OnWarningLogged) *WarningLogger {
	return &WarningLogger{
		validationOnly:  validationOnly,
		onWarningLogged: onWarningLogged,
		shouldLog:       true,
		logFilter:       map[string]struct{}{},
	}
}

func (w *WarningLogger) DisableLogging() {
	w.shouldLog = false
}

func (w *WarningLogger) GetWarnings() []error {
	return w.warnings
}

func (w *WarningLogger) LogWarning(ctx context.Context, msg string, err error) {
	if w.validationOnly && errors.IsUnsupported(err) {
		return
	}

	errKey := msg + " " + err.Error()

	_, logged := w.logFilter[errKey]
	if logged {
		return
	}
	w.logFilter[errKey] = struct{}{}

	if w.shouldLog {
		From(ctx).Warn(msg, zap.Error(err))
	}

	w.warnings = append(w.warnings, err)

	if w.onWarningLogged != nil {
		w.onWarningLogged(err)
	}
}
