////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package utils

import (
	"encoding/base64"
	"fmt"
	"strings"
	"syscall/js"
	"testing"
)

var testBytes = [][]byte{
	nil,
	{},
	{0},
	{0, 1, 2, 3},
	{214, 108, 207, 78, 229, 11, 42, 219, 42, 87, 205, 104, 252, 73, 223,
		229, 145, 209, 79, 111, 34, 96, 238, 127, 11, 105, 114, 62, 239,
		130, 145, 82, 3},
}

// Tests that a series of Uint8Array Javascript objects are correctly converted
// to base 64 strings with Uint8ArrayToBase64.
func TestUint8ArrayToBase64(t *testing.T) {
	t.Errorf("ERROR")
	for i, val := range testBytes {
		// Create Uint8Array and set each element individually
		jsBytes := Uint8Array.New(len(val))
		for j, v := range val {
			jsBytes.SetIndex(j, v)
		}

		jsB64 := Uint8ArrayToBase64(js.Value{}, []js.Value{jsBytes})

		expected := base64.StdEncoding.EncodeToString(val)

		if expected != jsB64 {
			t.Errorf("Did not receive expected base64 encoded string (%d)."+
				"\nexpected: %s\nreceived: %s", i, expected, jsB64)
		}
	}
}

// Tests that Base64ToUint8Array correctly decodes a series of base 64 encoded
// strings into Uint8Array.
func TestBase64ToUint8Array(t *testing.T) {
	for i, val := range testBytes {
		b64 := base64.StdEncoding.EncodeToString(val)
		jsArr, err := base64ToUint8Array(js.ValueOf(b64))
		if err != nil {
			t.Errorf("Failed to convert js.Value to base 64: %+v", err)
		}

		// Generate the expected string to match the output of toString() on a
		// Uint8Array
		expected := strings.ReplaceAll(fmt.Sprintf("%d", val), " ", ",")[1:]
		expected = expected[:len(expected)-1]

		// Get the string value of the Uint8Array
		jsString := jsArr.Call("toString").String()

		if expected != jsString {
			t.Errorf("Failed to recevie expected string representation of "+
				"the Uint8Array (%d).\nexpected: %s\nreceived: %s",
				i, expected, jsString)
		}
	}
}

// Tests that a base 64 encoded string decoded to Uint8Array via
// Base64ToUint8Array and back to a base 64 encoded string via
// Uint8ArrayToBase64 matches the original.
func TestBase64ToUint8ArrayUint8ArrayToBase64(t *testing.T) {
	for i, val := range testBytes {
		b64 := base64.StdEncoding.EncodeToString(val)
		jsArr, err := base64ToUint8Array(js.ValueOf(b64))
		if err != nil {
			t.Errorf("Failed to convert js.Value to base 64: %+v", err)
		}

		jsB64 := Uint8ArrayToBase64(js.Value{}, []js.Value{jsArr})

		if b64 != jsB64 {
			t.Errorf("JSON from Uint8Array does not match original (%d)."+
				"\nexpected: %s\nreceived: %s", i, b64, jsB64)
		}
	}
}

func TestUint8ArrayEquals(t *testing.T) {
	for i, val := range testBytes {
		// Create Uint8Array and set each element individually
		jsBytesA := Uint8Array.New(len(val))
		for j, v := range val {
			jsBytesA.SetIndex(j, v)
		}

		jsBytesB := CopyBytesToJS(val)

		if !Uint8ArrayEquals(js.Value{}, []js.Value{jsBytesA, jsBytesB}).(bool) {
			t.Errorf("Two equal byte slices were found to be different (%d)."+
				"\nexpected: %s\nreceived: %s", i,
				jsBytesA.Call("toString").String(),
				jsBytesB.Call("toString").String())
		}
	}
}
