//go:build js && wasm

// Package api provides stubs for tetratelabs/wazero/api package
package api

import "context"

// Module represents a WebAssembly module (stub)
type Module interface {
	Name() string
	Close(ctx context.Context) error
	ExportedFunction(name string) Function
	Memory() Memory
}

// Memory represents WebAssembly memory (stub)
type Memory interface {
	Read(offset, byteCount uint32) ([]byte, bool)
	Write(offset uint32, v []byte) bool
	Size() uint32
}

// Function represents a WebAssembly function (stub)
type Function interface {
	Call(ctx context.Context, params ...interface{}) ([]interface{}, error)
}

// stubMemory implements Memory interface
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
