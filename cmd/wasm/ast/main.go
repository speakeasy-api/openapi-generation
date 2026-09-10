//go:build js && wasm
// +build js,wasm

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"syscall/js"
	"time"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/fastAST"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/mode"
	"github.com/speakeasy-api/openapi/openapi"
)

// version is set at build time via -ldflags
var version = "dev"

func main() {
	js.Global().Set("Healthcheck", promisify(func(args []js.Value) (string, error) {
		js.Global().Get("console").Call("log", "💓 AST Health check in Go")
		return "{\"success\": true}", nil
	}))

	js.Global().Set("SerializeSandboxAST", promisify(func(args []js.Value) (string, error) {
		ctx := context.Background()
		ctx = mode.SetSpeakeasyExecutionContextValidation(ctx)
		ctx = mode.SetSpeakeasyExecutionContextEmbedded(ctx)
		js.Global().Get("console").Call("log", "🔄 SerializeAST called with", len(args), "arguments")

		if len(args) < 1 {
			return "", fmt.Errorf("expected at least one argument")
		}

		// Convert the JavaScript object to a JSON string
		schemaStr := args[0].Index(0).String()

		js.Global().Get("console").Call("log", "🔄 Validating schema")

		startBuildDocTime := time.Now()
		doc, _, err := openapi.Unmarshal(ctx, bytes.NewBufferString(schemaStr), openapi.WithSkipValidation())
		if err != nil {
			return "", err
		}

		js.Global().Get("console").Call("log", "🔄 BuildV3Model took:", time.Since(startBuildDocTime).String())

		startDocsASTTime := time.Now()
		// DocsAST initialization
		fastAST, err := fastAST.NewFastAST(ctx, doc)
		if err != nil {
			return "", err
		}

		js.Global().Get("console").Call("log", "🔄 FastAST initialization took:", time.Since(startDocsASTTime).String())

		js.Global().Get("console").Call("log", "🔄 Marshalling FastAST")
		startMarshallingTime := time.Now()
		marshalledJSON, err := json.Marshal(fastAST)
		if err != nil {
			return "", err
		}

		js.Global().Get("console").Call("log", "🔄 Marshalling FastAST took:", time.Since(startMarshallingTime).String())
		js.Global().Get("console").Call("log", "🔄 Returning FastAST")
		return string(marshalledJSON), nil
	}))

	<-make(chan bool)
}
