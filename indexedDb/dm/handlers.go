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
	"encoding/json"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/v4/dm"
	cryptoChannel "gitlab.com/elixxir/crypto/channel"
	"gitlab.com/elixxir/crypto/fastRNG"
	"gitlab.com/elixxir/xxdk-wasm/indexedDb"
	"gitlab.com/elixxir/xxdk-wasm/indexedDbWorker"
	mChannels "gitlab.com/elixxir/xxdk-wasm/indexedDbWorker/channels"
	mDm "gitlab.com/elixxir/xxdk-wasm/indexedDbWorker/dm"
	"gitlab.com/xx_network/crypto/csprng"
	"time"
)

// manager handles the event model and the message handler, which is used to
// send information between the event model and the main thread.
type manager struct {
	mh    *indexedDb.MessageHandler
	model dm.EventModel
}

// RegisterHandlers registers all the reception handlers to manage messages from
// the main thread for the channels.EventModel.
func (m *manager) RegisterHandlers() {

	m.mh.RegisterHandler(indexedDbWorker.NewWASMEventModelTag, m.newWASMEventModelHandler)
	m.mh.RegisterHandler(indexedDbWorker.ReceiveTag, m.receiveHandler)
	m.mh.RegisterHandler(indexedDbWorker.ReceiveTextTag, m.receiveTextHandler)
	m.mh.RegisterHandler(indexedDbWorker.ReceiveReplyTag, m.receiveReplyHandler)
	m.mh.RegisterHandler(indexedDbWorker.ReceiveReactionTag, m.receiveReactionHandler)
	m.mh.RegisterHandler(indexedDbWorker.UpdateSentStatusTag, m.updateSentStatusHandler)
}

// newWASMEventModelHandler is the handler for NewWASMEventModel.
func (m *manager) newWASMEventModelHandler(data []byte) []byte {
	var msg mChannels.NewWASMEventModelMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		jww.ERROR.Printf("Could not JSON unmarshal NewWASMEventModelMessage "+
			"from NewWASMEventModel in main thread: %+v", err)
		return []byte{}
	}

	// Create new encryption cipher
	rng := fastRNG.NewStreamGenerator(12, 1024, csprng.NewSystemRNG)
	encryption, err := cryptoChannel.NewCipherFromJSON(
		[]byte(msg.EncryptionJSON), rng.GetStream())
	if err != nil {
		jww.ERROR.Printf("Could not JSON unmarshal channel cipher from "+
			"main thread: %+v", err)
		return []byte{}
	}

	m.model, err = NewWASMEventModel(msg.Path, encryption,
		m.messageReceivedCallback, m.storeEncryptionStatus)
	if err != nil {
		return []byte(err.Error())
	}
	return []byte{}
}

// messageReceivedCallback sends calls to the MessageReceivedCallback in the
// main thread.
//
// messageReceivedCallback adhere to the MessageReceivedCallback type.
func (m *manager) messageReceivedCallback(
	uuid uint64, pubKey ed25519.PublicKey, update bool) {
	// Package parameters for sending
	msg := &mDm.MessageReceivedCallbackMessage{
		UUID:   uuid,
		PubKey: pubKey,
		Update: update,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		jww.ERROR.Printf(
			"Could not JSON marshal MessageReceivedCallbackMessage: %+v", err)
		return
	}

	// Send it to the main thread
	m.mh.SendResponse(indexedDbWorker.GetMessageTag, indexedDbWorker.InitID, data)
}

// storeEncryptionStatus augments the functionality of
// storage.StoreIndexedDbEncryptionStatus. It takes the database name and
// encryption status
//
// storeEncryptionStatus adheres to the storeEncryptionStatusFn type.
func (m *manager) storeEncryptionStatus(
	databaseName string, encryption bool) (bool, error) {
	// Package parameters for sending
	msg := &mChannels.EncryptionStatusMessage{
		DatabaseName:     databaseName,
		EncryptionStatus: encryption,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return false, err
	}

	// Register response handler with channel that will wait for the response
	responseChan := make(chan []byte)
	m.mh.RegisterHandler(indexedDbWorker.EncryptionStatusTag,
		func(data []byte) []byte {
			responseChan <- data
			return nil
		})

	// Send encryption status to main thread
	m.mh.SendResponse(indexedDbWorker.EncryptionStatusTag, indexedDbWorker.InitID, data)

	// Wait for response
	var response mChannels.EncryptionStatusReply
	select {
	case responseData := <-responseChan:
		if err = json.Unmarshal(responseData, &response); err != nil {
			return false, err
		}
	case <-time.After(indexedDbWorker.ResponseTimeout):
		return false, errors.Errorf("timed out after %s waiting for "+
			"response about the database encryption status from local "+
			"storage in the main thread", indexedDbWorker.ResponseTimeout)
	}

	// If the response contain an error, return it
	if response.Error != "" {
		return false, errors.New(response.Error)
	}

	// Return the encryption status
	return response.EncryptionStatus, nil
}

