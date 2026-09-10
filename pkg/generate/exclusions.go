package generate

import (
	"context"
	"os"
	"regexp"

	"github.com/dop251/goja"
	"github.com/speakeasy-api/openapi-generation/v2/internal/casing"
	"github.com/speakeasy-api/openapi-generation/v2/internal/executor"
	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/internal/licensing"
	"github.com/speakeasy-api/openapi-generation/v2/internal/sanitization"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/filetracking"
)

func (g *Generator) setupExclusions(ctx context.Context) error {
	e, err := executor.New(g.target, "exclusions.ts")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return err
	}

	caser := casing.New()
	if err := e.AddJSFuncs(map[string]func(call executor.CallContext) goja.Value{
		"caser": func(call executor.CallContext) goja.Value {
			return call.VM.ToValue(caser)
		},
		"addUntrackedPattern": func(call executor.CallContext) goja.Value {
			goRegex := call.Argument(0).String() // RE2 Regex pattern
			g.subsystem.FileTracker.AddUntrackedPattern(regexp.MustCompile(goRegex))
			return goja.Undefined()
		},
		"addHeaderPattern": func(call executor.CallContext) goja.Value {
			goRegex := call.Argument(0).String() // RE2 Regex pattern

			dontMatch := false
			if len(call.Arguments) > 1 {
				dontMatch = call.Argument(1).ToBoolean()
			}

			g.subsystem.FileTracker.AddHeaderPatterns = append(g.subsystem.FileTracker.AddHeaderPatterns, &filetracking.AddHeaderPattern{Pattern: regexp.MustCompile(goRegex), DontMatch: dontMatch})
			return goja.Undefined()
		},
		"accountHasFeatureAccess": func(call executor.CallContext) goja.Value {
			feat := features.FeatureFromString(call.Argument(0).String())
			return call.VM.ToValue(licensing.AccountHasFeatureAccess(ctx, feat))
		},
		"addIgnoreOverride": func(call executor.CallContext) goja.Value {
			g.ignore.AddIgnoreOverride(call.Argument(0).String()) // RE2 Regex pattern
			return goja.Undefined()
		},
		"sanitizeFile": func(call executor.CallContext) goja.Value {
			return call.VM.ToValue(sanitization.SanitizeFile(call.Argument(0).String(), call.Argument(1).String()))
		},
	}); err != nil {
		return err
	}

	if _, err := e.Run("setupExclusions", g.subsystem.Config.Generation.SDKClassName); err != nil {
		return err
	}

	return nil
}
