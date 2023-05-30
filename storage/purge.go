////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package storage

import (
	"sync/atomic"
	"syscall/js"

	"github.com/hack-pad/go-indexeddb/idb"
	jww "github.com/spf13/jwalterweatherman"

	"gitlab.com/elixxir/client/v4/storage/utility"
	"gitlab.com/elixxir/wasm-utils/exception"
	"gitlab.com/elixxir/wasm-utils/storage"
)

// numClientsRunning is an atomic that tracks the current number of Cmix
// followers that have been started. Every time one is started, this counter
// must be incremented and every time one is stopped, it must be decremented.
//
// This variable is an atomic. Only access it with atomic functions
var numClientsRunning uint64

// IncrementNumClientsRunning increments the number client tracker. This should
// be called when starting the network follower.
func IncrementNumClientsRunning() {
	atomic.AddUint64(&numClientsRunning, 1)
}

// DecrementNumClientsRunning decrements the number client tracker. This should
// be called when stopping the network follower.
func DecrementNumClientsRunning() {
	atomic.AddUint64(&numClientsRunning, ^uint64(0))
}

// Purge clears all local storage and indexedDb databases saved by this WASM
// binary. This can only occur when no cMix followers are running. The user's
// password is required.
//
// Parameters:
//   - args[0] - Storage directory path (string). This is the same directory
//     path passed into [wasm.NewCmix].
//   - args[1] - The user-supplied password (string). This is the same password
//     passed into [wasm.NewCmix].
//
// Returns:
//   - Throws an error if the password is incorrect or if not all cMix
//     followers have been stopped.
func Purge(_ js.Value, args []js.Value) any {
	storageDirectory := args[0].String()
	userPassword := args[1].String()

	// Check the password
	if !verifyPassword(userPassword) {
		exception.Throwf("invalid password")
		return nil
	}

	// Verify all Cmix followers are stopped
	if n := atomic.LoadUint64(&numClientsRunning); n != 0 {
		exception.Throwf("%d cMix followers running; all need to be stopped", n)
		return nil
	}

	// Get all indexedDb database names
	databaseList, err := GetIndexedDbList()
	if err != nil {
		exception.Throwf(
			"failed to get list of indexedDb database names: %+v", err)
		return nil
	}
	jww.DEBUG.Printf("[PURGE] Found %d databases to delete: %s",
		len(databaseList), databaseList)

	// Delete each database
	for dbName := range databaseList {
		_, err = idb.Global().DeleteDatabase(dbName)
		if err != nil {
			exception.Throwf(
				"failed to delete indexedDb database %q: %+v", dbName, err)
			return nil
		}
	}

	// Get local storage
	ls := storage.GetLocalStorage()

	// Clear all local storage saved by this WASM project
	n := ls.Clear()
	jww.DEBUG.Printf("[PURGE] Cleared %d WASM keys in local storage", n)

	// Clear all EKV from local storage
	keys := ls.LocalStorageUNSAFE().KeysPrefix(storageDirectory)
	n = len(keys)
	for _, keyName := range keys {
		ls.LocalStorageUNSAFE().RemoveItem(keyName)
	}
	jww.DEBUG.Printf("[PURGE] Cleared %d keys with the prefix %q (for EKV)",
		n, storageDirectory)

	// Clear all NDFs saved to local storage
	keys = ls.LocalStorageUNSAFE().KeysPrefix(utility.NdfStorageKeyNamePrefix)
	n = len(keys)
	for _, keyName := range keys {
		ls.LocalStorageUNSAFE().RemoveItem(keyName)
	}
	jww.DEBUG.Printf("[PURGE] Cleared %d keys with the prefix %q (for NDF)",
		n, utility.NdfStorageKeyNamePrefix)

	return nil
}