// receiveHandler is the handler for wasmModel.Receive.
func (m *manager) receiveHandler(data []byte) []byte {
	var msg mDm.TransferMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		jww.ERROR.Printf("Could not JSON unmarshal dm.TransferMessage from "+
			"Receive in main thread: %+v", err)
		return nil
	}

	uuid := m.model.Receive(
		msg.MessageID, msg.Nickname, msg.Text, msg.PubKey, msg.DmToken,
		msg.Codeset, msg.Timestamp, msg.Round, msg.MType, msg.Status)

	uuidData, err := json.Marshal(uuid)
	if err != nil {
		jww.ERROR.Printf(
			"Could not JSON marshal UUID from Receive: %+v", err)
		return nil
	}
	return uuidData
}

// receiveTextHandler is the handler for wasmModel.ReceiveText.
func (m *manager) receiveTextHandler(data []byte) []byte {
	var msg mDm.TransferMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		jww.ERROR.Printf("Could not JSON unmarshal dm.TransferMessage from "+
			"ReceiveText in main thread: %+v", err)
		return nil
	}

	uuid := m.model.ReceiveText(
		msg.MessageID, msg.Nickname, string(msg.Text), msg.PubKey, msg.DmToken,
		msg.Codeset, msg.Timestamp, msg.Round, msg.Status)

	uuidData, err := json.Marshal(uuid)
	if err != nil {
		jww.ERROR.Printf(
			"Could not JSON marshal UUID from ReceiveText: %+v", err)
		return nil
	}
	return uuidData
}

// receiveReplyHandler is the handler for wasmModel.ReceiveReply.
func (m *manager) receiveReplyHandler(data []byte) []byte {
	var msg mDm.TransferMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		jww.ERROR.Printf("Could not JSON unmarshal dm.TransferMessage from "+
			"ReceiveReply in main thread: %+v", err)
		return nil
	}

	uuid := m.model.ReceiveReply(msg.MessageID, msg.ReactionTo, msg.Nickname,
		string(msg.Text), msg.PubKey, msg.DmToken, msg.Codeset, msg.Timestamp,
		msg.Round, msg.Status)

	uuidData, err := json.Marshal(uuid)
	if err != nil {
		jww.ERROR.Printf(
			"Could not JSON marshal UUID from ReceiveReply: %+v", err)
		return nil
	}
	return uuidData
}

// receiveReactionHandler is the handler for wasmModel.ReceiveReaction.
func (m *manager) receiveReactionHandler(data []byte) []byte {
	var msg mDm.TransferMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		jww.ERROR.Printf("Could not JSON unmarshal dm.TransferMessage from "+
			"ReceiveReaction in main thread: %+v", err)
		return nil
	}

	uuid := m.model.ReceiveReaction(msg.MessageID, msg.ReactionTo, msg.Nickname,
		string(msg.Text), msg.PubKey, msg.DmToken, msg.Codeset, msg.Timestamp,
		msg.Round, msg.Status)

	uuidData, err := json.Marshal(uuid)
	if err != nil {
		jww.ERROR.Printf(
			"Could not JSON marshal UUID from ReceiveReaction: %+v", err)
		return nil
	}
	return uuidData
}

// updateSentStatusHandler is the handler for wasmModel.UpdateSentStatus.
func (m *manager) updateSentStatusHandler(data []byte) []byte {
	var msg mDm.TransferMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		jww.ERROR.Printf("Could not JSON unmarshal dm.TransferMessage from "+
			"UpdateSentStatus in main thread: %+v", err)
		return nil
	}

	m.model.UpdateSentStatus(
		msg.UUID, msg.MessageID, msg.Timestamp, msg.Round, msg.Status)
	return nil
}
