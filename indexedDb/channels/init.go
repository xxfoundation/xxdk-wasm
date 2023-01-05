////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package channels

import (
	"encoding/json"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/v4/channels"
	cryptoChannel "gitlab.com/elixxir/crypto/channel"
	"gitlab.com/elixxir/xxdk-wasm/indexedDb"
	"gitlab.com/elixxir/xxdk-wasm/storage"
	"gitlab.com/xx_network/primitives/id"
	"time"
)

// WorkerJavascriptFileURL is the URL of the script the worker will execute to
// launch the worker WASM binary. It must obey the same-origin policy.
const WorkerJavascriptFileURL = "/integrations/assets/indexedDbWorker.js"

// MessageReceivedCallback is called any time a message is received or updated.
//
// update is true if the row is old and was edited.
type MessageReceivedCallback func(uuid uint64, channelID *id.ID, update bool)

// NewWASMEventModelBuilder returns an EventModelBuilder which allows
// the channel manager to define the path but the callback is the same
// across the board.
func NewWASMEventModelBuilder(encryption cryptoChannel.Cipher,
	cb MessageReceivedCallback) channels.EventModelBuilder {
	fn := func(path string) (channels.EventModel, error) {
		return NewWASMEventModel(path, encryption, cb)
	}
	return fn
}

// NewWASMEventModelMessage is JSON marshalled and sent to the worker for
// [NewWASMEventModel].
type NewWASMEventModelMessage struct {
	Path           string `json:"path"`
	EncryptionJSON string `json:"encryptionJSON"`
}

// NewWASMEventModel returns a [channels.EventModel] backed by a wasmModel.
// The name should be a base64 encoding of the users public key.
func NewWASMEventModel(path string, encryption cryptoChannel.Cipher,
	cb MessageReceivedCallback) (channels.EventModel, error) {

	// TODO: bring in URL and name from caller
	wh, err := indexedDb.NewWorkerHandler(
		WorkerJavascriptFileURL, "channelsIndexedDb")
	if err != nil {
		return nil, err
	}

	// Register handler to manage messages for the MessageReceivedCallback
	wh.RegisterHandler(indexedDb.GetMessageTag, indexedDb.InitID, false,
		messageReceivedCallbackHandler(cb))

	// Register handler to manage checking encryption status from local storage
	wh.RegisterHandler(indexedDb.EncryptionStatusTag, indexedDb.InitID, false,
		checkDbEncryptionStatusHandler(wh))

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
	wh.SendMessage(indexedDb.NewWASMEventModelTag, payload, func(data []byte) {
		errChan <- string(data)
	})

	select {
	case workerErr := <-errChan:
		if workerErr != "" {
			return nil, errors.New(workerErr)
		}
	case <-time.After(indexedDb.ResponseTimeout):
		return nil, errors.Errorf("timed out after %s waiting for indexedDB "+
			"database in worker to intialize", indexedDb.ResponseTimeout)
	}

	return &wasmModel{wh}, nil
}

// MessageReceivedCallbackMessage is JSON marshalled and received from the
// worker for the [MessageReceivedCallback] callback.
type MessageReceivedCallbackMessage struct {
	UUID      uint64 `json:"uuid"`
	ChannelID *id.ID `json:"channelID"`
	Update    bool   `json:"update"`
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
		cb(msg.UUID, msg.ChannelID, msg.Update)
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
func checkDbEncryptionStatusHandler(wh *indexedDb.WorkerHandler) func(data []byte) {
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

		wh.SendMessage(indexedDb.EncryptionStatusTag, statusData, nil)
	}
}
