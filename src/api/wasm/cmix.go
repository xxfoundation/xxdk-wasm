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
)

// Cmix wraps the [bindings.Cmix] object so its methods can be wrapped to be
// Javascript compatible.
type Cmix struct {
	*bindings.Cmix
}

// NewCmix creates user storage, generates keys, connects, and registers with
// the network. Note that this does not register a username/identity, but merely
// creates a new cryptographic identity for adding such information at a later
// date.
//
// Users of this function should delete the storage directory on error.
//
// Parameters:
//   - args[0] - NDF JSON ([ndf.NetworkDefinition]) (string).
//   - args[1] - Storage directory path (string).
//   - args[2] - Password used for storage (Uint8Array).
//   - args[3] - Registration code (string).
//
// Returns a promise:
//   - Resolves on success.
//   - Rejected with an error if creating a new cMix client fails.
func NewCmix(ndfJSON, storageDir string, password []byte, registrationCode string) error {
	return bindings.NewCmix(ndfJSON, storageDir, password, registrationCode)
}

// LoadCmix will load an existing user storage from the storageDir using the
// password. This will fail if the user storage does not exist or the password
// is incorrect.
//
// The password is passed as a byte array so that it can be cleared from memory
// and stored as securely as possible using the MemGuard library.
//
// LoadCmix does not block on network connection and instead loads and starts
// subprocesses to perform network operations.
//
// Parameters:
//   - args[0] - Storage directory path (string).
//   - args[1] - Password used for storage (Uint8Array).
//   - args[2] - JSON of [xxdk.CMIXParams] (Uint8Array).
//
// Returns a promise:
//   - Resolves to a Javascript representation of the [Cmix] object.
//   - Rejected with an error if loading [Cmix] fails.
func LoadCmix(storageDir string, password []byte, cmixParamsJSON []byte) (*Cmix,
	error) {
	net, err := bindings.LoadCmix(storageDir, password, cmixParamsJSON)
	return &Cmix{net}, err
}
