package templates

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/dop251/goja"
	goversion "github.com/hashicorp/go-version"
	"github.com/mitchellh/mapstructure"
	"github.com/speakeasy-api/openapi-generation/v2/internal/executor"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/filesystem"
	config "github.com/speakeasy-api/sdk-gen-config"
)

func GetMinimumTargetVersion() map[string]string {
	return map[string]string{
		"java":       "v2",
		"python":     "v2",
		"typescript": "v2",
	}
}

// Deprecated: Use AllSupportedTemplateNames or GetSupportedSDKTemplateNames
// instead.
func GetSupportedLanguages() []string {
	return slices.Concat(
		GetSupportedSDKTemplateNames(),
		[]string{"terraform"},
	)
}

// Deprecated: Use IsSupportedTemplateName or IsSupportedSDKTemplateName
// instead.
func CheckLanguageSupported(language string) bool {
	return IsSupportedSDKTemplateName(language) || language == "terraform"
}

func GetTargetFromTargetString(target string) (types.Target, error) {
	configFields, err := GetLanguageConfigFields(types.NewTargetFromTemplate(target), true)
	if err != nil {
		return types.Target{}, err
	}

	templateVersion := ""
	for _, field := range configFields {
		if field.Name == "templateVersion" && field.DefaultValue != nil {
			templateVersion = (*field.DefaultValue).(string)
			break
		}
	}
	if templateVersion == "v1" {
		templateVersion = ""
	}

	if strings.HasSuffix(target, templateVersion) {
		cut, _ := strings.CutSuffix(target, templateVersion)
		return types.Target{
			Target:   cut,
			Template: target,
		}, nil
	}

	return types.Target{
		Target:   target,
		Template: fmt.Sprintf("%s%s", target, templateVersion),
	}, nil
}

// Deprecated: Use AllSupportedTargets or GetSupportedSDKTargets instead.
func GetSupportedTargets() []types.Target {
	return slices.Concat(
		GetSupportedSDKTargets(),
		[]types.Target{types.NewTargetFromTemplate("terraform")},
	)
}

func GetAvailableTemplates() []string {
	availableTemplateNames := []string{}

	for _, templateName := range templateNames() {
		if templateIsSunset(templateName) {
			continue
		}

		availableTemplateNames = append(availableTemplateNames, templateName)
	}

	return availableTemplateNames
}

func GetLanguageConfigFields(target types.Target, newSDK bool) ([]config.SDKGenConfigField, error) {
	commonFields, err := getCommonConfigFields(newSDK)
	if err != nil {
		return nil, err
	}

	e, err := executor.New(target, "config.ts")
	if err != nil {
		return nil, err
	}

	defVal, err := e.Run("getConfigFields", commonFields, newSDK)
	if err != nil {
		return nil, err
	}

	var configDetails map[string]config.SDKGenConfigField
	if err := mapstructure.Decode(defVal, &configDetails); err != nil {
		return nil, err
	}

	fields := []config.SDKGenConfigField{}
	for _, field := range configDetails {
		fields = append(fields, field)
	}

	return fields, nil
}

func getCommonConfigFields(newSDK bool) (map[string]config.SDKGenConfigField, error) {
	e, err := executor.New(types.NewTargetFromTemplate("common"), "common/config.ts")
	if err != nil {
		return nil, err
	}

	defVal, err := e.Run("getCommonConfigFields", newSDK)
	if err != nil {
		return nil, err
	}

	var configDetails map[string]config.SDKGenConfigField
	if err := mapstructure.Decode(defVal, &configDetails); err != nil {
		return nil, err
	}

	return configDetails, nil
}

func ResolveConfig(cfg *config.Configuration, target types.Target, outDir string, fs filesystem.FileSystem) error {
	e, err := executor.New(target, "config.ts")
	if err != nil {
		return fmt.Errorf("failed to create executor: %w", err)
	}
	err = e.AddJSFuncs(map[string]func(call executor.CallContext) goja.Value{
		"readFile": func(call executor.CallContext) goja.Value {
			// paths are relative to outdir
			absPath := filepath.Join(outDir, call.Argument(0).String())
			data, err := fs.ReadFile(absPath)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return call.VM.ToValue(nil)
				}
				panic(err)
			}
			return call.VM.ToValue(string(data))
		},
		"compareVersions": func(call executor.CallContext) goja.Value {
			a, err := goversion.NewVersion(call.Argument(0).String())
			if err != nil {
				panic(fmt.Errorf("compareVersions: invalid version %q: %w", call.Argument(0).String(), err))
			}
			b, err := goversion.NewVersion(call.Argument(1).String())
			if err != nil {
				panic(fmt.Errorf("compareVersions: invalid version %q: %w", call.Argument(1).String(), err))
			}
			return call.VM.ToValue(a.Compare(b))
		},
	})
	if err != nil {
		return fmt.Errorf("failed to add js funcs: %w", err)
	}

	// Check if the target has a language configuration
	langCfg, ok := cfg.Languages[target.Target]
	if !ok {
		return nil // No language config for this target
	}

	// Try to run resolveConfig function, but don't treat missing function as an error
	// Note: modifications are performed in-place. `cfg` is a reference so writes are visible.
	_, err = e.Run("resolveConfig", e.ToValue(langCfg.Cfg))
	if err != nil {
		// Check if the error is due to missing function
		if strings.Contains(err.Error(), "failed to find resolveConfig function") {
			// resolveConfig function doesn't exist, which is fine
			return nil
		}
		return fmt.Errorf("failed to resolve configuration: %w", err)
	}

	return nil
}
