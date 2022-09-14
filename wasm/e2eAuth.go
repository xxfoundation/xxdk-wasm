////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"gitlab.com/elixxir/xxdk-wasm/utils"
	"syscall/js"
)

// Request sends a contact request from the user identity in the imported E2e
// structure to the passed contact, as well as the passed facts (it will error
// if they are too long).
//
// The other party must accept the request by calling Confirm to be able to send
// messages using E2e.SendE2E. When the other party does so, the "confirm"
// callback will get called.
//
// The round the request is initially sent on will be returned, but the request
// will be listed as a critical message, so the underlying cMix client will auto
// resend it in the event of failure.
//
// A request cannot be sent for a contact who has already received a request or
// who is already a partner.
//
// The request sends as a critical message, if the round it sends on fails, it
// will be auto resent by the cMix client.
//
// Parameters:
//  - args[0] - marshalled bytes of the partner [contact.Contact] (Uint8Array).
//  - args[1] - JSON of [fact.FactList] (Uint8Array).
//
// Returns a promise:
//  - Resolves to the ID of the round (int).
//  - Rejected with an error if sending the request fails.
func (e *E2e) Request(_ js.Value, args []js.Value) interface{} {
	partnerContact := utils.CopyBytesToGo(args[0])
	factsListJson := utils.CopyBytesToGo(args[1])

	promiseFn := func(resolve, reject func(args ...interface{}) js.Value) {
		rid, err := e.api.Request(partnerContact, factsListJson)
		if err != nil {
			reject(utils.JsTrace(err))
		} else {
			resolve(rid)
		}
	}

	return utils.CreatePromise(promiseFn)
}

// Confirm sends a confirmation for a received request. It can only be called
// once. This both sends keying material to the other party and creates a
// channel in the e2e handler, after which e2e messages can be sent to the
// partner using E2e.SendE2E.
//
// The round the request is initially sent on will be returned, but the request
// will be listed as a critical message, so the underlying cMix client will auto
// resend it in the event of failure.
//
// A confirmation cannot be sent for a contact who has not sent a request or who
// is already a partner. This can only be called once for a specific contact.
// The confirmation sends as a critical message; if the round it sends on fails,
// it will be auto resent by the cMix client.
//
// If the confirmation must be resent, use ReplayConfirm.
//
// Parameters:
//  - args[0] - marshalled bytes of the partner [contact.Contact] (Uint8Array).
//
// Returns a promise:
//  - Resolves to the ID of the round (int).
//  - Rejected with an error if sending the confirmation fails.
func (e *E2e) Confirm(_ js.Value, args []js.Value) interface{} {
	partnerContact := utils.CopyBytesToGo(args[0])

	promiseFn := func(resolve, reject func(args ...interface{}) js.Value) {
		rid, err := e.api.Confirm(partnerContact)
		if err != nil {
			reject(utils.JsTrace(err))
		} else {
			resolve(rid)
		}
	}

	return utils.CreatePromise(promiseFn)
}

// Reset sends a contact reset request from the user identity in the imported
// e2e structure to the passed contact, as well as the passed facts (it will
// error if they are too long).
//
// This deletes all traces of the relationship with the partner from e2e and
// create a new relationship from scratch.
//
// The round the reset is initially sent on will be returned, but the request
// will be listed as a critical message, so the underlying cMix client will auto
// resend it in the event of failure.
//
// A request cannot be sent for a contact who has already received a request or
// who is already a partner.
//
// Parameters:
//  - args[0] - marshalled bytes of the partner [contact.Contact] (Uint8Array).
//
// Returns a promise:
//  - Resolves to the ID of the round (int).
//  - Rejected with an error if sending the reset fails.
func (e *E2e) Reset(_ js.Value, args []js.Value) interface{} {
	partnerContact := utils.CopyBytesToGo(args[0])

	promiseFn := func(resolve, reject func(args ...interface{}) js.Value) {
		rid, err := e.api.Reset(partnerContact)
		if err != nil {
			reject(utils.JsTrace(err))
		} else {
			resolve(rid)
		}
	}

	return utils.CreatePromise(promiseFn)
}

// ReplayConfirm resends a confirmation to the partner. It will fail to send if
// the send relationship with the partner has already ratcheted.
//
// The confirmation sends as a critical message; if the round it sends on fails,
// it will be auto resent by the cMix client.
//
// This will not be useful if either side has ratcheted.
//
// Parameters:
//  - args[0] - marshalled bytes of the partner [contact.Contact] (Uint8Array).
//
// Returns a promise:
//  - Resolves to the ID of the round (int).
//  - Rejected with an error if resending the confirmation fails.
func (e *E2e) ReplayConfirm(_ js.Value, args []js.Value) interface{} {
	partnerContact := utils.CopyBytesToGo(args[0])

	promiseFn := func(resolve, reject func(args ...interface{}) js.Value) {
		rid, err := e.api.ReplayConfirm(partnerContact)
		if err != nil {
			reject(utils.JsTrace(err))
		} else {
			resolve(rid)
		}
	}

	return utils.CreatePromise(promiseFn)
}

