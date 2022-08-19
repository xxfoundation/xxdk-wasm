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

// RestlikeRequest performs a normal restlike request.
//
// Parameters:
//  - args[0] - ID of Cmix object in tracker (int).
//  - args[1] - ID of Connection object in tracker (int).
//  - args[2] - JSON of [bindings.RestlikeMessage] (Uint8Array).
//  - args[3] - JSON of [xxdk.E2EParams] (Uint8Array).
//
// Returns:
//  - JSON of [bindings.RestlikeMessage] (Uint8Array).
//  - Throws a TypeError if parsing the parameters or making the request fails.
func RestlikeRequest(_ js.Value, args []js.Value) interface{} {
	cmixId := args[0].Int()
	connectionID := args[1].Int()
	request := CopyBytesToGo(args[2])
	e2eParamsJSON := CopyBytesToGo(args[3])

	msg, err := bindings.RestlikeRequest(
		cmixId, connectionID, request, e2eParamsJSON)
	if err != nil {
		Throw(TypeError, err)
		return nil
	}

	return CopyBytesToJS(msg)
}

// RestlikeRequestAuth performs an authenticated restlike request.
//
// Parameters:
//  - args[0] - ID of Cmix object in tracker (int).
//  - args[1] - ID of AuthenticatedConnection object in tracker (int).
//  - args[2] - JSON of [bindings.RestlikeMessage] (Uint8Array).
//  - args[3] - JSON of [xxdk.E2EParams] (Uint8Array).
//
// Returns:
//  - JSON of [bindings.RestlikeMessage] (Uint8Array).
//  - Throws a TypeError if parsing the parameters or making the request fails.
func RestlikeRequestAuth(_ js.Value, args []js.Value) interface{} {
	cmixId := args[0].Int()
	authConnectionID := args[1].Int()
	request := CopyBytesToGo(args[2])
	e2eParamsJSON := CopyBytesToGo(args[3])

	msg, err := bindings.RestlikeRequestAuth(
		cmixId, authConnectionID, request, e2eParamsJSON)
	if err != nil {
		Throw(TypeError, err)
		return nil
	}

	return CopyBytesToJS(msg)
}
