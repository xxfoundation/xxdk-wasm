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
	"time"

	"github.com/pkg/errors"

	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/v4/bindings"
	"gitlab.com/elixxir/client/v4/channels"

	cryptoChannel "gitlab.com/elixxir/crypto/channel"
	"gitlab.com/elixxir/xxdk-wasm/storage"
	"gitlab.com/elixxir/xxdk-wasm/worker"
)

// databaseSuffix is the suffix to be appended to the name of the database.
const databaseSuffix = "_speakeasy"

// eventUpdateCallback is the [bindings.ChannelUICallback] callback function
// it has a type ([bindings.NickNameUpdate] to [bindings.MessageDeleted]
// and json data that is the callback information.
type eventUpdateCallback func(eventType int64, jsonData []byte)

// NewWASMEventModelBuilder returns an EventModelBuilder which allows
// the channel manager to define the path but the callback is the same
// across the board.
func NewWASMEventModelBuilder(wasmJsPath string,
	encryption cryptoChannel.Cipher,
	channelCbs bindings.ChannelUICallbacks) channels.EventModelBuilder {
	fn := func(path string) (channels.EventModel, error) {
		return NewWASMEventModel(path, wasmJsPath, encryption,
			channelCbs)
	}
	return fn
}

// NewWASMEventModelMessage is JSON marshalled and sent to the worker for
// [NewWASMEventModel].
type NewWASMEventModelMessage struct {
	DatabaseName   string `json:"databaseName"`
	EncryptionJSON string `json:"encryptionJSON"`
}

// NewWASMEventModel returns a [channels.EventModel] backed by a wasmModel.
// The name should be a base64 encoding of the users public key.
func NewWASMEventModel(path, wasmJsPath string, encryption cryptoChannel.Cipher,
	channelCbs bindings.ChannelUICallbacks) (
	channels.EventModel, error) {
	databaseName := path + databaseSuffix

	wm, err := worker.NewManager(wasmJsPath, "channelsIndexedDb", true)
	if err != nil {
		return nil, err
	}

	// Register handler to manage messages for the MessageReceivedCallback
	wm.RegisterCallback(MessageReceivedCallbackTag,
		messageReceivedCallbackHandler(channelCbs.EventUpdate))

	// Register handler to manage messages for the DeletedMessageCallback
	wm.RegisterCallback(DeletedMessageCallbackTag,
		deletedMessageCallbackHandler(channelCbs.EventUpdate))

	// Register handler to manage messages for the MutedUserCallback
	wm.RegisterCallback(MutedUserCallbackTag,
		mutedUserCallbackHandler(channelCbs.EventUpdate))

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
	wm.SendMessage(NewWASMEventModelTag, payload,
		func(data []byte) { dataChan <- data })

	select {
	case data := <-dataChan:
		if len(data) > 0 {
			return nil, errors.New(string(data))
		}
	case <-time.After(worker.ResponseTimeout):
		return nil, errors.Errorf("timed out after %s waiting for indexedDB "+
			"database in worker to initialize", worker.ResponseTimeout)
	}

	return &wasmModel{wm}, nil
}

// MessageReceivedCallbackMessage is JSON marshalled and received from the
// worker for the [MessageReceivedCallback] callback.
type MessageReceivedCallbackMessage struct {
	UUID      int64  `json:"uuid"`
	ChannelID []byte `json:"channelID"`
	Update    bool   `json:"update"`
}

// messageReceivedCallbackHandler returns a handler to manage messages for the
// MessageReceivedCallback.
func messageReceivedCallbackHandler(cb eventUpdateCallback) func(data []byte) {
	return func(data []byte) {
		cb(bindings.MessageReceived, data)
	}
}

// deletedMessageCallbackHandler returns a handler to manage messages for the
// DeletedMessageCallback.
func deletedMessageCallbackHandler(cb eventUpdateCallback) func(data []byte) {
	return func(data []byte) {
		cb(bindings.MessageDeleted, data)
	}
}

// mutedUserCallbackHandler returns a handler to manage messages for the
// MutedUserCallback.
func mutedUserCallbackHandler(cb eventUpdateCallback) func(data []byte) {
	return func(data []byte) {
		cb(bindings.UserMuted, data)
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
