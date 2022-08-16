////////////////////////////////////////////////////////////////////////////////
// Copyright © 2020 xx network SEZC                                           //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file                                                               //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package bindings

import (
	"gitlab.com/elixxir/client/bindings"
	"syscall/js"
)

// GenerateSecret creates a secret password using a system-based pseudorandom
// number generator.
//
// Parameters:
//  - args[0] - The size of secret. It should be set to 32, but can be set
//   higher in certain cases, but not lower (int).
//
// Returns:
//  - secret password (Uint8Array).
func GenerateSecret(_ js.Value, args []js.Value) interface{} {
	secret := bindings.GenerateSecret(args[0].Int())
	return CopyBytesToJS(secret)
}
