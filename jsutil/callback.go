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
	"strings"
	"syscall/js"

	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
)

// WrapCB wraps a Javascript function in an object so that it can be called
// later with only the arguments and without specifying the function name.
//
// Panics if m is not a function.
func WrapCB(parent js.Value, m string) func(args ...any) js.Value {
	if parent.Get(m).Type() != js.TypeFunction {
		// Create the error separate from the print so stack trace is printed
		err := errors.Errorf("Function %q is not of type %s", m, js.TypeFunction)
		jww.FATAL.Panicf("%+v", err)
	}

	return func(args ...any) js.Value { return parent.Call(m, args...) }
}

// ErrFromPanic converts a recovered panic value into an error.
// Use this in defer/recover blocks to convert panics to errors.
//
// Example:
//
//	func MyFunc() (err error) {
//	    defer func() {
//	        if r := recover(); r != nil {
//	            err = utils.ErrFromPanic(r)
//	        }
//	    }()
//	    // ... code that might panic
//	    return nil
//	}
func ErrFromPanic(r any) error {
	// Suppress "Go program has already exited" spam - these flood the
	// console after a crash and hide the original crash trace
	panicStr := fmt.Sprintf("%v", r)
	if !strings.Contains(panicStr, "Go program has already exited") {
		jww.ERROR.Printf("Panic recovered: %+v", r)
	}
	return fmt.Errorf("panic: %v", r)
}

// FuncOf creates a js.Func with panic recovery.
// Unlike js.FuncOf, this returns an error if the function creation panics.
// The returned js.Func must be released when no longer needed.
//
// The callback function receives js.Value arguments directly. Any panic
// inside the callback is recovered and logged.
func FuncOf(fn func(this js.Value, args []js.Value) any) (js.Func, error) {
	var f js.Func
	var err error

	defer func() {
		if r := recover(); r != nil {
			err = ErrFromPanic(r)
		}
	}()

	f = js.FuncOf(func(this js.Value, args []js.Value) any {
		// Recover from panics inside the callback
		defer func() {
			if r := recover(); r != nil {
				ErrFromPanic(r)
			}
		}()
		return fn(this, args)
	})

	return f, err
}

// ReleaseFunc releases a js.Func.
func ReleaseFunc(f js.Func) {
	f.Release()
}

// Call invokes a method on a js.Value.
// Returns the result and any error (converted from panic).
func Call(v js.Value, method string, args ...any) (js.Value, error) {
	var result js.Value
	var err error

	defer func() {
		if r := recover(); r != nil {
			err = ErrFromPanic(r)
		}
	}()

	result = v.Call(method, args...)
	return result, err
}

// Get retrieves a property from a js.Value.
// Returns the value and any error (converted from panic).
func Get(v js.Value, property string) (js.Value, error) {
	var result js.Value
	var err error

	defer func() {
		if r := recover(); r != nil {
			err = ErrFromPanic(r)
		}
	}()

	result = v.Get(property)
	return result, err
}
