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
	"gitlab.com/elixxir/xxdk-wasm/utils"
	"syscall/js"
)

// timeSource wraps Javascript callbacks to adhere to the [netTime.TimeSource]
// interface.
type timeSource struct {
	nowMs func(args ...interface{}) js.Value
}

// NowMs returns the current time in milliseconds.
func (ts *timeSource) NowMs() int64 {
	return int64(ts.nowMs().Int())
}

// SetTimeSource will set the time source that will be used when retrieving the
// current time using [netTime.Now]. This should be called BEFORE Login()
// and only be called once. Using this after Login is undefined behavior that
// may result in a crash.
//
// Parameters:
//  - timeNow is an object which adheres to [netTime.TimeSource]. Specifically,
//    this object should a NowMs() method which return a 64-bit integer value.
func SetTimeSource(_ js.Value, args []js.Value) interface{} {
	bindings.SetTimeSource(&timeSource{utils.WrapCB(args[0], "NowMs")})
	return nil
}

// SetOffset will set an internal offset variable. All calls to [netTime.Now]
// will have this offset applied to this value.
//
// Parameters:
//  - args[0] - a time by which netTime.Now will be offset. This value may be
//    negative or positive. This expects a 64-bit integer value which will
//    represent the number in microseconds this offset will be (int).
func SetOffset(_ js.Value, args []js.Value) interface{} {
	bindings.SetOffset(int64(args[0].Int()))
	return nil
}
