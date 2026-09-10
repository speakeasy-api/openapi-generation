//go:build js && wasm

// Package instrument provides stubs for internal/instrument package
// This eliminates gRPC and OpenTelemetry dependencies for WASM
package instrument

import (
	"context"
)

// SetupOTelSDK is a stub that does nothing and returns a no-op shutdown function
func SetupOTelSDK(ctx context.Context, uri string) (shutdown func(context.Context) error, err error) {
	// Return a no-op shutdown function for WASM
	return func(context.Context) error { return nil }, nil
}
