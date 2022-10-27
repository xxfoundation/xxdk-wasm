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
	"gitlab.com/elixxir/xxdk-wasm/utils"
	"syscall/js"
)

// Cmix wraps the [bindings.Cmix] object so its methods can be wrapped to be
// Javascript compatible.
type Cmix struct {
	api *bindings.Cmix
}

// newCmixJS creates a new Javascript compatible object (map[string]interface{})
// that matches the [Cmix] structure.
func newCmixJS(api *bindings.Cmix) map[string]interface{} {
	c := Cmix{api}
	cmix := map[string]interface{}{
		// cmix.go
		"GetID": js.FuncOf(c.GetID),

		// identity.go
		"MakeReceptionIdentity": js.FuncOf(
			c.MakeReceptionIdentity),
		"MakeLegacyReceptionIdentity": js.FuncOf(
			c.MakeLegacyReceptionIdentity),
		"GetReceptionRegistrationValidationSignature": js.FuncOf(
			c.GetReceptionRegistrationValidationSignature),

		// follow.go
		"StartNetworkFollower":            js.FuncOf(c.StartNetworkFollower),
		"StopNetworkFollower":             js.FuncOf(c.StopNetworkFollower),
		"WaitForNetwork":                  js.FuncOf(c.WaitForNetwork),
		"ReadyToSend":                     js.FuncOf(c.ReadyToSend),
		"NetworkFollowerStatus":           js.FuncOf(c.NetworkFollowerStatus),
		"GetNodeRegistrationStatus":       js.FuncOf(c.GetNodeRegistrationStatus),
		"IsReady":                         js.FuncOf(c.IsReady),
		"PauseNodeRegistrations":          js.FuncOf(c.PauseNodeRegistrations),
		"ChangeNumberOfNodeRegistrations": js.FuncOf(c.ChangeNumberOfNodeRegistrations),
		"HasRunningProcessies":            js.FuncOf(c.HasRunningProcessies),
		"IsHealthy":                       js.FuncOf(c.IsHealthy),
		"GetRunningProcesses":             js.FuncOf(c.GetRunningProcesses),
		"AddHealthCallback":               js.FuncOf(c.AddHealthCallback),
		"RemoveHealthCallback":            js.FuncOf(c.RemoveHealthCallback),
		"RegisterClientErrorCallback":     js.FuncOf(c.RegisterClientErrorCallback),
		"TrackServices":                   js.FuncOf(c.TrackServices),

		// connect.go
		"Connect": js.FuncOf(c.Connect),

		// delivery.go
		"WaitForRoundResult": js.FuncOf(c.WaitForRoundResult),

		// authenticatedConnection.go
		"ConnectWithAuthentication": js.FuncOf(c.ConnectWithAuthentication),
	}

	return cmix
}

// NewCmix creates user storage, generates keys, connects, and registers with
// the network. Note that this does not register a username/identity, but merely
// creates a new cryptographic identity for adding such information at a later
// date.
//
// Users of this function should delete the storage directory on error.
//
// Parameters:
//  - args[0] - NDF JSON ([ndf.NetworkDefinition]) (string).
//  - args[1] - Storage directory path (string).
//  - args[2] - Password used for storage (Uint8Array).
//  - args[3] - Registration code (string).
//
// Returns:
//  - Throws a TypeError if creating new [Cmix] fails.
func NewCmix(_ js.Value, args []js.Value) interface{} {
	password := utils.CopyBytesToGo(args[2])

	err := bindings.NewCmix(
		args[0].String(), args[1].String(), password, args[3].String())
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return nil
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
//  - args[0] - Storage directory path (string).
//  - args[1] - Password used for storage (Uint8Array).
//  - args[2] - JSON of [xxdk.CMIXParams] (Uint8Array).
//
// Returns a promise:
//  - Resolves to a Javascript representation of the [Cmix] object.
//  - Rejected with an error if loading [Cmix] fails.
func LoadCmix(_ js.Value, args []js.Value) interface{} {
	storageDir := args[0].String()
	password := utils.CopyBytesToGo(args[1])
	cmixParamsJSON := utils.CopyBytesToGo(args[2])

	promiseFn := func(resolve, reject func(args ...interface{}) js.Value) {
		net, err := bindings.LoadCmix(storageDir, password, cmixParamsJSON)
		if err != nil {
			reject(utils.JsTrace(err))
		} else {
			resolve(newCmixJS(net))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// GetID returns the ID for this [bindings.Cmix] in the cmixTracker.
//
// Returns:
//  - Tracker ID (int).
func (c *Cmix) GetID(js.Value, []js.Value) interface{} {
	return c.api.GetID()
}
