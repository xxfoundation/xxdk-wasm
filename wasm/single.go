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

////////////////////////////////////////////////////////////////////////////////
// Public Wrapper Methods                                                     //
////////////////////////////////////////////////////////////////////////////////

// TransmitSingleUse transmits payload to recipient via single-use.
//
// Parameters:
//  - args[0] - ID of E2e object in tracker (int).
//  - args[1] - JSON of recipient [contact.Contact] (Uint8Array).
//  - args[2] - tag that identifies the single-use message (string).
//  - args[3] - message contents (Uint8Array).
//  - args[4] - JSON of [single.RequestParams] (Uint8Array).
//  - args[5] - the callback that will be called when a response is received. It
//    is a Javascript object that has functions that implement the
//    [bindings.SingleUseResponse] interface.
//
// Returns:
//  - JSON [bindings.SingleUseSendReport], which can be passed into
//    Cmix.WaitForRoundResult to see if the send succeeded (Uint8Array).
//  - Throws a TypeError if transmission fails.
func TransmitSingleUse(_ js.Value, args []js.Value) interface{} {
	e2eID := args[0].Int()
	recipient := CopyBytesToGo(args[1])
	tag := args[2].String()
	payload := CopyBytesToGo(args[3])
	paramsJSON := CopyBytesToGo(args[4])
	responseCB := &singleUseResponse{WrapCB(args[5], "Callback")}

	report, err := bindings.TransmitSingleUse(
		e2eID, recipient, tag, payload, paramsJSON, responseCB)
	if err != nil {
		Throw(TypeError, err.Error())
		return nil
	}

	return CopyBytesToJS(report)
}

// Listen starts a single-use listener on a given tag using the passed in E2e
// object and SingleUseCallback func.
//
// Parameters:
//  - args[0] - ID of E2e object in tracker (int).
//  - args[1] - tag that identifies the single-use message (string).
//  - args[2] - the callback that will be called when a response is received. It
//    is a Javascript object that has functions that implement the
//    [bindings.SingleUseCallback] interface.
//
// Returns:
//  - Javascript representation of the Stopper object, an interface containing a
//    function used to stop the listener.
//  - Throws a TypeError if listening fails.
func Listen(_ js.Value, args []js.Value) interface{} {
	cb := &singleUseCallback{WrapCB(args[2], "Callback")}
	api, err := bindings.Listen(args[0].Int(), args[1].String(), cb)
	if err != nil {
		Throw(TypeError, err.Error())
		return nil
	}

	return newStopperJS(api)
}

////////////////////////////////////////////////////////////////////////////////
// Function Types                                                             //
////////////////////////////////////////////////////////////////////////////////

// Stopper wraps the [bindings.Stopper] interface so its methods can be wrapped
// to be Javascript compatible.
type Stopper struct {
	api bindings.Stopper
}

// newStopperJS creates a new Javascript compatible object
// (map[string]interface{}) that matches the Stopper structure.
func newStopperJS(api bindings.Stopper) map[string]interface{} {
	s := Stopper{api}
	stopperMap := map[string]interface{}{
		"Stop": js.FuncOf(s.Stop),
	}

	return stopperMap
}

func (s *Stopper) Stop(js.Value, []js.Value) interface{} {
	s.api.Stop()
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Callback Wrappers                                                          //
////////////////////////////////////////////////////////////////////////////////

// singleUseCallback wraps Javascript callbacks to adhere to the
// [bindings.SingleUseCallback] interface.
type singleUseCallback struct {
	callback func(args ...interface{}) js.Value
}

func (suc *singleUseCallback) Callback(callbackReport []byte, err error) {
	suc.callback(CopyBytesToJS(callbackReport), err.Error())
}

// singleUseResponse wraps Javascript callbacks to adhere to the
// [bindings.SingleUseResponse] interface.
type singleUseResponse struct {
	callback func(args ...interface{}) js.Value
}

func (sur *singleUseResponse) Callback(responseReport []byte, err error) {
	sur.callback(CopyBytesToJS(responseReport), err.Error())
}
