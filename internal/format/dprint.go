//go:build !js || !wasm

package format

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

//go:embed dprint-plugin-ruff.wasm
var dprintRuffBin []byte

//go:embed dprint-plugin-typescript.wasm
var dprintTSBin []byte

//go:embed dprint-plugin-java.wasm
var dprintJavaBin []byte

//go:embed dprint-plugin-mago.wasm
var dprintMagoBin []byte

var (
	tsConfig = `{
  "lineWidth": 80,
  "indentWidth": 2,
  "useTabs": false,
  "newLineKind": "lf",
  "semiColons": "prefer",
  "trailingCommas": "onlyMultiLine",
  "quoteStyle": "alwaysDouble",
  "useBraces": "whenNotSingleLine",
  "preferSingleLine": true
}`

	pyConfig = `{}`

	javaConfig = `{
  "lineWidth": 120,
  "indentWidth": 4,
  "useTabs": false,
  "newLineKind": "lf"
}`

	phpConfig = `{
  "printWidth": 100,
  "tabWidth": 4,
  "useTabs": false,
  "endOfLine": "lf",
  "singleQuote": true,
  "trailingComma": true,
  "phpVersionMajor": 8,
  "phpVersionMinor": 0
}`
)

// dprintFormatter is the interface that both v3 and v4 dprint implementations satisfy.
type dprintFormatter interface {
	Format(ctx context.Context, fileName string, data []byte) ([]byte, error)
	Close(ctx context.Context) error
}

type dprint struct {
	runtime wazero.Runtime
	module  api.Module

	formatConfig string

	// wasm functions
	_format                     api.Function
	_getErrorText               api.Function
	_setFilePath                api.Function
	_getFormattedText           api.Function
	_addToSharedBytesFromBuffer api.Function
	_setBufferWithSharedBytes   api.Function
	_clearSharedBytes           api.Function
	_getWASMMemoryBuffer        api.Function
	_setWASMMemoryBuffer        api.Function
	_getWASMMemoryBufferSize    api.Function
	_resetConfig                api.Function
	_setGlobalConfig            api.Function
	_setPluginConfig            api.Function
}

// dprintV4 implements dprintFormatter for dprint plugin schema version 4.
// V4 plugins use get_shared_bytes_ptr for direct memory access, register_config/release_config
// for config lifecycle, and format(config_id) instead of format().
type dprintV4 struct {
	runtime wazero.Runtime
	module  api.Module

	formatConfig string
	configID     uint32

	// wasm functions
	_format            api.Function
	_getErrorText      api.Function
	_setFilePath       api.Function
	_getFormattedText  api.Function
	_clearSharedBytes  api.Function
	_getSharedBytesPtr api.Function
	_registerConfig    api.Function
	_releaseConfig     api.Function
	_setOverrideConfig api.Function
}

func newDprint(lang string) (dprintFormatter, error) {
	var bin []byte
	var config string
	switch lang {
	case "typescript":
		bin = dprintTSBin
		config = tsConfig
	case "python":
		bin = dprintRuffBin
		config = pyConfig
	case "java":
		bin = dprintJavaBin
		config = javaConfig
	case "php":
		bin = dprintMagoBin
		config = phpConfig
	default:
		return nil, fmt.Errorf("unsupported language: %s", lang)
	}

	ctx := context.Background()

	r := wazero.NewRuntime(ctx)

	mod, err := r.Instantiate(ctx, bin)
	if err != nil {
		return nil, fmt.Errorf("instantiate: %w", err)
	}

	// Detect plugin schema version: v4 plugins export dprint_plugin_version_4
	if mod.ExportedFunction("dprint_plugin_version_4") != nil {
		return newDprintV4WithModule(r, mod, config)
	}

	return newDprintWithModule(r, mod, config), nil
}

