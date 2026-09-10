package template

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/speakeasy-api/openapi/sequencedmap"
	"github.com/spewerspew/spew"
	"gopkg.in/yaml.v3"

	"github.com/Masterminds/sprig/v3"
	"github.com/dop251/goja"
	"github.com/speakeasy-api/easytemplate"
	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/buckettypes"
	"github.com/speakeasy-api/openapi-generation/v2/internal/casing"
	"github.com/speakeasy-api/openapi-generation/v2/internal/configuration"
	"github.com/speakeasy-api/openapi-generation/v2/internal/env"
	"github.com/speakeasy-api/openapi-generation/v2/internal/examples"
	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/internal/format"
	"github.com/speakeasy-api/openapi-generation/v2/internal/imports"
	"github.com/speakeasy-api/openapi-generation/v2/internal/licensing"
	"github.com/speakeasy-api/openapi-generation/v2/internal/readme"
	"github.com/speakeasy-api/openapi-generation/v2/internal/sanitization"
	"github.com/speakeasy-api/openapi-generation/v2/internal/subsystem"
	"github.com/speakeasy-api/openapi-generation/v2/internal/terraform"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/cms"
	pkgerrors "github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/filetracking"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/markdown"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/patchfiles"
	"github.com/speakeasy-api/openapi-generation/v2/templates"
	"github.com/speakeasy-api/openapi-generation/v2/templates/dependencies/faker"
	"github.com/speakeasy-api/openapi-generation/v2/templates/dependencies/jsonpointer"
	"go.opentelemetry.io/otel/trace"
)

type (
	WriteFileFunc func(ctx context.Context, filename string, data []byte, perm fs.FileMode, checkExisting bool) error
	ReadFileFunc  func(filename string) ([]byte, error)
)

type Targeter interface {
	GetTarget(ctx context.Context, target, outDir string) (types.Target, error)
}

type Interopter interface {
	ExecuteInterTemplateFunction(ctx context.Context, call easytemplate.CallContext, target types.Target, a *ast.AST, method string, args ...any) goja.Value
	ExecuteInterTemplateTarget(ctx context.Context, target string, outDir string, a *ast.AST, cfg map[string]any, enabledFeatures map[string]bool) error
	GetTargetDefaultTemplateConfig(ctx context.Context, target types.Target) (map[string]any, error)
}

type WarningLogger interface {
	LogWarning(ctx context.Context, msg string, err error)
}

type Config struct {
	TemplateDir               string   // The main template directory to use
	AdditionalSearchLocations []string // Additional directories to search for templates, templates/common is always added
	Subsystem                 *subsystem.Subsystem
	Debug                     bool
	DebugGojaPort             int // DAP debug port for goja JS runtime (0 = disabled)
	WriteFileFunc             WriteFileFunc
	ReadFileFunc              ReadFileFunc
	FileTracker               *filetracking.Tracker
	Ignore                    *filetracking.Ignore
	Comments                  *cms.CommentTracker
	Targeter                  Targeter
	Interopter                Interopter
	OutDir                    string
	ImportConfig              configuration.ImportConfig
	WatchTemplatesLocation    string
	Bucketer                  *buckettypes.Bucketer
}

type Engine struct {
	*easytemplate.Engine
}

func New(ctx context.Context, target types.Target, cfg Config) *Engine {
	searchLocations := []string{cfg.TemplateDir}
	searchLocations = append(searchLocations, cfg.AdditionalSearchLocations...)

	if !slices.Contains(searchLocations, "templates/common") {
		searchLocations = append(searchLocations, "templates/common")
	}

	// Allow full paths to be used
	if !slices.Contains(searchLocations, "") {
		searchLocations = append(searchLocations, "")
	}

	tracer := trace.SpanFromContext(ctx).TracerProvider().Tracer("github.com/speakeasy-api/easytemplate")

	var templateFS fs.FS = &templates.TemplateFileSystem{}
	if cfg.WatchTemplatesLocation != "" {
		templateFS = os.DirFS(cfg.WatchTemplatesLocation)
	}

	opts := []easytemplate.Opt{
		easytemplate.WithTracer(tracer),
		easytemplate.WithReadFileSystem(templateFS),
		easytemplate.WithTemplateFuncs(getTmplFuncsMap(ctx, cfg)),
		easytemplate.WithWriteFunc(getWriteFileFunc(ctx, target, cfg)),
		easytemplate.WithSearchLocations(searchLocations),
		easytemplate.WithJSFuncs(getJSFuncs(ctx, target, cfg)),
		easytemplate.WithJSFiles(map[string]string{
			"faker":       faker.JS,
			"jsonpointer": jsonpointer.JS,
		}),
		easytemplate.WithRandSource(rand.New(rand.NewSource(0)).Float64),
	}

	if cfg.Debug {
		opts = append(opts, easytemplate.WithDebug())
	}

	if cfg.DebugGojaPort > 0 {
		opts = append(opts, easytemplate.WithDebugger(cfg.DebugGojaPort))
	}

	return &Engine{
		Engine: easytemplate.New(opts...),
	}
}

