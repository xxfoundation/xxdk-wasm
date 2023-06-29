////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package dm

import (
	"crypto/ed25519"
	"encoding/json"

	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"

	"gitlab.com/elixxir/client/v4/bindings"
	"gitlab.com/elixxir/client/v4/dm"
	idbCrypto "gitlab.com/elixxir/crypto/indexedDb"
	"gitlab.com/elixxir/xxdk-wasm/logging"
	"gitlab.com/elixxir/xxdk-wasm/storage"
	"gitlab.com/elixxir/xxdk-wasm/worker"
)

// databaseSuffix is the suffix to be appended to the name of the database.
const databaseSuffix = "_speakeasy_dm"

// MessageReceivedCallback is called any time a message is received or updated.
//
// messageUpdate is true if the Message already exists and was edited.
// conversationUpdate is true if the Conversation was created or modified.
type MessageReceivedCallback func(uuid uint64, pubKey ed25519.PublicKey,
	messageUpdate, conversationUpdate bool)

// NewWASMEventModelMessage is JSON marshalled and sent to the worker for
// [NewWASMEventModel].
type NewWASMEventModelMessage struct {
	DatabaseName   string `json:"databaseName"`
	EncryptionJSON string `json:"encryptionJSON"`
}

// NewWASMEventModel returns a [channels.EventModel] backed by a wasmModel.
// The name should be a base64 encoding of the users public key.
func NewWASMEventModel(path, wasmJsPath string, encryption idbCrypto.Cipher,
	cbs bindings.DmCallbacks) (dm.EventModel, error) {
	databaseName := path + databaseSuffix

	wh, err := worker.NewManager(wasmJsPath, "dmIndexedDb", true)
	if err != nil {
		return nil, err
	}

	// Register handler to manage messages for the MessageReceivedCallback
	wh.RegisterCallback(
		MessageReceivedCallbackTag, messageReceivedCallbackHandler(cbs))

	// Create MessageChannel between worker and logger so that the worker logs
	// are saved
	err = worker.CreateMessageChannel(logging.GetLogger().Worker(), wh,
		"dmIndexedDbLogger", worker.LoggerTag)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create message channel "+
			"between DM indexedDb worker and logger")
	}

	// Store the database name
	err = storage.StoreIndexedDb(databaseName)
	if err != nil {
		return nil, err
	}

	// Check that the encryption status
	encryptionStatus := encryption != nil
	err = checkDbEncryptionStatus(databaseName, encryptionStatus)
	if err != nil {
		return nil, err
	}

	encryptionJSON, err := json.Marshal(encryption)
	if err != nil {
		return nil, err
	}

	msg := NewWASMEventModelMessage{
		DatabaseName:   databaseName,
		EncryptionJSON: string(encryptionJSON),
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	response, err := wh.SendMessage(NewWASMEventModelTag, payload)
	if err != nil {
		return nil, errors.Wrapf(err,
			"failed to send message %q", NewWASMEventModelTag)
	} else if len(response) > 0 {
		return nil, errors.New(string(response))
	}

	return &wasmModel{wh}, nil
}

// messageReceivedCallbackHandler returns a handler to manage messages for the
// MessageReceivedCallback.
func messageReceivedCallbackHandler(cbs bindings.DmCallbacks) worker.ReceiverCallback {
	return func(message []byte, _ func([]byte)) {
		cbs.EventUpdate(bindings.DmMessageReceived, message)
	}
}

// checkDbEncryptionStatus returns an error if the encryption status provided
// does not match the stored status for this database name.
func checkDbEncryptionStatus(databaseName string, encryptionStatus bool) error {
	// Pass message values to storage
	loadedEncryptionStatus, err := storage.StoreIndexedDbEncryptionStatus(
		databaseName, encryptionStatus)
	if err != nil {
		return err
	}

	// Verify encryption status does not change
	if encryptionStatus != loadedEncryptionStatus {
		return errors.New(
			"cannot load database with different encryption status")
	} else if !encryptionStatus {
		jww.WARN.Printf("IndexedDb encryption disabled!")
	}

	return nil
}
