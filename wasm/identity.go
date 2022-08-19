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

////////////////////////////////////////////////////////////////////////////////
// ReceptionIdentity                                                          //
////////////////////////////////////////////////////////////////////////////////

// StoreReceptionIdentity stores the given identity in Cmix storage with the
// given key. This is the ideal way to securely store identities, as the caller
// of this function is only required to store the given key separately rather
// than the keying material.
//
// Parameters:
//  - args[0] - storage key (string)
//  - args[1] - JSON of the [xxdk.ReceptionIdentity] object (Uint8Array)
//  - args[2] - ID of Cmix object in tracker (int)
//
// Returns:
//  - throws a TypeError if the identity cannot be stored in storage
func StoreReceptionIdentity(_ js.Value, args []js.Value) interface{} {
	identity := CopyBytesToGo(args[1])
	err := bindings.StoreReceptionIdentity(
		args[0].String(), identity, args[2].Int())

	if err != nil {
		Throw(TypeError, err)
		return nil
	}

	return nil
}

// LoadReceptionIdentity loads the given identity in Cmix storage with the given
// key.
//
// Parameters:
//  - args[0] - storage key (string)
//  - args[1] - ID of Cmix object in tracker (int)
//
// Returns:
//  - JSON of the stored [xxdk.ReceptionIdentity] object (Uint8Array)
//  - throws a TypeError if the identity cannot be retrieved from storage
func LoadReceptionIdentity(_ js.Value, args []js.Value) interface{} {
	ri, err := bindings.LoadReceptionIdentity(args[0].String(), args[1].Int())
	if err != nil {
		Throw(TypeError, err)
		return nil
	}

	return CopyBytesToJS(ri)
}

// MakeReceptionIdentity generates a new cryptographic identity for receiving
// messages.
//
// Returns:
//  - JSON of the [xxdk.ReceptionIdentity] object (Uint8Array)
//  - throws a TypeError if creating a new identity fails
func (c *Cmix) MakeReceptionIdentity(js.Value, []js.Value) interface{} {
	ri, err := c.api.MakeReceptionIdentity()
	if err != nil {
		Throw(TypeError, err)
		return nil
	}

	return CopyBytesToJS(ri)
}

// MakeLegacyReceptionIdentity generates the legacy identity for receiving
// messages.
//
// Returns:
//  - JSON of the [xxdk.ReceptionIdentity] object (Uint8Array)
//  - throws a TypeError if creating a new legacy identity fails
func (c *Cmix) MakeLegacyReceptionIdentity(js.Value, []js.Value) interface{} {
	ri, err := c.api.MakeLegacyReceptionIdentity()
	if err != nil {
		Throw(TypeError, err)
		return nil
	}

	return CopyBytesToJS(ri)
}

// GetReceptionRegistrationValidationSignature returns the signature provided by
// the xx network.
//
// Returns:
//  - signature (Uint8Array)
func (c *Cmix) GetReceptionRegistrationValidationSignature(
	js.Value, []js.Value) interface{} {
	return CopyBytesToJS(c.api.GetReceptionRegistrationValidationSignature())
}

////////////////////////////////////////////////////////////////////////////////
// Contact Functions                                                          //
////////////////////////////////////////////////////////////////////////////////

// GetIDFromContact returns the ID in the [contact.Contact] object.
//
// Parameters:
//  - args[0] - marshalled bytes of [contact.Contact] (string)
//
// Returns:
//  - marshalled [id.ID] object (Uint8Array)
//  - throws a TypeError if loading the ID from the contact file fails
func GetIDFromContact(_ js.Value, args []js.Value) interface{} {
	cID, err := bindings.GetIDFromContact([]byte(args[0].String()))
	if err != nil {
		Throw(TypeError, err)
		return nil
	}

	return CopyBytesToJS(cID)
}

// GetPubkeyFromContact returns the DH public key in the [contact.Contact]
// object.
//
// Parameters:
//  - args[0] - JSON of [contact.Contact] (string)
//
// Returns:
//  - bytes of the [cyclic.Int] object (Uint8Array)
//  - throws a TypeError if loading the public key from the contact file fails
func GetPubkeyFromContact(_ js.Value, args []js.Value) interface{} {
	key, err := bindings.GetPubkeyFromContact([]byte(args[0].String()))
	if err != nil {
		Throw(TypeError, err)
		return nil
	}

	return CopyBytesToJS(key)
}

////////////////////////////////////////////////////////////////////////////////
// Fact Functions                                                             //
////////////////////////////////////////////////////////////////////////////////

// SetFactsOnContact replaces the facts on the contact with the passed in facts
// pass in empty facts in order to clear the facts.
//
// Parameters:
//  - args[0] - JSON of [contact.Contact] (Uint8Array)
//  - args[1] - JSON of [fact.FactList] (Uint8Array)
//
// Returns:
//  - marshalled bytes of the modified [contact.Contact] (string)
//  - throws a TypeError if loading or modifying the contact fails
func SetFactsOnContact(_ js.Value, args []js.Value) interface{} {
	marshaledContact := CopyBytesToGo(args[0])
	factListJSON := CopyBytesToGo(args[1])
	c, err := bindings.SetFactsOnContact(marshaledContact, factListJSON)
	if err != nil {
		Throw(TypeError, err)
		return nil
	}

	return c
}

// GetFactsFromContact returns the fact list in the [contact.Contact] object.
//
// Parameters:
//  - args[0] - JSON of [contact.Contact] (Uint8Array)
//
// Returns:
//  - JSON of [fact.FactList] (Uint8Array)
//  - throws a TypeError if loading the contact fails
func GetFactsFromContact(_ js.Value, args []js.Value) interface{} {
	fl, err := bindings.GetFactsFromContact(CopyBytesToGo(args[0]))
	if err != nil {
		Throw(TypeError, err)
		return nil
	}

	return CopyBytesToJS(fl)
}