func getJSFuncs(ctx context.Context, target types.Target, cfg Config) map[string]func(call easytemplate.CallContext) goja.Value {
	caser := casing.NewWithSymbolCasingOverrides(getSymbolCasingOverrides(cfg))
	examples := examples.Examples{}

	return map[string]func(call easytemplate.CallContext) goja.Value{
		"sanitizeName": func(call easytemplate.CallContext) goja.Value {
			return call.VM.ToValue(sanitization.SanitizeName(call.Argument(0).String()))
		},
		"sanitizeFile": func(call easytemplate.CallContext) goja.Value {
			return call.VM.ToValue(sanitization.SanitizeFile(call.Argument(0).String(), call.Argument(1).String()))
		},
		"pythonRelativeImport": func(call easytemplate.CallContext) goja.Value {
			return call.VM.ToValue(imports.PythonRelativeImport(
				call.Argument(0).String(),
				call.Argument(1).String(),
			))
		},
		"caser": func(call easytemplate.CallContext) goja.Value {
			return call.VM.ToValue(caser)
		},
		"logger": func(call easytemplate.CallContext) goja.Value {
			return call.VM.ToValue(logging.From(ctx))
		},
		"logWarning": func(call easytemplate.CallContext) goja.Value {
			msg := call.Argument(0).String()
			lineNumber := int(call.Argument(1).ToInteger())
			logging.LogWarning(ctx, "validation warning", pkgerrors.NewValidationWarning(msg, &yaml.Node{Line: lineNumber, Column: 0}, nil))
			return call.VM.ToValue(nil)
		},
		"getDirectoryFiles": func(call easytemplate.CallContext) goja.Value {
			files, err := getDirectoryFiles(cfg.TemplateDir, call.Argument(0).String())
			if err != nil {
				panic(err)
			}

			return call.VM.ToValue(files)
		},
		"directoryExists": func(call easytemplate.CallContext) goja.Value {
			exists := directoryExists(cfg.TemplateDir, call.Argument(0).String())
			return call.VM.ToValue(exists)
		},
		"copy": func(call easytemplate.CallContext) goja.Value {
			if err := copyFile(ctx, cfg.TemplateDir, call.Argument(0).String(), call.Argument(1).String(), cfg.WriteFileFunc); err != nil {
				panic(err)
			}

			return call.VM.ToValue(nil)
		},
		"fileExists": func(call easytemplate.CallContext) goja.Value {
			_, err := cfg.ReadFileFunc(call.Argument(0).String())

			switch {
			case errors.Is(err, os.ErrNotExist):
				return call.VM.ToValue(false)
			case err != nil:
				panic(err)
			default:
				return call.VM.ToValue(true)
			}
		},
		"patchFileExists": func(call easytemplate.CallContext) goja.Value {
			// A patch under .speakeasy/patches pins its target file to
			// generator+patch ownership. Templates consult this so a
			// generated-once file whose content a patch owns keeps being
			// emitted deterministically: hermetic (fresh-tree) and
			// incremental regeneration then produce identical output.
			// Probed through ReadFileFunc, not os.Stat, so injected
			// filesystems (embedded/WASM generation) see their patches too.
			patchPath := patchfiles.RelPathFor(call.Argument(0).String())
			if _, err := cfg.ReadFileFunc(patchPath); err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return call.VM.ToValue(false)
				}
				panic(err)
			}
			return call.VM.ToValue(true)
		},
		"readFile": func(call easytemplate.CallContext) goja.Value {
			data, err := cfg.ReadFileFunc(call.Argument(0).String())
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return call.VM.ToValue(nil)
				}

				panic(err)
			}

			return call.VM.ToValue(string(data))
		},
		"listFiles": func(call easytemplate.CallContext) goja.Value {
			prefix := call.Argument(0).String()
			files := make(map[string]bool, 0)

			contents, _ := os.ReadDir(filepath.Join(cfg.OutDir, prefix))
			for _, entry := range contents {
				if !entry.IsDir() {
					files[entry.Name()] = true
				}
			}
			paths := make([]string, 0, len(files))
			for k := range files {
				paths = append(paths, k)
			}

			sort.Strings(paths)

			return call.VM.ToValue(paths)
		},
		"listFolders": func(call easytemplate.CallContext) goja.Value {
			prefix := call.Argument(0).String()
			subfolders := make(map[string]bool, 0)

			contents, _ := os.ReadDir(filepath.Join(cfg.OutDir, prefix))
			for _, entry := range contents {
				if entry.IsDir() {
					subfolders[entry.Name()] = true
				}
			}

			paths := make([]string, 0, len(subfolders))
			for k := range subfolders {
				paths = append(paths, k)
			}

			sort.Strings(paths)

			return call.VM.ToValue(paths)
		},
		"writeFile": func(call easytemplate.CallContext) goja.Value {
			outFile := call.Argument(0).String()

			if err := cfg.WriteFileFunc(ctx, outFile,
				[]byte(call.Argument(1).String()),
				fs.FileMode(call.Argument(2).ToInteger()),
				call.Argument(3).ToBoolean(),
			); err != nil {
				panic(err)
			}

			return goja.Undefined()
		},
		// These are temporary until goja supports dotAll
		"replaceReadmeBlock": func(call easytemplate.CallContext) goja.Value {
			prefix := call.Argument(0).String()
			title := call.Argument(1).String()
			id := call.Argument(2).String()
			suffix := call.Argument(3).String()
			contents := call.Argument(4).String()
			newTitle := call.Argument(5).String()
			newContents := call.Argument(6).String()
			adjustedContents, performedReplace := readme.ReplaceBlock(
				prefix,
				title,
				id,
				suffix,
				contents,
				newTitle,
				newContents,
			)
			if performedReplace {
				return call.VM.ToValue(adjustedContents)
			}

			return goja.Undefined()
		},
		"isReadmeSectionMarkedAsNoAction": func(call easytemplate.CallContext) goja.Value {
			prefix := call.Argument(0).String()
			title := call.Argument(1).String()
			id := call.Argument(2).String()
			suffix := call.Argument(3).String()
			contents := call.Argument(4).String()
			return call.VM.ToValue(readme.IsBlockDisabled(
				prefix,
				title,
				id,
				suffix,
				contents,
			))
		},
		"parseReadmeBlockIDs": func(call easytemplate.CallContext) goja.Value {
			prefix := call.Argument(0).String()
			suffix := call.Argument(1).String()
			contents := call.Argument(2).String()
			return call.VM.ToValue(readme.ParseBlockIDs(prefix, suffix, contents))
		},
		"parseReadmeBlock": func(call easytemplate.CallContext) goja.Value {
			prefix := call.Argument(0).String()
			id := call.Argument(1).String()
			suffix := call.Argument(2).String()
			contents := call.Argument(3).String()
			return call.VM.ToValue(readme.ParseBlockWithID(prefix, id, suffix, contents))
		},
		"generateReadmeTOC": func(call easytemplate.CallContext) goja.Value {
			contents := call.Argument(0).String()
			minLevel := int(call.Argument(1).ToInteger())
			maxLevel := int(call.Argument(2).ToInteger())
			exclude := strings.Split(call.Argument(3).String(), ",")
			return call.VM.ToValue(readme.GenerateTableOfContents(contents, minLevel, maxLevel, exclude))
		},
		"generateReadmeTOCEntry": func(call easytemplate.CallContext) goja.Value {
			header := call.Argument(0).String()
			level := int(call.Argument(1).ToInteger())
			return call.VM.ToValue(readme.GenerateTableOfContentsEntry(header, level, level, nil))
		},
		// Deprecated: registerCommentDef/registerComment + getComment form a
		// non-atomic register-then-read pair, so a colliding key written by a
		// concurrent render job can leak into the read. Kept for custom
		// templates; use resolveCommentDef/resolveComment instead.
		"registerCommentDef": func(call easytemplate.CallContext) goja.Value {
			source := call.Argument(0).String()
			ns := call.Argument(1).String()
			name := call.Argument(2).String()
			t := call.Argument(3).ExportType()
			switch {
			case t == nil:
				return call.VM.ToValue(
					cfg.Comments.RegisterComment(source, ns, name, &ast.Comment{}),
				)
			case t.AssignableTo(reflect.TypeOf(&ast.Comment{})):
				comment := call.Argument(3).Export().(*ast.Comment)
				return call.VM.ToValue(
					cfg.Comments.RegisterComment(source, ns, name, comment),
				)
			case t.AssignableTo(reflect.TypeOf(ast.Comment{})):
				comment := call.Argument(3).Export().(ast.Comment)
				return call.VM.ToValue(
					cfg.Comments.RegisterComment(source, ns, name, &comment),
				)
			}
			panic(fmt.Sprintf("invalid comment type for %s.%s: %v", ns, name, t))
		},
		"registerComment": func(call easytemplate.CallContext) goja.Value {
			source := call.Argument(0).String()
			ns := call.Argument(1).String()
			name := call.Argument(2).String()
			summary := call.Argument(3).String()
			description := call.Argument(4).String()
			return call.VM.ToValue(
				cfg.Comments.RegisterComment(source, ns, name, &ast.Comment{
					Summary:     summary,
					Description: description,
				}),
			)
		},
		"resolveCommentDef": func(call easytemplate.CallContext) goja.Value {
			source := call.Argument(0).String()
			ns := nullishString(call.Argument(1))
			kind := call.Argument(2).String()
			name := nullishString(call.Argument(3))

			var comment *ast.Comment
			t := call.Argument(4).ExportType()
			switch {
			case t == nil:
				comment = nil
			case t.AssignableTo(reflect.TypeOf(&ast.Comment{})):
				comment = call.Argument(4).Export().(*ast.Comment)
			case t.AssignableTo(reflect.TypeOf(ast.Comment{})):
				c := call.Argument(4).Export().(ast.Comment)
				comment = &c
			default:
				panic(fmt.Sprintf("invalid comment type for %s.%s: %v", ns, name, t))
			}

			return call.VM.ToValue(cfg.Comments.ResolveComment(source, ns, kind, name, comment))
		},
		"resolveComment": func(call easytemplate.CallContext) goja.Value {
			source := call.Argument(0).String()
			ns := nullishString(call.Argument(1))
			kind := call.Argument(2).String()
			name := nullishString(call.Argument(3))
			summary := call.Argument(4).String()
			description := call.Argument(5).String()

			return call.VM.ToValue(cfg.Comments.ResolveComment(source, ns, kind, name, &ast.Comment{
				Summary:     summary,
				Description: description,
			}))
		},
		"selectExampleOperations": func(call easytemplate.CallContext) goja.Value {
			sdk := call.Argument(0).Export().(*ast.SDK)
			limit := int(call.Argument(1).ToInteger())
			scopesArray := call.Argument(2).Export().([]interface{})
			isMainExample := call.Argument(3).ToBoolean()

			scopes := make([]ast.UsageExampleScope, 0, len(scopesArray))
			for _, arg := range scopesArray {
				args := arg.(map[string]interface{})
				scope := ast.UsageExampleScope{
					OpFilter: args["OpFilter"].(string),
					Feature:  args["Feature"].(string),
					IsGlobal: args["IsGlobal"].(bool),
				}
				value, ok := args["Value"]
				if ok {
					scope.Value = value
				}

				scopes = append(scopes, scope)
			}

			// Get ServerToShowInSnippets from configuration
			shouldIncludeServerSelection := cfg.Subsystem.Config.Generation.UsageSnippets != nil &&
				cfg.Subsystem.Config.Generation.UsageSnippets.ServerToShowInSnippets != ""

			ues := readme.SelectExampleOperations(sdk, scopes, limit, isMainExample, shouldIncludeServerSelection)
			return call.VM.ToValue(ues)
		},
		"fileUploadOpPredicate": func(call easytemplate.CallContext) goja.Value {
			sdk := call.Argument(0).Export().(*ast.SDK)
			op := call.Argument(1).Export().(*ast.Operation)
			res := readme.FileUploadOpPredicate(sdk, op)
			return call.VM.ToValue(res)
		},

		"getComment": func(call easytemplate.CallContext) goja.Value {
			name := call.Argument(0).String()
			return call.VM.ToValue(cfg.Comments.GetComment(name))
		},
		"isFeatureUsed": func(call easytemplate.CallContext) goja.Value {
			feature := call.Argument(0).String()
			return call.VM.ToValue(cfg.Subsystem.Features.IsFeatureUsed(features.FeatureFromString(feature)))
		},
		"registerReadmeSectionTemplated": func(call easytemplate.CallContext) goja.Value {
			id := call.Argument(0).String()
			feature := call.Argument(1).String()

			cfg.Subsystem.Features.RegisterReadmeSectionTemplated(id, features.FeatureFromString(feature))

			return goja.Undefined()
		},
		"adjustHeadings": func(call easytemplate.CallContext) goja.Value {
			level := call.Argument(0).ToInteger()
			content := call.Argument(1).String()
			return call.VM.ToValue(markdown.AdjustHeadings(int(level), content))
		},
		"interoptCallMethod": func(call easytemplate.CallContext) goja.Value {
			a := call.Argument(0).Export().(*ast.AST)
			target := call.Argument(1).String()
			method := call.Argument(2).String()
			args := make([]any, 0, len(call.Arguments[3:len(call.Arguments)]))
			for _, arg := range call.Arguments[3:len(call.Arguments)] {
				args = append(args, arg.Export())
			}

			t, err := cfg.Targeter.GetTarget(context.Background(), target, cfg.OutDir)
			if err != nil {
				panic(err)
			}

			return cfg.Interopter.ExecuteInterTemplateFunction(ctx, call, t, a, method, args...)
		},
		"interoptTemplateString": func(call easytemplate.CallContext) goja.Value {
			a := call.Argument(0).Export().(*ast.AST)
			target := call.Argument(1).String()
			templateFile := call.Argument(2).String()
			callCtx := call.Argument(3).Export()

			t, err := cfg.Targeter.GetTarget(ctx, target, cfg.OutDir)
			if err != nil {
				panic(err)
			}

			return cfg.Interopter.ExecuteInterTemplateFunction(ctx, call, t, a, "templateString",
				templateFile, callCtx)
		},
		"interoptTemplateTarget": func(call easytemplate.CallContext) goja.Value {
			target := call.Argument(0).String()
			outDir := call.Argument(1).String()
			a := call.Argument(2).Export().(*ast.AST)
			targetCfg := call.Argument(3).Export().(map[string]any)

			enabledFeaturesExport := call.Argument(4).Export()

			var enabledFeatures map[string]bool
			if enabledFeaturesExport != nil {
				enabledFeaturesAny := enabledFeaturesExport.(map[string]interface{})

				enabledFeatures = make(map[string]bool, len(enabledFeaturesAny))
				for k, v := range enabledFeaturesAny {
					enabledFeatures[k] = v.(bool)
				}
			}

			if err := cfg.Interopter.ExecuteInterTemplateTarget(ctx, target, outDir, a, targetCfg, enabledFeatures); err != nil {
				panic(err)
			}

			return goja.Undefined()
		},
		"getTargetDefaultTemplateConfig": func(call easytemplate.CallContext) goja.Value {
			target := call.Argument(0).Export().(string)

			t, err := cfg.Targeter.GetTarget(ctx, target, cfg.OutDir)
			if err != nil {
				panic(err)
			}

			cfg, err := cfg.Interopter.GetTargetDefaultTemplateConfig(ctx, t)
			if err != nil {
				panic(err)
			}

			return call.VM.ToValue(cfg)
		},
		"formatUsageSnippetOutput": func(call easytemplate.CallContext) goja.Value {
			fauxName, hasFauxName := usageSnippetFileName[target.Target]
			if !hasFauxName {
				fauxName = "example." + target.Target
			}
			snippet := call.Argument(0).String()
			output, err := format.Format(ctx, target, fauxName, []byte(snippet))
			if err != nil {
				logging.LogWarning(ctx, "failed to format a usage snippet", err)

				debugPath := filepath.Join("temp", fauxName)
				_ = cfg.WriteFileFunc(ctx, debugPath, []byte(snippet), defaultFileMode, true)

				return call.VM.ToValue(snippet)
			}
			return call.VM.ToValue(string(output))
		},
		"createUsageContext": func(call easytemplate.CallContext) goja.Value {
			var sdk *ast.SDK

			arg0 := call.Argument(0)
			if arg0.Export() != nil {
				sdkT := arg0.ExportType()
				switch {
				case sdkT.AssignableTo(reflect.TypeOf(&ast.SDK{})):
					sdk = arg0.Export().(*ast.SDK)
				case sdkT.AssignableTo(reflect.TypeOf(ast.SDK{})):
					// TODO where are we passing things not as pointers?
					s := arg0.Export().(ast.SDK)
					sdk = &s
				}
			}

			var operation *ast.Operation

			arg1 := call.Argument(1)
			if arg1.Export() != nil {
				operationT := arg1.ExportType()
				switch {
				case operationT.AssignableTo(reflect.TypeOf(&ast.Operation{})):
					operation = arg1.Export().(*ast.Operation)
				case operationT.AssignableTo(reflect.TypeOf(ast.Operation{})):
					// TODO where are we passing things not as pointers?
					o := arg1.Export().(ast.Operation)
					operation = &o
				}
			}

			t := call.Argument(2).ExportType()

			var exampleConfig *extensions.UsageExampleConfig

			switch {
			case t == nil:
				exampleConfig = &extensions.UsageExampleConfig{}
			case t.AssignableTo(reflect.TypeOf(&extensions.UsageExampleConfig{})):
				exampleConfig = call.Argument(2).Export().(*extensions.UsageExampleConfig)
			default:
				panic(fmt.Sprintf("invalid exampleConfig type: %v", call.Argument(2).ExportType()))
			}

			uc := ast.CreateUsageContext(sdk, operation, exampleConfig, cfg.Subsystem.Config.Generation.Tests.SkipResponseBodyAssertions)
			// Include server selection in usage snippets only when ServerToShowInSnippets is configured
			shouldIncludeServerSelection := cfg.Subsystem.Config.Generation.UsageSnippets != nil &&
				cfg.Subsystem.Config.Generation.UsageSnippets.ServerToShowInSnippets != ""
			uc.PopulateGlobalParameterScopes("", shouldIncludeServerSelection)

			return call.VM.ToValue(uc)
		},
		"sortTypeDefFieldCount": func(call easytemplate.CallContext) goja.Value {
			var slice ast.TypeDefs

			exportedSlice := call.Argument(0).Export()
			s, ok := exportedSlice.(ast.TypeDefs)
			if !ok {
				s2, ok := exportedSlice.(*ast.TypeDefs)
				if !ok {
					panic(fmt.Sprintf("invalid slice type: %v", call.Argument(0).ExportType()))
				}
				s = *s2
			}
			slice = s
			sliceCopy := make(ast.TypeDefs, len(slice))
			copy(sliceCopy, slice)
			fieldCountCompare := func(a *ast.TypeDef, b *ast.TypeDef) int {
				if a.Type == "class" && b.Type == "class" {
					return len(a.Fields) - len(b.Fields)
				}

				if a.Type == "class" {
					return -1
				}

				if b.Type == "class" {
					return 1
				}
				return 0
			}
			slices.SortFunc(sliceCopy, fieldCountCompare)
			return call.VM.ToValue(sliceCopy)
		},
		"sortTypeDefRequiredFieldsDescending": func(call easytemplate.CallContext) goja.Value {
			exportedSlice := call.Argument(0).Export()

			var sliceCopy ast.TypeDefs
			if typedSlice, ok := exportedSlice.(ast.TypeDefs); ok {
				// Direct slice - make a copy for sorting
				sliceCopy = make(ast.TypeDefs, len(typedSlice))
				copy(sliceCopy, typedSlice)
			} else if ptrSlice, ok := exportedSlice.(*ast.TypeDefs); ok {
				// Pointer to slice - make a copy for sorting
				slice := *ptrSlice
				sliceCopy = make(ast.TypeDefs, len(slice))
				copy(sliceCopy, slice)
			} else if anySlice, ok := exportedSlice.([]interface{}); ok {
				// Interface slice - convert directly to our target slice (no double copy)
				sliceCopy = make(ast.TypeDefs, len(anySlice))
				for i, item := range anySlice {
					sliceCopy[i] = item.(*ast.TypeDef)
				}
			} else {
				panic(fmt.Sprintf("invalid slice type: %v", call.Argument(0).ExportType()))
			}

			slices.SortStableFunc(sliceCopy, compareTypeDefsDescendingRequiredFields)
			return call.VM.ToValue(sliceCopy)
		},
		"sortDiscriminatorMappingRequiredFieldsDescending": func(call easytemplate.CallContext) goja.Value {
			exportedSlice := call.Argument(0).Export()

			var sliceCopy ast.DiscriminatorMappings
			if typedSlice, ok := exportedSlice.(ast.DiscriminatorMappings); ok {
				// Direct slice - make a copy for sorting
				sliceCopy = make(ast.DiscriminatorMappings, len(typedSlice))
				copy(sliceCopy, typedSlice)
			} else if ptrSlice, ok := exportedSlice.(*ast.DiscriminatorMappings); ok {
				// Pointer to slice - make a copy for sorting
				slice := *ptrSlice
				sliceCopy = make(ast.DiscriminatorMappings, len(slice))
				copy(sliceCopy, slice)
			} else if anySlice, ok := exportedSlice.([]interface{}); ok {
				// Interface slice - convert directly to our target slice (no double copy)
				sliceCopy = make(ast.DiscriminatorMappings, len(anySlice))
				for i, item := range anySlice {
					sliceCopy[i] = item.(*ast.DiscriminatorMapping)
				}
			} else {
				panic(fmt.Sprintf("invalid slice type: %v", call.Argument(0).ExportType()))
			}

			slices.SortStableFunc(sliceCopy, compareDiscriminatorMappingDescendingRequiredFields)
			return call.VM.ToValue(sliceCopy)
		},
		"getAllTypes": func(call easytemplate.CallContext) goja.Value {
			return call.VM.ToValue(cfg.Subsystem.Register.AllTypes())
		},
		"parseYaml": func(call easytemplate.CallContext) goja.Value {
			file, err := cfg.ReadFileFunc(call.Argument(0).String())
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return call.VM.ToValue(nil)
				}

				panic(err)
			}
			var data map[string]interface{}
			err = yaml.Unmarshal(file, &data)
			if err != nil {
				return call.VM.ToValue(nil)
			}

			return call.VM.ToValue(data)
		},
		"quote": func(call easytemplate.CallContext) goja.Value {
			quoted := fmt.Sprintf("%q", call.Argument(0).String())
			return call.VM.ToValue(quoted)
		},
		"withTimeout": func(call easytemplate.CallContext) goja.Value {
			callback := call.Argument(0)
			timeoutMs := call.Argument(1).ToInteger()
			jobId := call.Argument(2).String()
			jobFileName := call.Argument(3).String()

			// Execute the callback with a timeout - ensures the function completes within the specified time
			done := make(chan bool, 1)

			go func() {
				if callback != nil && !goja.IsUndefined(callback) && !goja.IsNull(callback) {
					if fn, ok := goja.AssertFunction(callback); ok {
						_, _ = fn(goja.Undefined())
					}
				}
				done <- true
			}()

			select {
			case <-done:
				// Function completed successfully within timeout
			case <-time.After(time.Duration(timeoutMs) * time.Millisecond):
				panic(fmt.Sprintf("withTimeout timed out for job %s %s", jobId, jobFileName))
				// Timeout occurred - the function may still be running but we don't wait for it
			}

			return goja.Undefined()
		},
		"debug": func(call easytemplate.CallContext) goja.Value {
			exported := make([]any, 0, len(call.Arguments))
			for _, arg := range call.Arguments {
				export := arg.Export()
				exported = append(exported, export)
			}
			builder := make([]string, 0, len(exported))
			// Put your breakpoint on the next line
			for _, arg := range exported {
				builder = append(builder, spew.Sprintf("%v", arg))
			}
			result := strings.Join(builder, ", ")
			return call.VM.ToValue(result)
		},
		"addUntrackedPattern": func(call easytemplate.CallContext) goja.Value {
			goRegex := call.Argument(0).String() // RE2 Regex pattern
			cfg.FileTracker.AddUntrackedPattern(regexp.MustCompile(goRegex))
			return goja.Undefined()
		},
		"accountHasFeatureAccess": func(call easytemplate.CallContext) goja.Value {
			feat := features.FeatureFromString(call.Argument(0).String())
			return call.VM.ToValue(licensing.AccountHasFeatureAccess(ctx, feat))
		},
		"getGenericMapTypeDef": func(call easytemplate.CallContext) goja.Value {
			return call.VM.ToValue(ast.NewType(&ast.TypeDef{
				Type: ast.DataTypeMap,
				ItemType: ast.NewType(&ast.TypeDef{
					Type: ast.DataTypeAny,
				}, nil),
			}, nil))
		},
		"getGenericArrayTypeDef": func(call easytemplate.CallContext) goja.Value {
			return call.VM.ToValue(ast.NewType(&ast.TypeDef{
				Type: ast.DataTypeArray,
				ItemType: ast.NewType(&ast.TypeDef{
					Type: ast.DataTypeAny,
				}, nil),
			}, nil))
		},
		"isRE2Regex": func(call easytemplate.CallContext) goja.Value {
			_, err := regexp.Compile(call.Argument(0).String())

			return call.VM.ToValue(err == nil)
		},
		"examplesHelper": func(call easytemplate.CallContext) goja.Value {
			return call.VM.ToValue(examples)
		},
		"isDebug": func(call easytemplate.CallContext) goja.Value {
			return call.VM.ToValue(env.IsDebug())
		},
		"addTypeToBucket": func(call easytemplate.CallContext) goja.Value {
			types := call.Argument(0).Export().(ast.BucketedTypes)
			model := call.Argument(1).String()
			typeDef := call.Argument(2).Export().(*ast.TypeDef)
			return call.VM.ToValue(buckettypes.AddTypeToBucket(types, model, typeDef))
		},
		"getOperationModels": func(call easytemplate.CallContext) goja.Value {
			operation := call.Argument(0).Export().(*ast.Operation)

			if operation == nil {
				return call.VM.ToValue(sequencedmap.New[string, *sequencedmap.Map[string, ast.TypeDefs]]())
			}

			ignoreResponseSubTypes := false
			if len(call.Arguments) > 1 {
				ignoreResponseSubTypes = call.Argument(1).ToBoolean()
			}
			ignoreResponseTypes := false
			if len(call.Arguments) > 2 {
				ignoreResponseTypes = call.Argument(2).ToBoolean()
			}
			considerFlattening := false
			if len(call.Arguments) > 3 {
				considerFlattening = call.Argument(3).ToBoolean()
			}
			return call.VM.ToValue(cfg.Bucketer.GetOperationModels(operation, ignoreResponseSubTypes, ignoreResponseTypes, considerFlattening))
		},
		"getOperationModelName": func(call easytemplate.CallContext) goja.Value {
			operation := call.Argument(0).Export().(*ast.Operation)
			return call.VM.ToValue(buckettypes.GetOperationModelName(cfg.Subsystem, operation))
		},
		"areCircular": func(call easytemplate.CallContext) goja.Value {
			t1 := call.Argument(0).Export().(*ast.TypeDef)
			t2 := call.Argument(1).Export().(*ast.TypeDef)
			return call.VM.ToValue(cfg.Subsystem.Cycles.AreCircular(t1, t2))
		},
		"isPartOfCycle": func(call easytemplate.CallContext) goja.Value {
			t := call.Argument(0).Export().(*ast.TypeDef)
			return call.VM.ToValue(cfg.Subsystem.Cycles.IsPartOfCycle(t))
		},
		"createBucket": func(call easytemplate.CallContext) goja.Value {
			existingBucket := call.Argument(0).Export().(ast.BucketedTypes)
			bucketName := call.Argument(1).String()

			existingBucket.Set(bucketName, sequencedmap.New[string, ast.TypeDefs]())

			return goja.Undefined()
		},
		"typeDefToFieldDef": func(call easytemplate.CallContext) goja.Value {
			typeDef := call.Argument(0).Export().(*ast.TypeDef)

			var parentFieldDef *ast.FieldDef
			if len(call.Arguments) > 1 {
				raw := call.Argument(1).Export()
				if raw != nil {
					parentFieldDef = raw.(*ast.FieldDef)
				}
			}
			optional := false
			if len(call.Arguments) > 2 {
				optional = call.Argument(2).ToBoolean()
			}
			nullable := false
			if len(call.Arguments) > 3 {
				nullable = call.Argument(3).ToBoolean()
			}

			name := typeDef.Name
			originalName := ""
			if parentFieldDef != nil {
				name = parentFieldDef.Name
				originalName = parentFieldDef.OriginalName
			}

			return call.VM.ToValue(&ast.FieldDef{
				Name:         name,
				OriginalName: originalName,
				Type:         typeDef,
				Optional:     optional,
				Nullable:     nullable,
			})
		},
		"base64Encode": func(call easytemplate.CallContext) goja.Value {
			return call.VM.ToValue(base64.StdEncoding.EncodeToString([]byte(call.Argument(0).String())))
		},
		"topologicalSortTypeDefs": func(call easytemplate.CallContext) goja.Value {
			types := call.Argument(0).Export().(ast.TypeDefs)
			return call.VM.ToValue(ast.TopologicalSortTypeDefs(types, call.Argument(1).ToBoolean()))
		},
		"createZeroTypeDef": func(call easytemplate.CallContext) goja.Value {
			return call.VM.ToValue(&ast.TypeDef{
				Extensions: &ast.TypeDefExtensions{
					All: make(map[string]any),
				},
			})
		},
		"jsonToHCLExpression": func(call easytemplate.CallContext) goja.Value {
			return call.VM.ToValue(terraform.JSONToHCLExpression(call.Argument(0).String()))
		},
	}
}

