////////////////////////////////////////////////////////////////////////////////
// Copyright © 2020 xx network SEZC                                           //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file                                                               //
////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package utils

import (
	"encoding/json"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"syscall/js"
)

var (
	Error      = js.Global().Get("Error")
	JSON       = js.Global().Get("JSON")
	Promise    = js.Global().Get("Promise")
	Uint8Array = js.Global().Get("Uint8Array")
)

// CopyBytesToGo copies the Uint8Array stored in the js.Value to []byte. This is
// a wrapper for js.CopyBytesToGo to make it more convenient.
func CopyBytesToGo(src js.Value) []byte {
	b := make([]byte, src.Length())
	js.CopyBytesToGo(b, src)
	return b
}

// CopyBytesToJS copies the []byte to a Uint8Array stored in a js.Value. This is
// a wrapper for js.CopyBytesToJS to make it more convenient.
func CopyBytesToJS(src []byte) js.Value {
	dst := Uint8Array.New(len(src))
	js.CopyBytesToJS(dst, src)
	return dst
}

// WrapCB wraps a Javascript function in an object so that it can be called
// later with only the arguments and without specifying the function name.
//
// Panics if m is not a function.
func WrapCB(parent js.Value, m string) func(args ...interface{}) js.Value {
	if parent.Get(m).Type() != js.TypeFunction {
		// Create the error separate from the print so stack trace is printed
		err := errors.Errorf("Function %q is not of type %s", m, js.TypeFunction)
		jww.FATAL.Panicf("%+v", err)
	}

	return func(args ...interface{}) js.Value {
		return parent.Call(m, args...)
	}
}

// JsonToJS is a helper that converts JSON bytes input
// to a [js.Value] of the object subtype.
func JsonToJS(inputJson []byte) (js.Value, error) {
	jsObj := make(map[string]interface{})
	err := json.Unmarshal(inputJson, &jsObj)
	if err != nil {
		return js.Value{}, err
	}
	return js.ValueOf(jsObj), nil
}

// JsToJson converts the Javascript value to JSON.
func JsToJson(value js.Value) string {
	return JSON.Call("stringify", value).String()
}

type PromiseFn func(resolve, reject func(args ...interface{}) js.Value)

// CreatePromise creates a Javascript promise to return the value of a blocking
// Go function to Javascript.
func CreatePromise(f PromiseFn) interface{} {
	// Create handler for promise (this will be a Javascript function)
	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Spawn a new go routine to perform the blocking function
		go func(resolve, reject js.Value) {
			f(resolve.Invoke, reject.Invoke)
		}(args[0], args[1])

		return nil
	})

	// Create and return the Promise object
	return Promise.New(handler)
}
