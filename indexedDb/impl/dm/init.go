////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package main

import (
	"crypto/ed25519"
	"gitlab.com/elixxir/crypto/database"
	"syscall/js"

	"github.com/hack-pad/go-indexeddb/idb"
	jww "github.com/spf13/jwalterweatherman"

	"gitlab.com/elixxir/client/v4/dm"
	"gitlab.com/elixxir/xxdk-wasm/indexedDb/impl"
)

// currentVersion is the current version of the IndexedDb runtime. Used for
// migration purposes.
const currentVersion uint = 1

// MessageReceivedCallback is called any time a message is received or updated.
//
// messageUpdate is true if the Message already exists and was edited.
// conversationUpdate is true if the Conversation was created or modified.
type MessageReceivedCallback func(
	uuid uint64, pubKey ed25519.PublicKey, messageUpdate, conversationUpdate bool)

// NewWASMEventModel returns a [channels.EventModel] backed by a wasmModel.
// The name should be a base64 encoding of the users public key. Returns the
// EventModel based on IndexedDb and the database name as reported by IndexedDb.
func NewWASMEventModel(databaseName string, encryption database.Cipher,
	cb MessageReceivedCallback) (dm.EventModel, error) {
	return newWASMModel(databaseName, encryption, cb)
}

// newWASMModel creates the given [idb.Database] and returns a wasmModel.
func newWASMModel(databaseName string, encryption database.Cipher,
	cb MessageReceivedCallback) (*wasmModel, error) {
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

	wrapper := &wasmModel{db: db, receivedMessageCB: cb, cipher: encryption}
	return wrapper, nil
}

// v1Upgrade performs the v0 -> v1 database upgrade.
//
// This can never be changed without permanently breaking backwards
// compatibility.
func v1Upgrade(db *idb.Database) error {
	indexOpts := idb.IndexOptions{
		Unique:     false,
		MultiEntry: false,
	}

	// Build Message ObjectStore and Indexes
	messageStoreOpts := idb.ObjectStoreOptions{
		KeyPath:       js.ValueOf(msgPkeyName),
		AutoIncrement: true,
	}
	messageStore, err := db.CreateObjectStore(messageStoreName, messageStoreOpts)
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
	_, err = messageStore.CreateIndex(messageStoreConversationIndex,
		js.ValueOf(messageStoreConversation), indexOpts)
	if err != nil {
		return err
	}
	_, err = messageStore.CreateIndex(messageStoreSenderIndex,
		js.ValueOf(messageStoreSender), indexOpts)
	if err != nil {
		return err
	}

	// Build Channel ObjectStore
	conversationStoreOpts := idb.ObjectStoreOptions{
		KeyPath:       js.ValueOf(convoPkeyName),
		AutoIncrement: false,
	}
	_, err = db.CreateObjectStore(conversationStoreName, conversationStoreOpts)
	if err != nil {
		return err
	}

	return nil
}
