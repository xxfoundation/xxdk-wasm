////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package worker

// message is the outer message that contains the contents of each message sent
// to the worker. It is transmitted as JSON.
type message struct {
	Tag      Tag    `json:"tag"`
	ID       uint64 `json:"id"`
	DeleteCB bool   `json:"deleteCB"`
	Data     []byte `json:"data"`
}
