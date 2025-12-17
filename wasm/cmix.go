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
	"sync"
	"syscall/js"

	"gitlab.com/elixxir/client/v4/bindings"
	"gitlab.com/elixxir/xxdk-wasm/indexedDb/worker/kv"
	utils "gitlab.com/elixxir/xxdk-wasm/jsutil"
)

// Cmix wraps the [bindings.Cmix] object so its methods can be wrapped to be
// Javascript compatible.
type Cmix struct {
	api *bindings.Cmix
}

// GenericKeyValue implements [bindings.GenericKeyValue] by wrapping a kv.Store.
// The kv.Store communicates with the KV Worker via MessageChannel, providing
// persistent IndexedDB storage for all components.
type GenericKeyValue struct {
	store kv.Store
}

// Global GenericKeyValue instance (singleton since there's only one KV Worker)
var (
	genericKV   *GenericKeyValue
	genericKVMu sync.Mutex
)

// getGenericKeyValue returns a GenericKeyValue that wraps the global kv.Store.
// This requires SetKVWorkerManager to have been called first.
// The kvPath parameter is passed for error messages; the underlying EKV handles namespacing.
func getGenericKeyValue(kvPath string) *GenericKeyValue {
	genericKVMu.Lock()
	defer genericKVMu.Unlock()

	if genericKV != nil {
		return genericKV
	}

	// Get the global store from the kv package (set by SetKVWorkerManager)
	store := kv.GetStore()
	if store == nil {
		panic("GenericKeyValue: KV Worker not initialized. " +
			"Ensure SetKVWorkerManager was called before NewCmix/LoadCmix (kvPath: " + kvPath + ")")
	}

	genericKV = &GenericKeyValue{store: store}
	return genericKV
}

// Get implements [bindings.GenericKeyValue.Get]
func (g *GenericKeyValue) Get(key string) ([]byte, error) {
	return g.store.Get(key)
}

// Set implements [bindings.GenericKeyValue.Set]
func (g *GenericKeyValue) Set(key string, value []byte) error {
	return g.store.Set(key, value)
}

// Delete implements [bindings.GenericKeyValue.Delete]
func (g *GenericKeyValue) Delete(key string) error {
	return g.store.Delete(key)
}

// Keys implements [bindings.GenericKeyValue.Keys]
func (g *GenericKeyValue) Keys() ([]byte, error) {
	return g.store.Keys()
}

// newCmixJS creates a new Javascript compatible object (js.Value) that
// matches the [Cmix] structure. Returns js.Value directly to avoid encoding issues.
func newCmixJS(api *bindings.Cmix) any {
	c := Cmix{api}

	// Create JavaScript object directly to avoid Go map encoding issues
	obj := js.Global().Get("Object").New()

	// cmix.go
	obj.Set("GetID", js.FuncOf(c.GetID))
	obj.Set("GetReceptionID", js.FuncOf(c.GetReceptionID))
	obj.Set("EKVGet", js.FuncOf(c.EKVGet))
	obj.Set("EKVSet", js.FuncOf(c.EKVSet))

	// identity.go
	obj.Set("MakeReceptionIdentity", js.FuncOf(c.MakeReceptionIdentity))
	obj.Set("MakeLegacyReceptionIdentity", js.FuncOf(c.MakeLegacyReceptionIdentity))
	obj.Set("GetReceptionRegistrationValidationSignature", js.FuncOf(c.GetReceptionRegistrationValidationSignature))

	// follow.go
	obj.Set("StartNetworkFollower", js.FuncOf(c.StartNetworkFollower))
	obj.Set("StopNetworkFollower", js.FuncOf(c.StopNetworkFollower))
	obj.Set("SetTrackNetworkPeriod", js.FuncOf(c.SetTrackNetworkPeriod))
	obj.Set("WaitForNetwork", js.FuncOf(c.WaitForNetwork))
	obj.Set("ReadyToSend", js.FuncOf(c.ReadyToSend))
	obj.Set("NetworkFollowerStatus", js.FuncOf(c.NetworkFollowerStatus))
	obj.Set("GetNodeRegistrationStatus", js.FuncOf(c.GetNodeRegistrationStatus))
	obj.Set("IsReady", js.FuncOf(c.IsReady))
	obj.Set("PauseNodeRegistrations", js.FuncOf(c.PauseNodeRegistrations))
	obj.Set("ChangeNumberOfNodeRegistrations", js.FuncOf(c.ChangeNumberOfNodeRegistrations))
	obj.Set("HasRunningProcessies", js.FuncOf(c.HasRunningProcessies))
	obj.Set("IsHealthy", js.FuncOf(c.IsHealthy))
	obj.Set("GetRunningProcesses", js.FuncOf(c.GetRunningProcesses))
	obj.Set("AddHealthCallback", js.FuncOf(c.AddHealthCallback))
	obj.Set("RemoveHealthCallback", js.FuncOf(c.RemoveHealthCallback))
	obj.Set("RegisterClientErrorCallback", js.FuncOf(c.RegisterClientErrorCallback))
	obj.Set("TrackServicesWithIdentity", js.FuncOf(c.TrackServicesWithIdentity))
	obj.Set("TrackServices", js.FuncOf(c.TrackServices))

	// connect.go
	obj.Set("Connect", js.FuncOf(c.Connect))

	// delivery.go
	obj.Set("WaitForRoundResult", js.FuncOf(c.WaitForRoundResult))

	// authenticatedConnection.go
	obj.Set("ConnectWithAuthentication", js.FuncOf(c.ConnectWithAuthentication))

	// Return as any - js.Value will be passed through without encoding
	return obj
}

