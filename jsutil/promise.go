////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package jsutil

import (
	"fmt"
	"syscall/js"

	jww "github.com/spf13/jwalterweatherman"
)

// CreatePromise creates a JavaScript Promise that properly handles panics.
//
// The callback function receives resolve and reject functions that can be called
// to resolve or reject the Promise. The callback runs in a goroutine with panic
// recovery - any panic will automatically reject the Promise with a JavaScript Error.
//
// Usage:
//
//	func MyWasmFunc(_ js.Value, args []js.Value) any {
//	    return CreatePromise(func(resolve, reject func(args ...any) js.Value) {
//	        // Copy args immediately at start of promise callback
//	        key := args[0].String()
//
//	        result, err := someOperation(key)
//	        if err != nil {
//	            errorConstructor := js.Global().Get("Error")
//	            errorObject := errorConstructor.New(err.Error())
//	            reject(errorObject)
//	            return
//	        }
//
//	        resolve(result)
//	    })
//	}
func CreatePromise(fn func(resolve, reject func(args ...any) js.Value)) js.Value {
	// Channel to signal when the goroutine is done and handler can be released
	done := make(chan struct{})

	handler := js.FuncOf(func(_ js.Value, promiseArgs []js.Value) any {
		resolve := promiseArgs[0]
		reject := promiseArgs[1]

		// Run in a goroutine to avoid blocking JS event loop and to prevent
		// GC issues that occur when heavy computation (like argon2) runs
		// inside a synchronous JS callback.
		//
		// IMPORTANT: Callers MUST copy all js.Value data to Go types BEFORE
		// calling CreatePromise. The goroutine runs later when js.Value
		// references from the original call may be invalid.
		go func() {
			defer close(done)

			// Panic recovery - catch any panic and reject Promise
			defer func() {
				if r := recover(); r != nil {
					// Panic occurred - create JavaScript Error and reject Promise
					errorMsg := fmt.Sprintf("Go panic: %v", r)
					jww.ERROR.Printf("Panic recovered: %+v", r)

					errorConstructor := js.Global().Get("Error")
					errorObject := errorConstructor.New(errorMsg)
					reject.Invoke(errorObject)
				}
			}()

			// Call user's function with wrapper functions for resolve/reject
			fn(func(args ...any) js.Value {
				if len(args) == 0 {
					resolve.Invoke(js.Undefined())
				} else {
					resolve.Invoke(args[0])
				}
				return js.Undefined()
			}, func(args ...any) js.Value {
				if len(args) > 0 {
					reject.Invoke(args[0])
				}
				return js.Undefined()
			})
		}()

		return js.Undefined()
	})

	promise := Promise.New(handler)

	// Release handler in a goroutine after the main goroutine is done
	// This ensures the handler stays valid while resolve/reject might be called
	go func() {
		<-done
		handler.Release()
	}()

	return promise
}

// RejectWithError is a helper function that creates a proper JavaScript Error
// object and rejects a Promise with it. This standardizes error handling across
// all WASM bindings.
//
// Usage:
//
//	return CreatePromise(func(resolve, reject func(args ...any) js.Value) {
//	    result, err := someOperation()
//	    if err != nil {
//	        RejectWithError(reject, err)
//	        return
//	    }
//	    resolve(result)
//	})
func RejectWithError(reject func(...any) js.Value, err error) {
	errorConstructor := js.Global().Get("Error")
	errorObject := errorConstructor.New(err.Error())
	reject(errorObject)
}

// Await waits on a Javascript value. It blocks until the awaitable successfully
// resolves to the result or rejects to err.
//
// If there is a result, err will be nil and vice versa.
func Await(awaitable js.Value) (result []js.Value, err []js.Value) {
	then := make(chan []js.Value)
	defer close(then)

	thenFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		then <- args
		return js.Undefined()
	})
	defer thenFunc.Release()

	catch := make(chan []js.Value)
	defer close(catch)

	catchFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		catch <- args
		return js.Undefined()
	})
	defer catchFunc.Release()

	awaitable.Call("then", thenFunc).Call("catch", catchFunc)

	select {
	case result = <-then:
		return result, nil
	case err = <-catch:
		return nil, err
	}
}
