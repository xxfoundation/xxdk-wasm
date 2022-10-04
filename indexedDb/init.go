////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package indexedDb

import (
	"syscall/js"

	"github.com/hack-pad/go-indexeddb/idb"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"

	"gitlab.com/elixxir/client/channels"
	"gitlab.com/xx_network/primitives/id"
)

const (
	// databaseSuffix is the suffix to be appended to the name of
	// the database.
	databaseSuffix = "_speakeasy"

	// currentVersion is the current version of the IndexDb
	// runtime. Used for migration purposes.
	currentVersion uint = 1
)

// MessageReceivedCallback is called any time a message is received or updated
// update is true if the row is old and was edited
type MessageReceivedCallback func(uuid uint64, channelID *id.ID, update bool)

// NewWASMEventModelBuilder returns an EventModelBuilder which allows
// the channel manager to define the path but the callback is the same
// across the board.
func NewWASMEventModelBuilder(
	cb MessageReceivedCallback) channels.EventModelBuilder {
	fn := func(path string) (channels.EventModel, error) {
		return NewWASMEventModel(path, cb)
	}
	return fn
}

// NewWASMEventModel returns a [channels.EventModel] backed by a wasmModel.
// The name should be a base64 encoding of the users public key.
func NewWASMEventModel(path string, cb MessageReceivedCallback) (
	channels.EventModel, error) {
	databaseName := path + databaseSuffix
	return newWASMModel(databaseName, cb)
}

// newWASMModel creates the given [idb.Database] and returns a wasmModel.
func newWASMModel(databaseName string, cb MessageReceivedCallback) (
	*wasmModel, error) {
	// Attempt to open database object
	ctx, cancel := newContext()
	defer cancel()
	openRequest, _ := idb.Global().Open(ctx, databaseName, currentVersion,
		func(db *idb.Database, oldVersion, newVersion uint) error {
			if oldVersion == newVersion {
				jww.INFO.Printf("IndexDb version is current: v%d", newVersion)
				return nil
			}

			jww.INFO.Printf(
				"IndexDb upgrade required: v%d -> v%d", oldVersion, newVersion)

			if oldVersion == 0 && newVersion == 1 {
				return v1Upgrade(db)
			}

			return errors.Errorf("Invalid version upgrade path: v%d -> v%d",
				oldVersion, newVersion)
		})

	// Wait for database open to finish
	db, err := openRequest.Await(ctx)

	return &wasmModel{db: db, receivedMessageCB: cb}, err
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
		js.ValueOf(messageStoreMessage), indexOpts)
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
	_, err = db.CreateObjectStore(channelsStoreName, storeOpts)
	if err != nil {
		return err
	}

	return nil
}
