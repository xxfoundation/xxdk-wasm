////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 Privategrity Corporation                                   /
//                                                                             /
// All rights reserved.                                                        /
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm
// +build js,wasm

package indexedDb

import (
	"context"
	"github.com/hack-pad/go-indexeddb/idb"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"syscall/js"

	"gitlab.com/elixxir/client/channels"
)

const (
	// databaseSuffix to be appended to the name of the database
	databaseSuffix = "_messenger"
	// currentVersion of the IndexDb runtime. Used for migration purposes.
	currentVersion uint = 1
)

// NewWasmEventModel returns a [channels.EventModel] backed by a wasmModel
func NewWasmEventModel(username string) (channels.EventModel, error) {
	ctx := context.Background()
	databaseName := username + databaseSuffix

	// Attempt to open database object
	openRequest, _ := idb.Global().Open(ctx, databaseName, currentVersion,
		func(db *idb.Database, oldVersion, newVersion uint) error {
			if oldVersion == newVersion {
				jww.INFO.Printf("IndexDb version is current: v%d",
					newVersion)
				return nil
			}

			jww.INFO.Printf("IndexDb upgrade required: v%d -> v%d",
				oldVersion, newVersion)

			if oldVersion == 0 && newVersion == 1 {
				return v1Upgrade(db)
			}

			return errors.Errorf("Invalid version upgrade path: v%d -> v%d",
				oldVersion, newVersion)
		})

	// Wait for database open to finish
	db, err := openRequest.Await(ctx)

	return &wasmModel{db: db}, err
}

// v1Upgrade performs the v0 -> v1 database upgrade.
// This can never be changed without permanently breaking backwards compatibility.
func v1Upgrade(db *idb.Database) error {
	storeOpts := idb.ObjectStoreOptions{
		KeyPath:       js.ValueOf(pkeyName),
		AutoIncrement: false,
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

	// Build User ObjectStore
	_, err = db.CreateObjectStore(userStoreName, storeOpts)
	if err != nil {
		return err
	}

	// Build Channel ObjectStore
	_, err = db.CreateObjectStore(channelsStoreName, storeOpts)
	if err != nil {
		return err
	}

	return nil
}