// CallAllReceivedRequests will iterate through all pending contact requests and
// replay them on the callbacks.
func (e *E2e) CallAllReceivedRequests(js.Value, []js.Value) interface{} {
	e.api.CallAllReceivedRequests()
	return nil
}

// DeleteRequest deletes sent or received requests for a specific partner ID.
//
// Parameters:
//  - args[0] - marshalled bytes of the partner [contact.Contact] (Uint8Array).
//
// Returns:
//  - Throws TypeError if the deletion fails.
func (e *E2e) DeleteRequest(_ js.Value, args []js.Value) interface{} {
	partnerContact := utils.CopyBytesToGo(args[0])
	err := e.api.DeleteRequest(partnerContact)
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return nil
}

// DeleteAllRequests clears all requests from auth storage.
//
// Returns:
//  - Throws TypeError if the deletion fails.
func (e *E2e) DeleteAllRequests(js.Value, []js.Value) interface{} {
	err := e.api.DeleteAllRequests()
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return nil
}

// DeleteSentRequests clears all sent requests from auth storage.
//
// Returns:
//  - Throws TypeError if the deletion fails.
func (e *E2e) DeleteSentRequests(js.Value, []js.Value) interface{} {
	err := e.api.DeleteSentRequests()
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return nil
}

// DeleteReceiveRequests clears all received requests from auth storage.
//
// Returns:
//  - Throws TypeError if the deletion fails.
func (e *E2e) DeleteReceiveRequests(js.Value, []js.Value) interface{} {
	err := e.api.DeleteReceiveRequests()
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return nil
}

// GetReceivedRequest returns a contact if there is a received request for it.
//
// Parameters:
//  - args[0] - marshalled bytes of the partner [contact.Contact] (Uint8Array).
//
// Returns:
//  - The marshalled bytes of [contact.Contact] (Uint8Array).
//  - Throws TypeError if getting the received request fails.
func (e *E2e) GetReceivedRequest(_ js.Value, args []js.Value) interface{} {
	partnerContact := utils.CopyBytesToGo(args[0])
	c, err := e.api.GetReceivedRequest(partnerContact)
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return utils.CopyBytesToJS(c)
}

// VerifyOwnership checks if the received ownership proof is valid.
//
// Parameters:
//  - args[0] - marshalled bytes of the received [contact.Contact] (Uint8Array).
//  - args[1] - marshalled bytes of the verified [contact.Contact] (Uint8Array).
//  - args[2] - ID of E2e object in tracker (int)
//
// Returns:
//  - Returns true if the ownership is valid (boolean)
//  - Throws TypeError if loading the parameters fails.
func (e *E2e) VerifyOwnership(_ js.Value, args []js.Value) interface{} {
	receivedContact := utils.CopyBytesToGo(args[0])
	verifiedContact := utils.CopyBytesToGo(args[1])
	isValid, err := e.api.VerifyOwnership(
		receivedContact, verifiedContact, args[2].Int())
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return isValid
}

// AddPartnerCallback adds a new callback that overrides the generic auth
// callback for the given partner ID.
//
// Parameters:
//  - args[0] - marshalled bytes of the partner [id.ID] (Uint8Array).
//  - args[1] - Javascript object that has functions that implement the
//    [bindings.AuthCallbacks] interface
//
// Returns:
//  - Throws TypeError if the [id.ID] cannot be unmarshalled.
func (e *E2e) AddPartnerCallback(_ js.Value, args []js.Value) interface{} {
	partnerID := utils.CopyBytesToGo(args[0])
	callbacks := newAuthCallbacks(args[1])
	err := e.api.AddPartnerCallback(partnerID, callbacks)
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return nil
}

// DeletePartnerCallback deletes the callback that overrides the generic
// auth callback for the given partner ID.
//
// Parameters:
//  - args[0] - marshalled bytes of the partner [id.ID] (Uint8Array).
//
// Returns:
//  - Throws TypeError if the [id.ID] cannot be unmarshalled.
func (e *E2e) DeletePartnerCallback(_ js.Value, args []js.Value) interface{} {
	partnerID := utils.CopyBytesToGo(args[0])
	err := e.api.DeletePartnerCallback(partnerID)
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return nil
}
