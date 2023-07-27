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
	"gitlab.com/elixxir/wasm-utils/exception"
	"gitlab.com/elixxir/wasm-utils/utils"
	"syscall/js"
)

////////////////////////////////////////////////////////////////////////////////
// Structs and Interfaces                                                     //
////////////////////////////////////////////////////////////////////////////////

// UserDiscovery wraps the [bindings.UserDiscovery] object so its methods can be
// wrapped to be Javascript compatible.
type UserDiscovery struct {
	api *bindings.UserDiscovery
}

// newE2eJS creates a new Javascript compatible object (map[string]any) that
// matches the [E2e] structure.
func newUserDiscoveryJS(api *bindings.UserDiscovery) map[string]any {
	ud := UserDiscovery{api}
	udMap := map[string]any{
		"GetID":                  js.FuncOf(ud.GetID),
		"GetFacts":               js.FuncOf(ud.GetFacts),
		"GetContact":             js.FuncOf(ud.GetContact),
		"ConfirmFact":            js.FuncOf(ud.ConfirmFact),
		"SendRegisterFact":       js.FuncOf(ud.SendRegisterFact),
		"PermanentDeleteAccount": js.FuncOf(ud.PermanentDeleteAccount),
		"RemoveFact":             js.FuncOf(ud.RemoveFact),
	}

	return udMap
}

// GetID returns the ID for this [UserDiscovery] in the [UserDiscovery] tracker.
//
// Returns:
//   - Tracker ID (int).
func (ud *UserDiscovery) GetID(js.Value, []js.Value) any {
	return ud.api.GetID()
}

// udNetworkStatus wraps Javascript callbacks to adhere to the
// [bindings.UdNetworkStatus] interface.
type udNetworkStatus struct {
	udNetworkStatus func(args ...any) js.Value
}

// UdNetworkStatus returns the status of UD.
//
// Returns:
//   - UD status (int).
func (uns *udNetworkStatus) UdNetworkStatus() int {
	return uns.udNetworkStatus().Int()
}

////////////////////////////////////////////////////////////////////////////////
// Manager functions                                                          //
////////////////////////////////////////////////////////////////////////////////

// NewOrLoadUd loads an existing Manager from storage or creates a new one if
// there is no extant storage information. Parameters need be provided to
// specify how to connect to the User Discovery service. These parameters may be
// used to contact either the UD server hosted by the xx network team or a
// custom third-party operated server. For the former, all the information may
// be pulled from the NDF using the bindings.
//
// Params
//   - args[0] - ID of [E2e] object in tracker (int).
//   - args[1] - Javascript object that has functions that implement the
//     [bindings.UdNetworkStatus] interface. This is the network follower
//     function wrapped in [bindings.UdNetworkStatus].
//   - args[2] - The username the user wants to register with UD. If the user is
//     already registered, this field may be blank (string).
//   - args[3] - The registration validation signature; a signature provided by
//     the network (i.e., the client registrar). This may be nil; however, UD
//     may return an error in some cases (e.g., in a production level
//     environment) (Uint8Array).
//   - args[4] - The TLS certificate for the UD server this call will connect
//     with. You may use the UD server run by the xx network team by using
//     [E2e.GetUdCertFromNdf] (Uint8Array).
//   - args[5] - Marshalled bytes of the [contact.Contact] of the server this
//     call will connect with. You may use the UD server run by the xx network
//     team by using [E2e.GetUdContactFromNdf] (Uint8Array).
//   - args[6] - the IP address of the UD server this call will connect with.
//     You may use the UD server run by the xx network team by using
//     [E2e.GetUdAddressFromNdf] (string).
//
// Returns:
//   - Javascript representation of the [UserDiscovery] object that is
//     registered to the specified UD service.
//   - Throws an error if creating or loading fails.
func NewOrLoadUd(_ js.Value, args []js.Value) any {
	e2eID := args[0].Int()
	follower := &udNetworkStatus{utils.WrapCB(args[1], "UdNetworkStatus")}
	username := args[2].String()
	registrationValidationSignature := utils.CopyBytesToGo(args[3])
	cert := utils.CopyBytesToGo(args[4])
	contactFile := utils.CopyBytesToGo(args[5])
	address := args[6].String()

	api, err := bindings.NewOrLoadUd(e2eID, follower, username,
		registrationValidationSignature, cert, contactFile, address)
	if err != nil {
		exception.ThrowTrace(err)
		return nil
	}

	return newUserDiscoveryJS(api)
}

