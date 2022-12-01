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

// AuthenticatedConnection wraps the [bindings.AuthenticatedConnection] object
// so its methods can be wrapped to be Javascript compatible.
type AuthenticatedConnection struct {
	api *bindings.AuthenticatedConnection
}

// newAuthenticatedConnectionJS creates a new Javascript compatible object
// (map[string]any) that matches the [AuthenticatedConnection] structure.
func newAuthenticatedConnectionJS(
	api *bindings.AuthenticatedConnection) map[string]any {
	ac := AuthenticatedConnection{api}
	acMap := map[string]any{
		"IsAuthenticated":  js.FuncOf(ac.IsAuthenticated),
		"GetId":            js.FuncOf(ac.GetId),
		"SendE2E":          js.FuncOf(ac.SendE2E),
		"Close":            js.FuncOf(ac.Close),
		"GetPartner":       js.FuncOf(ac.GetPartner),
		"RegisterListener": js.FuncOf(ac.RegisterListener),
	}

	return acMap
}

// IsAuthenticated returns true.
//
// Returns:
//   - True (boolean).
func (ac *AuthenticatedConnection) IsAuthenticated(js.Value, []js.Value) any {
	return ac.api.IsAuthenticated()
}

// GetId returns the ID for this [bindings.AuthenticatedConnection] in the
// authenticatedConnectionTracker.
//
// Returns:
//   - Tracker ID (int).
func (ac *AuthenticatedConnection) GetId(js.Value, []js.Value) any {
	return ac.api.GetId()
}

// SendE2E is a wrapper for sending specifically to the
// [AuthenticatedConnection]'s [partner.Manager].
//
// Parameters:
//   - args[0] - Message type from [catalog.MessageType] (int).
//   - args[1] - Message payload (Uint8Array).
//
// Returns a promise:
//   - Resolves to the JSON of [bindings.E2ESendReport], which can be passed
//     into [Cmix.WaitForRoundResult] to see if the send succeeded (Uint8Array).
//   - Rejected with an error if sending fails.
func (ac *AuthenticatedConnection) SendE2E(_ js.Value, args []js.Value) any {
	mt := args[0].Int()
	payload := utils.CopyBytesToGo(args[1])

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		sendReport, err := ac.api.SendE2E(mt, payload)
		if err != nil {
			reject(utils.JsTrace(err))
		} else {
			resolve(utils.CopyBytesToJS(sendReport))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// Close deletes this [AuthenticatedConnection]'s [partner.Manager] and releases
// resources.
//
// Returns:
//   - Throws a TypeError if closing fails.
func (ac *AuthenticatedConnection) Close(js.Value, []js.Value) any {
	return ac.api.Close()
}

// GetPartner returns the [partner.Manager] for this [AuthenticatedConnection].
//
// Returns:
//   - Marshalled bytes of the partner's [id.ID] (Uint8Array).
func (ac *AuthenticatedConnection) GetPartner(js.Value, []js.Value) any {
	return utils.CopyBytesToJS(ac.api.GetPartner())
}

// RegisterListener is used for E2E reception and allows for reading data sent
// from the [partner.Manager].
//
// Parameters:
//   - args[0] - Message type from [catalog.MessageType] (int).
//   - args[1] - Javascript object that has functions that implement the
//     [bindings.Listener] interface.
//
// Returns:
//   - Throws a TypeError is registering the listener fails.
func (ac *AuthenticatedConnection) RegisterListener(
	_ js.Value, args []js.Value) any {
	err := ac.api.RegisterListener(args[0].Int(),
		&listener{utils.WrapCB(args[1], "Hear"), utils.WrapCB(args[1], "Name")})
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return nil
}

// ConnectWithAuthentication is called by the client (i.e., the one establishing
// connection with the server). Once a [connect.Connection] has been established
// with the server, it then authenticates their identity to the server.
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
func (c *Cmix) ConnectWithAuthentication(_ js.Value, args []js.Value) any {
	e2eID := args[0].Int()
	recipientContact := utils.CopyBytesToGo(args[1])
	e2eParamsJSON := utils.CopyBytesToGo(args[2])

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		ac, err := c.api.ConnectWithAuthentication(
			e2eID, recipientContact, e2eParamsJSON)
		if err != nil {
			reject(utils.JsTrace(err))
		} else {
			resolve(newAuthenticatedConnectionJS(ac))
		}
	}

	return utils.CreatePromise(promiseFn)
}
