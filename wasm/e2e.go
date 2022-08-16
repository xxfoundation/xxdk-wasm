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

// E2e wraps the [bindings.E2e] object so its methods can be wrapped to be
// Javascript compatible.
type E2e struct {
	api *bindings.E2e
}

// newE2eJS creates a new Javascript compatible object (map[string]interface{})
// that matches the E2e structure.
func newE2eJS(api *bindings.E2e) map[string]interface{} {
	e := E2e{api}
	e2e := map[string]interface{}{
		// e2e.go
		"GetID":               js.FuncOf(e.GetID),
		"GetContact":          js.FuncOf(e.GetContact),
		"GetUdAddressFromNdf": js.FuncOf(e.GetUdAddressFromNdf),
		"GetUdCertFromNdf":    js.FuncOf(e.GetUdCertFromNdf),
		"GetUdContactFromNdf": js.FuncOf(e.GetUdContactFromNdf),

		// e2eHandler.go
		"GetReceptionID":          js.FuncOf(e.GetReceptionID),
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

	return e2e
}

// GetID returns the ID for this [bindings.E2e] in the e2eTracker.
//
// Returns:
//  - int of the ID
func (e *E2e) GetID(js.Value, []js.Value) interface{} {
	return e.api.GetID()
}

// Login creates and returns a new E2e object and adds it to the
// e2eTrackerSingleton. Identity should be created via
// [Cmix.MakeReceptionIdentity] and passed in here. If callbacks is left nil, a
// default [auth.Callbacks] will be used.
//
// Parameters:
//  - args[0] - ID of Cmix object in tracker (int)
//  - args[1] - Javascript object that has functions that implement the
//    [bindings.AuthCallbacks] interface
//  - args[2] - JSON of the [xxdk.ReceptionIdentity] object (Uint8Array)
//  - args[3] - JSON of [xxdk.E2EParams] (Uint8Array)
//
// Returns:
//  - Javascript representation of the E2e object
//  - Throws a TypeError if logging in fails
func Login(_ js.Value, args []js.Value) interface{} {
	callbacks := newAuthCallbacks(args[1])
	identity := CopyBytesToGo(args[2])
	e2eParamsJSON := CopyBytesToGo(args[3])

	newE2E, err := bindings.Login(
		args[0].Int(), callbacks, identity, e2eParamsJSON)
	if err != nil {
		Throw(TypeError, err.Error())
		return nil
	}

	return newE2eJS(newE2E)
}

// LoginEphemeral creates and returns a new ephemeral E2e object and adds it to
// the e2eTrackerSingleton. Identity should be created via
// [Cmix.MakeReceptionIdentity] or [Cmix.MakeLegacyReceptionIdentity] and passed
// in here. If callbacks is left nil, a default [auth.Callbacks] will be used.
//
// Parameters:
//  - args[0] - ID of Cmix object in tracker (int)
//  - args[1] - Javascript object that has functions that implement the
//    [bindings.AuthCallbacks] interface
//  - args[2] - JSON of the [xxdk.ReceptionIdentity] object (Uint8Array)
//  - args[3] - JSON of [xxdk.E2EParams] (Uint8Array)
//
// Returns:
//  - Javascript representation of the E2e object
//  - Throws a TypeError if logging in fails
func LoginEphemeral(_ js.Value, args []js.Value) interface{} {
	callbacks := newAuthCallbacks(args[1])
	identity := CopyBytesToGo(args[2])
	e2eParamsJSON := CopyBytesToGo(args[3])

	newE2E, err := bindings.LoginEphemeral(
		args[0].Int(), callbacks, identity, e2eParamsJSON)
	if err != nil {
		Throw(TypeError, err.Error())
		return nil
	}

	return newE2eJS(newE2E)
}

// GetContact returns a [contact.Contact] object for the E2e ReceptionIdentity.
//
// Returns:
//  - Marshalled [contact.Contact] (Uint8Array)
func (e *E2e) GetContact(js.Value, []js.Value) interface{} {
	return CopyBytesToJS(e.api.GetContact())
}

// GetUdAddressFromNdf retrieve the User Discovery's network address fom the
// NDF.
//
// Returns:
//  - User Discovery's address (string)
func (e *E2e) GetUdAddressFromNdf(js.Value, []js.Value) interface{} {
	return e.api.GetUdAddressFromNdf()
}

// GetUdCertFromNdf retrieves the User Discovery's TLS certificate from the NDF.
//
// Returns:
//  - Public certificate in PEM format (Uint8Array)
func (e *E2e) GetUdCertFromNdf(js.Value, []js.Value) interface{} {
	return CopyBytesToJS(e.api.GetUdCertFromNdf())
}

// GetUdContactFromNdf assembles the User Discovery's contact file from the data
// within the NDF.
//
// Returns
//  - Marshalled [contact.Contact] (Uint8Array)
//  - Throws a TypeError if the contact file cannot be loaded
func (e *E2e) GetUdContactFromNdf(js.Value, []js.Value) interface{} {
	b, err := e.api.GetUdContactFromNdf()
	if err != nil {
		Throw(TypeError, err.Error())
		return nil
	}

	return CopyBytesToJS(b)
}

////////////////////////////////////////////////////////////////////////////////
// Auth Callbacks                                                             //
////////////////////////////////////////////////////////////////////////////////

// authCallbacks wraps Javascript callbacks to adhere to the
// [bindings.AuthCallbacks] interface.
type authCallbacks struct {
	request func(args ...interface{}) js.Value
	confirm func(args ...interface{}) js.Value
	reset   func(args ...interface{}) js.Value
}

// newAuthCallbacks adds all the callbacks from the Javascript object. If a
// callback is not defined, it is skipped.
func newAuthCallbacks(value js.Value) *authCallbacks {
	a := &authCallbacks{}

	request := value.Get("Request")
	if !request.IsUndefined() {
		a.request = request.Invoke
	}

	confirm := value.Get("Confirm")
	if !confirm.IsUndefined() {
		a.confirm = confirm.Invoke
	}

	reset := value.Get("Reset")
	if !reset.IsUndefined() {
		a.confirm = reset.Invoke
	}

	return a
}

func (a *authCallbacks) Request(contact, receptionId []byte, ephemeralId, roundId int64) {
	if a.request != nil {
		a.request(contact, receptionId, ephemeralId, roundId)
	}
}

func (a *authCallbacks) Confirm(contact, receptionId []byte, ephemeralId, roundId int64) {
	if a.confirm != nil {
		a.confirm(contact, receptionId, ephemeralId, roundId)
	}

}
func (a *authCallbacks) Reset(contact, receptionId []byte, ephemeralId, roundId int64) {
	if a.reset != nil {
		a.reset(contact, receptionId, ephemeralId, roundId)
	}
}
