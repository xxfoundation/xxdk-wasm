////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package worker

// Tag describes how a message sent to or from the worker should be handled.
type Tag string

// Generic tags used by all workers.
const (
	readyTag Tag = "Ready"
)
