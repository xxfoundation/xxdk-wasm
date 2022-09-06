////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"gitlab.com/elixxir/client/bindings"
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

// newE2eJS creates a new Javascript compatible object (map[string]interface{})
// that matches the E2e structure.
func newUserDiscoveryJS(api *bindings.UserDiscovery) map[string]interface{} {
	ud := UserDiscovery{api}
	udMap := map[string]interface{}{
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

// GetID returns the ID for this UserDiscovery in the UserDiscovery tracker.
//
// Returns:
//  - int
func (ud *UserDiscovery) GetID(js.Value, []js.Value) interface{} {
	return ud.api.GetID()
}

// udNetworkStatus wraps Javascript callbacks to adhere to the
// [bindings.UdNetworkStatus] interface.
type udNetworkStatus struct {
	udNetworkStatus func(args ...interface{}) js.Value
}

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
//  - args[0] - ID of E2e object in tracker (int).
//  - args[1] - Javascript object that has functions that implement the
//    [bindings.UdNetworkStatus] interface. This is the network follower
//    function wrapped in [bindings.UdNetworkStatus].
//  - args[2] - the username the user wants to register with UD. If the user is
//    already registered, this field may be blank (string).
//  - args[3] - the registration validation signature; a signature provided by
//    the network (i.e., the client registrar). This may be nil; however, UD may
//    return an error in some cases (e.g., in a production level environment)
//    (Uint8Array).
//  - args[4] - the TLS certificate for the UD server this call will connect
//    with. You may use the UD server run by the xx network team by using
//    E2e.GetUdCertFromNdf (Uint8Array).
//  - args[5] - marshalled [contact.Contact]. This represents the contact file
//    of the server this call will connect with. You may use the UD server run
//    by the xx network team by using E2e.GetUdContactFromNdf (Uint8Array).
//  - args[6] - the IP address of the UD server this call will connect with. You
//    may use the UD server run by the xx network team by using
//    E2e.GetUdAddressFromNdf (string).
//
// Returns:
//  - Javascript representation of the UserDiscovery object that is registered
//    to the specified UD service.
//  - Throws a TypeError if creating or loading fails.
func NewOrLoadUd(_ js.Value, args []js.Value) interface{} {
	e2eID := args[0].Int()
	follower := &udNetworkStatus{WrapCB(args[1], "UdNetworkStatus")}
	username := args[2].String()
	registrationValidationSignature := CopyBytesToGo(args[3])
	cert := CopyBytesToGo(args[4])
	contactFile := CopyBytesToGo(args[5])
	address := args[6].String()

	api, err := bindings.NewOrLoadUd(e2eID, follower, username,
		registrationValidationSignature, cert, contactFile, address)
	if err != nil {
		Throw(TypeError, err)
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
//  - args[0] - ID of E2e object in tracker (int).
//  - args[1] - Javascript object that has functions that implement the
//    [bindings.UdNetworkStatus] interface. This is the network follower
//    function wrapped in [bindings.UdNetworkStatus].
//  - args[2] - JSON of [fact.Fact] username that is registered with UD
//    (Uint8Array).
//  - args[3] - JSON of [fact.Fact] email address that is registered with UD
//    (Uint8Array).
//  - args[4] - JSON of [fact.Fact] phone number that is registered with UD
//    (Uint8Array).
//  - args[5] - the TLS certificate for the UD server this call will connect
//    with. You may use the UD server run by the xx network team by using
//    E2e.GetUdCertFromNdf (Uint8Array).
//  - args[6] - marshalled [contact.Contact]. This represents the contact file
//    of the server this call will connect with. You may use the UD server run
//    by the xx network team by using E2e.GetUdContactFromNdf (Uint8Array).
//  - args[7] - the IP address of the UD server this call will connect with. You
//    may use the UD server run by the xx network team by using
//    E2e.GetUdAddressFromNdf (string).
//
// Returns:
//  - Javascript representation of the UserDiscovery object that is loaded from
//    backup.
//  - Throws a TypeError if getting UD from backup fails.
func NewUdManagerFromBackup(_ js.Value, args []js.Value) interface{} {
	e2eID := args[0].Int()
	follower := &udNetworkStatus{WrapCB(args[1], "UdNetworkStatus")}
	usernameFactJson := CopyBytesToGo(args[2])
	emailFactJson := CopyBytesToGo(args[3])
	phoneFactJson := CopyBytesToGo(args[4])
	cert := CopyBytesToGo(args[5])
	contactFile := CopyBytesToGo(args[6])
	address := args[7].String()

	api, err := bindings.NewUdManagerFromBackup(e2eID, follower, usernameFactJson, emailFactJson,
		phoneFactJson, cert, contactFile, address)
	if err != nil {
		Throw(TypeError, err)
		return nil
	}

	return newUserDiscoveryJS(api)
}

// GetFacts returns a JSON marshalled list of [fact.Fact] objects that exist
// within the Store's registeredFacts map.
//
// Returns:
//  - JSON of [fact.FactList] (Uint8Array).
func (ud *UserDiscovery) GetFacts(js.Value, []js.Value) interface{} {
	return CopyBytesToJS(ud.api.GetFacts())
}

// GetContact returns the marshalled bytes of the [contact.Contact] for UD as
// retrieved from the NDF.
//
// Returns:
//  - JSON of [contact.Contact] (Uint8Array).
//  - Throws TypeError if getting the contact fails.
func (ud *UserDiscovery) GetContact(js.Value, []js.Value) interface{} {
	c, err := ud.api.GetContact()
	if err != nil {
		Throw(TypeError, err)
		return nil
	}

	return CopyBytesToJS(c)
}

// ConfirmFact confirms a fact first registered via SendRegisterFact. The
// confirmation ID comes from SendRegisterFact while the code will come over the
// associated communications system.
//
// Parameters:
//  - args[0] - confirmation ID (string).
//  - args[1] - code (string).
//
// Returns:
//  - Throws TypeError if confirming the fact fails.
func (ud *UserDiscovery) ConfirmFact(_ js.Value, args []js.Value) interface{} {
	err := ud.api.ConfirmFact(args[0].String(), args[1].String())
	if err != nil {
		Throw(TypeError, err)
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
//  - args[0] - JSON of [fact.Fact] (Uint8Array).
//
// Returns:
//  - The confirmation ID (string).
//  - Throws TypeError if sending the fact fails.
func (ud *UserDiscovery) SendRegisterFact(_ js.Value, args []js.Value) interface{} {
	confirmationID, err := ud.api.SendRegisterFact(CopyBytesToGo(args[0]))
	if err != nil {
		Throw(TypeError, err)
		return nil
	}

	return confirmationID
}

// PermanentDeleteAccount removes the username associated with this user from
// the UD service. This will only take a username type fact, and the fact must
// be associated with this user.
//
// Parameters:
//  - args[0] - JSON of [fact.Fact] (Uint8Array).
//
// Returns:
//  - Throws TypeError if deletion fails.
func (ud *UserDiscovery) PermanentDeleteAccount(_ js.Value, args []js.Value) interface{} {
	err := ud.api.PermanentDeleteAccount(CopyBytesToGo(args[0]))
	if err != nil {
		Throw(TypeError, err)
		return nil
	}

	return nil
}

// RemoveFact removes a previously confirmed fact. This will fail if the fact
// passed in is not UD service does not associate this fact with this user.
//
// Parameters:
//  - args[0] - JSON of [fact.Fact] (Uint8Array).
//
// Returns:
//  - Throws TypeError if removing the fact fails.
func (ud *UserDiscovery) RemoveFact(_ js.Value, args []js.Value) interface{} {
	err := ud.api.RemoveFact(CopyBytesToGo(args[0]))
	if err != nil {
		Throw(TypeError, err)
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
	callback func(args ...interface{}) js.Value
}

func (ulc *udLookupCallback) Callback(contactBytes []byte, err error) {
	ulc.callback(CopyBytesToJS(contactBytes), err.Error())
}

// LookupUD returns the public key of the passed ID as known by the user
// discovery system or returns by the timeout.
//
// Parameters:
//  - args[0] - ID of E2e object in tracker (int).
//  - args[1] - JSON of User Discovery's [contact.Contact] (Uint8Array).
//  - args[2] - Javascript object that has functions that implement the
//    [bindings.UdLookupCallback] interface.
//  - args[3] - JSON of [id.ID] for the user to look up (Uint8Array).
//  - args[4] - JSON of [single.RequestParams] (Uint8Array).
//
// Returns:
//  - JSON of [bindings.SingleUseSendReport], which can be passed into
//    Cmix.WaitForRoundResult to see if the send succeeded.
//  - Throws a TypeError if the lookup fails.
func LookupUD(_ js.Value, args []js.Value) interface{} {
	e2eID := args[0].Int()
	udContact := CopyBytesToGo(args[1])
	cb := &udLookupCallback{WrapCB(args[2], "Callback")}
	lookupId := CopyBytesToGo(args[3])
	singleRequestParamsJSON := CopyBytesToGo(args[4])

	report, err := bindings.LookupUD(
		e2eID, udContact, cb, lookupId, singleRequestParamsJSON)
	if err != nil {
		Throw(TypeError, err)
		return nil
	}

	return CopyBytesToJS(report)
}

////////////////////////////////////////////////////////////////////////////////
// User Discovery Search                                                      //
////////////////////////////////////////////////////////////////////////////////

// udSearchCallback wraps Javascript callbacks to adhere to the
// [bindings.UdSearchCallback] interface.
type udSearchCallback struct {
	callback func(args ...interface{}) js.Value
}

func (usc *udSearchCallback) Callback(contactListJSON []byte, err error) {
	usc.callback(CopyBytesToJS(contactListJSON), err.Error())
}

// SearchUD searches user discovery for the passed Facts. The searchCallback
// will return a list of contacts, each having the facts it hit against. This is
// NOT intended to be used to search for multiple users at once; that can have a
// privacy reduction. Instead, it is intended to be used to search for a user
// where multiple pieces of information is known.
//
// Parameters:
//  - args[0] - ID of E2e object in tracker (int).
//  - args[1] - JSON of User Discovery's [contact.Contact] (Uint8Array).
//  - args[2] - JSON of [fact.FactList] (Uint8Array).
//  - args[4] - JSON of [single.RequestParams] (Uint8Array).
//
// Returns:
//  - JSON of [bindings.SingleUseSendReport], which can be passed into
//    Cmix.WaitForRoundResult to see if the send succeeded.
//  - Throws a TypeError if the search fails.
func SearchUD(_ js.Value, args []js.Value) interface{} {
	e2eID := args[0].Int()
	udContact := CopyBytesToGo(args[1])
	cb := &udSearchCallback{WrapCB(args[2], "Callback")}
	factListJSON := CopyBytesToGo(args[3])
	singleRequestParamsJSON := CopyBytesToGo(args[4])

	report, err := bindings.SearchUD(
		e2eID, udContact, cb, factListJSON, singleRequestParamsJSON)
	if err != nil {
		Throw(TypeError, err)
		return nil
	}

	return CopyBytesToJS(report)
}
