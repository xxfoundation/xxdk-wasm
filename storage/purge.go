////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package storage

import (
	"github.com/hack-pad/go-indexeddb/idb"
	"github.com/pkg/errors"
	"gitlab.com/elixxir/client/storage/utility"
	"gitlab.com/elixxir/xxdk-wasm/utils"
	"sync/atomic"
	"syscall/js"
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
//  - args[0] - Storage directory path (string). This is the same directory path
//    passed into [wasm.NewCmix].
//  - args[1] - The user-supplied password (string). This is the same password
//    passed into [wasm.NewCmix].
//
// Returns:
//  - Throws a TypeError if the password is incorrect or if not all cMix
//    followers have been stopped.
func Purge(_ js.Value, args []js.Value) interface{} {
	storageDirectory := args[0].String()
	userPassword := args[1].String()
	// Clear all EKV from local storage
	GetLocalStorage().ClearPrefix("speakeasyapp")

	// Check the password
	if !verifyPassword(userPassword) {
		utils.Throw(utils.TypeError, errors.New("invalid password"))
		return nil
	}

	// Verify all Cmix followers are stopped
	// if n := atomic.LoadUint64(&numClientsRunning); n != 0 {
	// 	utils.Throw(utils.TypeError, errors.Errorf(
	// 		"%d cMix followers running; all need to be stopped", n))
	// 	return nil
	// }

	// Get all indexedDb database names
	databaseList, err := GetIndexedDbList()
	if err != nil {
		utils.Throw(utils.TypeError, errors.Errorf(
			"failed to get list of indexedDb database names: %+v", err))
		return nil
	}

	// Delete each database
	for dbName := range databaseList {
		_, err = idb.Global().DeleteDatabase(dbName)
		if err != nil {
			utils.Throw(utils.TypeError, errors.Errorf(
				"failed to delete indexedDb database %q: %+v", dbName, err))
			return nil
		}
	}

	// Get local storage
	ls := GetLocalStorage()

	// Clear all local storage saved by this WASM project
	ls.ClearWASM()

	// Clear all EKV from local storage
	ls.ClearPrefix(storageDirectory)

	// Clear all NDFs saved to local storage
	ls.ClearPrefix(utility.NdfStorageKeyNamePrefix)

	return nil
}
