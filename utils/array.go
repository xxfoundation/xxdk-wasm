////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package utils

import (
	"bytes"
	"encoding/base64"
	"syscall/js"
)

// Uint8ArrayToBase64 encodes an uint8 array to a base 64 string.
//
// Parameters:
//  - args[0] - Javascript 8-bit unsigned integer array (Uint8Array).
//
// Returns:
//  - Base 64 encoded string (string).
func Uint8ArrayToBase64(_ js.Value, args []js.Value) interface{} {
	return base64.StdEncoding.EncodeToString(CopyBytesToGo(args[0]))
}

// Base64ToUint8Array decodes a base 64 encoded string to a Uint8Array.
//
// Parameters:
//  - args[0] - Base 64 encoded string (string).
//
// Returns:
//  - Javascript 8-bit unsigned integer array (Uint8Array).
//  - Throws TypeError if decoding the string fails.
func Base64ToUint8Array(_ js.Value, args []js.Value) interface{} {
	b, err := base64ToUint8Array(args[0])
	if err != nil {
		Throw(TypeError, err)
	}

	return b
}

// base64ToUint8Array is a helper function that returns an error instead of
// throwing it.
func base64ToUint8Array(base64String js.Value) (js.Value, error) {
	b, err := base64.StdEncoding.DecodeString(base64String.String())
	if err != nil {
		return js.Value{}, err
	}

	return CopyBytesToJS(b), nil
}

// Uint8ArrayEquals returns true if the two Uint8Array are equal and false
// otherwise.
//
// Parameters:
//  - args[0] - Array A (Uint8Array).
//  - args[1] - Array B (Uint8Array).
//
// Returns:
//  - If the two arrays are equal (boolean).
func Uint8ArrayEquals(_ js.Value, args []js.Value) interface{} {
	a := CopyBytesToGo(args[0])
	b := CopyBytesToGo(args[1])

	return bytes.Equal(a, b)
}
