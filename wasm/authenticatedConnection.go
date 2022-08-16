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

// AuthenticatedConnection wraps the [bindings.AuthenticatedConnection] object
// so its methods can be wrapped to be Javascript compatible.
type AuthenticatedConnection struct {
	api *bindings.AuthenticatedConnection
}

// newAuthenticatedConnectionJS creates a new Javascript compatible object
// (map[string]interface{}) that matches the AuthenticatedConnection structure.
func newAuthenticatedConnectionJS(
	api *bindings.AuthenticatedConnection) map[string]interface{} {
	ac := AuthenticatedConnection{api}
	acMap := map[string]interface{}{
		"IsAuthenticated": js.FuncOf(ac.IsAuthenticated),
	}

	return acMap
}

func (ac *AuthenticatedConnection) IsAuthenticated(js.Value, []js.Value) interface{} {
	return ac.api.IsAuthenticated()
}

// ConnectWithAuthentication is called by the client (i.e., the one establishing
// connection with the server). Once a connect.Connection has been established
// with the server, it then authenticates their identity to the server.
// accepts a marshalled ReceptionIdentity and contact.Contact object
//
// Parameters:
//  - args[0] - ID of E2e object in tracker (int)
//  - args[1] - marshalled recipient [contact.Contact] (Uint8Array).
//  - args[3] - JSON of [xxdk.E2EParams] (Uint8Array).
//
// Returns:
//  - Javascript representation of the Connection object
//  - throws a TypeError if creating loading the parameters or connecting fails
func (c *Cmix) ConnectWithAuthentication(_ js.Value, args []js.Value) interface{} {
	recipientContact := CopyBytesToGo(args[1])
	e2eParamsJSON := CopyBytesToGo(args[2])

	ac, err := c.api.ConnectWithAuthentication(
		args[0].Int(), recipientContact, e2eParamsJSON)
	if err != nil {
		Throw(TypeError, err.Error())
		return nil
	}

	return newAuthenticatedConnectionJS(ac)
}
