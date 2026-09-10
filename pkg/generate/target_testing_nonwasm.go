//go:build !js && !wasm

package generate

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func newMockServerDockerProcess(ctx context.Context, mockServerDirectory string) (process *mockServerProcess, err error) {
	defer func() {
		if r := recover(); r != nil {
			if panicErr, ok := r.(error); ok {
				err = fmt.Errorf("failed to start mock server: %w", panicErr)
			} else {
				err = fmt.Errorf("failed to start mock server: %v", r)
			}
		}
	}()

	if _, err := exec.LookPath("docker"); err != nil {
		return nil, errors.New("docker is not installed")
	}

	testDataDir := filepath.Join(mockServerDirectory, "testdata")

	// Prevent Docker error: COPY failed: file not found in build context or excluded by .dockerignore: stat testdata: file does not exist
	// Especially if generation occurs separate from testing while using a
	// version control system, an empty directory may not be present. This final
	// check prevents the need for the generator and customer to manage a keeper
	// file or similar to ensure the directory exists.
	if _, err := os.Stat(testDataDir); err != nil && errors.Is(err, fs.ErrNotExist) {
		if err := os.MkdirAll(testDataDir, 0o755); err != nil {
			return nil, fmt.Errorf("could not create mockserver testdata directory: %w", err)
		}
	}

	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    mockServerDirectory,
			Dockerfile: "Dockerfile",
		},
		ExposedPorts: []string{"18080/tcp"},
		WaitingFor:   wait.ForLog("starting server"),
	}
	mockServerC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("could not construct mock server container: %w", err)
	}

	return &mockServerProcess{
		container: mockServerC,
	}, nil
}
