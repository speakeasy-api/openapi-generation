package executor

import (
	"fmt"
	"path"
	"strings"
	"sync"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	esbuild "github.com/evanw/esbuild/pkg/api"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/templates"
)

type Executor struct {
	vm     *goja.Runtime
	mutex  sync.Mutex
	target types.Target
	file   string
}

// Cache for esbuild TypeScript-to-JavaScript transform results.
// Templates are read from embed.FS and are immutable during a process
// lifetime, so the transform output for a given path never changes.
var (
	transformCache      = make(map[string][]byte)
	transformCacheMutex sync.RWMutex
)

type CallContext struct {
	goja.FunctionCall
	VM *goja.Runtime
}

func New(target types.Target, file string) (*Executor, error) {
	vm := goja.New()
	registry := require.NewRegistryWithLoader(templateSourceLoader)
	// NOTE: embed.FS access must always use forward slashes, even on Windows.
	scriptPath := path.Join("templates", target.Template, file)

	registry.Enable(vm)
	console.Enable(vm)

	jsData, err := templateSourceLoader(scriptPath)
	if err != nil {
		return nil, err
	}

	if _, err := vm.RunScript(scriptPath, string(jsData)); err != nil {
		return nil, err
	}

	return &Executor{
		vm:     vm,
		target: target,
		file:   file,
	}, nil
}

func NewWithVM(vm *goja.Runtime) *Executor {
	return &Executor{
		vm: vm,
	}
}

func NewWithEntrypoint(content string) *Executor {
	vm := goja.New()

	if _, err := vm.RunScript("entrypoint.js", content); err != nil {
		panic(err)
	}

	return &Executor{
		vm: vm,
	}
}

func (e *Executor) AddJSFuncs(funcs map[string]func(call CallContext) goja.Value) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	for name, fn := range funcs {
		wrappedFn := func(fn func(call CallContext) goja.Value) func(call goja.FunctionCall) goja.Value {
			return func(call goja.FunctionCall) goja.Value {
				return fn(CallContext{
					FunctionCall: call,
					VM:           e.vm,
				})
			}
		}(fn)

		if err := e.vm.Set(name, wrappedFn); err != nil {
			return fmt.Errorf("failed to set js function %s: %w", name, err)
		}
	}

	return nil
}

func (e *Executor) Run(fnName string, args ...any) (any, error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	fn, ok := goja.AssertFunction(e.vm.Get(fnName))
	if !ok {
		return nil, fmt.Errorf("failed to find %s function %s %s", fnName, e.target.Template, e.file)
	}

	gojaArgs := make([]goja.Value, len(args))
	for i, arg := range args {
		gojaArgs[i] = e.vm.ToValue(arg)
	}

	val, err := fn(goja.Undefined(), gojaArgs...)
	if err != nil {
		return nil, err
	}

	return val.Export(), nil
}

func (e *Executor) ToValue(i any) goja.Value {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	return e.vm.ToValue(i)
}

// Enables require() of template TypeScript source files.
func templateSourceLoader(path string) ([]byte, error) {
	tsData, err := templates.TemplateFS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read template path %s: %w", path, err)
	}

	jsData, err := transformTypeScript(path, tsData)
	if err != nil {
		return nil, fmt.Errorf("unable to convert TypeScript to JavaScript: %w", err)
	}

	return jsData, nil
}

// Converts the given source TypeScript code into JavaScript code.
// Results are cached since templates are embedded and immutable.
func transformTypeScript(path string, data []byte) ([]byte, error) {
	transformCacheMutex.RLock()
	if cached, ok := transformCache[path]; ok {
		transformCacheMutex.RUnlock()
		return cached, nil
	}
	transformCacheMutex.RUnlock()

	result := esbuild.Transform(string(data), esbuild.TransformOptions{
		Target:    esbuild.ES2015,
		Loader:    esbuild.LoaderTS,
		Sourcemap: esbuild.SourceMapExternal,
	})

	if len(result.Errors) > 0 {
		var msg strings.Builder

		for _, errMsg := range result.Errors {
			if errMsg.Location == nil {
				fmt.Fprintf(&msg, "%v @ %v;", errMsg.Text, path)
			} else {
				fmt.Fprintf(&msg, "%v @ %v %v:%v;", errMsg.Text, path, errMsg.Location.Line, errMsg.Location.Column)
			}
		}

		return nil, fmt.Errorf("script compilation failed: %s", msg.String())
	}

	transformCacheMutex.Lock()
	transformCache[path] = result.Code
	transformCacheMutex.Unlock()

	return result.Code, nil
}
