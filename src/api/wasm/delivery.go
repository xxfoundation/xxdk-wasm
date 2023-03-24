////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

// MessageDeliveryCallback wraps a Javascript object so that it implements the
// [bindings.MessageDeliveryCallback] interface.
type MessageDeliveryCallback struct {
	EventCallbackFn func(delivered, timedOut bool, roundResults []byte) `wasm:"EventCallback"`
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
//     unsuccessful. Returns true if ALL rounds were successful.
//   - timedOut - Returns true if any of the rounds timed out while being
//     monitored. Returns false if all rounds statuses were returned.
//   - roundResults - rounds contains a mapping of all previously requested
//     rounds to their respective round results. Marshalled bytes of
//     map[[id.Round]][cmix.RoundResult].
func (mdc *MessageDeliveryCallback) EventCallback(
	delivered, timedOut bool, roundResults []byte) {
	mdc.EventCallbackFn(delivered, timedOut, roundResults)
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
//   - roundList - JSON of [bindings.RoundsList] or JSON of any send report that
//     inherits a [bindings.RoundsList] object.
//   - mdc - Javascript object that matches the [MessageDeliveryCallback]
//     struct.
//   - timeoutMS - Timeout when the callback will return if no state update
//     occurs, in milliseconds.
func (c *Cmix) WaitForRoundResult(
	roundList []byte, mdc *MessageDeliveryCallback, timeoutMS int) error {
	return c.Cmix.WaitForRoundResult(roundList, mdc, timeoutMS)
}
