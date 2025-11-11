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
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"

	"gitlab.com/elixxir/wasm-utils/storage"
	"gitlab.com/elixxir/wasm-utils/utils"
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
	return utils.SafeFunc(func(this js.Value, args []js.Value) (any, error) {
		userPassword := args[0].String()

		// Check the password
		if !verifyPassword(userPassword) {
			return nil, errors.New("invalid password")
		}

		// Verify all Cmix followers are stopped
		if n := atomic.LoadUint64(&numClientsRunning); n != 0 {
			return nil, errors.Errorf("%d cMix followers running; all need to be stopped", n)
		}

		// Get all indexedDb database names
		databaseList, err := GetIndexedDbList()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get list of indexedDb database names")
		}
		jww.DEBUG.Printf("[PURGE] Found %d databases to delete: %s",
			len(databaseList), databaseList)

		// Delete each database
		for dbName := range databaseList {
			_, err = idb.Global().DeleteDatabase(dbName)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to delete indexedDb database %q", dbName)
			}
		}

		// Get local storage
		ls := storage.GetLocalStorage()

		// Clear all local storage saved by this WASM project
		n := ls.Clear()
		jww.DEBUG.Printf("[PURGE] Cleared %d WASM keys in local storage", n)

		return nil, nil
	}).Invoke(js.Value{}, args)
}
