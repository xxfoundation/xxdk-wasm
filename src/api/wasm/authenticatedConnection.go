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

// AuthenticatedConnection wraps the [bindings.AuthenticatedConnection] object
// so its methods can be wrapped to be Javascript compatible.
type AuthenticatedConnection struct {
	*bindings.AuthenticatedConnection
}

// RegisterListener is used for E2E reception and allows for reading data sent
// from the [partner.Manager].
//
// Parameters:
//   - messageType - Message type from [catalog.MessageType].
//   - newListener - Javascript object that matches the [Listener] struct.
func (ac *AuthenticatedConnection) RegisterListener(messageType int, newListener *Listener) error {
	return ac.AuthenticatedConnection.RegisterListener(messageType, newListener)
}

// ConnectWithAuthentication is called by the client (i.e., the one establishing
// connection with the server). Once a [connect.Connection] has been established
// with the server, it then authenticates their identity to the server.
//
// Parameters:
//   - e2eID - ID of [E2e] object in tracker.
//   - recipientContact - Marshalled bytes of the recipient [contact.Contact].
//   - e2eParamsJSON - JSON of [xxdk.E2EParams].
func (c *Cmix) ConnectWithAuthentication(e2eID int, recipientContact,
	e2eParamsJSON []byte) (*AuthenticatedConnection, error) {
	ac, err :=
		c.Cmix.ConnectWithAuthentication(e2eID, recipientContact, e2eParamsJSON)
	return &AuthenticatedConnection{ac}, err
}
