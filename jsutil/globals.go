////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package jsutil

import (
	"syscall/js"
)

var (
	// JSON is the Javascript JSON type. It is used to perform JSON operations
	// on the Javascript layer.
	JSON = js.Global().Get("JSON")

	// Object is the Javascript Object type. It is used to perform Object
	// operations on the Javascript layer.
	Object = js.Global().Get("Object")

	// Promise is the Javascript Promise type. It is used to generate new
	// promises.
	Promise = js.Global().Get("Promise")

	// Uint8Array is the Javascript Uint8Array type. It is used to create new
	// Uint8Array.
	Uint8Array = js.Global().Get("Uint8Array")
)
