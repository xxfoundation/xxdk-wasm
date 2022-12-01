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

// E2e wraps the [bindings.E2e] object so its methods can be wrapped to be
// Javascript compatible.
type E2e struct {
	api *bindings.E2e
}

// newE2eJS creates a new Javascript compatible object (map[string]any) that
// matches the [E2e] structure.
func newE2eJS(api *bindings.E2e) map[string]any {
	e := E2e{api}
	e2eMap := map[string]any{
		// e2e.go
		"GetID":               js.FuncOf(e.GetID),
		"GetContact":          js.FuncOf(e.GetContact),
		"GetUdAddressFromNdf": js.FuncOf(e.GetUdAddressFromNdf),
		"GetUdCertFromNdf":    js.FuncOf(e.GetUdCertFromNdf),
		"GetUdContactFromNdf": js.FuncOf(e.GetUdContactFromNdf),

		// e2eHandler.go
		"GetReceptionID":          js.FuncOf(e.GetReceptionID),
		"DeleteContact":           js.FuncOf(e.DeleteContact),
		"GetAllPartnerIDs":        js.FuncOf(e.GetAllPartnerIDs),
		"PayloadSize":             js.FuncOf(e.PayloadSize),
		"SecondPartitionSize":     js.FuncOf(e.SecondPartitionSize),
		"PartitionSize":           js.FuncOf(e.PartitionSize),
		"FirstPartitionSize":      js.FuncOf(e.FirstPartitionSize),
		"GetHistoricalDHPrivkey":  js.FuncOf(e.GetHistoricalDHPrivkey),
		"GetHistoricalDHPubkey":   js.FuncOf(e.GetHistoricalDHPubkey),
		"HasAuthenticatedChannel": js.FuncOf(e.HasAuthenticatedChannel),
		"RemoveService":           js.FuncOf(e.RemoveService),
		"SendE2E":                 js.FuncOf(e.SendE2E),
		"AddService":              js.FuncOf(e.AddService),
		"RegisterListener":        js.FuncOf(e.RegisterListener),

		// e2eAuth.go
		"Request":                 js.FuncOf(e.Request),
		"Confirm":                 js.FuncOf(e.Confirm),
		"Reset":                   js.FuncOf(e.Reset),
		"ReplayConfirm":           js.FuncOf(e.ReplayConfirm),
		"CallAllReceivedRequests": js.FuncOf(e.CallAllReceivedRequests),
		"DeleteRequest":           js.FuncOf(e.DeleteRequest),
		"DeleteAllRequests":       js.FuncOf(e.DeleteAllRequests),
		"DeleteSentRequests":      js.FuncOf(e.DeleteSentRequests),
		"DeleteReceiveRequests":   js.FuncOf(e.DeleteReceiveRequests),
		"GetReceivedRequest":      js.FuncOf(e.GetReceivedRequest),
		"VerifyOwnership":         js.FuncOf(e.VerifyOwnership),
		"AddPartnerCallback":      js.FuncOf(e.AddPartnerCallback),
		"DeletePartnerCallback":   js.FuncOf(e.DeletePartnerCallback),
	}

	return e2eMap
}

// GetID returns the ID for this [E2e] in the [E2e] tracker.
//
// Returns:
//   - Tracker ID (int).
func (e *E2e) GetID(js.Value, []js.Value) any {
	return e.api.GetID()
}

// Login creates and returns a new [E2e] object and adds it to the
// e2eTrackerSingleton. Identity should be created via
// [Cmix.MakeReceptionIdentity] and passed in here. If callbacks is left nil, a
// default [auth.Callbacks] will be used.
//
// Parameters:
//   - args[0] - ID of [Cmix] object in tracker (int).
//   - args[1] - Javascript object that has functions that implement the
//     [bindings.AuthCallbacks] interface.
//   - args[2] - JSON of the [xxdk.ReceptionIdentity] (Uint8Array).
//   - args[3] - JSON of [xxdk.E2EParams] (Uint8Array).
//
// Returns:
//   - Javascript representation of the [E2e] object.
//   - Throws a TypeError if logging in fails.
func Login(_ js.Value, args []js.Value) any {
	callbacks := newAuthCallbacks(args[1])
	identity := utils.CopyBytesToGo(args[2])
	e2eParamsJSON := utils.CopyBytesToGo(args[3])

	newE2E, err := bindings.Login(
		args[0].Int(), callbacks, identity, e2eParamsJSON)
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return newE2eJS(newE2E)
}

