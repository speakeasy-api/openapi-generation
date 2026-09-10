//go:build js && wasm

// Package wazero provides stubs for tetratelabs/wazero package
// This is needed because we don't need WebAssembly runtime inside WebAssembly
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

// Any methods that api.Module needs can be stubbed here
func (m *stubModule) Name() string {
	return "stub"
}

func (m *stubModule) Close(ctx context.Context) error {
	return nil
}