func newDprintWithModule(r wazero.Runtime, module api.Module, formatConfig string) *dprint {
	return &dprint{
		runtime: r,
		module:  module,

		formatConfig: formatConfig,

		_format:                     module.ExportedFunction("format"),
		_getErrorText:               module.ExportedFunction("get_error_text"),
		_setFilePath:                module.ExportedFunction("set_file_path"),
		_getFormattedText:           module.ExportedFunction("get_formatted_text"),
		_addToSharedBytesFromBuffer: module.ExportedFunction("add_to_shared_bytes_from_buffer"),
		_setBufferWithSharedBytes:   module.ExportedFunction("set_buffer_with_shared_bytes"),
		_clearSharedBytes:           module.ExportedFunction("clear_shared_bytes"),
		_getWASMMemoryBuffer:        module.ExportedFunction("get_wasm_memory_buffer"),
		_setWASMMemoryBuffer:        module.ExportedFunction("set_wasm_memory_buffer"),
		_getWASMMemoryBufferSize:    module.ExportedFunction("get_wasm_memory_buffer_size"),
		_resetConfig:                module.ExportedFunction("reset_config"),
		_setGlobalConfig:            module.ExportedFunction("set_global_config"),
		_setPluginConfig:            module.ExportedFunction("set_plugin_config"),
	}
}

func (f *dprint) Close(ctx context.Context) error {
	return f.runtime.Close(ctx)
}

func (f *dprint) Format(ctx context.Context, fileName string, data []byte) ([]byte, error) {
	_, err := f._resetConfig.Call(ctx)
	if err != nil {
		return nil, fmt.Errorf("reset config: %w", err)
	}

	err = f.putString(ctx, f.formatConfig)
	if err != nil {
		return nil, fmt.Errorf("write global config: %w", err)
	}

	_, err = f._setGlobalConfig.Call(ctx)
	if err != nil {
		return nil, fmt.Errorf("set global config: %w", err)
	}

	err = f.putString(ctx, "{}")
	if err != nil {
		return nil, fmt.Errorf("write plugin config: %w", err)
	}

	_, err = f._setPluginConfig.Call(ctx)
	if err != nil {
		return nil, fmt.Errorf("set plugin config: %w", err)
	}

	err = f.putString(ctx, fileName)
	if err != nil {
		return nil, fmt.Errorf("write file path: %w", err)
	}

	_, err = f._setFilePath.Call(ctx)
	if err != nil {
		return nil, fmt.Errorf("set file path: %w", err)
	}

	err = f.putString(ctx, string(data))
	if err != nil {
		return nil, fmt.Errorf("write file content: %w", err)
	}

	fres, err := f._format.Call(ctx)
	if err != nil {
		return nil, fmt.Errorf("format: %w", err)
	}

	if len(fres) == 0 {
		return nil, errors.New("missing format result")
	}

	code := fres[0]

	switch code {
	case 0: // content did not need formatting
		return data, nil

	case 1: // Success. Nothing to do here, we'll read out the result below

	case 2: // Error occurred.
		lastError, err := f.getLastError(ctx)
		if err != nil {
			return nil, fmt.Errorf("get last error: %w", err)
		}
		return nil, fmt.Errorf("format: %s", lastError)

	default:
		return nil, fmt.Errorf("format failed with code: %v", code)
	}

	count, err := f._getFormattedText.Call(ctx)
	if err != nil {
		return nil, fmt.Errorf("get formatted text length: %w", err)
	}

	out, err := f.pullString(ctx, uint32(count[0]))
	if err != nil {
		return nil, fmt.Errorf("get formatted text: %w", err)
	}

	return []byte(out), nil
}

func (f *dprint) putString(ctx context.Context, content string) error {
	_, err := f._clearSharedBytes.Call(ctx, uint64(len(content)))
	if err != nil {
		return fmt.Errorf("clear shared bytes: %w", err)
	}

	bufsize, err := f.getBufferSize(ctx)
	if err != nil {
		return fmt.Errorf("get wasm memory buffer size: %w", err)
	}

	src := []byte(content)
	length := uint64(len(src))

	var n uint64
	for i := uint64(0); i < length; i += n {
		if i >= length {
			break
		}

		writeLen := min(length-i, uint64(bufsize))

		// We assume the uint32 conversion below is safe because the maximum
		// buffer size is within uint32 and the min() call above sets the upper
		// bound of writeLen to that size.
		buf, err := f.getBuffer(ctx, uint32(writeLen))
		if err != nil {
			return fmt.Errorf("get wasm buffer: %w", err)
		}
		if len(buf) == 0 {
			return errors.New("wasm buffer empty")
		}

		n = uint64(copy(buf, src[i:i+writeLen]))

		_, err = f._addToSharedBytesFromBuffer.Call(ctx, writeLen)
		if err != nil {
			return fmt.Errorf("add to shared bytes from buffer: %w", err)
		}
	}

	return nil
}

