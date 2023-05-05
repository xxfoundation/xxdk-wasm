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
	readyTag Tag = "<WW>Ready</WW>"
)

// Channel is a tag given to a MessageChannel between two workers to identify
// traffic sent and received on its ports.
type Channel string

const (
	ChannelsIndexedDbLogging = "ChannelsIndexedDbLogging"
	DmIndexedDbLogging       = "dmIndexedDbLogging"
)
