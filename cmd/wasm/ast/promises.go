//go:build js && wasm
// +build js,wasm

package main

import (
	"fmt"
	"sync"
	"syscall/js"
)

func promisify(fn func(args []js.Value) (string, error)) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		// Handler for the Promise
		handler := js.FuncOf(func(this js.Value, promiseArgs []js.Value) interface{} {
			resolve := promiseArgs[0]
			reject := promiseArgs[1]

			// Run this code asynchronously
			go func() {
				defer func() {
					// recover from panics and return them as promise rejections
					if r := recover(); r != nil {
						errorConstructor := js.Global().Get("Error")
						errorObject := errorConstructor.New(fmt.Sprintf("panic occurred: %v", r))
						reject.Invoke(errorObject)
						return
					}
				}()

				result, err := fn(args)
				if err != nil {
					errorConstructor := js.Global().Get("Error")
					errorObject := errorConstructor.New(err.Error())
					reject.Invoke(errorObject)
					return
				}

				resolve.Invoke(result)
			}()

			// The handler of a Promise doesn't return any value
			return nil
		})

		// Create and return the Promise object
		promiseConstructor := js.Global().Get("Promise")
		return promiseConstructor.New(handler)
	})
}

func promiseToChannel(promise js.Value) chan struct{} {
	done := make(chan struct{})

	// Use a waitgroup to ensure proper cleanup
	var wg sync.WaitGroup
	wg.Add(1)

	// Use a single cleanup function to be called when either promise completes
	cleanup := sync.Once{}
	releaseCallbacks := func() {
		cleanup.Do(func() {
			// Signal done and close the channel
			done <- struct{}{}
			close(done)
			wg.Done()
		})
	}

	// Create callbacks that can safely be released
	successFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go releaseCallbacks()
		return nil
	})

	failureFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		js.Global().Get("console").Call("log", "❌ Promise rejected with error:")
		if len(args) > 0 {
			js.Global().Get("console").Call("error", args[0])
		}
		go releaseCallbacks()
		return nil
	})

	// Setup the promise callbacks
	promise.Call("then", successFunc).Call("catch", failureFunc)

	// Start a cleanup goroutine that will release the functions
	// after the promise resolves
	go func() {
		wg.Wait()
		// Only release after all operations are guaranteed to be done
		successFunc.Release()
		failureFunc.Release()
	}()

	return done
}

func await(awaitable js.Value) ([]js.Value, []js.Value) {
	then := make(chan []js.Value)
	defer close(then)
	thenFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		then <- args
		return nil
	})
	defer thenFunc.Release()

	catch := make(chan []js.Value)
	defer close(catch)
	catchFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		catch <- args
		return nil
	})
	defer catchFunc.Release()

	awaitable.Call("then", thenFunc).Call("catch", catchFunc)

	select {
	case result := <-then:
		return result, nil
	case err := <-catch:
		return nil, err
	}
}
