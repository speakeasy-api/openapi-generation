//go:build js && wasm

package generate

import (
	"context"
	"fmt"
	"os/exec"
)

// Stub implementation for WASM environments
func newMockServerDockerProcess(ctx context.Context, mockServerDirectory string) (process *mockServerProcess, err error) {
	return nil, fmt.Errorf("Docker not supported in WASM environment")
}

// Assign process group is a no-op in WASM
func assignProcessGroup(cmd *exec.Cmd) {
	// No-op on WASM
}

// Kill process is a no-op in WASM
func killProcess(cmd *exec.Cmd) {
	// No-op on WASM
}
