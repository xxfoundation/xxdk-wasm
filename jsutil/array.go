////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package jsutil

import (
	"bytes"
	"encoding/base64"
	"syscall/js"
)

// Uint8ArrayToBase64 encodes an uint8 array to a base 64 string.
//
// Parameters:
//   - args[0] - Javascript 8-bit unsigned integer array (Uint8Array).
//
// Returns:
//   - Base 64 encoded string (string).
func Uint8ArrayToBase64(_ js.Value, args []js.Value) any {
	return base64.StdEncoding.EncodeToString(CopyBytesToGo(args[0]))
}

// Base64ToUint8Array decodes a base 64 encoded string to a Uint8Array.
//
// Parameters:
//   - args[0] - Base 64 encoded string (string).
//
// Returns:
//   - Promise that resolves to Javascript 8-bit unsigned integer array (Uint8Array).
//   - Rejects with error if decoding the string fails.
func Base64ToUint8Array(_ js.Value, args []js.Value) any {
	// Copy args immediately to avoid race conditions with js.Value lifecycle
	input := args[0].String()

	return CreatePromise(func(resolve, reject func(...any) js.Value) {
		b, err := base64.StdEncoding.DecodeString(input)
		if err != nil {
			errorConstructor := js.Global().Get("Error")
			errorObject := errorConstructor.New(err.Error())
			reject(errorObject)
			return
		}
		resolve(CopyBytesToJS(b))
	})
}

// Uint8ArrayEquals returns true if the two Uint8Array are equal and false
// otherwise.
//
// Parameters:
//   - args[0] - Array A (Uint8Array).
//   - args[1] - Array B (Uint8Array).
//
// Returns:
//   - If the two arrays are equal (boolean).
func Uint8ArrayEquals(_ js.Value, args []js.Value) any {
	a := CopyBytesToGo(args[0])
	b := CopyBytesToGo(args[1])

	return bytes.Equal(a, b)
}
