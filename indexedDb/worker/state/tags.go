////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package dm

import "gitlab.com/elixxir/xxdk-wasm/worker"

// List of tags that can be used when sending a message or registering a handler
// to receive a message.
const (
	NewStateTag worker.Tag = "NewState"
	SetTag      worker.Tag = "Set"
	GetTag      worker.Tag = "Get"
)