// NewUdManagerFromBackup builds a new user discover manager from a backup. It
// will construct a manager that is already registered and restore already
// registered facts into store.
//
// Note that it can take in both ane email address and a phone number or both.
// However, at least one fact must be specified; providing no facts will return
// an error.
//
// Parameters:
//   - args[0] - ID of [E2e] object in tracker (int).
//   - args[1] - Javascript object that has functions that implement the
//     [bindings.UdNetworkStatus] interface. This is the network follower
//     function wrapped in [bindings.UdNetworkStatus].
//   - args[2] - The TLS certificate for the UD server this call will connect
//     with. You may use the UD server run by the xx network team by using
//     [E2e.GetUdCertFromNdf] (Uint8Array).
//   - args[3] - Marshalled bytes of the [contact.Contact] of the server this
//     call will connect with. You may use the UD server run by the xx network
//     team by using [E2e.GetUdContactFromNdf] (Uint8Array).
//   - args[4] - The IP address of the UD server this call will connect with.
//     You may use the UD server run by the xx network team by using
//     [E2e.GetUdAddressFromNdf] (string).
//
// Returns:
//   - Javascript representation of the [UserDiscovery] object that is loaded
//     from backup.
//   - Throws an error if getting UD from backup fails.
func NewUdManagerFromBackup(_ js.Value, args []js.Value) any {
	e2eID := args[0].Int()
	follower := &udNetworkStatus{utils.WrapCB(args[1], "UdNetworkStatus")}
	cert := utils.CopyBytesToGo(args[5])
	contactFile := utils.CopyBytesToGo(args[6])
	address := args[7].String()

	api, err := bindings.NewUdManagerFromBackup(
		e2eID, follower, cert,
		contactFile, address)
	if err != nil {
		exception.ThrowTrace(err)
		return nil
	}

	return newUserDiscoveryJS(api)
}

// GetFacts returns a JSON marshalled list of [fact.Fact] objects that exist
// within the Store's registeredFacts map.
//
// Returns:
//   - JSON of [fact.FactList] (Uint8Array).
func (ud *UserDiscovery) GetFacts(js.Value, []js.Value) any {
	return utils.CopyBytesToJS(ud.api.GetFacts())
}

// GetContact returns the marshalled bytes of the [contact.Contact] for UD as
// retrieved from the NDF.
//
// Returns:
//   - Marshalled bytes of [contact.Contact] (Uint8Array).
//   - Throws TypeError if getting the contact fails.
func (ud *UserDiscovery) GetContact(js.Value, []js.Value) any {
	c, err := ud.api.GetContact()
	if err != nil {
		exception.ThrowTrace(err)
		return nil
	}

	return utils.CopyBytesToJS(c)
}

// ConfirmFact confirms a fact first registered via
// [UserDiscovery.SendRegisterFact]. The confirmation ID comes from
// [UserDiscovery.SendRegisterFact] while the code will come over the associated
// communications system.
//
// Parameters:
//   - args[0] - Confirmation ID (string).
//   - args[1] - Code (string).
//
// Returns:
//   - Throws TypeError if confirming the fact fails.
func (ud *UserDiscovery) ConfirmFact(_ js.Value, args []js.Value) any {
	err := ud.api.ConfirmFact(args[0].String(), args[1].String())
	if err != nil {
		exception.ThrowTrace(err)
		return nil
	}

	return nil
}