func (f *dprint) pullString(ctx context.Context, length uint32) (string, error) {
	bufferSize, err := f.getBufferSize(ctx)
	if err != nil {
		return "", fmt.Errorf("get wasm memory buffer size: %w", err)
	}

	buffer := make([]byte, length)

	var n uint64
	l := uint64(length)
	for i := uint64(0); i < l; i += n {
		readLen := min(l-i, uint64(bufferSize))

		_, err = f._setBufferWithSharedBytes.Call(ctx, i, readLen)
		if err != nil {
			return "", fmt.Errorf("set buffer with shared bytes: %w", err)
		}

		// We assume the uint32 conversion below is safe because the maximum
		// buffer size is within uint32 and the min() call above sets the upper
		// bound of readLen to that size.
		wasmBuffer, err := f.getBuffer(ctx, uint32(readLen))
		if err != nil {
			return "", fmt.Errorf("get wasm buffer: %w", err)
		}

		n = uint64(copy(buffer[i:], wasmBuffer))
	}

	return string(buffer), nil
}

func (f *dprint) getBuffer(ctx context.Context, length uint32) ([]byte, error) {
	res, err := f._getWASMMemoryBuffer.Call(ctx)
	if err != nil {
		return nil, err
	}

	if len(res) != 1 {
		return nil, fmt.Errorf("invalid result length: %d", len(res))
	}

	offset, err := uint64ToUint32(res[0])
	if err != nil {
		return nil, fmt.Errorf("invalid offset: %v: %w", offset, err)
	}

	mem, ok := f.module.Memory().Read(offset, length)
	if !ok {
		return nil, fmt.Errorf("read: out of bounds: offset %d length %d", offset, length)
	}

	return mem, nil
}

func (f *dprint) getBufferSize(ctx context.Context) (uint32, error) {
	res, err := f._getWASMMemoryBufferSize.Call(ctx)
	if err != nil {
		return 0, err
	}
	if len(res) != 1 {
		return 0, fmt.Errorf("invalid length: %d", len(res))
	}

	sizeu32, err := uint64ToUint32(res[0])
	if err != nil {
		return 0, fmt.Errorf("invalid size: %v: %w", sizeu32, err)
	}

	return sizeu32, nil
}

func (f *dprint) getLastError(ctx context.Context) (string, error) {
	res, err := f._getErrorText.Call(ctx)
	if err != nil {
		return "", err
	}
	if len(res) != 1 {
		return "", fmt.Errorf("invalid length: %d", len(res))
	}

	length, err := uint64ToUint32(res[0])
	if err != nil {
		return "", fmt.Errorf("invalid length: %v: %w", res[0], err)
	}

	return f.pullString(ctx, length)
}

func uint64ToUint32(v uint64) (uint32, error) {
	out := uint32(v)
	if uint64(out) != v {
		return 0, errors.New("out of bounds")
	}
	return out, nil
}

// newDprintV4WithModule creates a dprintV4 formatter from an already-instantiated WASM module
// that uses dprint plugin schema version 4.
func newDprintV4WithModule(r wazero.Runtime, module api.Module, formatConfig string) (*dprintV4, error) {
	d := &dprintV4{
		runtime: r,
		module:  module,

		formatConfig: formatConfig,

		_format:            module.ExportedFunction("format"),
		_getErrorText:      module.ExportedFunction("get_error_text"),
		_setFilePath:       module.ExportedFunction("set_file_path"),
		_getFormattedText:  module.ExportedFunction("get_formatted_text"),
		_clearSharedBytes:  module.ExportedFunction("clear_shared_bytes"),
		_getSharedBytesPtr: module.ExportedFunction("get_shared_bytes_ptr"),
		_registerConfig:    module.ExportedFunction("register_config"),
		_releaseConfig:     module.ExportedFunction("release_config"),
		_setOverrideConfig: module.ExportedFunction("set_override_config"),
	}

	ctx := context.Background()

	// Register config once during initialization.
	// The v4 plugin expects config as {"plugin": {...}, "global": {...}}.
	configJSON := fmt.Sprintf(`{"plugin":{},"global":%s}`, formatConfig)
	if err := d.putString(ctx, configJSON); err != nil {
		return nil, fmt.Errorf("write config: %w", err)
	}

	_, err := d._registerConfig.Call(ctx, 0)
	if err != nil {
		return nil, fmt.Errorf("register config: %w", err)
	}
	d.configID = 0

	return d, nil
}

