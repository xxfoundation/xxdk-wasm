////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"syscall/js"

	"gitlab.com/elixxir/client/v4/bindings"
	utils "gitlab.com/elixxir/xxdk-wasm/jsutil"
)

// DownloadAndVerifySignedNdfWithUrl retrieves the NDF from a specified URL.
// The NDF is processed into a protobuf containing a signature that is verified
// using the cert string passed in. The NDF is returned as marshaled byte data
// that may be used to start a client.
//
// Parameters:
//   - args[0] - The URL to download from (string).
//   - args[1] - The NDF certificate (string).
//
// Returns a promise:
//   - Resolves to the JSON of the NDF ([ndf.NetworkDefinition]) (Uint8Array).
//   - Rejected with an error if downloading fails.
func DownloadAndVerifySignedNdfWithUrl(_ js.Value, args []js.Value) any {
	// ✅ Parse ALL args BEFORE CreatePromise to avoid race conditions
	url := args[0].String()
	cert := args[1].String()

	return utils.CreatePromise(func(resolve, reject func(...any) js.Value) {
		ndf, err := bindings.DownloadAndVerifySignedNdfWithUrl(url, cert)
		if err != nil {
			errorConstructor := js.Global().Get("Error")
			errorObject := errorConstructor.New(err.Error())
			reject(errorObject)
			return
		}
		resolve(utils.CopyBytesToJS(ndf))
	})
}