// SendRegisterFact adds a fact for the user to user discovery. Will only
// succeed if the user is already registered and the system does not have the
// fact currently registered for any user.
//
// This does not complete the fact registration process, it returns a
// confirmation ID instead. Over the communications system the fact is
// associated with, a code will be sent. This confirmation ID needs to be called
// along with the code to finalize the fact.
//
// Parameters:
//   - args[0] - JSON of [fact.Fact] (Uint8Array).
//
// Returns:
//   - The confirmation ID (string).
//   - Throws TypeError if sending the fact fails.
func (ud *UserDiscovery) SendRegisterFact(_ js.Value, args []js.Value) any {
	confirmationID, err := ud.api.SendRegisterFact(utils.CopyBytesToGo(args[0]))
	if err != nil {
		exception.ThrowTrace(err)
		return nil
	}

	return confirmationID
}

// PermanentDeleteAccount removes the username associated with this user from
// the UD service. This will only take a username type fact, and the fact must
// be associated with this user.
//
// Parameters:
//   - args[0] - JSON of [fact.Fact] (Uint8Array).
//
// Returns:
//   - Throws TypeError if deletion fails.
func (ud *UserDiscovery) PermanentDeleteAccount(_ js.Value, args []js.Value) any {
	err := ud.api.PermanentDeleteAccount(utils.CopyBytesToGo(args[0]))
	if err != nil {
		exception.ThrowTrace(err)
		return nil
	}

	return nil
}

