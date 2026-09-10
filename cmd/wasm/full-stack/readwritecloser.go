//go:build js && wasm
// +build js,wasm

package main

import (
	"fmt"
	"io"
	"regexp"
	"syscall/js"
)

// JSReadWriteCloser implements io.ReadWriteCloser using JavaScript callbacks
type JSReadWriteCloser struct {
	asyncReadReady js.Value
	syncRead       js.Value
	notify         js.Value
	onAbort        js.Value

	chunkList       js.Value
	currentChunk    string
	currentChunkPos int  // Current position in buffer
	pendingRead     bool // Whether we have pending data to read

	// Content length pattern
	clPattern *regexp.Regexp
}

func NewJSReadWriteCloser(asyncReadReady, syncRead, notify, onAbort js.Value) *JSReadWriteCloser {
	return &JSReadWriteCloser{
		asyncReadReady: asyncReadReady,
		syncRead:       syncRead,
		notify:         notify,
		onAbort:        onAbort,
		pendingRead:    false,
		clPattern:      regexp.MustCompile(`Content-Length: (\d+)`),
	}
}

// Read implementation for io.ReadWriteCloser
func (rw *JSReadWriteCloser) Read(p []byte) (int, error) {
	// If we have data in our buffer, return it
	if rw.pendingRead && rw.currentChunkPos < len(rw.currentChunk) {
		copySize := min(len(p), len(rw.currentChunk)-rw.currentChunkPos)
		copy(p, rw.currentChunk[rw.currentChunkPos:rw.currentChunkPos+copySize])
		rw.currentChunkPos += copySize

		// If we've emptied the buffer, reset it
		if rw.currentChunkPos >= len(rw.currentChunk) {
			rw.pendingRead = false
			rw.currentChunkPos = 0
		}

		return copySize, nil
	}

	// Wait until read is ready
	promiseObj := rw.asyncReadReady.Invoke()

	// Only wait if the promise is not undefined/null
	if promiseObj.Type() != js.TypeUndefined && promiseObj.Type() != js.TypeNull {
		done := promiseToChannel(promiseObj)
		<-done
	}

	val := rw.syncRead.Invoke()

	if val.Type() == js.TypeNull {
		return 0, io.EOF
	}
	// val should be an array of strings
	isArray := js.Global().Get("Array").Call("isArray", val)
	if !isArray.Bool() {
		return 0, fmt.Errorf("expected array, got %v", val.Type())
	}
	// shift() the first one
	rw.pendingRead = true
	rw.currentChunk = val.Call("shift").String()

	rw.currentChunkPos = 0

	return rw.Read(p)
}

// Write implementation for io.ReadWriteCloser
func (rw *JSReadWriteCloser) Write(p []byte) (int, error) {
	message := string(p)

	// Log a summary of what we're sending
	if len(message) > 100 {
		js.Global().Get("console").Call("log", fmt.Sprintf("📤 Write: Sending %d bytes: %s... [truncated]", len(p), message[:100]))
	} else {
		js.Global().Get("console").Call("log", fmt.Sprintf("📤 Write: Sending %d bytes: %s", len(p), message))
	}

	// Send the message to JavaScript
	rw.notify.Invoke(message)

	return len(p), nil
}

// Close implementation for io.ReadWriteCloser
func (rw *JSReadWriteCloser) Close() error {
	js.Global().Get("console").Call("log", "🛑 Connection closing")
	rw.onAbort.Invoke()
	return nil
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