// NewCmix creates user storage, generates keys, connects, and registers with
// the network. Note that this does not register a username/identity, but merely
// creates a new cryptographic identity for adding such information at a later date.
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
	// ✅ Parse ALL args BEFORE CreatePromise to avoid race conditions
	ndfJSON := args[0].String()
	storageDir := args[1].String()
	password := utils.CopyBytesToGo(args[2])
	registrationCode := args[3].String()

	return utils.CreatePromise(func(resolve, reject func(...any) js.Value) {
		err := bindings.NewCmix(ndfJSON, storageDir, password, registrationCode)
		if err != nil {
			errorConstructor := js.Global().Get("Error")
			errorObject := errorConstructor.New(err.Error())
			reject(errorObject)
			return
		}
		resolve(js.Undefined())
	})
}

// NewCmixWithKV creates user storage, generates keys, connects, and registers with
// the network using a GenericKeyValue for storage. Note that this does not
// register a username/identity, but merely creates a new cryptographic identity
// for adding such information at a later date.
//
// Users of this function should delete the storage directory on error.
//
// Parameters:
//   - args[0] - Global path string (e.g., "__xxdkKvInstance") where the KV object is stored.
//   - args[1] - NDF JSON ([ndf.NetworkDefinition]) (string).
//   - args[2] - Storage directory path (string).
//   - args[3] - Password used for storage (Uint8Array).
//   - args[4] - Registration code (string).
//
// Returns a promise:
//   - Resolves on success.
//   - Rejected with an error if creating a new cMix client fails.
func NewCmixWithKV(_ js.Value, args []js.Value) any {
	// ✅ Parse ALL args BEFORE CreatePromise to avoid race conditions
	kvPath := args[0].String()
	ndfJSON := args[1].String()
	storageDir := args[2].String()
	password := utils.CopyBytesToGo(args[3])
	registrationCode := args[4].String()

	return utils.CreatePromise(func(resolve, reject func(...any) js.Value) {
		fmt.Println("[DEBUG] NewCmixWithKV called, storageDir:", storageDir, "kvPath:", kvPath)

		// Get the GenericKeyValue that wraps the KV Worker store
		kvStore := getGenericKeyValue(kvPath)

		fmt.Println("[DEBUG] Calling bindings.NewCmixWithKV")
		err := bindings.NewCmixWithKV(kvStore, ndfJSON, storageDir, password, registrationCode)
		if err != nil {
			fmt.Println("[DEBUG] bindings.NewCmixWithKV error:", err)
			errorConstructor := js.Global().Get("Error")
			errorObject := errorConstructor.New(err.Error())
			reject(errorObject)
			return
		}

		fmt.Println("[DEBUG] bindings.NewCmixWithKV succeeded")
		resolve(js.Undefined())
	})
}

