//go:build js && wasm

// Package format provides stubs for internal/format package
// This eliminates wazero and other heavy formatting dependencies for WASM
package format

import (
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
)

// Format is a stub that returns the input unchanged
// In WASM we skip formatting to reduce binary size
func Format(target types.Target, fileName string, data []byte) ([]byte, error) {
	// For WASM builds, just return the data unchanged to avoid heavy formatting dependencies
	return data, nil
}

// Any other exported functions from the original format package can be stubbed here