func (f *dprintV4) Close(ctx context.Context) error {
	if f._releaseConfig != nil {
		_, _ = f._releaseConfig.Call(ctx, uint64(f.configID))
	}
	return f.runtime.Close(ctx)
}

func (f *dprintV4) Format(ctx context.Context, fileName string, data []byte) ([]byte, error) {
	// Set override config (empty -- no per-file overrides)
	if err := f.putString(ctx, "{}"); err != nil {
		return nil, fmt.Errorf("write override config: %w", err)
	}
	_, err := f._setOverrideConfig.Call(ctx)
	if err != nil {
		return nil, fmt.Errorf("set override config: %w", err)
	}

	// Set file path
	if err := f.putString(ctx, fileName); err != nil {
		return nil, fmt.Errorf("write file path: %w", err)
	}
	_, err = f._setFilePath.Call(ctx)
	if err != nil {
		return nil, fmt.Errorf("set file path: %w", err)
	}

	// Write file content (must be last before format call)
	if err := f.putString(ctx, string(data)); err != nil {
		return nil, fmt.Errorf("write file content: %w", err)
	}

	// Format with config_id
	fres, err := f._format.Call(ctx, uint64(f.configID))
	if err != nil {
		return nil, fmt.Errorf("format: %w", err)
	}

	if len(fres) == 0 {
		return nil, errors.New("missing format result")
	}

	code := fres[0]

	switch code {
	case 0: // content did not need formatting
		return data, nil

	case 1: // Success

	case 2: // Error occurred
		lastError, err := f.getLastError(ctx)
		if err != nil {
			return nil, fmt.Errorf("get last error: %w", err)
		}
		return nil, fmt.Errorf("format: %s", lastError)

	default:
		return nil, fmt.Errorf("format failed with code: %v", code)
	}

	count, err := f._getFormattedText.Call(ctx)
	if err != nil {
		return nil, fmt.Errorf("get formatted text length: %w", err)
	}

	out, err := f.pullString(ctx, uint32(count[0]))
	if err != nil {
		return nil, fmt.Errorf("get formatted text: %w", err)
	}

	return []byte(out), nil
}

// putString writes a string into the shared bytes buffer using the v4 direct-pointer approach.
func (f *dprintV4) putString(ctx context.Context, content string) error {
	src := []byte(content)
	length := uint64(len(src))

	// clear_shared_bytes(size) allocates space and returns a pointer
	res, err := f._clearSharedBytes.Call(ctx, length)
	if err != nil {
		return fmt.Errorf("clear shared bytes: %w", err)
	}

	if len(res) == 0 {
		return errors.New("clear_shared_bytes returned no result")
	}

	ptr, err := uint64ToUint32(res[0])
	if err != nil {
		return fmt.Errorf("invalid shared bytes pointer: %w", err)
	}

	// Write directly to WASM memory at the returned pointer
	if !f.module.Memory().Write(ptr, src) {
		return fmt.Errorf("write: out of bounds: offset %d length %d", ptr, length)
	}

	return nil
}

// pullString reads a string from the shared bytes buffer using the v4 direct-pointer approach.
func (f *dprintV4) pullString(ctx context.Context, length uint32) (string, error) {
	res, err := f._getSharedBytesPtr.Call(ctx)
	if err != nil {
		return "", fmt.Errorf("get shared bytes ptr: %w", err)
	}

	if len(res) == 0 {
		return "", errors.New("get_shared_bytes_ptr returned no result")
	}

	ptr, err := uint64ToUint32(res[0])
	if err != nil {
		return "", fmt.Errorf("invalid shared bytes pointer: %w", err)
	}

	data, ok := f.module.Memory().Read(ptr, length)
	if !ok {
		return "", fmt.Errorf("read: out of bounds: offset %d length %d", ptr, length)
	}

	return string(data), nil
}

func (f *dprintV4) getLastError(ctx context.Context) (string, error) {
	res, err := f._getErrorText.Call(ctx)
	if err != nil {
		return "", err
	}
	if len(res) != 1 {
		return "", fmt.Errorf("invalid length: %d", len(res))
	}

	length, err := uint64ToUint32(res[0])
	if err != nil {
		return "", fmt.Errorf("invalid length: %v: %w", res[0], err)
	}

	return f.pullString(ctx, length)
}
