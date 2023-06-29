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
	"gitlab.com/elixxir/client/v4/bindings"
	"gitlab.com/elixxir/client/v4/channels"
	idbCrypto "gitlab.com/elixxir/crypto/indexedDb"
	"gitlab.com/elixxir/xxdk-wasm/logging"
	"gitlab.com/elixxir/xxdk-wasm/storage"
	"gitlab.com/elixxir/xxdk-wasm/worker"
)

// databaseSuffix is the suffix to be appended to the name of the database.
const databaseSuffix = "_speakeasy"

// eventUpdateCallback is the [bindings.ChannelUICallback] callback function
// it has a type [bindings.NickNameUpdate] to [bindings.MessageDeleted]
// and json data that is the callback information.
type eventUpdateCallback func(eventType int64, jsonData []byte)

// NewWASMEventModelBuilder returns an EventModelBuilder which allows
// the channel manager to define the path but the callback is the same
// across the board.
func NewWASMEventModelBuilder(wasmJsPath string, encryption idbCrypto.Cipher,
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
func NewWASMEventModel(path, wasmJsPath string, encryption idbCrypto.Cipher,
	channelCbs bindings.ChannelUICallbacks) (channels.EventModel, error) {
	databaseName := path + databaseSuffix

	wm, err := worker.NewManager(wasmJsPath, "channelsIndexedDb", true)
	if err != nil {
		return nil, err
	}

	// Register handler to manage messages for the EventUpdate
	wm.RegisterCallback(EventUpdateCallbackTag,
		messageReceivedCallbackHandler(channelCbs.EventUpdate))

	// Create MessageChannel between worker and logger so that the worker logs
	// are saved
	err = worker.CreateMessageChannel(logging.GetLogger().Worker(), wm,
		"channelsIndexedDbLogger", worker.LoggerTag)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create message channel "+
			"between channel indexedDb worker and logger")
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

	response, err := wm.SendMessage(NewWASMEventModelTag, payload)
	if err != nil {
		return nil, errors.Wrapf(err,
			"failed to send message %q", NewWASMEventModelTag)
	} else if len(response) > 0 {
		return nil, errors.New(string(response))
	}

	return &wasmModel{wm}, nil
}

// EventUpdateCallbackMessage is JSON marshalled and received from the worker
// for the [EventUpdate] callback.
type EventUpdateCallbackMessage struct {
	EventType int64  `json:"eventType"`
	JsonData  []byte `json:"jsonData"`
}

// messageReceivedCallbackHandler returns a handler to manage messages for the
// MessageReceivedCallback.
func messageReceivedCallbackHandler(cb eventUpdateCallback) worker.ReceiverCallback {
	return func(message []byte, _ func([]byte)) {
		var msg EventUpdateCallbackMessage
		err := json.Unmarshal(message, &msg)
		if err != nil {
			jww.ERROR.Printf(
				"Failed to JSON unmarshal %T from worker: %+v", msg, err)
			return
		}

		cb(msg.EventType, msg.JsonData)
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
