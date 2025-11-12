////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"fmt"
	"sync/atomic"
	"syscall/js"

	"gitlab.com/elixxir/client/v4/bindings"
	"gitlab.com/elixxir/wasm-utils/utils"
)

// initializing prevents a synchronized Cmix object from being loaded while one
// is being initialized.
var initializing atomic.Bool

// Cmix wraps the [bindings.Cmix] object so its methods can be wrapped to be
// Javascript compatible.
type Cmix struct {
	api *bindings.Cmix
}

// GenericKeyValue implements [bindings.GenericKeyValue] by wrapping a JavaScript object.
// It stores the parent js.Value to prevent the JavaScript callbacks from being garbage collected.
type GenericKeyValue struct {
	parent js.Value // Keep the parent object alive to prevent callback GC
	get    func(args ...any) js.Value
	set    func(args ...any) js.Value
	delete func(args ...any) js.Value
	keys   func(args ...any) js.Value
}

// kvRegistry stores references to prevent JS GC
var kvRegistry = js.Global().Get("Map").New()
var kvCounter int

// newGenericKeyValue maps the functions of the Javascript object matching
// [bindings.GenericKeyValue] to a GenericKeyValue.
func newGenericKeyValue(arg js.Value) *GenericKeyValue {
	fmt.Println("[DEBUG] newGenericKeyValue: arg type:", arg.Type())

	// Register in JS Map to prevent GC
	kvCounter++
	kvRegistry.Call("set", kvCounter, arg)
	fmt.Println("[DEBUG] newGenericKeyValue: Registered KV with ID:", kvCounter)

	return &GenericKeyValue{
		parent: arg, // Store parent to keep callbacks alive!
		get:    utils.WrapCB(arg, "Get"),
		set:    utils.WrapCB(arg, "Set"),
		delete: utils.WrapCB(arg, "Delete"),
		keys:   utils.WrapCB(arg, "Keys"),
	}
}

// Get implements [bindings.GenericKeyValue.Get]
func (kv *GenericKeyValue) Get(key string) ([]byte, error) {
	v, awaitErr := utils.Await(kv.get(key))
	if awaitErr != nil {
		return nil, js.Error{Value: awaitErr[0]}
	}
	return utils.CopyBytesToGo(v[0]), nil
}

// Set implements [bindings.GenericKeyValue.Set]
func (kv *GenericKeyValue) Set(key string, value []byte) error {
	_, awaitErr := utils.Await(kv.set(key, utils.CopyBytesToJS(value)))
	if awaitErr != nil {
		return js.Error{Value: awaitErr[0]}
	}
	return nil
}

// Delete implements [bindings.GenericKeyValue.Delete]
func (kv *GenericKeyValue) Delete(key string) error {
	_, awaitErr := utils.Await(kv.delete(key))
	if awaitErr != nil {
		return js.Error{Value: awaitErr[0]}
	}
	return nil
}

// Keys implements [bindings.GenericKeyValue.Keys]
func (kv *GenericKeyValue) Keys() ([]byte, error) {
	v, awaitErr := utils.Await(kv.keys())
	if awaitErr != nil {
		return nil, js.Error{Value: awaitErr[0]}
	}
	return utils.CopyBytesToGo(v[0]), nil
}

// newCmixJS creates a new Javascript compatible object (map[string]any) that
// matches the [Cmix] structure.
func newCmixJS(api *bindings.Cmix) map[string]any {
	c := Cmix{api}
	cmix := map[string]any{
		// cmix.go
		"GetID":          js.FuncOf(c.GetID),
		"GetReceptionID": js.FuncOf(c.GetReceptionID),
		"EKVGet":         utils.SafeFunc(c.EKVGet),
		"EKVSet":         utils.SafeFunc(c.EKVSet),

		// identity.go
		"MakeReceptionIdentity":                       utils.SafeFunc(c.MakeReceptionIdentity),
		"MakeLegacyReceptionIdentity":                 utils.SafeFunc(c.MakeLegacyReceptionIdentity),
		"GetReceptionRegistrationValidationSignature": js.FuncOf(c.GetReceptionRegistrationValidationSignature),

		// follow.go
		"StartNetworkFollower":            utils.SafeFunc(c.StartNetworkFollower),
		"StopNetworkFollower":             utils.SafeFunc(c.StopNetworkFollower),
		"SetTrackNetworkPeriod":           js.FuncOf(c.SetTrackNetworkPeriod),
		"WaitForNetwork":                  utils.SafeFunc(c.WaitForNetwork),
		"ReadyToSend":                     js.FuncOf(c.ReadyToSend),
		"NetworkFollowerStatus":           js.FuncOf(c.NetworkFollowerStatus),
		"GetNodeRegistrationStatus":       utils.SafeFunc(c.GetNodeRegistrationStatus),
		"IsReady":                         utils.SafeFunc(c.IsReady),
		"PauseNodeRegistrations":          utils.SafeFunc(c.PauseNodeRegistrations),
		"ChangeNumberOfNodeRegistrations": utils.SafeFunc(c.ChangeNumberOfNodeRegistrations),
		"HasRunningProcessies":            js.FuncOf(c.HasRunningProcessies),
		"IsHealthy":                       js.FuncOf(c.IsHealthy),
		"GetRunningProcesses":             utils.SafeFunc(c.GetRunningProcesses),
		"AddHealthCallback":               js.FuncOf(c.AddHealthCallback),
		"RemoveHealthCallback":            js.FuncOf(c.RemoveHealthCallback),
		"RegisterClientErrorCallback":     js.FuncOf(c.RegisterClientErrorCallback),
		"TrackServicesWithIdentity":       utils.SafeFunc(c.TrackServicesWithIdentity),
		"TrackServices":                   js.FuncOf(c.TrackServices),

		// connect.go
		"Connect": utils.SafeFunc(c.Connect),

		// delivery.go
		"WaitForRoundResult": utils.SafeFunc(c.WaitForRoundResult),

		// authenticatedConnection.go
		"ConnectWithAuthentication": utils.SafeFunc(c.ConnectWithAuthentication),
	}

	return cmix
}

