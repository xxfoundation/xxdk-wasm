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

// RestlikeRequest performs a normal restlike request.
//
// Parameters:
//   - args[0] - ID of [Cmix] object in tracker (int).
//   - args[1] - ID of [Connection] object in tracker (int).
//   - args[2] - JSON of [bindings.RestlikeMessage] (Uint8Array).
//   - args[3] - JSON of [xxdk.E2EParams] (Uint8Array).
//
// Returns a promise:
//   - Resolves to the JSON of the [bindings.RestlikeMessage], which can be
//     passed into [Cmix.WaitForRoundResult] to see if the send succeeded
//     (Uint8Array).
//   - Rejected with an error if parsing the parameters or making the request
//     fails.
func RestlikeRequest(_ js.Value, args []js.Value) any {
	// ✅ Parse ALL args BEFORE CreatePromise to avoid race conditions
	cmixId := args[0].Int()
	connectionID := args[1].Int()
	request := utils.CopyBytesToGo(args[2])
	e2eParamsJSON := utils.CopyBytesToGo(args[3])

	return utils.CreatePromise(func(resolve, reject func(...any) js.Value) {
		msg, err := bindings.RestlikeRequest(
			cmixId, connectionID, request, e2eParamsJSON)
		if err != nil {
			errorConstructor := js.Global().Get("Error")
			errorObject := errorConstructor.New(err.Error())
			reject(errorObject)
			return
		}

		resolve(utils.CopyBytesToJS(msg))
	})
}

// RestlikeRequestAuth performs an authenticated restlike request.
//
// Parameters:
//   - args[0] - ID of [Cmix] object in tracker (int).
//   - args[1] - ID of [AuthenticatedConnection] object in tracker (int).
//   - args[2] - JSON of [bindings.RestlikeMessage] (Uint8Array).
//   - args[3] - JSON of [xxdk.E2EParams] (Uint8Array).
//
// Returns a promise:
//   - Resolves to the JSON of the [bindings.RestlikeMessage], which can be
//     passed into [Cmix.WaitForRoundResult] to see if the send succeeded
//     (Uint8Array).
//   - Rejected with an error if parsing the parameters or making the request
//     fails.
func RestlikeRequestAuth(_ js.Value, args []js.Value) any {
	// ✅ Parse ALL args BEFORE CreatePromise to avoid race conditions
	cmixId := args[0].Int()
	authConnectionID := args[1].Int()
	request := utils.CopyBytesToGo(args[2])
	e2eParamsJSON := utils.CopyBytesToGo(args[3])

	return utils.CreatePromise(func(resolve, reject func(...any) js.Value) {
		msg, err := bindings.RestlikeRequestAuth(
			cmixId, authConnectionID, request, e2eParamsJSON)
		if err != nil {
			errorConstructor := js.Global().Get("Error")
			errorObject := errorConstructor.New(err.Error())
			reject(errorObject)
			return
		}

		resolve(utils.CopyBytesToJS(msg))
	})
}
