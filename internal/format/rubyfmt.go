//go:build !js || !wasm

package format

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/experimental"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/tetratelabs/wazero/sys"
	"go.uber.org/zap"
)

//go:embed rubyfmt.wasm
var rubyfmtBin []byte

var (
	rbOnce     sync.Once
	rbRuntime  wazero.Runtime
	rbCompiled wazero.CompiledModule
	errRbInit  error
)

func ensureRubyfmt() error {
	rbOnce.Do(func() {
		ctx := context.Background()
		rbRuntime = wazero.NewRuntime(ctx)
		wasi_snapshot_preview1.MustInstantiate(ctx, rbRuntime)

		// Provide stub for env.__wasi_init_tp (thread pointer init) needed by Rust WASI libc.
		// The binary imports this as type () -> () (no params).
		_, errRbInit = rbRuntime.NewHostModuleBuilder("env").
			NewFunctionBuilder().
			WithFunc(func() {}).
			Export("__wasi_init_tp").
			Instantiate(ctx)
		if errRbInit != nil {
			return
		}

		rbCompiled, errRbInit = rbRuntime.CompileModule(ctx, rubyfmtBin)
	})
	return errRbInit
}

var (
	rbCounter   uint64
	rbAllocator = &pooledAllocator{}
)

func formatRuby(ctx context.Context, _ string, data []byte) ([]byte, error) {
	if err := ensureRubyfmt(); err != nil {
		return nil, fmt.Errorf("rubyfmt init: %w", err)
	}

	stdin := bytes.NewReader(data)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// Each instantiation needs a unique name in wazero
	n := atomic.AddUint64(&rbCounter, 1)

	cfg := wazero.NewModuleConfig().
		WithStdin(stdin).
		WithStdout(&stdout).
		WithStderr(&stderr).
		WithName(fmt.Sprintf("rubyfmt-%d", n))

	ctx = experimental.WithMemoryAllocator(ctx, rbAllocator)
	mod, err := rbRuntime.InstantiateModule(ctx, rbCompiled, cfg)
	if mod != nil {
		defer func() { _ = mod.Close(ctx) }()
	}
	if err != nil {
		var exitErr *sys.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 0 {
			// WASI proc_exit(0) means success
		} else {
			// Unexpected failure — log warning and return original data unchanged
			logging.From(ctx).Warn("rubyfmt formatting failed", zap.Error(err), zap.String("stderr", stderr.String()))
			return data, nil
		}
	}

	if stdout.Len() == 0 {
		return data, nil
	}

	return stdout.Bytes(), nil
}

func init() {
	go func() { _ = ensureRubyfmt() }()
}
