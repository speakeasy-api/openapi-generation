//go:build js && wasm

// Package wait provides stubs for the wait package in testcontainers-go
package wait

// ForLog returns a condition that waits for a specified log message
func ForLog(message string) interface{} {
	return nil
}
