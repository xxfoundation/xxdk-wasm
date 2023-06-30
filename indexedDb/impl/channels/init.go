////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package main

import (
	"syscall/js"

	"github.com/hack-pad/go-indexeddb/idb"
	jww "github.com/spf13/jwalterweatherman"

	"gitlab.com/elixxir/client/v4/channels"
	idbCrypto "gitlab.com/elixxir/crypto/indexedDb"
	"gitlab.com/elixxir/xxdk-wasm/indexedDb/impl"
)

// currentVersion is the current version of the IndexedDb runtime. Used for
// migration purposes.
const currentVersion uint = 1

// eventUpdate takes an event type and JSON object from
// bindings/channelsCallbacks.go.
type eventUpdate func(eventType int64, jsonMarshallable any)

// NewWASMEventModel returns a [channels.EventModel] backed by a wasmModel.
// The name should be a base64 encoding of the users public key. Returns the
// EventModel based on IndexedDb and the database name as reported by IndexedDb.
func NewWASMEventModel(databaseName string, encryption idbCrypto.Cipher,
	eu eventUpdate) (channels.EventModel, error) {
	return newWASMModel(databaseName, encryption, eu)
}

// newWASMModel creates the given [idb.Database] and returns a wasmModel.
func newWASMModel(databaseName string, encryption idbCrypto.Cipher,
	eu eventUpdate) (*wasmModel, error) {
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

	wrapper := &wasmModel{
		db:     db,
		cipher: encryption,
		eu:     eu,
	}
	return wrapper, nil
}

// v1Upgrade performs the v0 -> v1 database upgrade.
//
// This can never be changed without permanently breaking backwards
// compatibility.
func v1Upgrade(db *idb.Database) error {
	storeOpts := idb.ObjectStoreOptions{
		KeyPath:       js.ValueOf(pkeyName),
		AutoIncrement: true,
	}
	indexOpts := idb.IndexOptions{
		Unique:     false,
		MultiEntry: false,
	}

	// Build Message ObjectStore and Indexes
	messageStore, err := db.CreateObjectStore(messageStoreName, storeOpts)
	if err != nil {
		return err
	}
	_, err = messageStore.CreateIndex(messageStoreMessageIndex,
		js.ValueOf(messageStoreMessage),
		idb.IndexOptions{
			Unique:     true,
			MultiEntry: false,
		})
	if err != nil {
		return err
	}
	_, err = messageStore.CreateIndex(messageStoreChannelIndex,
		js.ValueOf(messageStoreChannel), indexOpts)
	if err != nil {
		return err
	}
	_, err = messageStore.CreateIndex(messageStoreParentIndex,
		js.ValueOf(messageStoreParent), indexOpts)
	if err != nil {
		return err
	}
	_, err = messageStore.CreateIndex(messageStoreTimestampIndex,
		js.ValueOf(messageStoreTimestamp), indexOpts)
	if err != nil {
		return err
	}
	_, err = messageStore.CreateIndex(messageStorePinnedIndex,
		js.ValueOf(messageStorePinned), indexOpts)
	if err != nil {
		return err
	}

	// Build Channel ObjectStore
	_, err = db.CreateObjectStore(channelStoreName, storeOpts)
	if err != nil {
		return err
	}

	// Build File ObjectStore
	_, err = db.CreateObjectStore(fileStoreName, idb.ObjectStoreOptions{
		KeyPath:       js.ValueOf(pkeyName),
		AutoIncrement: false,
	})
	return err
}
