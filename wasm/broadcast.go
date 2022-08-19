////////////////////////////////////////////////////////////////////////////////
// Copyright © 2020 xx network SEZC                                           //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file                                                               //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"gitlab.com/elixxir/client/bindings"
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
	channelDefinition := CopyBytesToGo(args[1])

	api, err := bindings.NewBroadcastChannel(args[0].Int(), channelDefinition)
	if err != nil {
		Throw(TypeError, err.Error())
		return nil
	}

	return newChannelJS(api)
}

// broadcastListener wraps Javascript callbacks to adhere to the
// [bindings.BroadcastListener] interface.
type broadcastListener struct {
	callback func(args ...interface{}) js.Value
}

func (bl *broadcastListener) Callback(payload []byte, err error) {
	bl.callback(CopyBytesToJS(payload), err.Error())
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
		&broadcastListener{WrapCB(args[0], "Callback")}, args[1].Int())
	if err != nil {
		Throw(TypeError, err.Error())
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
// Returns:
//  - JSON of [bindings.BroadcastReport], which can be passed into
//    Cmix.WaitForRoundResult to see if the broadcast succeeded (Uint8Array).
//  - Throws a TypeError if broadcasting fails.
func (c *Channel) Broadcast(_ js.Value, args []js.Value) interface{} {
	report, err := c.api.Broadcast(CopyBytesToGo(args[0]))
	if err != nil {
		Throw(TypeError, err.Error())
		return nil
	}

	return CopyBytesToJS(report)
}

// BroadcastAsymmetric sends a given payload over the broadcast channel using
// asymmetric broadcast. This mode of encryption requires a private key.
//
// Parameters:
//  - args[0] - payload (Uint8Array).
//  - args[1] - private key (Uint8Array).
//
// Returns:
//  - JSON of [bindings.BroadcastReport], which can be passed into
//    Cmix.WaitForRoundResult to see if the broadcast succeeded (Uint8Array).
//  - Throws a TypeError if broadcasting fails.
func (c *Channel) BroadcastAsymmetric(_ js.Value, args []js.Value) interface{} {
	report, err := c.api.BroadcastAsymmetric(
		CopyBytesToGo(args[0]), CopyBytesToGo(args[1]))
	if err != nil {
		Throw(TypeError, err.Error())
		return nil
	}

	return CopyBytesToJS(report)
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
		Throw(TypeError, err.Error())
		return nil
	}

	return CopyBytesToJS(def)
}

// Stop stops the channel from listening for more messages.
func (c *Channel) Stop(js.Value, []js.Value) interface{} {
	c.api.Stop()
	return nil
}
