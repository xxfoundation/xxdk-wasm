////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package main

import (
	"github.com/hack-pad/go-indexeddb/idb"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/xxdk-wasm/indexedDb/impl"
	"syscall/js"
)

// currentVersion is the current version of the IndexedDb runtime. Used for
// migration purposes.
const currentVersion uint = 1

// NewState returns a [ClientState] backed by IndexedDb.
// The name should be a base64 encoding of the users public key.
func NewState(databaseName string) (ClientState, error) {
	return newState(databaseName)
}

// newState creates the given [idb.Database] and returns a stateModel.
func newState(databaseName string) (*stateModel, error) {
	// Attempt to open database object
	ctx, cancel := impl.NewContext()
	defer cancel()
	openRequest, err := idb.Global().Open(ctx, databaseName, currentVersion,
		func(db *idb.Database, oldVersion, newVersion uint) error {
			if oldVersion == newVersion {
				jww.INFO.Printf("IndexDb version for %s is current: v%d",
					databaseName, newVersion)
				return nil
			}

			jww.INFO.Printf("IndexDb upgrade required for %s: v%d -> v%d",
				databaseName, oldVersion, newVersion)

			if oldVersion == 0 && newVersion >= 1 {
				err := v1Upgrade(db)
				if err != nil {
					return err
				}
				oldVersion = 1
			}

			// if oldVersion == 1 && newVersion >= 2 { v2Upgrade(), oldVersion = 2 }
			return nil
		})
	if err != nil {
		return nil, err
	}

	// Wait for database open to finish
	db, err := openRequest.Await(ctx)
	if err != nil {
		return nil, err
	} else if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	wrapper := &stateModel{db: db}
	return wrapper, nil
}

// v1Upgrade performs the v0 -> v1 database upgrade.
//
// This can never be changed without permanently breaking backwards
// compatibility.
func v1Upgrade(db *idb.Database) error {
	storeOpts := idb.ObjectStoreOptions{
		KeyPath:       js.ValueOf(pkeyName),
		AutoIncrement: false,
	}
	_, err := db.CreateObjectStore(stateStoreName, storeOpts)
	return err
}
