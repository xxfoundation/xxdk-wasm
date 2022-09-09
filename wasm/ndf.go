////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"gitlab.com/elixxir/client/bindings"
	"gitlab.com/elixxir/xxdk-wasm/utils"
	"syscall/js"
)

// DownloadAndVerifySignedNdfWithUrl retrieves the NDF from a specified URL.
// The NDF is processed into a protobuf containing a signature that is verified
// using the cert string passed in. The NDF is returned as marshaled byte data
// that may be used to start a client.
//
// Parameters:
//  - args[0] - The URL to download from (string).
//  - args[1] - The NDF certificate (string).
// Returns:
//  - JSON of the NDF (Uint8Array).
func DownloadAndVerifySignedNdfWithUrl(_ js.Value, args []js.Value) interface{} {
	ndf, err := bindings.DownloadAndVerifySignedNdfWithUrl(
		args[0].String(), args[1].String())
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return utils.CopyBytesToJS(ndf)
}
