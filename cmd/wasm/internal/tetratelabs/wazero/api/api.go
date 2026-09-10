//go:build js && wasm

// Package api provides stubs for tetratelabs/wazero/api package
package api

import "context"

// Module represents a WebAssembly module (stub)
type Module interface {
	Name() string
	Close(ctx context.Context) error
	// Add other methods as needed
	ExportedFunction(name string) Function
}

// Function represents a WebAssembly function (stub)
type Function interface {
	Call(ctx context.Context, params ...interface{}) ([]interface{}, error)
}

// stubFunction implements Function interface
type stubFunction struct{}

func (f *stubFunction) Call(ctx context.Context, params ...interface{}) ([]interface{}, error) {
	return nil, nil
}
