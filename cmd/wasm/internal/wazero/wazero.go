//go:build js && wasm

// Package wazero provides stubs for tetratelabs/wazero package
package wazero

import (
	"context"

	"github.com/tetratelabs/wazero/api"
)

// Runtime represents a WebAssembly runtime (stub)
type Runtime interface {
	Instantiate(ctx context.Context, binary []byte) (api.Module, error)
	Close(ctx context.Context) error
}

// stubRuntime is a no-op implementation
type stubRuntime struct{}

// NewRuntime creates a new stub runtime
func NewRuntime(ctx context.Context) Runtime {
	return &stubRuntime{}
}

// Instantiate returns a stub module
func (r *stubRuntime) Instantiate(ctx context.Context, binary []byte) (api.Module, error) {
	return &stubModule{}, nil
}

// Close is a no-op
func (r *stubRuntime) Close(ctx context.Context) error {
	return nil
}

// stubModule implements api.Module interface
type stubModule struct{}

// Name returns a stub name
func (m *stubModule) Name() string {
	return "stub"
}

// Close is a no-op
func (m *stubModule) Close(ctx context.Context) error {
	return nil
}

// ExportedFunction returns a stub function
func (m *stubModule) ExportedFunction(name string) api.Function {
	return &stubFunction{}
}

// Memory returns stub memory
func (m *stubModule) Memory() api.Memory {
	return &stubMemory{}
}

// stubFunction implements api.Function interface
type stubFunction struct{}

// Call returns empty results - return uint32/uint64 values for compatibility
func (f *stubFunction) Call(ctx context.Context, params ...interface{}) ([]interface{}, error) {
	// Return some default values that might be expected - both uint32 and uint64
	return []interface{}{uint32(0), uint64(0)}, nil
}

// stubMemory implements api.Memory interface
type stubMemory struct{}

func (m *stubMemory) Read(offset, byteCount uint32) ([]byte, bool) {
	return make([]byte, byteCount), true
}

func (m *stubMemory) Write(offset uint32, v []byte) bool {
	return true
}

func (m *stubMemory) Size() uint32 {
	return 1024 * 1024 // 1MB stub
}
