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

// Channel wraps the [bindings.Channel] object so its methods can be wrapped to
// be Javascript compatible.
type Channel struct {
	api *bindings.Channel
}

// newE2eJS creates a new Javascript compatible object (map[string]interface{})
// that matches the E2e structure.
func newChannelJS(api *bindings.Channel) map[string]interface{} {
	c := Channel{api}
	chMap := map[string]interface{}{
		"Listen":                   c.Listen,
		"Broadcast":                c.Broadcast,
		"BroadcastAsymmetric":      c.BroadcastAsymmetric,
		"MaxPayloadSize":           c.MaxPayloadSize,
		"MaxAsymmetricPayloadSize": c.MaxAsymmetricPayloadSize,
		"Get":                      c.Get,
		"Stop":                     c.Stop,
	}

	return chMap
}

// NewBroadcastChannel creates a bindings-layer broadcast channel and starts
// listening for new messages
//
// Parameters:
//  - args[0] - ID of Cmix object in tracker (int).
//  - args[1] - JSON of [bindings.ChannelDef] (Uint8Array).
//
// Returns:
//  - Javascript representation of the Channel object.
//  - Throws a TypeError if creation fails.
func NewBroadcastChannel(_ js.Value, args []js.Value) interface{} {
	channelDefinition := utils.CopyBytesToGo(args[1])

	api, err := bindings.NewBroadcastChannel(args[0].Int(), channelDefinition)
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return newChannelJS(api)
}

// broadcastListener wraps Javascript callbacks to adhere to the
// [bindings.BroadcastListener] interface.
type broadcastListener struct {
	callback func(args ...interface{}) js.Value
}

// Callback is used to listen for broadcast messages.
//
// Parameters:
//  - payload - returns the JSON of [bindings.E2ESendReport], which can be
//    passed into cmix.WaitForRoundResult to see if the send succeeded
//    (Uint8Array).
//  - err - returns an error on failure (Error).
func (bl *broadcastListener) Callback(payload []byte, err error) {
	bl.callback(utils.CopyBytesToJS(payload), utils.JsTrace(err))
}

// Listen registers a BroadcastListener for a given method. This allows users to
// handle incoming broadcast messages.
//
// Parameters:
//  - args[0] - Javascript object that has functions that implement the
//    [bindings.BroadcastListener] interface.
//  - args[1] - number corresponding to broadcast.Method constant, 0 for
//    symmetric or 1 for asymmetric (int).
//
// Returns:
//  - Throws a TypeError if registering the listener fails.
func (c *Channel) Listen(_ js.Value, args []js.Value) interface{} {
	err := c.api.Listen(
		&broadcastListener{utils.WrapCB(args[0], "Callback")}, args[1].Int())
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return nil
}

// Broadcast sends a given payload over the broadcast channel using symmetric
// broadcast.
//
// Parameters:
//  - args[0] - payload (Uint8Array).
//
// Returns a promise:
//  - Resolves to the JSON of the [bindings.BroadcastReport], which can be
//    passed into Cmix.WaitForRoundResult to see if the send succeeded
//    (Uint8Array).
//  - Rejected with an error if broadcasting fails.
func (c *Channel) Broadcast(_ js.Value, args []js.Value) interface{} {
	payload := utils.CopyBytesToGo(args[0])

	promiseFn := func(resolve, reject func(args ...interface{}) js.Value) {
		sendReport, err := c.api.Broadcast(payload)
		if err != nil {
			reject(utils.JsTrace(err))
		} else {
			resolve(utils.CopyBytesToJS(sendReport))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// BroadcastAsymmetric sends a given payload over the broadcast channel using
// asymmetric broadcast. This mode of encryption requires a private key.
//
// Parameters:
//  - args[0] - payload (Uint8Array).
//  - args[1] - private key (Uint8Array).
//
// Returns a promise:
//  - Resolves to the JSON of the [bindings.BroadcastReport], which can be
//    passed into Cmix.WaitForRoundResult to see if the send succeeded
//    (Uint8Array).
//  - Rejected with an error if broadcasting fails.
func (c *Channel) BroadcastAsymmetric(_ js.Value, args []js.Value) interface{} {
	payload := utils.CopyBytesToGo(args[0])
	privateKey := utils.CopyBytesToGo(args[1])

	promiseFn := func(resolve, reject func(args ...interface{}) js.Value) {
		sendReport, err := c.api.BroadcastAsymmetric(payload, privateKey)
		if err != nil {
			reject(utils.JsTrace(err))
		} else {
			resolve(utils.CopyBytesToJS(sendReport))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// MaxPayloadSize returns the maximum possible payload size which can be
// broadcast.
//
// Returns:
//  - int
func (c *Channel) MaxPayloadSize(js.Value, []js.Value) interface{} {
	return c.api.MaxPayloadSize()
}

// MaxAsymmetricPayloadSize returns the maximum possible payload size which can
// be broadcast.
//
// Returns:
//  - int
func (c *Channel) MaxAsymmetricPayloadSize(js.Value, []js.Value) interface{} {
	return c.api.MaxAsymmetricPayloadSize()
}

// Get returns the JSON of the channel definition.
//
// Returns:
//  - JSON of [bindings.ChannelDef] (Uint8Array).
//  - Throws a TypeError if marshalling fails.
func (c *Channel) Get(js.Value, []js.Value) interface{} {
	def, err := c.api.Get()
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return utils.CopyBytesToJS(def)
}

// Stop stops the channel from listening for more messages.
func (c *Channel) Stop(js.Value, []js.Value) interface{} {
	c.api.Stop()
	return nil
}
