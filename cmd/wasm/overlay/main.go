//go:build js && wasm
// +build js,wasm

package main

import (
	"syscall/js"
)

func main() {
	js.Global().Set("Healthcheck", promisify(func(args []js.Value) (string, error) {
		js.Global().Get("console").Call("log", "💓 Overlay Health check in Go")
		return "{\"success\": true}", nil
	}))

	js.Global().Set("CalculateOverlay", promisify(func(args []js.Value) (string, error) {
		mainArgs := args[0].Index(0)

		originalStr := mainArgs.Get("original").String()
		modifiedStr := mainArgs.Get("modified").String()
		existingOverlayStr := mainArgs.Get("overlay").String()

		return CalculateOverlay(originalStr, modifiedStr, existingOverlayStr)
	}))

	js.Global().Set("ApplyOverlay", promisify(func(args []js.Value) (string, error) {
		mainArgs := args[0].Index(0)
		original := mainArgs.Get("original").String()
		overlay := mainArgs.Get("overlay").String()

		return ApplyOverlay(original, overlay)
	}))

	js.Global().Set("FormatYAML", promisify(func(args []js.Value) (string, error) {
		mainArgs := args[0].Index(0)
		yamlContent := mainArgs.Get("yamlContent").String()

		return FormatYAML(yamlContent)
	}))

	<-make(chan bool)
}
