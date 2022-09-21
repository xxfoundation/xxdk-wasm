////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"gitlab.com/elixxir/client/bindings"
	"gitlab.com/elixxir/xxdk-wasm/utils"
	"syscall/js"
)

// DummyTraffic wraps the [bindings.DummyTraffic] object so its methods can be
// wrapped to be Javascript compatible.
type DummyTraffic struct {
	api *bindings.DummyTraffic
}

// newDummyTrafficJS creates a new Javascript compatible object
// (map[string]interface{}) that matches the [DummyTraffic] structure.
func newDummyTrafficJS(newDT *bindings.DummyTraffic) map[string]interface{} {
	dt := DummyTraffic{newDT}
	dtMap := map[string]interface{}{
		"SetStatus": js.FuncOf(dt.SetStatus),
		"GetStatus": js.FuncOf(dt.GetStatus),
	}

	return dtMap
}

// NewDummyTrafficManager creates a [DummyTraffic] manager and initialises the
// dummy traffic sending thread. Note that the manager does not start sending
// dummy traffic until true is passed into [DummyTraffic.SetStatus]. The time
// duration between each sending operation and the amount of messages sent each
// interval are randomly generated values with bounds defined by the given
// parameters below.
//
// Parameters:
//  - args[0] - A [Cmix] object ID in the tracker (int).
//  - args[1] - The maximum number of the random number of messages sent each
//    sending cycle (int).
//  - args[2] - The average duration, in milliseconds, to wait between sends
//    (int).
//  - args[3] - The upper bound of the interval between sending cycles, in
//    milliseconds. Sends occur every average send (args[2]) +/- a random
//    duration with an upper bound of args[3] (int).
//
// Returns:
//  - Javascript representation of the DummyTraffic object.
//  - Throws a TypeError if creating the manager fails.
func NewDummyTrafficManager(_ js.Value, args []js.Value) interface{} {
	dt, err := bindings.NewDummyTrafficManager(
		args[0].Int(), args[1].Int(), args[2].Int(), args[3].Int())
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return newDummyTrafficJS(dt)
}

// SetStatus sets the state of the [DummyTraffic] manager's send thread by
// passing in a boolean parameter. There may be a small delay in between this
// call and the status of the sending thread to change accordingly. For example,
// passing false into this call while the sending thread is currently sending
// messages will not cancel nor halt the sending operation, but will pause the
// thread once that operation has completed.
//
// Parameters:
//  - args[0] - Input should be true if you want to send dummy messages and
//    false if you want to pause dummy messages (boolean).
//
// Returns:
//  - Throws a TypeError if the [DummyTraffic.SetStatus] is called too
//    frequently, causing the internal status channel to fill.
func (dt *DummyTraffic) SetStatus(_ js.Value, args []js.Value) interface{} {
	err := dt.api.SetStatus(args[0].Bool())
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return nil
}

// GetStatus returns the current state of the [DummyTraffic] manager's sending
// thread. Note that this function does not return the status set by the most
// recent call to [SetStatus]. Instead, this call returns the current status of
// the sending thread. This is due to the small delay that may occur between
// calling [SetStatus] and the sending thread taking into effect that status
// change.
//
// Returns:
//   - Returns true if sending thread is sending dummy messages and false if
//     sending thread is paused/stopped and is not sending dummy messages
//     (boolean).
func (dt *DummyTraffic) GetStatus(js.Value, []js.Value) interface{} {
	return dt.api.GetStatus()
}
