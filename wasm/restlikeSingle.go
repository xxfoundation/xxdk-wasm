////////////////////////////////////////////////////////////////////////////////
// Copyright © 2020 xx network SEZC                                           //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file                                                               //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"gitlab.com/elixxir/client/bindings"
	"syscall/js"
)

// RestlikeCallback wraps Javascript callbacks to adhere to the
// [bindings.RestlikeCallback] interface.
type restlikeCallback struct {
	callback func(args ...interface{}) js.Value
}

func (rlc *restlikeCallback) Callback(payload []byte, err error) {
	rlc.callback(CopyBytesToJS(payload), err.Error())
}

// RequestRestLike sends a restlike request to a given contact.
//
// Parameters:
//  - args[0] - ID of E2e object in tracker (int).
//  - args[1] - marshalled recipient [contact.Contact] (Uint8Array).
//  - args[2] - JSON of [bindings.RestlikeMessage] (Uint8Array).
//  - args[3] - JSON of [single.RequestParams] (Uint8Array).
//
// Returns:
//  - JSON of [restlike.Message] (Uint8Array).
//  - Throws a TypeError if parsing the parameters or making the request fails.
func RequestRestLike(_ js.Value, args []js.Value) interface{} {
	e2eID := args[0].Int()
	recipient := CopyBytesToGo(args[1])
	request := CopyBytesToGo(args[2])
	paramsJSON := CopyBytesToGo(args[3])

	msg, err := bindings.RequestRestLike(e2eID, recipient, request, paramsJSON)
	if err != nil {
		Throw(TypeError, err)
		return nil
	}

	return CopyBytesToJS(msg)
}

// AsyncRequestRestLike sends an asynchronous restlike request to a given
// contact.
//
// The RestlikeCallback will be called with the results of JSON marshalling the
// response when received.
//
// Parameters:
//  - args[0] - ID of E2e object in tracker (int).
//  - args[1] - marshalled recipient [contact.Contact] (Uint8Array).
//  - args[2] - JSON of [bindings.RestlikeMessage] (Uint8Array).
//  - args[3] - JSON of [single.RequestParams] (Uint8Array).
//  - args[4] - Javascript object that has functions that implement the
//    [bindings.RestlikeCallback] interface.
//
// Returns:
//  - Throws a TypeError if parsing the parameters or making the request fails.
func AsyncRequestRestLike(_ js.Value, args []js.Value) interface{} {
	e2eID := args[0].Int()
	recipient := CopyBytesToGo(args[1])
	request := CopyBytesToGo(args[2])
	paramsJSON := CopyBytesToGo(args[3])
	cb := &restlikeCallback{WrapCB(args[4], "Callback")}

	err := bindings.AsyncRequestRestLike(
		e2eID, recipient, request, paramsJSON, cb)
	if err != nil {
		Throw(TypeError, err)
		return nil
	}

	return nil
}
