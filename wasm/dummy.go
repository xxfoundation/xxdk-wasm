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
	"syscall/js"
)

// DummyTraffic wraps the [bindings.DummyTraffic] object so its methods can be
// wrapped to be Javascript compatible.
type DummyTraffic struct {
	api *bindings.DummyTraffic
}

// newDummyTrafficJS creates a new Javascript compatible object (map[string]any)
// that matches the [DummyTraffic] structure.
func newDummyTrafficJS(newDT *bindings.DummyTraffic) map[string]any {
	dt := DummyTraffic{newDT}
	dtMap := map[string]any{
		"Pause":     js.FuncOf(dt.Pause),
		"Start":     js.FuncOf(dt.Start),
		"GetStatus": js.FuncOf(dt.GetStatus),
	}

	return dtMap
}

// NewDummyTrafficManager creates a [DummyTraffic] manager and initialises the
// dummy traffic sending thread. Note that the manager is by default paused,
// and as such the sending thread must be started by calling [DummyTraffic.Start].
// The time duration between each sending operation and the amount of messages
// sent each interval are randomly generated values with bounds defined by the
// given parameters below.
//
// Parameters:
//   - args[0] - A [Cmix] object ID in the tracker (int).
//   - args[1] - The maximum number of the random number of messages sent each
//     sending cycle (int).
//   - args[2] - The average duration, in milliseconds, to wait between sends
//     (int).
//   - args[3] - The upper bound of the interval between sending cycles, in
//     milliseconds. Sends occur every average send (args[2]) +/- a random
//     duration with an upper bound of args[3] (int).
//
// Returns:
//   - Javascript representation of the DummyTraffic object.
//   - Throws an error if creating the manager fails.
func NewDummyTrafficManager(_ js.Value, args []js.Value) any {
	dt, err := bindings.NewDummyTrafficManager(
		args[0].Int(), args[1].Int(), args[2].Int(), args[3].Int())
	if err != nil {
		exception.ThrowTrace(err)
		return nil
	}

	return newDummyTrafficJS(dt)
}

// Pause will pause the [DummyTraffic]'s sending thread, meaning messages will
// no longer be sent. After calling Pause, the sending thread may only be
// resumed by calling Resume.
//
// There may be a small delay between this call and the pause taking effect.
// This is because Pause will not cancel the thread when it is in the process
// of sending messages, but will instead wait for that thread to complete. The
// thread will then be prevented from beginning another round of sending.
//
// Returns:
//   - Throws an error if it fails to send a pause signal to the sending
//     thread.
func (dt *DummyTraffic) Pause(js.Value, []js.Value) any {
	err := dt.api.Pause()
	if err != nil {
		exception.ThrowTrace(err)
		return nil
	}

	return nil
}

// Start will start up the [DummyTraffic]'s sending thread, meaning messages
// will be sent. This should be called after calling [NewDummyTrafficManager],
// by default the thread is paused. This may also be called after a call to
// [DummyTraffic.Pause].
//
// This will re-initialize the sending thread with a new randomly generated
// interval between sending dummy messages. This means that there is zero
// guarantee that the sending interval prior to pausing will be the same
// sending interval after a call to Start.
//
// Returns:
//   - Throws an error if it fails to send a start signal to the sending
//     thread.
func (dt *DummyTraffic) Start(js.Value, []js.Value) any {
	err := dt.api.Start()
	if err != nil {
		exception.ThrowTrace(err)
		return nil
	}

	return nil
}

// GetStatus returns the current state of the [DummyTraffic] manager's sending
// thread. Note that the status returned here may lag behind a user's earlier
// call to pause the sending thread. This is a result of a small delay (see
// [DummyTraffic.Pause] for more details)
//
// Returns:
//   - Returns true ([dummy.Running]) if the sending thread is sending
//     messages and false ([dummy.Paused]) if the sending thread is not sending
//     messages.
func (dt *DummyTraffic) GetStatus(js.Value, []js.Value) any {
	return dt.api.GetStatus()
}