// LoadCmixWithKV will load an existing user storage backed by a key-value store from
// the storageDir using the password. This will fail if the user storage does not exist
// or the password is incorrect.
//
// The password is passed as a byte array so that it can be cleared from memory
// and stored as securely as possible using the MemGuard library.
//
// LoadCmixWithKV does not block on network connection and instead loads and starts
// subprocesses to perform network operations.
//
// Parameters:
//   - args[0] - Global path string (e.g., "__xxdkKvInstance") where the KV object is stored.
//   - args[1] - Storage directory path (string).
//   - args[2] - Password used for storage (Uint8Array).
//   - args[3] - JSON of [xxdk.CMIXParams] (Uint8Array).
//
// Returns a promise:
//   - Resolves to a Javascript representation of the [Cmix] object.
//   - Rejected with an error if loading [Cmix] fails.
func LoadCmixWithKV(_ js.Value, args []js.Value) any {
	// ✅ Parse ALL args BEFORE CreatePromise to avoid race conditions
	kvPath := args[0].String()
	storageDir := args[1].String()
	password := utils.CopyBytesToGo(args[2])
	cmixParamsJSON := utils.CopyBytesToGo(args[3])

	return utils.CreatePromise(func(resolve, reject func(...any) js.Value) {
		fmt.Println("[DEBUG] LoadCmixWithKV called, kvPath:", kvPath)
		fmt.Println("[DEBUG] LoadCmixWithKV CreatePromise starting")

		kvStore := getGenericKeyValue(kvPath)

		fmt.Println("[DEBUG] Calling bindings.LoadCmixWithKV")
		net, err := bindings.LoadCmixWithKV(kvStore, storageDir, password, cmixParamsJSON)
		if err != nil {
			fmt.Println("[DEBUG] bindings.LoadCmixWithKV error:", err)
			errorConstructor := js.Global().Get("Error")
			errorObject := errorConstructor.New(err.Error())
			reject(errorObject)
			return
		}

		fmt.Println("[DEBUG] bindings.LoadCmixWithKV succeeded, returning newCmixJS")
		resolve(newCmixJS(net))
	})
}

// LoadCmix will load an existing user storage from the storageDir using the password.
// This will fail if the user storage does not exist or the password is incorrect.
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
	// ✅ Parse ALL args BEFORE CreatePromise to avoid race conditions
	storageDir := args[0].String()
	password := utils.CopyBytesToGo(args[1])
	cmixParamsJSON := utils.CopyBytesToGo(args[2])

	return utils.CreatePromise(func(resolve, reject func(...any) js.Value) {
		net, err := bindings.LoadCmix(storageDir, password, cmixParamsJSON)
		if err != nil {
			errorConstructor := js.Global().Get("Error")
			errorObject := errorConstructor.New(err.Error())
			reject(errorObject)
			return
		}
		resolve(newCmixJS(net))
	})
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
func (c *Cmix) EKVGet(_ js.Value, args []js.Value) any {
	// ✅ Parse ALL args BEFORE CreatePromise to avoid race conditions
	key := args[0].String()

	return utils.CreatePromise(func(resolve, reject func(...any) js.Value) {
		val, err := c.api.EKVGet(key)
		if err != nil {
			errorConstructor := js.Global().Get("Error")
			errorObject := errorConstructor.New(err.Error())
			reject(errorObject)
			return
		}

		resolve(utils.CopyBytesToJS(val))
	})
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
	// ✅ Parse ALL args BEFORE CreatePromise to avoid race conditions
	key := args[0].String()
	val := utils.CopyBytesToGo(args[1])

	return utils.CreatePromise(func(resolve, reject func(...any) js.Value) {
		err := c.api.EKVSet(key, val)
		if err != nil {
			errorConstructor := js.Global().Get("Error")
			errorObject := errorConstructor.New(err.Error())
			reject(errorObject)
			return
		}

		resolve(js.Undefined())
	})
}
