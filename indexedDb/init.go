////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package indexedDb

import (
	"encoding/json"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/v4/channels"
	cryptoChannel "gitlab.com/elixxir/crypto/channel"
	"gitlab.com/xx_network/primitives/id"
	"time"
)

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

// MessageReceivedCallbackMessage is JSON marshalled and received from the
// worker for the [MessageReceivedCallback] callback.
type MessageReceivedCallbackMessage struct {
	UUID      uint64 `json:"uuid"`
	ChannelID *id.ID `json:"channelID"`
	Update    bool   `json:"update"`
}

// NewWASMEventModel returns a [channels.EventModel] backed by a wasmModel.
// The name should be a base64 encoding of the users public key.
func NewWASMEventModel(path string, encryption cryptoChannel.Cipher,
	cb MessageReceivedCallback) (channels.EventModel, error) {

	wh, err := newWorkerHandler("indexedDbWorker.js", "indexedDbWorker")
	if err != nil {
		return nil, err
	}

	// Register handler to manage messages for the MessageReceivedCallback
	wh.registerHandler(GetMessageTag, InitialID, false, func(data []byte) {
		var msg MessageReceivedCallbackMessage
		err2 := json.Unmarshal(data, &msg)
		if err2 != nil {
			jww.ERROR.Printf("Failed to JSON unmarshal "+
				"MessageReceivedCallback message from worker: %+v", err2)
		}
		cb(msg.UUID, msg.ChannelID, msg.Update)
	})

	encryptionJSON, err := json.Marshal(encryption)
	if err != nil {
		return nil, errors.Errorf("could not JSON marshal channel.Cipher: %+v", err)
	}

	message := NewWASMEventModelMessage{
		Path:           path,
		EncryptionJSON: string(encryptionJSON),
	}

	payload, err := json.Marshal(message)
	if err != nil {
		return nil, errors.Errorf(
			"could not JSON marshal payload for NewWASMEventModel: %+v", err)
	}

	errChan := make(chan string)
	wh.sendMessage(NewWASMEventModelTag, payload, func(data []byte) {
		errChan <- string(data)
	})

	select {
	case workerErr := <-errChan:
		if workerErr != "" {
			return nil, errors.New(workerErr)
		}
	case <-time.After(databaseInitialConnectionTimeout):
		return nil, errors.Errorf("timed out after %s waiting for indexedDB "+
			"database in worker to intialize", databaseInitialConnectionTimeout)
	}

	return &wasmModel{wh}, nil
}
