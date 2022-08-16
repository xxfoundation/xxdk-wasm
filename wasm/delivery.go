////////////////////////////////////////////////////////////////////////////////
// Copyright © 2020 xx network SEZC                                           //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file                                                               //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"syscall/js"
)

// messageDeliveryCallback wraps Javascript callbacks to adhere to the
// [bindings.MessageDeliveryCallback] interface.
type messageDeliveryCallback struct {
	eventCallback func(args ...interface{}) js.Value
}

func (mdc *messageDeliveryCallback) EventCallback(
	delivered, timedOut bool, roundResults []byte) {
	mdc.eventCallback(delivered, timedOut, CopyBytesToJS(roundResults))
}

// WaitForRoundResult allows the caller to get notified if the rounds a message was sent in successfully completed. Under the hood, this uses an API
// that uses the internal round data, network historical round lookup, and
// waiting on network events to determine what has (or will) occur.
//
// This function takes the marshaled send report to ensure a memory leak does
// not occur as a result of both sides of the bindings holding a reference to
// the same pointer.
//
// roundList is a JSON marshalled RoundsList or any JSON marshalled send report
// that inherits a RoundsList object.
//
// Parameters:
//  - args[0] - JSON of [bindings.RoundsList] or JSON of any send report that
//    inherits a [bindings.RoundsList] object (Uint8Array)
//  - args[1] - Javascript object that has functions that implement the
//    [bindings.MessageDeliveryCallback] interface
//  - args[2] - timeout when the callback will return if no state update occurs,
//    in milliseconds (int)
//
// Returns:
//  - throws a TypeError if the parameters are invalid or getting round results
//    fails
func (c *Cmix) WaitForRoundResult(_ js.Value, args []js.Value) interface{} {
	roundList := CopyBytesToGo(args[0])
	mdc := &messageDeliveryCallback{args[1].Get("EventCallback").Invoke}

	err := c.api.WaitForRoundResult(roundList, mdc, args[2].Int())
	if err != nil {
		Throw(TypeError, err.Error())
		return nil
	}

	return nil
}
