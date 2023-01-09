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
	mChannels "gitlab.com/elixxir/xxdk-wasm/indexedDbWorker/channels"
	mDm "gitlab.com/elixxir/xxdk-wasm/indexedDbWorker/dm"
	"gitlab.com/elixxir/xxdk-wasm/worker"
	"gitlab.com/xx_network/crypto/csprng"
	"time"
)

var zeroUUID = []byte{0, 0, 0, 0, 0, 0, 0, 0}

// manager handles the event model and the message handler, which is used to
// send information between the event model and the main thread.
type manager struct {
	mh    *worker.ThreadManager
	model dm.EventModel
}

// RegisterHandlers registers all the reception handlers to manage messages from
// the main thread for the channels.EventModel.
func (m *manager) RegisterHandlers() {
	m.mh.RegisterCallback(mDm.NewWASMEventModelTag, m.newWASMEventModelHandler)
	m.mh.RegisterCallback(mDm.ReceiveTag, m.receiveHandler)
	m.mh.RegisterCallback(mDm.ReceiveTextTag, m.receiveTextHandler)
	m.mh.RegisterCallback(mDm.ReceiveReplyTag, m.receiveReplyHandler)
	m.mh.RegisterCallback(mDm.ReceiveReactionTag, m.receiveReactionHandler)
	m.mh.RegisterCallback(mDm.UpdateSentStatusTag, m.updateSentStatusHandler)
}

// newWASMEventModelHandler is the handler for NewWASMEventModel. Returns an
// empty slice on success or an error message on failure.
func (m *manager) newWASMEventModelHandler(data []byte) ([]byte, error) {
	var msg mChannels.NewWASMEventModelMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return []byte{}, errors.Errorf(
			"failed to JSON unmarshal %T from main thread: %+v", msg, err)
	}

	// Create new encryption cipher
	rng := fastRNG.NewStreamGenerator(12, 1024, csprng.NewSystemRNG)
	encryption, err := cryptoChannel.NewCipherFromJSON(
		[]byte(msg.EncryptionJSON), rng.GetStream())
	if err != nil {
		return []byte{}, errors.Errorf("failed to JSON unmarshal channel "+
			"cipher from main thread: %+v", err)
	}

	m.model, err = NewWASMEventModel(msg.Path, encryption,
		m.messageReceivedCallback, m.storeDatabaseName, m.storeEncryptionStatus)
	if err != nil {
		return []byte(err.Error()), nil
	}
	return []byte{}, nil
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
	m.mh.SendMessage(mDm.MessageReceivedCallbackTag, data)
}

// storeDatabaseName sends the database name to the main thread and waits for
// the response. This function mocks the behavior of storage.StoreIndexedDb.
//
// storeDatabaseName adheres to the storeDatabaseNameFn type.
func (m *manager) storeDatabaseName(databaseName string) error {
	// Register response handler with channel that will wait for the response
	responseChan := make(chan []byte)
	m.mh.RegisterCallback(mDm.StoreDatabaseNameTag,
		func(data []byte) ([]byte, error) {
			responseChan <- data
			return nil, nil
		})

	// Send encryption status to main thread
	m.mh.SendMessage(mDm.StoreDatabaseNameTag, []byte(databaseName))

	// Wait for response
	select {
	case response := <-responseChan:
		if len(response) > 0 {
			return errors.New(string(response))
		}
	case <-time.After(worker.ResponseTimeout):
		return errors.Errorf("[WW] Timed out after %s waiting for response "+
			"about storing the database name in local storage in the main "+
			"thread", worker.ResponseTimeout)
	}

	return nil
}

// storeEncryptionStatus sends the database name and encryption status to the
// main thread and waits for the response. If the value has not been previously
// saved, it returns the saves encryption status. This function mocks the
// behavior of storage.StoreIndexedDbEncryptionStatus.
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
	m.mh.RegisterCallback(mDm.EncryptionStatusTag,
		func(data []byte) ([]byte, error) {
			responseChan <- data
			return nil, nil
		})

	// Send encryption status to main thread
	m.mh.SendMessage(mDm.EncryptionStatusTag, data)

	// Wait for response
	var response mChannels.EncryptionStatusReply
	select {
	case responseData := <-responseChan:
		if err = json.Unmarshal(responseData, &response); err != nil {
			return false, err
		}
	case <-time.After(worker.ResponseTimeout):
		return false, errors.Errorf("timed out after %s waiting for "+
			"response about the database encryption status from local "+
			"storage in the main thread", worker.ResponseTimeout)
	}

	// If the response contain an error, return it
	if response.Error != "" {
		return false, errors.New(response.Error)
	}

	// Return the encryption status
	return response.EncryptionStatus, nil
}

