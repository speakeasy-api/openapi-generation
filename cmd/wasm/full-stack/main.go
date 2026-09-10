//go:build js && wasm
// +build js,wasm

package main

import (
	"context"
	contextpkg "context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"syscall/js"

	"github.com/iancoleman/strcase"
	generationaccess "github.com/speakeasy-api/generation-context/access"
	"github.com/speakeasy-api/openapi-generation/v2/internal/analytics"
	"github.com/speakeasy-api/openapi-generation/v2/internal/docsData"
	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi-generation/v2/languageserver/server"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/mode"
	config "github.com/speakeasy-api/sdk-gen-config"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// version is set at build time via -ldflags
var version = "dev"

func main() {
	// Language Server initialization
	js.Global().Set("InitializeLS", promisify(func(args []js.Value) (string, error) {
		js.Global().Get("console").Call("log", "📥 InitializeLS called with", len(args), "arguments")

		if len(args) < 1 {
			err := fmt.Errorf("callback object with functions required")
			js.Global().Get("console").Call("error", "❌ Error:", err.Error())
			return "", err
		}

		callbacks := args[0]
		if callbacks.Type() != js.TypeObject {
			err := fmt.Errorf("first argument must be an object with functions")
			js.Global().Get("console").Call("error", "❌ Error:", err.Error())
			return "", err
		}

		js.Global().Get("console").Call("log", "📦 Got callbacks object")

		// Extract the callbacks
		asyncReadReady := callbacks.Get("asyncReadReady")
		syncRead := callbacks.Get("syncRead")
		notify := callbacks.Get("notify")
		onAbort := callbacks.Get("onAbort")

		js.Global().Get("console").Call("log", "📢 Retrieved all callback functions")

		// Validate the callbacks are functions
		hasAllFunctions := true

		if asyncReadReady.Type() != js.TypeFunction {
			js.Global().Get("console").Call("error", "❌ asyncReadReady is not a function")
			hasAllFunctions = false
		}

		if syncRead.Type() != js.TypeFunction {
			js.Global().Get("console").Call("error", "❌ syncRead is not a function")
			hasAllFunctions = false
		}

		if notify.Type() != js.TypeFunction {
			js.Global().Get("console").Call("error", "❌ notify is not a function")
			hasAllFunctions = false
		}

		if onAbort.Type() != js.TypeFunction {
			js.Global().Get("console").Call("error", "❌ onAbort is not a function")
			hasAllFunctions = false
		}

		if !hasAllFunctions {
			err := fmt.Errorf("callbacks must all be functions")
			js.Global().Get("console").Call("error", "❌ Error:", err.Error())
			return "", err
		}

		js.Global().Get("console").Call("log", "🚀 Starting Speakeasy Language Server")

		// Create a console logger
		jsLogger := &JSConsoleLogger{}

		// Set up context
		ctx := generationaccess.WithDirect(contextpkg.Background(), generationaccess.DisableTelemetry(), generationaccess.ElectAGPL())

		// Create the ReadWriteCloser with buffering support
		rwc := NewJSReadWriteCloser(
			asyncReadReady,
			syncRead,
			notify,
			onAbort,
		)

		logger := logging.NewLogger(zap.InfoLevel)
		fs := NewFileSystem(logger)

		// Create and start the language server with version from build
		ls := server.NewSpeakeasyServer(version, fs)

		// Set the logger and context on the server
		ls.Server.Log = jsLogger
		ls.Server.Context = ctx

		jsLogger.Info("Server capabilities initialized")
		js.Global().Get("console").Call("log", "✅ Language server created with console logging")

		// Serve using the JS interface
		go func() {
			js.Global().Get("console").Call("log", "🔄 Starting language server stream")
			ls.Server.ServeStream(rwc)
			js.Global().Get("console").Call("log", "🛑 Language server stream ended")
		}()

		js.Global().Get("console").Call("log", "✅ Speakeasy Language Server initialization complete")

		return "{\"success\": true}", nil
	}))

	// Health check
	js.Global().Set("Healthcheck", promisify(func(args []js.Value) (string, error) {
		js.Global().Get("console").Call("log", "💓 Full-Stack Health check in Go")
		return "{\"success\": true}", nil
	}))

	// Usage Snippets Generation
	js.Global().Set("GenerateUsageSnippets", promisify(func(args []js.Value) (string, error) {
		js.Global().Get("console").Call("log", "🔄 GenerateUsageSnippets called with", len(args), "arguments")

		if len(args) < 1 {
			return "", fmt.Errorf("expected at least one argument")
		}

		mainArgs := args[0].Index(0)
		js.Global().Get("console").Call("log", "🔄 GenerateUsageSnippets called with", args[0])

		schemaStr := mainArgs.Get("schema").String()
		operationIds := mainArgs.Get("operationIds").Call("join", ",")
		operationIdsArray := strings.Split(operationIds.String(), ",")
		genYamlConfig := mainArgs.Get("config")

		js.Global().Get("console").Call("debug", "🔄 Config:", genYamlConfig)
		js.Global().Get("console").Call("debug", "🔄 Operation IDs:", operationIds)

		target := mainArgs.Get("target").String()
		js.Global().Get("console").Call("debug", "🔄 Target:", target)

		ctx := context.Background()
		ctx = mode.SetSpeakeasyExecutionContextEmbedded(ctx)
		ctx = mode.SetSpeakeasyExecutionContextValidation(ctx)
		ctx = generationaccess.WithDirect(ctx, generationaccess.DisableTelemetry(), generationaccess.ElectAGPL())

		os.Setenv("SPEAKEASY_DEBUG_CONCURRENCY", "1")

		if len(operationIdsArray) > 0 {
			os.Setenv("SPEAKEASY_DEBUG_OPERATION_FILTER", strings.Join(operationIdsArray, ","))
		}

		logger := logging.NewLogger(zap.InfoLevel)
		fs := NewFileSystem(logger)

		targetType, err := generate.GetTargetFromTargetString(target)

		cfg, err := config.GetDefaultConfig(true, generate.GetLanguageConfigDefaults, map[string]bool{targetType.Target: true})
		if err != nil {
			return "", err
		}

		opts := []generate.GeneratorOptions{
			generate.WithFileSystem(fs),
			generate.WithLogger(logger),
		}

		if len(operationIdsArray) > 0 {
			js.Global().Get("console").Call("log", "🔄 Operation IDs length:", len(operationIdsArray))
			opts = append(opts, generate.WithUsageSnippetArgsByOperationID(strings.Join(operationIdsArray, ",")))
		}

		g, err := generate.New(opts...)
		if err != nil {
			return "", err
		}

		languageCfg, err := g.GetTargetDefaultTemplateConfig(ctx, targetType)
		if err != nil {
			return "", err
		}

		hasConfig := !genYamlConfig.IsUndefined()

		// Create a new map to ensure we're not modifying a read-only map
		newCfg := make(map[string]any, len(languageCfg))
		for k, v := range languageCfg {
			newCfg[strcase.ToLowerCamel(k)] = v
		}

		if hasConfig {
			if !genYamlConfig.Get("modulePath").IsUndefined() && !genYamlConfig.Get("modulePath").IsNull() {
				js.Global().Get("console").Call("log", "🔄 Setting modulePath to:", genYamlConfig.Get("modulePath").String())
				newCfg["modulePath"] = genYamlConfig.Get("modulePath").String()
			}
			if !genYamlConfig.Get("packageName").IsUndefined() && !genYamlConfig.Get("packageName").IsNull() {
				js.Global().Get("console").Call("log", "🔄 Setting packageName to:", genYamlConfig.Get("packageName").String())
				newCfg["packageName"] = genYamlConfig.Get("packageName").String()
			}
			if !genYamlConfig.Get("sdkPackageName").IsUndefined() && !genYamlConfig.Get("sdkPackageName").IsNull() {
				js.Global().Get("console").Call("log", "🔄 Setting sdkPackageName to:", genYamlConfig.Get("sdkPackageName").String())
				newCfg["sdkPackageName"] = genYamlConfig.Get("sdkPackageName").String()
			}

			js.Global().Get("console").Call("log", "🔄 Setting sdkClassName to:", genYamlConfig.Get("sdkClassName").String())
			cfg.Generation.SDKClassName = genYamlConfig.Get("sdkClassName").String()
		}
		languageCfg = newCfg

		cfg.Languages[targetType.Target] = config.LanguageConfig{
			Version: "2.0.0",
			Cfg:     languageCfg,
		}

		yamlBytes, err := yaml.Marshal(cfg)
		if err != nil {
			return "", err
		}

		fs.WriteFile("gen.yaml", yamlBytes, 0o644)

		genLock := `lockVersion: "2.0.0"
id: "speakeasy-openapi-generator"`
		fs.WriteFile("gen.lock", []byte(genLock), 0o644)

		if errs := g.Generate(ctx, []byte(schemaStr), "", targetType.Template, "", false, false); len(errs) > 0 {
			isValidationError := false
			for _, err := range errs {
				if strings.Contains(err.Error(), "validation error:") {
					isValidationError = true
				}
			}

			if isValidationError {
				return "", fmt.Errorf("failed to generate usage snippets: the provided openapi spec is not valid for sdk generation: %+v", errs)
			}

			return "", fmt.Errorf("failed to generate usage snippets: %+v", errs)
		}

		jsonFiles, err := fs.OutputSnippetsAsJSON()
		if err != nil {
			return "", err
		}

		js.Global().Get("console").Call("log", "🔄 Returning usage snippets", jsonFiles)

		return jsonFiles, nil
	}))

	js.Global().Set("SerializeDocsData", promisify(func(args []js.Value) (string, error) {
		ctx := context.Background()
		ctx = mode.SetSpeakeasyExecutionContextEmbedded(ctx)
		ctx = mode.SetSpeakeasyExecutionContextValidation(ctx)

		ctx = generationaccess.WithDirect(ctx, generationaccess.DisableTelemetry(), generationaccess.ElectAGPL())

		// Note: if we start running into errors cause we're trying to make syscalls
		// we can fix it by adding:
		//
		// ctx = mode.SetSpeakeasyExecutionContextEmbedded(ctx)
		//
		// We can't enable that today though because it requires a gen.yaml file to
		// work correctly. I ran into issues generating a default/empty one, and
		// this raises further questions about what a default gen.yaml file actually
		// means. Down the road, we'll likely want a non default one, but rather
		// we'll want one that was used to build every SDK, since we'll want to use
		// this info to document SDK specific namespaces, classes, functions,
		// properties and all that good stuff.

		if len(args) < 1 {
			return "", fmt.Errorf("expected at least one argument")
		}

		// Convert the JavaScript object to a JSON string
		schemaStr := args[0].String()
		logger := logging.NewLogger(zap.InfoLevel)

		fs := NewFileSystem(logger)

		gen, err := generate.New(generate.WithDontWrite(), generate.WithFileSystem(fs))
		if err != nil {
			return "", err
		}

		// Convert js.Value args to expected types
		docInfo := &document.DocumentInfo{
			Schema:     []byte(schemaStr),
			SchemaPath: "foo.json",
			IsRemote:   false,
		}

		// Initialize the generator with "." as outDir for file tracking
		err = gen.Init(ctx, "go", ".")
		if err != nil {
			return "", err
		}

		ad := &analytics.Data{}
		if _, err := gen.LoadAndValidateDoc(ctx, docInfo, ad, "", true); err != nil {
			return "", err
		}

		ast, err := gen.GenerateAST(ctx, docInfo, ad)
		if err != nil {
			return "", err
		}

		docs, err := docsData.NewDocsData(ast)
		if err != nil {
			return "", err
		}

		marshalledJSON, err := json.Marshal(docs)
		if err != nil {
			return "", err
		}

		return string(marshalledJSON), nil
	}))

	<-make(chan bool)
}
