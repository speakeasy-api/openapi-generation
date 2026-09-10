//go:build js && wasm
// +build js,wasm

package main

import (
	"fmt"
	"syscall/js"

	"github.com/speakeasy-api/openapi-generation/v2/languageserver/server"
)

// JSConsoleLogger implements Logger interface using JavaScript console
type JSConsoleLogger struct{}

// validate JSConsoleLogger implements Logger interface
var _ server.Logger = &JSConsoleLogger{}

func (l *JSConsoleLogger) Error(message string, keysAndValues ...any) {
	msg := fmt.Sprintf("❌ SERVER: %s", message)
	js.Global().Get("console").Call("error", msg)

	if len(keysAndValues) > 0 {
		js.Global().Get("console").Call("error", keysAndValues)
	}
}

func (l *JSConsoleLogger) Errorf(format string, args ...any) {
	l.Error(fmt.Sprintf(format, args))
}

func (l *JSConsoleLogger) Warning(message string, keysAndValues ...any) {
	msg := fmt.Sprintf("⚠️ SERVER: %s", message)
	js.Global().Get("console").Call("warn", msg)

	if len(keysAndValues) > 0 {
		js.Global().Get("console").Call("warn", keysAndValues)
	}
}

func (l *JSConsoleLogger) Warningf(format string, args ...any) {
	l.Warning(fmt.Sprintf(format, args))
}

func (l *JSConsoleLogger) Info(message string, keysAndValues ...any) {
	msg := fmt.Sprintf("ℹ️ SERVER: %s", message)
	js.Global().Get("console").Call("info", msg)

	if len(keysAndValues) > 0 {
		js.Global().Get("console").Call("info", keysAndValues)
	}
}

func (l *JSConsoleLogger) Infof(format string, args ...any) {
	l.Info(fmt.Sprintf(format, args))
}

func (l *JSConsoleLogger) Debug(message string, keysAndValues ...any) {
	msg := fmt.Sprintf("🔧 SERVER: %s", message)
	js.Global().Get("console").Call("debug", msg)

	if len(keysAndValues) > 0 {
		js.Global().Get("console").Call("debug", keysAndValues)
	}
}

func (l *JSConsoleLogger) Debugf(format string, args ...any) {
	l.Debug(fmt.Sprintf(format, args))
}