func nullishString(value goja.Value) string {
	if goja.IsNull(value) || goja.IsUndefined(value) {
		return ""
	}

	return value.String()
}

func getSymbolCasingOverrides(cfg Config) []string {
	if cfg.Subsystem == nil || cfg.Subsystem.Config == nil {
		return nil
	}

	rawCustomCasings := cfg.Subsystem.Config.GetLanguageConfigValue("customCasings")
	return casing.CustomCasingOverrides(rawCustomCasings)
}

// cachedSprigFuncs is computed once since sprig.FuncMap() is deterministic and
// allocates a new map with 100+ entries each time it's called.
var cachedSprigFuncs = sprig.FuncMap()

func getTmplFuncsMap(ctx context.Context, cfg Config) template.FuncMap {
	// Start with our custom funcs, then merge cached sprig funcs in.
	ourFuncs := map[string]any{
		"isFeatureUsed": func(feature string) bool {
			return cfg.Subsystem.Features.IsFeatureUsed(features.FeatureFromString(feature))
		},
		"adjustHeadings": markdown.AdjustHeadings,
		// skipFormat opts a template file out of the post-render formatting pipeline.
		// WARNING: the function must be the very first call in a .stmpl file, e.g.:
		// ```async.ts.stmpl
		// {{- skipFormat "a short string documenting the reason for skipping formatting"}}
		// [... rest of template ..]
		// ```
		"skipFormat": func(_reason string) string {
			return format.NoFormatMarker
		},
		"accountHasFeatureAccess": func(feat string) bool {
			return licensing.AccountHasFeatureAccess(ctx, features.FeatureFromString(feat))
		},
		// indent lines only if they contain content
		"indentPresent": func(amount int, block string) string {
			lines := strings.Split(block, "\n")
			finalLines := make([]string, len(lines))
			for i, line := range lines {
				if len(line) > 0 {
					finalLines[i] = strings.Repeat(" ", amount) + line
				} else {
					finalLines[i] = line
				}
			}
			return strings.Join(finalLines, "\n")
		},
	}

	// merge the two maps
	for k, v := range cachedSprigFuncs {
		if _, ok := ourFuncs[k]; ok {
			panic("duplicate function: " + k)
		}

		ourFuncs[k] = v
	}

	return ourFuncs
}

// TODO: Move this somewhere more appropriate
func compareTypeDefsDescendingRequiredFields(a, b *ast.TypeDef) int {
	countRequiredFields := func(t *ast.TypeDef) int {
		count := 0
		for _, field := range t.Fields {
			if !field.Optional {
				count++
			}
		}
		return count
	}

	if a.Type == "class" && b.Type == "class" {
		return countRequiredFields(b) - countRequiredFields(a)
	}

	if a.Type == "class" && countRequiredFields(a) > 0 {
		return -1
	}

	if b.Type == "class" && countRequiredFields(b) > 0 {
		return 1
	}

	return 0
}

func compareDiscriminatorMappingDescendingRequiredFields(a, b *ast.DiscriminatorMapping) int {
	return compareTypeDefsDescendingRequiredFields(a.Type, b.Type)
}
