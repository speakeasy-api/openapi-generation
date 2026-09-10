//go:build js && wasm

// Package testcontainers provides empty stubs for the testcontainers-go package
// This is needed because the real package uses Unix system calls that are not available in WASM
package testcontainers

import (
	"context"
	"time"
)

// TerminateOptions is a type that holds the options for terminating a container.
type TerminateOptions struct {
	ctx         context.Context
	stopTimeout *time.Duration
	volumes     []string
}

// TerminateOption is a type that represents an option for terminating a container.
type TerminateOption func(*TerminateOptions)

// Container represents a generic container
type Container interface {
	Endpoint(context.Context, string) (string, error)
	Terminate(context.Context, ...TerminateOption) error
}

// ContainerRequest represents a request to create a container
type ContainerRequest struct {
	FromDockerfile FromDockerfile
	ExposedPorts   []string
	WaitingFor     interface{}
}

// FromDockerfile contains options for building a container from a Dockerfile
type FromDockerfile struct {
	Context    string
	Dockerfile string
}

// GenericContainerRequest is a request to create a generic container
type GenericContainerRequest struct {
	ContainerRequest ContainerRequest
	Started          bool
}

// GenericContainer creates a generic container (stub implementation)
func GenericContainer(ctx context.Context, req GenericContainerRequest) (Container, error) {
	return nil, nil
}

// Wait package stubs
type Wait struct{}

// ForLog returns a condition that waits for a specified log message
func ForLog(message string) interface{} {
	return nil
}