// NewCmix creates user storage, generates keys, connects, and registers with
// the network using a GenericKeyValue for storage. Note that this does not
// register a username/identity, but merely creates a new cryptographic identity
// for adding such information at a later date.
//
// Users of this function should delete the storage directory on error.
//
// Parameters:
//   - args[0] - Javascript [GenericKeyValue] implementation.
//   - args[1] - NDF JSON ([ndf.NetworkDefinition]) (string).
//   - args[2] - Storage directory path (string).
//   - args[3] - Password used for storage (Uint8Array).
//   - args[4] - Registration code (string).
//
// Returns a promise:
//   - Resolves on success.
//   - Rejected with an error if creating a new cMix client fails.
func NewCmix(_ js.Value, args []js.Value) any {
	return utils.SafeFunc(func(this js.Value, args []js.Value) (any, error) {
		fmt.Println("[DEBUG] NewCmix: Starting")
		kv := newGenericKeyValue(args[0])
		fmt.Println("[DEBUG] NewCmix: Created GenericKeyValue, parent stored:", !kv.parent.IsUndefined())
		ndfJSON := args[1].String()
		storageDir := args[2].String()
		password := utils.CopyBytesToGo(args[3])
		registrationCode := args[4].String()

		fmt.Println("[DEBUG] NewCmix: Calling bindings.NewCmixWithKV")
		err := bindings.NewCmixWithKV(kv, ndfJSON, storageDir, password, registrationCode)
		fmt.Println("[DEBUG] NewCmix: Returned from bindings.NewCmixWithKV, err:", err)
		if err != nil {
			return nil, err
		}
		return js.Undefined(), nil
	}).Invoke(jsArgsToAny(args)...)
}

// LoadCmix will load an existing user storage backed by a key-value store from
// the storageDir using the
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
	return utils.SafeFunc(func(this js.Value, args []js.Value) (any, error) {
		storageDir := args[0].String()
		password := utils.CopyBytesToGo(args[1])
		cmixParamsJSON := utils.CopyBytesToGo(args[2])

		net, err := bindings.LoadCmix(storageDir, password, cmixParamsJSON)
		if err != nil {
			return nil, err
		}
		return newCmixJS(net), nil
	}).Invoke(jsArgsToAny(args)...)
}


// UnloadCmix will unload an existing cMix instance
//
// Parameters:
//   - args[0] - ID of [Cmix] object in tracker (int). This can be retrieved
//     using [Cmix.GetID].
//
// Returns error or nil
func UnloadCmix(_ js.Value, args []js.Value) any {
	cmixID := args[0].Int()
	return bindings.DeleteCmixInstance(cmixID)
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


// EKVGet allows access to a value inside the secure encrypted key value store.
//
// Parameters:
//   - args[0] - Key (string).
//
// Returns a promise:
//   - Resolves to the value (Uint8Array)
//   - Rejected with an error if accessing the KV fails.
func (c *Cmix) EKVGet(this js.Value, args []js.Value) (any, error) {
	key := args[0].String()

	val, err := c.api.EKVGet(key)
	if err != nil {
		return nil, err
	}
	return utils.CopyBytesToJS(val), nil
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
func (c *Cmix) EKVSet(this js.Value, args []js.Value) (any, error) {
	key := args[0].String()
	val := utils.CopyBytesToGo(args[1])

	err := c.api.EKVSet(key, val)
	if err != nil {
		return nil, err
	}
	return js.Undefined(), nil
}
