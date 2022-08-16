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

// Connection wraps the [bindings.Connection] object so its methods can be
// wrapped to be Javascript compatible.
type Connection struct {
	api *bindings.Connection
}

// newConnectJS creates a new Javascript compatible object
// (map[string]interface{}) that matches the Connection structure.
func newConnectJS(api *bindings.Connection) map[string]interface{} {
	c := Connection{api}
	connectionMap := map[string]interface{}{
		// connect.go
		"GetID":            js.FuncOf(c.GetID),
		"SendE2E":          js.FuncOf(c.SendE2E),
		"Close":            js.FuncOf(c.Close),
		"GetPartner":       js.FuncOf(c.GetPartner),
		"RegisterListener": js.FuncOf(c.RegisterListener),
	}

	return connectionMap
}

// GetID returns the ID for this [bindings.Connection] in the connectionTracker.
//
// Returns:
//  - int of the ID
func (c *Connection) GetID(js.Value, []js.Value) interface{} {
	return c.api.GetId()
}

// Connect performs auth key negotiation with the given recipient and returns a
// Connection object for the newly created [partner.Manager].
//
// This function is to be used sender-side and will block until the
// [partner.Manager] is confirmed.
//
// Parameters:
//  - args[0] - ID of the E2E object in the E2E tracker (int).
//  - args[1] - marshalled recipient [contact.Contact] (Uint8Array).
//  - args[3] - JSON of [xxdk.E2EParams] (Uint8Array).
//
// Returns:
//  - Javascript representation of the Connection object
//  - throws a TypeError if creating loading the parameters or connecting fails
func (c *Cmix) Connect(_ js.Value, args []js.Value) interface{} {
	recipientContact := CopyBytesToGo(args[1])
	e2eParamsJSON := CopyBytesToGo(args[2])
	api, err := c.api.Connect(args[0].Int(), recipientContact, e2eParamsJSON)
	if err != nil {
		Throw(TypeError, err.Error())
		return nil
	}

	return newConnectJS(api)
}

// SendE2E is a wrapper for sending specifically to the Connection's
// [partner.Manager].
//
// Returns:
//  - []byte - the JSON marshalled bytes of the E2ESendReport object, which can
//    be passed into WaitForRoundResult to see if the send succeeded.
//
// Parameters:
//  - args[0] - message type from [catalog.MessageType] (int)
//  - args[1] - message payload (Uint8Array)
//
// Returns:
//  - JSON of [bindings.E2ESendReport], which can be passed into
//    cmix.WaitForRoundResult to see if the send succeeded (Uint8Array)
//  - throws a TypeError if sending fails
func (c *Connection) SendE2E(_ js.Value, args []js.Value) interface{} {
	sendReport, err := c.api.SendE2E(args[0].Int(), CopyBytesToGo(args[1]))
	if err != nil {
		Throw(TypeError, err.Error())
		return nil
	}
	return CopyBytesToJS(sendReport)
}

// Close deletes this Connection's partner.Manager and releases resources.
//
// Returns:
//  - throws a TypeError if closing fails
func (c *Connection) Close(js.Value, []js.Value) interface{} {
	err := c.api.Close()
	if err != nil {
		Throw(TypeError, err.Error())
		return nil
	}

	return nil
}

// GetPartner returns the partner.Manager for this Connection.
//
// Returns:
//  - bytes of the partner's [id.ID] (Uint8Array)
func (c *Connection) GetPartner(js.Value, []js.Value) interface{} {
	return CopyBytesToJS(c.api.GetPartner())
}

// listener adheres to the [bindings.Listener] interface.
type listener struct {
	hear func(args ...interface{}) js.Value
	name func(args ...interface{}) js.Value
}

func (l *listener) Hear(item []byte) { l.hear(CopyBytesToJS(item)) }
func (l *listener) Name() string     { return l.name().String() }

// RegisterListener is used for E2E reception and allows for reading data sent
// from the partner.Manager.
//
// Parameters:
//  - args[0] - message type from [catalog.MessageType] (int)
//  - args[1] - Javascript object that has functions that implement the
//    [bindings.Listener] interface
//
// Returns:
//  - throws a TypeError is registering the listener fails
func (c *Connection) RegisterListener(_ js.Value, args []js.Value) interface{} {
	err := c.api.RegisterListener(args[0].Int(),
		&listener{args[1].Get("Hear").Invoke, args[1].Get("Name").Invoke})
	if err != nil {
		Throw(TypeError, err.Error())
		return nil
	}

	return nil
}
