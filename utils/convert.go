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
	"syscall/js"
)

// CopyBytesToGo copies the [Uint8Array] stored in the [js.Value] to []byte.
// This is a wrapper for [js.CopyBytesToGo] to make it more convenient.
func CopyBytesToGo(src js.Value) []byte {
	b := make([]byte, src.Length())
	js.CopyBytesToGo(b, src)
	return b
}

// CopyBytesToJS copies the []byte to a [Uint8Array] stored in a [js.Value].
// This is a wrapper for [js.CopyBytesToJS] to make it more convenient.
func CopyBytesToJS(src []byte) js.Value {
	dst := Uint8Array.New(len(src))
	js.CopyBytesToJS(dst, src)
	return dst
}

// JsToJson converts the Javascript value to JSON.
func JsToJson(value js.Value) string {
	return JSON.Call("stringify", value).String()
}

// JsonToJS converts a JSON bytes input to a [js.Value] of the object subtype.
func JsonToJS(inputJson []byte) (js.Value, error) {
	var jsObj map[string]any
	err := json.Unmarshal(inputJson, &jsObj)
	if err != nil {
		return js.ValueOf(nil), err
	}

	return js.ValueOf(jsObj), nil
}