// RemoveFact removes a previously confirmed fact. This will fail if the fact
// passed in is not UD service does not associate this fact with this user.
//
// Parameters:
//   - args[0] - JSON of [fact.Fact] (Uint8Array).
//
// Returns:
//   - Throws TypeError if removing the fact fails.
func (ud *UserDiscovery) RemoveFact(_ js.Value, args []js.Value) any {
	err := ud.api.RemoveFact(utils.CopyBytesToGo(args[0]))
	if err != nil {
		exception.ThrowTrace(err)
		return nil
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// User Discovery Lookup                                                      //
////////////////////////////////////////////////////////////////////////////////

// udLookupCallback wraps Javascript callbacks to adhere to the
// [bindings.UdLookupCallback] interface.
type udLookupCallback struct {
	callback func(args ...any) js.Value
}

// Callback is called by [LookupUD] to return the contact that matches the
// passed in ID.
//
// Parameters:
//   - contactBytes - Marshalled bytes of the [contact.Contact] returned from
//     the lookup, or nil if an error occurs (Uint8Array).
//   - err - Returns an error on failure (Error).
func (ulc *udLookupCallback) Callback(contactBytes []byte, err error) {
	ulc.callback(utils.CopyBytesToJS(contactBytes), exception.NewTrace(err))
}

// LookupUD returns the public key of the passed ID as known by the user
// discovery system or returns by the timeout.
//
// Parameters:
//   - args[0] - ID of [E2e] object in tracker (int).
//   - args[1] - Marshalled bytes of the User Discovery's [contact.Contact]
//     (Uint8Array).
//   - args[2] - Javascript object that has functions that implement the
//     [bindings.UdLookupCallback] interface.
//   - args[3] - Marshalled bytes of the [id.ID] for the user to look up
//     (Uint8Array).
//   - args[4] - JSON of [single.RequestParams] (Uint8Array).
//
// Returns a promise:
//   - Resolves to the JSON of the [bindings.SingleUseSendReport], which can be
//     passed into [Cmix.WaitForRoundResult] to see if the send succeeded
//     (Uint8Array).
//   - Rejected with an error if the lookup fails.
func LookupUD(_ js.Value, args []js.Value) any {
	e2eID := args[0].Int()
	udContact := utils.CopyBytesToGo(args[1])
	cb := &udLookupCallback{utils.WrapCB(args[2], "Callback")}
	lookupId := utils.CopyBytesToGo(args[3])
	singleRequestParamsJSON := utils.CopyBytesToGo(args[4])

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		sendReport, err := bindings.LookupUD(
			e2eID, udContact, cb, lookupId, singleRequestParamsJSON)
		if err != nil {
			reject(exception.NewTrace(err))
		} else {
			resolve(utils.CopyBytesToJS(sendReport))
		}
	}

	return utils.CreatePromise(promiseFn)
}

////////////////////////////////////////////////////////////////////////////////
// User Discovery Search                                                      //
////////////////////////////////////////////////////////////////////////////////

// udSearchCallback wraps Javascript callbacks to adhere to the
// [bindings.UdSearchCallback] interface.
type udSearchCallback struct {
	callback func(args ...any) js.Value
}

// Callback is called by [SearchUD] to return a list of [contact.Contact]
// objects that match the list of facts passed into [SearchUD].
//
// Parameters:
//   - contactListJSON - JSON of an array of [contact.Contact], or nil if an
//     error occurs (Uint8Array).
//   - err - Returns any error that occurred in the search (Error).
//
// JSON Example:
//
//	{
//	  "<xxc(2)F8dL9EC6gy+RMJuk3R+Au6eGExo02Wfio5cacjBcJRwDEgB7Ugdw/BAr6RkCABkWAFV1c2VybmFtZTA7c4LzV05sG+DMt+rFB0NIJg==xxc>",
//	  "<xxc(2)eMhAi/pYkW5jCmvKE5ZaTglQb+fTo1D8NxVitr5CCFADEgB7Ugdw/BAr6RoCABkWAFV1c2VybmFtZTE7fElAa7z3IcrYrrkwNjMS2w==xxc>",
//	  "<xxc(2)d7RJTu61Vy1lDThDMn8rYIiKSe1uXA/RCvvcIhq5Yg4DEgB7Ugdw/BAr6RsCABkWAFV1c2VybmFtZTI7N3XWrxIUpR29atpFMkcR6A==xxc>"
//	}
func (usc *udSearchCallback) Callback(contactListJSON []byte, err error) {
	usc.callback(utils.CopyBytesToJS(contactListJSON), exception.NewTrace(err))
}

// SearchUD searches user discovery for the passed Facts. The searchCallback
// will return a list of contacts, each having the facts it hit against. This is
// NOT intended to be used to search for multiple users at once; that can have a
// privacy reduction. Instead, it is intended to be used to search for a user
// where multiple pieces of information is known.
//
// Parameters:
//   - args[0] - ID of [E2e] object in tracker (int).
//   - args[1] - Marshalled bytes of the User Discovery's [contact.Contact]
//     (Uint8Array).
//   - args[2] - JSON of [fact.FactList] (Uint8Array).
//   - args[4] - JSON of [single.RequestParams] (Uint8Array).
//
// Returns a promise:
//   - Resolves to the JSON of the [bindings.SingleUseSendReport], which can be
//     passed into [Cmix.WaitForRoundResult] to see if the send succeeded
//     (Uint8Array).
//   - Rejected with an error if the search fails.
func SearchUD(_ js.Value, args []js.Value) any {
	e2eID := args[0].Int()
	udContact := utils.CopyBytesToGo(args[1])
	cb := &udSearchCallback{utils.WrapCB(args[2], "Callback")}
	factListJSON := utils.CopyBytesToGo(args[3])
	singleRequestParamsJSON := utils.CopyBytesToGo(args[4])

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		sendReport, err := bindings.SearchUD(
			e2eID, udContact, cb, factListJSON, singleRequestParamsJSON)
		if err != nil {
			reject(exception.NewTrace(err))
		} else {
			resolve(utils.CopyBytesToJS(sendReport))
		}
	}

	return utils.CreatePromise(promiseFn)
}
