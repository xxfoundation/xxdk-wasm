////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"gitlab.com/elixxir/client/v4/xxdk"
	"gitlab.com/elixxir/xxdk-wasm/src/api/utils"
	"syscall/js"
)

////////////////////////////////////////////////////////////////////////////////
// ReceptionIdentity                                                          //
////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
// Contact Functions                                                          //
////////////////////////////////////////////////////////////////////////////////

// GetContactFromReceptionIdentity returns the [contact.Contact] object from the
// [xxdk.ReceptionIdentity].
//
// Parameters:
//   - args[0] - JSON of [xxdk.ReceptionIdentity] (Uint8Array).
//
// Returns:
//   - Marshalled bytes of [contact.Contact] (string).
//   - Throws a TypeError if unmarshalling the identity fails.
func GetContactFromReceptionIdentity(_ js.Value, args []js.Value) any {
	// Note that this function does not appear in normal bindings
	identityJSON := utils.CopyBytesToGo(args[0])
	identity, err := xxdk.UnmarshalReceptionIdentity(identityJSON)
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return utils.CopyBytesToJS(identity.GetContact().Marshal())
}
