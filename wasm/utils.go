////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import "syscall/js"

// jsArgsToAny converts a []js.Value slice to []any for use with js.Value.Invoke()
// which expects variadic ...any arguments.
//
// This is necessary because js.Value.Invoke() has the signature Invoke(args ...any),
// and Go 1.25+ strictly validates that arguments passed to syscall/js.ValueOf() are
// of supported types. Passing a []js.Value directly would cause a panic.
func jsArgsToAny(args []js.Value) []any {
	result := make([]any, len(args))
	for i, arg := range args {
		result[i] = arg
	}
	return result
}
