////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package storage

import (
	"fmt"
	"sync/atomic"
	"syscall/js"

	json "github.com/goccy/go-json"
	"github.com/hack-pad/go-indexeddb/idb"
	jww "github.com/spf13/jwalterweatherman"

	utils "gitlab.com/elixxir/xxdk-wasm/jsutil"
	"gitlab.com/elixxir/xxdk-wasm/indexedDb/worker/kv"
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
//   - args[0] - The user-supplied password (string). This is the same password
//     passed into [wasm.NewCmix].
//
// Returns a promise:
//   - Resolves on success.
//   - Rejects with an error if the password is incorrect or if not all cMix followers
//     have been stopped.
func Purge(_ js.Value, args []js.Value) any {
	return utils.CreatePromise(func(resolve, reject func(...any) js.Value) {
		userPassword := args[0].String()

		// Check the password
		if !verifyPassword(userPassword) {
			errorConstructor := js.Global().Get("Error")
			errorObject := errorConstructor.New("invalid password")
			reject(errorObject)
			return
		}

		// Verify all Cmix followers are stopped
		if n := atomic.LoadUint64(&numClientsRunning); n != 0 {
			errorConstructor := js.Global().Get("Error")
			errorObject := errorConstructor.New(fmt.Sprintf("%d cMix followers running; all need to be stopped", n))
			reject(errorObject)
			return
		}

		// Get all indexedDb database names
		databaseList, err := GetIndexedDbList()
		if err != nil {
			errorConstructor := js.Global().Get("Error")
			errorObject := errorConstructor.New(fmt.Sprintf("failed to get list of indexedDb database names: %+v", err))
			reject(errorObject)
			return
		}
		jww.DEBUG.Printf("[PURGE] Found %d databases to delete: %s",
			len(databaseList), databaseList)

		// Delete each database
		for dbName := range databaseList {
			_, err = idb.Global().DeleteDatabase(dbName)
			if err != nil {
				errorConstructor := js.Global().Get("Error")
				errorObject := errorConstructor.New(fmt.Sprintf("failed to delete indexedDb database %q: %+v", dbName, err))
				reject(errorObject)
				return
			}
		}

		// Clear all KV Worker keys
		store := kv.GetStore()
		if store != nil {
			keysBytes, err := store.Keys()
			if err == nil {
				var keys []string
				if err := json.Unmarshal(keysBytes, &keys); err == nil {
					for _, key := range keys {
						_ = store.Delete(key)
					}
					jww.DEBUG.Printf("[PURGE] Deleted %d KV Worker keys", len(keys))
				}
			}
		} else {
			jww.WARN.Print("[PURGE] KV store not available, skipping KV cleanup")
		}

		resolve(js.Undefined())
	})
}
