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
	"encoding/json"
	"gitlab.com/elixxir/xxdk-wasm/indexedDbWorker"
	"time"

	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/v4/dm"
	cryptoChannel "gitlab.com/elixxir/crypto/channel"
	"gitlab.com/elixxir/xxdk-wasm/storage"
)

// WorkerJavascriptFileURL is the URL of the script the worker will execute to
// launch the worker WASM binary. It must obey the same-origin policy.
const WorkerJavascriptFileURL = "/integrations/assets/indexedDbWorker.js"

// MessageReceivedCallback is called any time a message is received or updated.
//
// update is true if the row is old and was edited.
type MessageReceivedCallback func(
	uuid uint64, pubKey ed25519.PublicKey, update bool)

// NewWASMEventModelMessage is JSON marshalled and sent to the worker for
// [NewWASMEventModel].
type NewWASMEventModelMessage struct {
	Path           string `json:"path"`
	EncryptionJSON string `json:"encryptionJSON"`
}

// NewWASMEventModel returns a [channels.EventModel] backed by a wasmModel.
// The name should be a base64 encoding of the users public key.
func NewWASMEventModel(path string, encryption cryptoChannel.Cipher,
	cb MessageReceivedCallback) (dm.EventModel, error) {

	// TODO: bring in URL and name from caller
	wh, err := indexedDbWorker.NewWorkerHandler(
		WorkerJavascriptFileURL, "dmIndexedDb")
	if err != nil {
		return nil, err
	}

	// Register handler to manage messages for the MessageReceivedCallback
	wh.RegisterHandler(indexedDbWorker.GetMessageTag, indexedDbWorker.InitID,
		false, messageReceivedCallbackHandler(cb))

	// Register handler to manage checking encryption status from local storage
	wh.RegisterHandler(indexedDbWorker.EncryptionStatusTag,
		indexedDbWorker.InitID, false, checkDbEncryptionStatusHandler(wh))

	// Register handler to manage the storage of the database name
	wh.RegisterHandler(indexedDbWorker.StoreDatabaseNameTag,
		indexedDbWorker.InitID, false, storeDatabaseNameHandler(wh))

	encryptionJSON, err := json.Marshal(encryption)
	if err != nil {
		return nil, err
	}

	message := NewWASMEventModelMessage{
		Path:           path,
		EncryptionJSON: string(encryptionJSON),
	}

	payload, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}

	errChan := make(chan string)
	wh.SendMessage(indexedDbWorker.NewWASMEventModelTag, payload,
		func(data []byte) { errChan <- string(data) })

	select {
	case workerErr := <-errChan:
		if workerErr != "" {
			return nil, errors.New(workerErr)
		}
	case <-time.After(indexedDbWorker.ResponseTimeout):
		return nil, errors.Errorf("timed out after %s waiting for indexedDB "+
			"database in worker to intialize", indexedDbWorker.ResponseTimeout)
	}

	return &wasmModel{wh}, nil
}

// MessageReceivedCallbackMessage is JSON marshalled and received from the
// worker for the [MessageReceivedCallback] callback.
type MessageReceivedCallbackMessage struct {
	UUID   uint64            `json:"uuid"`
	PubKey ed25519.PublicKey `json:"pubKey"`
	Update bool              `json:"update"`
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
		cb(msg.UUID, msg.PubKey, msg.Update)
	}
}

// EncryptionStatusMessage is JSON marshalled and received from the worker when
// the database checks if it is encrypted.
type EncryptionStatusMessage struct {
	DatabaseName     string `json:"databaseName"`
	EncryptionStatus bool   `json:"encryptionStatus"`
}

// EncryptionStatusReply is JSON marshalled and sent to the worker is response
// to the [EncryptionStatusMessage].
type EncryptionStatusReply struct {
	EncryptionStatus bool   `json:"encryptionStatus"`
	Error            string `json:"error"`
}

// checkDbEncryptionStatusHandler returns a handler to manage checking
// encryption status from local storage.
func checkDbEncryptionStatusHandler(
	wh *indexedDbWorker.WorkerHandler) func(data []byte) {
	return func(data []byte) {
		// Unmarshal received message
		var msg EncryptionStatusMessage
		err := json.Unmarshal(data, &msg)
		if err != nil {
			jww.ERROR.Printf("Failed to JSON unmarshal "+
				"EncryptionStatusMessage message from worker: %+v", err)
			return
		}

		// Pass message values to storage
		loadedEncryptionStatus, err := storage.StoreIndexedDbEncryptionStatus(
			msg.DatabaseName, msg.EncryptionStatus)
		var reply EncryptionStatusReply
		if err != nil {
			reply.Error = err.Error()
		} else {
			reply.EncryptionStatus = loadedEncryptionStatus
		}

		// Return response
		statusData, err := json.Marshal(reply)
		if err != nil {
			jww.ERROR.Printf(
				"Failed to JSON marshal EncryptionStatusReply: %+v", err)
			return
		}

		wh.SendMessage(indexedDbWorker.EncryptionStatusTag, statusData, nil)
	}
}

// storeDatabaseNameHandler returns a handler that stores the database name to
// storage when it is received from the worker.
func storeDatabaseNameHandler(
	wh *indexedDbWorker.WorkerHandler) func(data []byte) {
	return func(data []byte) {
		var returnData []byte

		// Get the database name and save it to storage
		if err := storage.StoreIndexedDb(string(data)); err != nil {
			returnData = []byte(err.Error())
		}

		wh.SendMessage(indexedDbWorker.StoreDatabaseNameTag, returnData, nil)
	}
}