// LoginEphemeral creates and returns a new ephemeral [E2e] object and adds it
// to the e2eTrackerSingleton. Identity should be created via
// [Cmix.MakeReceptionIdentity] or [Cmix.MakeLegacyReceptionIdentity] and passed
// in here. If callbacks is left nil, a default [auth.Callbacks] will be used.
//
// Parameters:
//   - args[0] - ID of [Cmix] object in tracker (int).
//   - args[1] - Javascript object that has functions that implement the
//     [bindings.AuthCallbacks] interface.
//   - args[2] - JSON of the [xxdk.ReceptionIdentity] object (Uint8Array).
//   - args[3] - JSON of [xxdk.E2EParams] (Uint8Array).
//
// Returns:
//   - Javascript representation of the [E2e] object.
//   - Throws a TypeError if logging in fails.
func LoginEphemeral(_ js.Value, args []js.Value) any {
	callbacks := newAuthCallbacks(args[1])
	identity := utils.CopyBytesToGo(args[2])
	e2eParamsJSON := utils.CopyBytesToGo(args[3])

	newE2E, err := bindings.LoginEphemeral(
		args[0].Int(), callbacks, identity, e2eParamsJSON)
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return newE2eJS(newE2E)
}

// GetContact returns a [contact.Contact] object for the [E2e]
// [bindings.ReceptionIdentity].
//
// Returns:
//   - Marshalled bytes of [contact.Contact] (Uint8Array).
func (e *E2e) GetContact(js.Value, []js.Value) any {
	return utils.CopyBytesToJS(e.api.GetContact())
}

// GetUdAddressFromNdf retrieve the User Discovery's network address fom the
// NDF.
//
// Returns:
//   - User Discovery's address (string).
func (e *E2e) GetUdAddressFromNdf(js.Value, []js.Value) any {
	return e.api.GetUdAddressFromNdf()
}

// GetUdCertFromNdf retrieves the User Discovery's TLS certificate from the NDF.
//
// Returns:
//   - Public certificate in PEM format (Uint8Array).
func (e *E2e) GetUdCertFromNdf(js.Value, []js.Value) any {
	return utils.CopyBytesToJS(e.api.GetUdCertFromNdf())
}

// GetUdContactFromNdf assembles the User Discovery's contact file from the data
// within the NDF.
//
// Returns
//   - Marshalled bytes of [contact.Contact] (Uint8Array).
//   - Throws a TypeError if the contact file cannot be loaded.
func (e *E2e) GetUdContactFromNdf(js.Value, []js.Value) any {
	b, err := e.api.GetUdContactFromNdf()
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return utils.CopyBytesToJS(b)
}

////////////////////////////////////////////////////////////////////////////////
// Auth Callbacks                                                             //
////////////////////////////////////////////////////////////////////////////////

// authCallbacks wraps Javascript callbacks to adhere to the
// [bindings.AuthCallbacks] interface.
type authCallbacks struct {
	request func(args ...any) js.Value
	confirm func(args ...any) js.Value
	reset   func(args ...any) js.Value
}

// newAuthCallbacks adds all the callbacks from the Javascript object.
func newAuthCallbacks(value js.Value) *authCallbacks {
	return &authCallbacks{
		request: utils.WrapCB(value, "Request"),
		confirm: utils.WrapCB(value, "Confirm"),
		reset:   utils.WrapCB(value, "Reset"),
	}
}

// Request will be called when an auth Request message is processed.
//
// Parameters:
//   - contact - Returns the marshalled bytes of the [contact.Contact] of the
//     sender (Uint8Array).
//   - receptionId - Returns the marshalled bytes of the sender's [id.ID]
//     (Uint8Array).
//   - ephemeralId - Returns the ephemeral ID of the sender (int).
//   - roundId - Returns the ID of the round the request was sent on (int).
func (a *authCallbacks) Request(
	contact, receptionId []byte, ephemeralId, roundId int64) {
	if a.request != nil {
		a.request(utils.CopyBytesToJS(contact), utils.CopyBytesToJS(receptionId),
			ephemeralId, roundId)
	}
}

// Confirm will be called when an auth Confirm message is processed.
//
// Parameters:
//   - contact - Returns the marshalled bytes of the [contact.Contact] of the
//     sender (Uint8Array).
//   - receptionId - Returns the marshalled bytes of the sender's [id.ID]
//     (Uint8Array).
//   - ephemeralId - Returns the ephemeral ID of the sender (int).
//   - roundId - Returns the ID of the round the confirmation was sent on (int).
func (a *authCallbacks) Confirm(
	contact, receptionId []byte, ephemeralId, roundId int64) {
	if a.confirm != nil {
		a.confirm(utils.CopyBytesToJS(contact), utils.CopyBytesToJS(receptionId),
			ephemeralId, roundId)
	}
}

// Reset will be called when an auth Reset operation occurs.
//
// Parameters:
//   - contact - Returns the marshalled bytes of the [contact.Contact] of the
//     sender (Uint8Array).
//   - receptionId - Returns the marshalled bytes of the sender's [id.ID]
//     (Uint8Array).
//   - ephemeralId - Returns the ephemeral ID of the sender (int).
//   - roundId - Returns the ID of the round the reset was sent on (int).
func (a *authCallbacks) Reset(
	contact, receptionId []byte, ephemeralId, roundId int64) {
	if a.reset != nil {
		a.reset(utils.CopyBytesToJS(contact), utils.CopyBytesToJS(receptionId),
			ephemeralId, roundId)
	}
}
