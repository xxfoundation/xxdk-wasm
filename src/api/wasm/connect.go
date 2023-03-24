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
)

// Connection wraps the [bindings.Connection] object so its methods can be
// wrapped to be Javascript compatible.
type Connection struct {
	*bindings.Connection
}

// Connect performs auth key negotiation with the given recipient and returns a
// [Connection] object for the newly created [partner.Manager].
//
// This function is to be used sender-side and will block until the
// [partner.Manager] is confirmed.
//
// Parameters:
//   - e2eID - ID of [E2e] object in tracker.
//   - recipientContact - Marshalled bytes of the recipient [contact.Contact].
//   - e2eParamsJSON - JSON of [xxdk.E2EParams].
func (c *Cmix) Connect(e2eID int, recipientContact, e2eParamsJSON []byte) (
	*Connection, error) {
	api, err := c.Cmix.Connect(e2eID, recipientContact, e2eParamsJSON)
	return &Connection{api}, err
}

// Listener wraps a Javascript object so that it implements the
// [bindings.Listener] interface.
type Listener struct {
	HearFn func(item []byte) `wasm:"Hear"`
	NameFn func() string     `wasm:"Name"`
}

// Hear is called to receive a message in the UI.
//
// Parameters:
//   - item - Returns the JSON of [bindings.Message] (Uint8Array).
func (l *Listener) Hear(item []byte) { l.HearFn(item) }

// Name returns a name; used for debugging.
//
// Returns:
//   - Name (string).
func (l *Listener) Name() string { return l.NameFn() }

// RegisterListener is used for E2E reception and allows for reading data sent
// from the [partner.Manager].
//
// Parameters:
//   - messageType - Message type from [catalog.MessageType].
//   - newListener - Javascript object that has functions that implement the
//     [bindings.Listener] interface.
func (c *Connection) RegisterListener(messageType int, newListener *Listener) error {
	return c.Connection.RegisterListener(messageType, newListener)
}
