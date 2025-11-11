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
	"gitlab.com/elixxir/wasm-utils/utils"
)

// RestlikeCallback wraps Javascript callbacks to adhere to the
// [bindings.RestlikeCallback] interface.
type restlikeCallback struct {
	callback func(args ...any) js.Value
}

// Callback returns the response from an asynchronous restlike request.
//
// Parameters:
//   - payload - JSON of [restlike.Message] (Uint8Array).
//   - err - Returns an error on failure (Error).
func (rlc *restlikeCallback) Callback(payload []byte, err error) {
	var errVal js.Value
	if err != nil {
		errVal = js.Global().Get("Error").New(err.Error())
	} else {
		errVal = js.Undefined()
	}
	rlc.callback(utils.CopyBytesToJS(payload), errVal)
}

// RequestRestLike sends a restlike request to a given contact.
//
// Parameters:
//   - args[0] - ID of [E2e] object in tracker (int).
//   - args[1] - Marshalled bytes of the recipient [contact.Contact]
//     (Uint8Array).
//   - args[2] - JSON of [bindings.RestlikeMessage] (Uint8Array).
//   - args[3] - JSON of [single.RequestParams] (Uint8Array).
//
// Returns a promise:
//   - Resolves to the JSON of the [bindings.Message], which can be passed into
//     [Cmix.WaitForRoundResult] to see if the send succeeded (Uint8Array).
//   - Rejected with an error if parsing the parameters or making the request
//     fails.
func RequestRestLike(_ js.Value, args []js.Value) (any, error) {
	e2eID := args[0].Int()
	recipient := utils.CopyBytesToGo(args[1])
	request := utils.CopyBytesToGo(args[2])
	paramsJSON := utils.CopyBytesToGo(args[3])

	msg, err := bindings.RequestRestLike(
		e2eID, recipient, request, paramsJSON)
	if err != nil {
		return nil, err
	}

	return utils.CopyBytesToJS(msg), nil
}

// AsyncRequestRestLike sends an asynchronous restlike request to a given
// contact.
//
// The RestlikeCallback will be called with the results of JSON marshalling the
// response when received.
//
// Parameters:
//   - args[0] - ID of [E2e] object in tracker (int).
//   - args[1] - Marshalled bytes of the recipient [contact.Contact]
//     (Uint8Array).
//   - args[2] - JSON of [bindings.RestlikeMessage] (Uint8Array).
//   - args[3] - JSON of [single.RequestParams] (Uint8Array).
//   - args[4] - Javascript object that has functions that implement the
//     [bindings.RestlikeCallback] interface.
//
// Returns:
//   - Throws an error if parsing the parameters or making the request fails.
func AsyncRequestRestLike(_ js.Value, args []js.Value) (any, error) {
	e2eID := args[0].Int()
	recipient := utils.CopyBytesToGo(args[1])
	request := utils.CopyBytesToGo(args[2])
	paramsJSON := utils.CopyBytesToGo(args[3])
	cb := &restlikeCallback{utils.WrapCB(args[4], "Callback")}

	err := bindings.AsyncRequestRestLike(
		e2eID, recipient, request, paramsJSON, cb)
	if err != nil {
		return nil, err
	}

	return js.Undefined(), nil
}
