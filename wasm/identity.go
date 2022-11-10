////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"gitlab.com/elixxir/client/v5/bindings"
	"gitlab.com/elixxir/client/v5/xxdk"
	"gitlab.com/elixxir/xxdk-wasm/utils"
	"syscall/js"
)

////////////////////////////////////////////////////////////////////////////////
// ReceptionIdentity                                                          //
////////////////////////////////////////////////////////////////////////////////

// StoreReceptionIdentity stores the given identity in [Cmix] storage with the
// given key. This is the ideal way to securely store identities, as the caller
// of this function is only required to store the given key separately rather
// than the keying material.
//
// Parameters:
//  - args[0] - Storage key (string).
//  - args[1] - JSON of the [xxdk.ReceptionIdentity] object (Uint8Array).
//  - args[2] - ID of [Cmix] object in tracker (int).
//
// Returns:
//  - Throws a TypeError if the identity cannot be stored in storage.
func StoreReceptionIdentity(_ js.Value, args []js.Value) interface{} {
	identity := utils.CopyBytesToGo(args[1])
	err := bindings.StoreReceptionIdentity(
		args[0].String(), identity, args[2].Int())

	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return nil
}

// LoadReceptionIdentity loads the given identity in [Cmix] storage with the
// given key.
//
// Parameters:
//  - args[0] - Storage key (string).
//  - args[1] - ID of [Cmix] object in tracker (int).
//
// Returns:
//  - JSON of the stored [xxdk.ReceptionIdentity] object (Uint8Array).
//  - Throws a TypeError if the identity cannot be retrieved from storage.
func LoadReceptionIdentity(_ js.Value, args []js.Value) interface{} {
	ri, err := bindings.LoadReceptionIdentity(args[0].String(), args[1].Int())
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return utils.CopyBytesToJS(ri)
}

// MakeReceptionIdentity generates a new cryptographic identity for receiving
// messages.
//
// Returns a promise:
//  - Resolves to the JSON of the [xxdk.ReceptionIdentity] object (Uint8Array).
//  - Rejected with an error if creating a new identity fails.
func (c *Cmix) MakeReceptionIdentity(js.Value, []js.Value) interface{} {
	promiseFn := func(resolve, reject func(args ...interface{}) js.Value) {
		ri, err := c.api.MakeReceptionIdentity()
		if err != nil {
			reject(utils.JsTrace(err))
		} else {
			resolve(utils.CopyBytesToJS(ri))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// MakeLegacyReceptionIdentity generates the legacy identity for receiving
// messages.
//
// Returns a promise:
//  - Resolves to the JSON of the [xxdk.ReceptionIdentity] object (Uint8Array).
//  - Rejected with an error if creating a new legacy identity fails.
func (c *Cmix) MakeLegacyReceptionIdentity(js.Value, []js.Value) interface{} {
	promiseFn := func(resolve, reject func(args ...interface{}) js.Value) {
		ri, err := c.api.MakeLegacyReceptionIdentity()
		if err != nil {
			reject(utils.JsTrace(err))
		} else {
			resolve(utils.CopyBytesToJS(ri))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// GetReceptionRegistrationValidationSignature returns the signature provided by
// the xx network.
//
// Returns:
//  - Reception registration validation signature (Uint8Array).
func (c *Cmix) GetReceptionRegistrationValidationSignature(
	js.Value, []js.Value) interface{} {
	return utils.CopyBytesToJS(
		c.api.GetReceptionRegistrationValidationSignature())
}

////////////////////////////////////////////////////////////////////////////////
// Contact Functions                                                          //
////////////////////////////////////////////////////////////////////////////////

// GetContactFromReceptionIdentity returns the [contact.Contact] object from the
// [xxdk.ReceptionIdentity].
//
// Parameters:
//  - args[0] - JSON of [xxdk.ReceptionIdentity] (Uint8Array).
//
// Returns:
//  - Marshalled bytes of [contact.Contact] (string).
//  - Throws a TypeError if unmarshalling the identity fails.
func GetContactFromReceptionIdentity(_ js.Value, args []js.Value) interface{} {
	// Note that this function does not appear in normal bindings
	identityJSON := utils.CopyBytesToGo(args[0])
	identity, err := xxdk.UnmarshalReceptionIdentity(identityJSON)
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return utils.CopyBytesToJS(identity.GetContact().Marshal())
}

// GetIDFromContact returns the ID in the [contact.Contact] object.
//
// Parameters:
//  - args[0] - Marshalled bytes of [contact.Contact] (Uint8Array).
//
// Returns:
//  - Marshalled bytes of [id.ID] (Uint8Array).
//  - Throws a TypeError if loading the ID from the contact file fails.
func GetIDFromContact(_ js.Value, args []js.Value) interface{} {
	cID, err := bindings.GetIDFromContact(utils.CopyBytesToGo(args[0]))
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return utils.CopyBytesToJS(cID)
}

// GetPubkeyFromContact returns the DH public key in the [contact.Contact]
// object.
//
// Parameters:
//  - args[0] - Marshalled [contact.Contact] (string).
//
// Returns:
//  - Bytes of the [cyclic.Int] object (Uint8Array).
//  - Throws a TypeError if loading the public key from the contact file fails.
func GetPubkeyFromContact(_ js.Value, args []js.Value) interface{} {
	key, err := bindings.GetPubkeyFromContact([]byte(args[0].String()))
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return utils.CopyBytesToJS(key)
}

////////////////////////////////////////////////////////////////////////////////
// Fact Functions                                                             //
////////////////////////////////////////////////////////////////////////////////

// SetFactsOnContact replaces the facts on the contact with the passed in facts
// pass in empty facts in order to clear the facts.
//
// Parameters:
//  - args[0] - Marshalled bytes of [contact.Contact] (Uint8Array).
//  - args[1] - JSON of [fact.FactList] (Uint8Array).
//
// Returns:
//  - Marshalled bytes of the modified [contact.Contact] (string).
//  - Throws a TypeError if loading or modifying the contact fails.
func SetFactsOnContact(_ js.Value, args []js.Value) interface{} {
	marshaledContact := utils.CopyBytesToGo(args[0])
	factListJSON := utils.CopyBytesToGo(args[1])
	c, err := bindings.SetFactsOnContact(marshaledContact, factListJSON)
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return c
}

// GetFactsFromContact returns the fact list in the [contact.Contact] object.
//
// Parameters:
//  - args[0] - Marshalled bytes of [contact.Contact] (Uint8Array).
//
// Returns:
//  - JSON of [fact.FactList] (Uint8Array).
//  - Throws a TypeError if loading the contact fails.
func GetFactsFromContact(_ js.Value, args []js.Value) interface{} {
	fl, err := bindings.GetFactsFromContact(utils.CopyBytesToGo(args[0]))
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return utils.CopyBytesToJS(fl)
}
