////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"syscall/js"

	"gitlab.com/elixxir/client/v4/bindings"
	"gitlab.com/elixxir/wasm-utils/exception"
	"gitlab.com/elixxir/wasm-utils/utils"
	"syscall/js"
)

// Cmix wraps the [bindings.Cmix] object so its methods can be wrapped to be
// Javascript compatible.
type Cmix struct {
	api *bindings.Cmix
}

// newCmixJS creates a new Javascript compatible object (map[string]any) that
// matches the [Cmix] structure.
func newCmixJS(api *bindings.Cmix) map[string]any {
	c := Cmix{api}
	cmix := map[string]any{
		// cmix.go
		"GetID":          js.FuncOf(c.GetID),
		"GetReceptionID": js.FuncOf(c.GetReceptionID),
		"GetRemoteKV":    js.FuncOf(c.GetRemoteKV),
		"EKVGet":         js.FuncOf(c.EKVGet),
		"EKVSet":         js.FuncOf(c.EKVSet),

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
		"SetTrackNetworkPeriod":           js.FuncOf(c.SetTrackNetworkPeriod),
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
		"TrackServicesWithIdentity":       js.FuncOf(c.TrackServicesWithIdentity),
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
//   - args[0] - NDF JSON ([ndf.NetworkDefinition]) (string).
//   - args[1] - Storage directory path (string).
//   - args[2] - Password used for storage (Uint8Array).
//   - args[3] - Registration code (string).
//
// Returns a promise:
//   - Resolves on success.
//   - Rejected with an error if creating a new cMix client fails.
func NewCmix(_ js.Value, args []js.Value) any {
	ndfJSON := args[0].String()
	storageDir := args[1].String()
	password := utils.CopyBytesToGo(args[2])
	registrationCode := args[3].String()

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		err := bindings.NewCmix(ndfJSON, storageDir, password, registrationCode)
		if err != nil {
			reject(exception.NewTrace(err))
		} else {
			resolve()
		}
	}

	return utils.CreatePromise(promiseFn)
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
func LoadCmix(_ js.Value, args []js.Value) any {
	storageDir := args[0].String()
	password := utils.CopyBytesToGo(args[1])
	cmixParamsJSON := utils.CopyBytesToGo(args[2])

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		net, err := bindings.LoadCmix(storageDir, password, cmixParamsJSON)
		if err != nil {
			reject(exception.NewTrace(err))
		} else {
			resolve(newCmixJS(net))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// LoadSynchronizedCmix will [LoadCmix] using a RemoteStore to establish
// a synchronized RemoteKV.
//
// Parameters:
//   - args[0] - Storage directory path (string).
//   - args[1] - Password used for storage (Uint8Array).
//   - args[2] - Javascript [RemoteStore] implementation.
//   - args[3] - JSON of [xxdk.CMIXParams] (Uint8Array).
//
// Returns a promise:
//   - Resolves to a Javascript representation of the [Cmix] object.
//   - Rejected with an error if loading [Cmix] fails.
func LoadSynchronizedCmix(_ js.Value, args []js.Value) any {
	storageDir := args[0].String()
	password := utils.CopyBytesToGo(args[1])
	rs := newRemoteStore(args[2])
	cmixParamsJSON := utils.CopyBytesToGo(args[3])

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		net, err := bindings.LoadSynchronizedCmix(storageDir, password,
			rs, cmixParamsJSON)
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
//   - Tracker ID (int).
func (c *Cmix) GetID(js.Value, []js.Value) any {
	return c.api.GetID()
}

// GetReceptionID returns the default reception identity for this cMix instance.
//
// Returns:
//   - Marshalled bytes of [id.ID] (Uint8Array).
func (c *Cmix) GetReceptionID(js.Value, []js.Value) any {
	return utils.CopyBytesToJS(c.api.GetReceptionID())
}

// GetRemoteKV returns the cMix RemoteKV
//
// Returns a promise:
//   - Resolves with the RemoteKV object.
func (c *Cmix) GetRemoteKV(_ js.Value, args []js.Value) any {

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		kv := c.api.GetRemoteKV()
		resolve(newRemoteKvJS(kv))
	}

	return utils.CreatePromise(promiseFn)
}

// EKVGet allows access to a value inside the secure encrypted key value store.
//
// Parameters:
//   - args[0] - Key (string).
//
// Returns a promise:
//   - Resolves to the value (Uint8Array)
//   - Rejected with an error if accessing the KV fails.
func (c *Cmix) EKVGet(_ js.Value, args []js.Value) any {
	key := args[0].String()

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		val, err := c.api.EKVGet(key)
		if err != nil {
			reject(exception.NewTrace(err))
		} else {
			resolve(utils.CopyBytesToJS(val))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// EKVSet sets a value inside the secure encrypted key value store.
//
// Parameters:
//   - args[0] - Key (string).
//   - args[1] - Value (Uint8Array).
//
// Returns a promise:
//   - Resolves on a successful save (void).
//   - Rejected with an error if saving fails.
func (c *Cmix) EKVSet(_ js.Value, args []js.Value) any {
	key := args[0].String()
	val := utils.CopyBytesToGo(args[1])

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		err := c.api.EKVSet(key, val)
		if err != nil {
			reject(exception.NewTrace(err))
		} else {
			resolve(nil)
		}
	}

	return utils.CreatePromise(promiseFn)
}
