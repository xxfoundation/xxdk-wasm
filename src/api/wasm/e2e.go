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
	"syscall/js"
)

// E2e wraps the [bindings.E2e] object so its methods can be wrapped to be
// Javascript compatible.
type E2e struct {
	*bindings.E2e
}

// newE2eJS creates a new Javascript compatible object (map[string]any) that
// matches the [E2e] structure.
func newE2eJS(api *bindings.E2e) map[string]any {
	e := E2e{api}
	e2eMap := map[string]any{

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

// Login creates and returns a new [E2e] object and adds it to the
// e2eTrackerSingleton. Identity should be created via
// [Cmix.MakeReceptionIdentity] and passed in here. If callbacks is left nil, a
// default [auth.Callbacks] will be used.
//
// Parameters:
//   - cmixID - ID of [Cmix] object in tracker.
//   - callbacks - Javascript object that matches the [AuthCallbacks] struct.
//   - identity - JSON of the [xxdk.ReceptionIdentity].
//   - e2eParamsJSON - JSON of [xxdk.E2EParams].
func Login(cmixID int, callbacks *AuthCallbacks, identity,
	e2eParamsJSON []byte) (*E2e, error) {
	api, err := bindings.Login(cmixID, callbacks, identity, e2eParamsJSON)
	return &E2e{api}, err
}

// LoginEphemeral creates and returns a new ephemeral [E2e] object and adds it
// to the e2eTrackerSingleton. Identity should be created via
// [Cmix.MakeReceptionIdentity] or [Cmix.MakeLegacyReceptionIdentity] and passed
// in here. If callbacks is left nil, a default [auth.Callbacks] will be used.
//
// Parameters:
//   - cmixID - ID of [Cmix] object in tracker.
//   - callbacks - Javascript object that matches the [AuthCallbacks] struct.
//   - identity - JSON of the [xxdk.ReceptionIdentity].
//   - e2eParamsJSON - JSON of [xxdk.E2EParams].
//
// Returns:
//   - Javascript representation of the [E2e] object.
//   - Throws a TypeError if logging in fails.
func LoginEphemeral(cmixId int, callbacks *AuthCallbacks, identity,
	e2eParamsJSON []byte) (*E2e, error) {
	api, err :=
		bindings.LoginEphemeral(cmixId, callbacks, identity, e2eParamsJSON)
	return &E2e{api}, err
}

////////////////////////////////////////////////////////////////////////////////
// Auth Callbacks                                                             //
////////////////////////////////////////////////////////////////////////////////

// AuthCallbacks wraps a Javascript object so that it implements the
// [bindings.AuthCallbacks] interface.
type AuthCallbacks struct {
	RequestFn func(contact, receptionId []byte, ephemeralId, roundId int64) `wasm:"Request"`
	ConfirmFn func(contact, receptionId []byte, ephemeralId, roundId int64) `wasm:"Confirm"`
	ResetFn   func(contact, receptionId []byte, ephemeralId, roundId int64) `wasm:"Reset"`
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
func (a *AuthCallbacks) Request(
	contact, receptionId []byte, ephemeralId, roundId int64) {
	a.RequestFn(contact, receptionId, ephemeralId, roundId)
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
func (a *AuthCallbacks) Confirm(
	contact, receptionId []byte, ephemeralId, roundId int64) {
	a.ConfirmFn(contact, receptionId, ephemeralId, roundId)
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
func (a *AuthCallbacks) Reset(
	contact, receptionId []byte, ephemeralId, roundId int64) {
	a.ResetFn(contact, receptionId, ephemeralId, roundId)
}
