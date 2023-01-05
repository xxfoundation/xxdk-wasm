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
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/v4/channels"
	cryptoChannel "gitlab.com/elixxir/crypto/channel"
	"gitlab.com/elixxir/xxdk-wasm/indexedDb"
	"gitlab.com/xx_network/primitives/id"
	"syscall/js"
	"time"
)

const (
	// databaseSuffix is the suffix to be appended to the name of
	// the database.
	databaseSuffix = "_speakeasy"

	// currentVersion is the current version of the IndexDb
	// runtime. Used for migration purposes.
	currentVersion uint = 1
)

// MessageReceivedCallback is called any time a message is received or updated.
//
// update is true if the row is old and was edited.
type MessageReceivedCallback func(uuid uint64, channelID *id.ID, update bool)

// NewWASMEventModel returns a [channels.EventModel] backed by a wasmModel.
// The name should be a base64 encoding of the users public key.
func NewWASMEventModel(path string, encryption cryptoChannel.Cipher,
	cb MessageReceivedCallback, storeEncryptionStatus storeEncryptionStatusFn) (
	channels.EventModel, error) {
	databaseName := path + databaseSuffix
	return newWASMModel(databaseName, encryption, cb, storeEncryptionStatus)
}

// storeEncryptionStatusFn matches storage.StoreIndexedDbEncryptionStatus so
// that the data can be sent between the worker and main thread.
type storeEncryptionStatusFn func(
	databaseName string, encryptionStatus bool) (bool, error)

// newWASMModel creates the given [idb.Database] and returns a wasmModel.
func newWASMModel(databaseName string, encryption cryptoChannel.Cipher,
	cb MessageReceivedCallback, storeEncryptionStatus storeEncryptionStatusFn) (
	*wasmModel, error) {
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
	loadedEncryptionStatus, err2 :=
		storeEncryptionStatus(databaseName, encryptionStatus)
	if err2 != nil {
		return nil, err2
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
	_, err = db.CreateObjectStore(channelsStoreName, storeOpts)
	if err != nil {
		return err
	}

	// Get the database name and save it to storage
	if databaseName, err2 := db.Name(); err2 != nil {
		return err2
	} else if err = storeDatabaseName(databaseName); err != nil {
		return err
	}

	return nil
}

// hackTestDb is a horrible function that exists as the result of an extremely
// long discussion about why initializing the IndexedDb sometimes silently
// fails. It ultimately tries to prevent an unrecoverable situation by actually
// inserting some nonsense data and then checking to see if it persists.
// If this function still exists in 2023, god help us all. Amen.
func (w *wasmModel) hackTestDb() error {
	testMessage := &Message{
		ID:        0,
		Nickname:  "test",
		MessageID: id.DummyUser.Marshal(),
	}
	msgId, helper := w.receiveHelper(testMessage, false)
	if helper != nil {
		return helper
	}
	result, err := indexedDb.Get(w.db, messageStoreName, js.ValueOf(msgId))
	if err != nil {
		return err
	}
	if result.IsUndefined() {
		return errors.Errorf("Failed to test db, record not present")
	}
	return nil
}

// storeDatabaseName sends the database name to storage.StoreIndexedDb in the
// main thread to be stored in localstorage and waits for the error to be
// returned.
//
// The function specified below is a placeholder until set by
// registerDatabaseNameStore. registerDatabaseNameStore must be called before
// storeDatabaseName.
var storeDatabaseName = func(databaseName string) error { return nil }

// RegisterDatabaseNameStore sets storeDatabaseName to send the database to
// storage.StoreIndexedDb in the main thread when called and registers a handler
// to listen for the response.
func RegisterDatabaseNameStore(m *manager) {
	storeDatabaseNameResponseChan := make(chan []byte)
	// Register handler
	m.mh.RegisterHandler(indexedDb.StoreDatabaseNameTag, func(data []byte) []byte {
		storeDatabaseNameResponseChan <- data
		return nil
	})

	storeDatabaseName = func(databaseName string) error {
		m.mh.SendResponse(indexedDb.StoreDatabaseNameTag, indexedDb.InitID,
			[]byte(databaseName))

		// Wait for response
		select {
		case response := <-storeDatabaseNameResponseChan:
			if len(response) > 0 {
				return errors.New(string(response))
			}
		case <-time.After(indexedDb.ResponseTimeout):
			return errors.Errorf("timed out after %s waiting for "+
				"response about storing the database name in local "+
				"storage in the main thread", indexedDb.ResponseTimeout)
		}
		return nil
	}
}
