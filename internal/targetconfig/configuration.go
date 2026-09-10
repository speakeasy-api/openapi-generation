package targetconfig

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dop251/goja"
	"github.com/speakeasy-api/openapi-generation/v2/internal/executor"
	"github.com/speakeasy-api/openapi-generation/v2/internal/sanitization"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
)

// Creates a Configuration by calling into the target config.ts file
// getGeneratorConfig function.
func NewConfiguration(ctx context.Context, target types.Target, genConfig GeneratorInitialConfiguration) (*Configuration, error) {
	e, err := executor.New(target, "config.ts")

	if err != nil {
		return nil, fmt.Errorf("unable to create %s target config.ts executor: %w", target.Target, err)
	}

	// Shared functions not worth replicating in JavaScript or setting up
	// proper JavaScript imports in the runtime.
	configurationFuncs := map[string]func(call executor.CallContext) goja.Value{
		"directoryExists": func(call executor.CallContext) goja.Value {
			info, err := os.Stat(filepath.Join(genConfig.OutDir, call.Argument(0).String()))

			if err != nil {
				return call.VM.ToValue(false)
			}

			return call.VM.ToValue(info.IsDir())
		},
		"fileExists": func(call executor.CallContext) goja.Value {
			info, err := os.Stat(filepath.Join(genConfig.OutDir, call.Argument(0).String()))

			if err != nil {
				return call.VM.ToValue(false)
			}

			return call.VM.ToValue(!info.IsDir())
		},
		"sanitizeFile": func(call executor.CallContext) goja.Value {
			return call.VM.ToValue(sanitization.SanitizeFile(call.Argument(0).String(), call.Argument(1).String()))
		},
	}

	if err := e.AddJSFuncs(configurationFuncs); err != nil {
		return nil, fmt.Errorf("unable to add functions to %s target config.ts executor: %w", target.Target, err)
	}

	rawGojaConfiguration, err := e.Run("getGeneratorConfig", genConfig)

	if err != nil {
		return nil, fmt.Errorf("unable to call %s target config.ts getGeneratorConfig function: %w", target.Target, err)
	}

	gojaConfiguration, err := NewGojaConfiguration(rawGojaConfiguration)

	if err != nil {
		return nil, fmt.Errorf("unable to create %s target goja configuration from config.ts getGeneratorConfig returned value: %w", target.Target, err)
	}

	targetConfig, err := gojaConfiguration.Convert(ctx)

	if err != nil {
		return nil, fmt.Errorf("unable to convert %s target goja configuration: %w", target.Target, err)
	}

	return targetConfig, nil
}

// Represents the target implementation configuration presented to the generator
// from the target.
type Configuration struct {
	// Processes and dependencies necessary to compile the generated target
	// code. If unset, the target does not implement code compilation.
	Compile *CompileConfiguration

	// Target feature information.
	//
	// TODO: Switch from individual template features.ts function calls.
	// Features *FeatureConfiguration

	// Configuration for file tracking.
	//
	// TODO: Switch from exclusions.ts setupExclusions.
	// FileTracking *FileTrackingConfiguration

	// Enabled if the target should not be shown as supported for generation
	// externally.
	//
	// TODO: Switch from template HIDDEN file.
	// Hidden bool

	// Processes and dependencies necessary to perform static analysis of the
	// generated target code. If unset, the target does not implement code
	// linting.
	Lint *LintConfiguration

	// Configuration for mock server.
	//
	// TODO: Switch from mockserver.ts getMockServerDirectory.
	// MockServer *MockServerConfiguration

	// Configuration for the target README file.
	//
	// TODO: Switch from features.ts isReadmeSectionImplemented and
	// isReadmeSectionIgnored.
	// Readme *ReadmeConfiguration

	// Enabled if the target has been removed.
	//
	// TODO: Switch from template SUNSET file.
	// Sunset bool

	// Configuration for target testing.
	Testing *TestingConfiguration
}
