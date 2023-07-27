////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"gitlab.com/elixxir/client/v4/bindings"
	"gitlab.com/elixxir/wasm-utils/exception"
	"gitlab.com/elixxir/wasm-utils/utils"
	"syscall/js"
)

// SetDashboardURL is a function which modifies the base dashboard URL that is
// returned as part of any send report. Internally, this is defaulted to
// "https://dashboard.xx.network". This should only be called if the user
// explicitly wants to modify the dashboard URL. This function is not
// thread-safe, and as such should only be called on setup.
//
// Parameters:
//   - args[0] - A valid URL that will be used for round look up on any send
//     report (string).
func SetDashboardURL(_ js.Value, args []js.Value) any {
	bindings.SetDashboardURL(args[0].String())

	return nil
}

// messageDeliveryCallback wraps Javascript callbacks to adhere to the
// [bindings.MessageDeliveryCallback] interface.
type messageDeliveryCallback struct {
	eventCallback func(args ...any) js.Value
}

// EventCallback gets called on the determination if all events related to a
// message send were successful.
//
// If delivered == true, timedOut == false && roundResults != nil
//
// If delivered == false, roundResults == nil
//
// If timedOut == true, delivered == false && roundResults == nil
//
// Parameters:
//   - delivered - Returns false if any rounds in the round map were
//     unsuccessful. Returns true if ALL rounds were successful (boolean).
//   - timedOut - Returns true if any of the rounds timed out while being
//     monitored. Returns false if all rounds statuses were returned (boolean).
//   - roundResults - rounds contains a mapping of all previously requested
//     rounds to their respective round results. Marshalled bytes of
//     map[[id.Round]][cmix.RoundResult] (Uint8Array).
func (mdc *messageDeliveryCallback) EventCallback(
	delivered, timedOut bool, roundResults []byte) {
	mdc.eventCallback(delivered, timedOut, utils.CopyBytesToJS(roundResults))
}

// WaitForRoundResult allows the caller to get notified if the rounds a message
// was sent in successfully completed. Under the hood, this uses an API that
// uses the internal round data, network historical round lookup, and waiting on
// network events to determine what has (or will) occur.
//
// This function takes the marshaled send report to ensure a memory leak does
// not occur as a result of both sides of the bindings holding a reference to
// the same pointer.
//
// roundList is a JSON marshalled [bindings.RoundsList] or any JSON marshalled
// send report that inherits a [bindings.RoundsList] object.
//
// Parameters:
//   - args[0] - JSON of [bindings.RoundsList] or JSON of any send report that
//     inherits a [bindings.RoundsList] object (Uint8Array).
//   - args[1] - Javascript object that has functions that implement the
//     [bindings.MessageDeliveryCallback] interface.
//   - args[2] - Timeout when the callback will return if no state update
//     occurs, in milliseconds (int).
//
// Returns:
//   - Throws an error if the parameters are invalid or getting round results
//     fails.
func (c *Cmix) WaitForRoundResult(_ js.Value, args []js.Value) any {
	roundList := utils.CopyBytesToGo(args[0])
	mdc := &messageDeliveryCallback{utils.WrapCB(args[1], "EventCallback")}

	err := c.api.WaitForRoundResult(roundList, mdc, args[2].Int())
	if err != nil {
		exception.ThrowTrace(err)
		return nil
	}

	return nil
}
