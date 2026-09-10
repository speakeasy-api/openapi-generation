//go:build js && wasm

package format

import (
	"context"
	"sync"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
)

// WASM-specific implementation that avoids wazero dependency
// For WASM builds, we skip complex formatting to reduce binary size

var (
	initOnce      sync.Once
	tfFormatter   = tfFormatStub{}
	tsFormatter   *dprintStub
	pyFormatter   *dprintStub
	javaFormatter *dprintStub
	phpFormatter  *dprintStub
	mtx           sync.Mutex
)

var (
	formattingEnabled   = map[string]bool{}
	formattingEnabledMu sync.RWMutex
)

func SetFormattingEnabled(target string, enabled bool) {
	formattingEnabledMu.Lock()
	defer formattingEnabledMu.Unlock()
	formattingEnabled[target] = enabled
}

func isFormattingEnabled(target string) bool {
	formattingEnabledMu.RLock()
	defer formattingEnabledMu.RUnlock()
	enabled, ok := formattingEnabled[target]
	return ok && enabled
}

// tfFormatStub is a no-op implementation for terraform formatting
type tfFormatStub struct{}

// Format returns the input unchanged for terraform files
func (tf tfFormatStub) Format(input []byte) ([]byte, error) {
	return input, nil
}

// dprintStub is a no-op implementation for WASM
type dprintStub struct{}

// Format returns the input unchanged
func (d *dprintStub) Format(input string) (string, error) {
	return input, nil
}

// formatString returns the input unchanged
func (d *dprintStub) formatString(input string) (string, error) {
	return input, nil
}

// close is a no-op
func (d *dprintStub) close() {
	// No-op
}

// newDprintTypeScript returns a stub formatter
func newDprintTypeScript() (*dprintStub, error) {
	return &dprintStub{}, nil
}

// newDprintPython returns a stub formatter
func newDprintPython() (*dprintStub, error) {
	return &dprintStub{}, nil
}

// ensureDPrint is a no-op for WASM
func ensureDPrint() {
	// No-op for WASM - formatters are already stubbed
}

// Format is the main entry point - returns input unchanged for WASM (aside
// from stripping the no-format marker) so binary size stays small.
func Format(_ context.Context, target types.Target, fileName string, data []byte) ([]byte, error) {
	if skip, cleaned := ShouldSkipFormat(data); skip {
		return cleaned, nil
	}
	return data, nil
}