// receiveHandler is the handler for wasmModel.Receive. Returns a UUID of 0 on
// error or the JSON marshalled UUID (uint64) on success.
func (m *manager) receiveHandler(data []byte) ([]byte, error) {
	var msg mDm.TransferMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return zeroUUID, errors.Errorf(
			"failed to JSON unmarshal %T from main thread: %+v", msg, err)
	}

	uuid := m.model.Receive(
		msg.MessageID, msg.Nickname, msg.Text, msg.PubKey, msg.DmToken,
		msg.Codeset, msg.Timestamp, msg.Round, msg.MType, msg.Status)

	uuidData, err := json.Marshal(uuid)
	if err != nil {
		return zeroUUID, errors.Errorf("failed to JSON marshal UUID : %+v", err)
	}
	return uuidData, nil
}

// receiveTextHandler is the handler for wasmModel.ReceiveText. Returns a UUID
// of 0 on error or the JSON marshalled UUID (uint64) on success.
func (m *manager) receiveTextHandler(data []byte) ([]byte, error) {
	var msg mDm.TransferMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return []byte{}, errors.Errorf(
			"failed to JSON unmarshal %T from main thread: %+v", msg, err)
	}

	uuid := m.model.ReceiveText(
		msg.MessageID, msg.Nickname, string(msg.Text), msg.PubKey, msg.DmToken,
		msg.Codeset, msg.Timestamp, msg.Round, msg.Status)

	uuidData, err := json.Marshal(uuid)
	if err != nil {
		return []byte{}, errors.Errorf("failed to JSON marshal UUID : %+v", err)
	}

	return uuidData, nil
}

// receiveReplyHandler is the handler for wasmModel.ReceiveReply. Returns a UUID
// of 0 on error or the JSON marshalled UUID (uint64) on success.
func (m *manager) receiveReplyHandler(data []byte) ([]byte, error) {
	var msg mDm.TransferMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return zeroUUID, errors.Errorf(
			"failed to JSON unmarshal %T from main thread: %+v", msg, err)
	}

	uuid := m.model.ReceiveReply(msg.MessageID, msg.ReactionTo, msg.Nickname,
		string(msg.Text), msg.PubKey, msg.DmToken, msg.Codeset, msg.Timestamp,
		msg.Round, msg.Status)

	uuidData, err := json.Marshal(uuid)
	if err != nil {
		return zeroUUID, errors.Errorf("failed to JSON marshal UUID : %+v", err)
	}

	return uuidData, nil
}

// receiveReactionHandler is the handler for wasmModel.ReceiveReaction. Returns
// a UUID of 0 on error or the JSON marshalled UUID (uint64) on success.
func (m *manager) receiveReactionHandler(data []byte) ([]byte, error) {
	var msg mDm.TransferMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return zeroUUID, errors.Errorf(
			"failed to JSON unmarshal %T from main thread: %+v", msg, err)
	}

	uuid := m.model.ReceiveReaction(msg.MessageID, msg.ReactionTo, msg.Nickname,
		string(msg.Text), msg.PubKey, msg.DmToken, msg.Codeset, msg.Timestamp,
		msg.Round, msg.Status)

	uuidData, err := json.Marshal(uuid)
	if err != nil {
		return zeroUUID, errors.Errorf("failed to JSON marshal UUID : %+v", err)
	}

	return uuidData, nil
}

// updateSentStatusHandler is the handler for wasmModel.UpdateSentStatus. Always
// returns nil; meaning, no response is supplied (or expected).
func (m *manager) updateSentStatusHandler(data []byte) ([]byte, error) {
	var msg mDm.TransferMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, errors.Errorf(
			"failed to JSON unmarshal %T from main thread: %+v", msg, err)
	}

	m.model.UpdateSentStatus(
		msg.UUID, msg.MessageID, msg.Timestamp, msg.Round, msg.Status)

	return nil, nil
}
