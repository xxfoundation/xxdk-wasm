////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package channelEventModel

import (
	"crypto/ed25519"
	"syscall/js"

	"github.com/hack-pad/go-indexeddb/idb"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/v4/dm"
	cryptoChannel "gitlab.com/elixxir/crypto/channel"
	"gitlab.com/elixxir/xxdk-wasm/indexedDb"
	"gitlab.com/elixxir/xxdk-wasm/storage"
)

const (
	// databaseSuffix is the suffix to be appended to the name of
	// the database.
	databaseSuffix = "_speakeasy_dm"

	// currentVersion is the current version of the IndexDb
	// runtime. Used for migration purposes.
	currentVersion uint = 1
)

// MessageReceivedCallback is called any time a message is received or updated.
//
// update is true if the row is old and was edited.
type MessageReceivedCallback func(
	uuid uint64, pubKey ed25519.PublicKey, update bool)

// NewWASMEventModel returns a [channels.EventModel] backed by a wasmModel.
// The name should be a base64 encoding of the users public key.
func NewWASMEventModel(path string, encryption cryptoChannel.Cipher,
	cb MessageReceivedCallback) (dm.EventModel, error) {
	databaseName := path + databaseSuffix
	return newWASMModel(databaseName, encryption, cb)
}

// newWASMModel creates the given [idb.Database] and returns a wasmModel.
func newWASMModel(databaseName string, encryption cryptoChannel.Cipher,
	cb MessageReceivedCallback) (*wasmModel, error) {
	// Attempt to open database object
	ctx, cancel := indexedDb.NewContext()
	defer cancel()
	openRequest, err := idb.Global().Open(ctx, databaseName, currentVersion,
		func(db *idb.Database, oldVersion, newVersion uint) error {
			if oldVersion == newVersion {
				jww.INFO.Printf("IndexDb version is current: v%d",
					newVersion)
				return nil
			}

			jww.INFO.Printf("IndexDb upgrade required: v%d -> v%d",
				oldVersion, newVersion)

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
	}

	// Save the encryption status to storage
	encryptionStatus := encryption != nil
	loadedEncryptionStatus, err := storage.StoreIndexedDbEncryptionStatus(
		databaseName, encryptionStatus)
	if err != nil {
		return nil, err
	}

	// Verify encryption status does not change
	if encryptionStatus != loadedEncryptionStatus {
		return nil, errors.New(
			"Cannot load database with different encryption status.")
	} else if !encryptionStatus {
		jww.WARN.Printf("IndexedDb encryption disabled!")
	}

	// Attempt to ensure the database has been properly initialized
	openRequest, err = idb.Global().Open(ctx, databaseName, currentVersion,
		func(db *idb.Database, oldVersion, newVersion uint) error {
			return nil
		})
	if err != nil {
		return nil, err
	}
	// Wait for database open to finish
	db, err = openRequest.Await(ctx)
	if err != nil {
		return nil, err
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

	// Build Channel ObjectStore
	conversationStoreOpts := idb.ObjectStoreOptions{
		KeyPath:       js.ValueOf(convoPkeyName),
		AutoIncrement: false,
	}
	_, err = db.CreateObjectStore(conversationStoreName, conversationStoreOpts)
	if err != nil {
		return err
	}

	// Get the database name and save it to storage
	if databaseName, err := db.Name(); err != nil {
		return err
	} else if err = storage.StoreIndexedDb(databaseName); err != nil {
		return err
	}

	return nil
}
