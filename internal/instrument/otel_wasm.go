//go:build js && wasm

package instrument

import (
	"context"
)

// SetupOTelSDK is a stub that does nothing and returns a no-op shutdown function
// For WASM builds, we skip telemetry to reduce binary size
func SetupOTelSDK(ctx context.Context, uri string) (shutdown func(context.Context) error, err error) {
	// Return a no-op shutdown function for WASM
	return func(context.Context) error { return nil }, nil
}
