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
	"time"

	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"

	"gitlab.com/elixxir/client/v4/dm"
	cryptoChannel "gitlab.com/elixxir/crypto/channel"
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
func NewWASMEventModel(path, wasmJsPath string, encryption cryptoChannel.Cipher,
	cb MessageReceivedCallback) (dm.EventModel, error) {
	databaseName := path + databaseSuffix

	wh, err := worker.NewManager(wasmJsPath, "dmIndexedDb", true)
	if err != nil {
		return nil, err
	}

	// Register handler to manage messages for the MessageReceivedCallback
	wh.RegisterCallback(
		MessageReceivedCallbackTag, messageReceivedCallbackHandler(cb))

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

	dataChan := make(chan []byte)
	wh.SendMessage(NewWASMEventModelTag, payload,
		func(data []byte) { dataChan <- data })

	select {
	case data := <-dataChan:
		if data != nil {
			return nil, errors.New(string(data))
		}
	case <-time.After(worker.ResponseTimeout):
		return nil, errors.Errorf("timed out after %s waiting for indexedDB "+
			"database in worker to initialize", worker.ResponseTimeout)
	}

	return &wasmModel{wh}, nil
}

// MessageReceivedCallbackMessage is JSON marshalled and received from the
// worker for the [MessageReceivedCallback] callback.
type MessageReceivedCallbackMessage struct {
	UUID               uint64            `json:"uuid"`
	PubKey             ed25519.PublicKey `json:"pubKey"`
	MessageUpdate      bool              `json:"message_update"`
	ConversationUpdate bool              `json:"conversation_update"`
}

// messageReceivedCallbackHandler returns a handler to manage messages for the
// MessageReceivedCallback.
func messageReceivedCallbackHandler(cb MessageReceivedCallback) func(data []byte) {
	return func(data []byte) {
		var msg MessageReceivedCallbackMessage
		err := json.Unmarshal(data, &msg)
		if err != nil {
			jww.ERROR.Printf("Failed to JSON unmarshal "+
				"MessageReceivedCallback message from worker: %+v", err)
			return
		}
		cb(msg.UUID, msg.PubKey, msg.MessageUpdate, msg.ConversationUpdate)
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
