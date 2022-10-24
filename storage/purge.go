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
	"gitlab.com/elixxir/xxdk-wasm/utils"
	"sync/atomic"
	"syscall/js"
)

// NumClientsRunning is an atomic that tracks the current number of Cmix
// followers that have been started. Every time one is started, this counter
// must be incremented and every time one is stopped, it must be decremented.
//
// This variable is an atomic. Only access it with atomic functions
var NumClientsRunning uint64

// Purge clears all local storage and indexedDb databases saved by this WASM
// binary. All Cmix followers must be closed and the user's password is
// required.
//
// Warning: This deletes all storage local to the webpage running this WASM.
// Only use if you want to destroy everything.
//
// Parameters:
//  - args[0] - Storage directory path (string).
//  - args[1] - Password used for storage (Uint8Array).
//
// Returns:
//  - Throws a TypeError if the password is incorrect or if not all Cmix
//    followers have been stopped.
func Purge(_ js.Value, args []js.Value) interface{} {
	// Check the password
	if !verifyPassword(args[1].String()) {
		utils.Throw(utils.TypeError, errors.New("invalid password"))
		return nil
	}

	// Verify all Cmix followers are stopped
	if n := atomic.LoadUint64(&NumClientsRunning); n != 0 {
		utils.Throw(
			utils.TypeError, errors.Errorf("%d Cmix followers running", n))
		return nil
	}

	// Get all indexedDb database names
	databaseList, err := GetIndexedDbList()
	if err != nil {
		utils.Throw(
			utils.TypeError, errors.Errorf(
				"failed to get list of indexedDb database names: %+v", err))
		return nil
	}

	// Delete each database
	for dbName := range databaseList {
		_, err = idb.Global().DeleteDatabase(dbName)
		if err != nil {
			utils.Throw(
				utils.TypeError, errors.Errorf(
					"failed to delete indexedDb database %q: %+v", dbName, err))
			return nil
		}
	}

	// Clear WASM local storage and EKV
	ls := GetLocalStorage()
	ls.ClearWASM()
	ls.ClearPrefix(args[0].String())

	return nil
}
