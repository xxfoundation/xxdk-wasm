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

// Connection wraps the [bindings.Connection] object so its methods can be
// wrapped to be Javascript compatible.
type Connection struct {
	api *bindings.Connection
}

// newConnectJS creates a new Javascript compatible object (map[string]any) that
// matches the [Connection] structure.
func newConnectJS(api *bindings.Connection) map[string]any {
	c := Connection{api}
	connectionMap := map[string]any{
		// connect.go
		"GetId":            js.FuncOf(c.GetId),
		"SendE2E":          js.FuncOf(c.SendE2E),
		"Close":            js.FuncOf(c.Close),
		"GetPartner":       js.FuncOf(c.GetPartner),
		"RegisterListener": js.FuncOf(c.RegisterListener),
	}

	return connectionMap
}

// GetId returns the ID for this [bindings.Connection] in the connectionTracker.
//
// Returns:
//   - Tracker ID (int).
func (c *Connection) GetId(js.Value, []js.Value) any {
	return c.api.GetId()
}

// Connect performs auth key negotiation with the given recipient and returns a
// [Connection] object for the newly created [partner.Manager].
//
// This function is to be used sender-side and will block until the
// [partner.Manager] is confirmed.
//
// Parameters:
//   - args[0] - ID of [E2e] object in tracker (int).
//   - args[1] - Marshalled bytes of the recipient [contact.Contact]
//     (Uint8Array).
//   - args[3] - JSON of [xxdk.E2EParams] (Uint8Array).
//
// Returns a promise:
//   - Resolves to a Javascript representation of the [Connection] object.
//   - Rejected with an error if loading the parameters or connecting fails.
func (c *Cmix) Connect(_ js.Value, args []js.Value) any {
	e2eID := args[0].Int()
	recipientContact := utils.CopyBytesToGo(args[1])
	e2eParamsJSON := utils.CopyBytesToGo(args[2])

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		api, err := c.api.Connect(e2eID, recipientContact, e2eParamsJSON)
		if err != nil {
			reject(utils.JsTrace(err))
		} else {
			resolve(newConnectJS(api))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// SendE2E is a wrapper for sending specifically to the [Connection]'s
// [partner.Manager].
//
// Parameters:
//   - args[0] - Message type from [catalog.MessageType] (int).
//   - args[1] - Message payload (Uint8Array).
//
// Returns a promise:
//   - Resolves to the JSON of the [bindings.E2ESendReport], which can be passed
//     into [Cmix.WaitForRoundResult] to see if the send succeeded (Uint8Array).
//   - Rejected with an error if sending fails.
func (c *Connection) SendE2E(_ js.Value, args []js.Value) any {
	e2eID := args[0].Int()
	payload := utils.CopyBytesToGo(args[1])

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		sendReport, err := c.api.SendE2E(e2eID, payload)
		if err != nil {
			reject(utils.JsTrace(err))
		} else {
			resolve(utils.CopyBytesToJS(sendReport))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// Close deletes this [Connection]'s [partner.Manager] and releases resources.
//
// Returns:
//   - Throws a TypeError if closing fails.
func (c *Connection) Close(js.Value, []js.Value) any {
	err := c.api.Close()
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return nil
}

// GetPartner returns the [partner.Manager] for this [Connection].
//
// Returns:
//   - Marshalled bytes of the partner's [id.ID] (Uint8Array).
func (c *Connection) GetPartner(js.Value, []js.Value) any {
	return utils.CopyBytesToJS(c.api.GetPartner())
}

// listener adheres to the [bindings.Listener] interface.
type listener struct {
	hear func(args ...any) js.Value
	name func(args ...any) js.Value
}

// Hear is called to receive a message in the UI.
//
// Parameters:
//   - item - Returns the JSON of [bindings.Message] (Uint8Array).
func (l *listener) Hear(item []byte) { l.hear(utils.CopyBytesToJS(item)) }

// Name returns a name; used for debugging.
//
// Returns:
//   - Name (string).
func (l *listener) Name() string { return l.name().String() }

// RegisterListener is used for E2E reception and allows for reading data sent
// from the [partner.Manager].
//
// Parameters:
//   - args[0] - message type from [catalog.MessageType] (int).
//   - args[1] - Javascript object that has functions that implement the
//     [bindings.Listener] interface.
//
// Returns:
//   - Throws a TypeError is registering the listener fails.
func (c *Connection) RegisterListener(_ js.Value, args []js.Value) any {
	err := c.api.RegisterListener(args[0].Int(),
		&listener{utils.WrapCB(args[1], "Hear"), utils.WrapCB(args[1], "Name")})
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return nil
}
