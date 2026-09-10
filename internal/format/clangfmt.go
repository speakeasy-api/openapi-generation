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
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/experimental"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/tetratelabs/wazero/sys"
	"go.uber.org/zap"
)

//go:embed clangfmt.wasm
var clangfmtBin []byte

// clangFormatStyle configures clang-format to approximate dotnet format output.
//
//   - BasedOnStyle: Microsoft                        — Allman braces, 4-space indent
//   - NamespaceIndentation: All                      — indent contents inside namespace {}
//   - ColumnLimit: 0                                 — don't reflow/wrap lines
//   - AllowShortIfStatementsOnASingleLine: WithoutElse — preserve single-line if statements
//   - IndentCaseLabels: true                         — indent case labels inside switch
//   - BreakBeforeCloseBracketFunction: true          — closing ) on its own line (LLVM 22+)
const clangFormatStyle = "{BasedOnStyle: Microsoft, NamespaceIndentation: All, ColumnLimit: 0, AllowShortIfStatementsOnASingleLine: WithoutElse, IndentCaseLabels: true, BreakBeforeCloseBracketFunction: true}"

var (
	cfOnce     sync.Once
	cfRuntime  wazero.Runtime
	cfCompiled wazero.CompiledModule
	errCfInit  error
)

func ensureClangfmt() error {
	cfOnce.Do(func() {
		ctx := context.Background()
		cfRuntime = wazero.NewRuntime(ctx)
		wasi_snapshot_preview1.MustInstantiate(ctx, cfRuntime)

		// Stub out Emscripten's env module imports.
		// The standalone WASM binary from Emscripten imports a handful of
		// syscalls and helpers from "env". Since we only use stdin/stdout
		// for formatting, these can all be no-ops / return error.
		_, errCfInit = cfRuntime.NewHostModuleBuilder("env").
			// Memory growth notification (no-op).
			NewFunctionBuilder().
			WithFunc(func(_ context.Context, _ api.Module, _ uint32) {}).
			Export("emscripten_notify_memory_growth").
			// Filesystem syscalls – return -1 (ENOSYS). Never called during formatting.
			NewFunctionBuilder().
			WithFunc(func(_ context.Context, _ api.Module, _, _, _, _ uint32) uint32 { return ^uint32(0) }).
			Export("__syscall_faccessat").
			NewFunctionBuilder().
			WithFunc(func(_ context.Context, _ api.Module, _ uint32) uint32 { return ^uint32(0) }).
			Export("__syscall_chdir").
			NewFunctionBuilder().
			WithFunc(func(_ context.Context, _ api.Module, _, _ uint32) uint32 { return ^uint32(0) }).
			Export("__syscall_getcwd").
			NewFunctionBuilder().
			WithFunc(func(_ context.Context, _ api.Module, _, _, _ uint32) uint32 { return ^uint32(0) }).
			Export("__syscall_getdents64").
			NewFunctionBuilder().
			WithFunc(func(_ context.Context, _ api.Module, _, _, _, _ uint32) uint32 { return ^uint32(0) }).
			Export("__syscall_readlinkat").
			NewFunctionBuilder().
			WithFunc(func(_ context.Context, _ api.Module, _, _, _ uint32) uint32 { return ^uint32(0) }).
			Export("__syscall_unlinkat").
			NewFunctionBuilder().
			WithFunc(func(_ context.Context, _ api.Module, _, _, _ uint32) uint32 { return ^uint32(0) }).
			Export("__syscall_statfs64").
			Instantiate(ctx)
		if errCfInit != nil {
			return
		}

		cfCompiled, errCfInit = cfRuntime.CompileModule(ctx, clangfmtBin)
	})
	return errCfInit
}

var (
	cfCounter   uint64
	cfAllocator = &pooledAllocator{}
)

func formatCSharp(ctx context.Context, fileName string, data []byte) ([]byte, error) {
	if err := ensureClangfmt(); err != nil {
		return nil, fmt.Errorf("clangfmt init: %w", err)
	}

	stdin := bytes.NewReader(data)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// Each instantiation needs a unique name in wazero.
	n := atomic.AddUint64(&cfCounter, 1)

	cfg := wazero.NewModuleConfig().
		WithStdin(stdin).
		WithStdout(&stdout).
		WithStderr(&stderr).
		WithName(fmt.Sprintf("clangfmt-%d", n)).
		// argv[0] = program name, argv[1] = style, argv[2] = filename hint
		// Style config targets dotnet format parity. See clangfmt_test.go for
		// documented divergences that cannot be resolved via config alone.
		WithArgs("clang-format-wasi", clangFormatStyle, fileName)

	ctx = experimental.WithMemoryAllocator(ctx, cfAllocator)
	mod, err := cfRuntime.InstantiateModule(ctx, cfCompiled, cfg)
	if mod != nil {
		defer func() { _ = mod.Close(ctx) }()
	}
	if err != nil {
		var exitErr *sys.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 0 {
			// WASI proc_exit(0) means success
		} else {
			// Unexpected failure — log warning and return original data unchanged.
			logging.From(ctx).Warn("clangfmt formatting failed",
				zap.Error(err),
				zap.String("stderr", stderr.String()),
				zap.String("file", fileName))
			return data, nil
		}
	}

	if stdout.Len() == 0 {
		return data, nil
	}

	return stdout.Bytes(), nil
}

func init() {
	go func() { _ = ensureClangfmt() }()
}
