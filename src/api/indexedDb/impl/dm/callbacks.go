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
	"time"

	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/v4/dm"
	cryptoChannel "gitlab.com/elixxir/crypto/channel"
	"gitlab.com/elixxir/crypto/fastRNG"
	wDm "gitlab.com/elixxir/xxdk-wasm/src/api/indexedDb/worker/dm"
	"gitlab.com/elixxir/xxdk-wasm/src/api/worker"
	"gitlab.com/xx_network/crypto/csprng"
)

var zeroUUID = []byte{0, 0, 0, 0, 0, 0, 0, 0}

// manager handles the event model and the message callbacks, which is used to
// send information between the event model and the main thread.
type manager struct {
	mh    *worker.ThreadManager
	model dm.EventModel
}

// registerCallbacks registers all the reception callbacks to manage messages
// from the main thread for the channels.EventModel.
func (m *manager) registerCallbacks() {
	m.mh.RegisterCallback(wDm.NewWASMEventModelTag, m.newWASMEventModelCB)
	m.mh.RegisterCallback(wDm.ReceiveTag, m.receiveCB)
	m.mh.RegisterCallback(wDm.ReceiveTextTag, m.receiveTextCB)
	m.mh.RegisterCallback(wDm.ReceiveReplyTag, m.receiveReplyCB)
	m.mh.RegisterCallback(wDm.ReceiveReactionTag, m.receiveReactionCB)
	m.mh.RegisterCallback(wDm.UpdateSentStatusTag, m.updateSentStatusCB)

	m.mh.RegisterCallback(wDm.BlockSenderTag, m.blockSenderCB)
	m.mh.RegisterCallback(wDm.UnblockSenderTag, m.unblockSenderCB)
	m.mh.RegisterCallback(wDm.GetConversationTag, m.getConversationCB)
	m.mh.RegisterCallback(wDm.GetConversationsTag, m.getConversationsCB)
}

// newWASMEventModelCB is the callback for NewWASMEventModel. Returns an empty
// slice on success or an error message on failure.
func (m *manager) newWASMEventModelCB(data []byte) ([]byte, error) {
	var msg wDm.NewWASMEventModelMessage
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
func (m *manager) messageReceivedCallback(uuid uint64, pubKey ed25519.PublicKey,
	messageUpdate, conversationUpdate bool) {
	// Package parameters for sending
	msg := &wDm.MessageReceivedCallbackMessage{
		UUID:               uuid,
		PubKey:             pubKey,
		MessageUpdate:      messageUpdate,
		ConversationUpdate: conversationUpdate,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		jww.ERROR.Printf(
			"Could not JSON marshal MessageReceivedCallbackMessage: %+v", err)
		return
	}

	// Send it to the main thread
	m.mh.SendMessage(wDm.MessageReceivedCallbackTag, data)
}

// storeDatabaseName sends the database name to the main thread and waits for
// the response. This function mocks the behavior of storage.StoreIndexedDb.
//
// storeDatabaseName adheres to the storeDatabaseNameFn type.
func (m *manager) storeDatabaseName(databaseName string) error {
	// Register response callback with channel that will wait for the response
	responseChan := make(chan []byte)
	m.mh.RegisterCallback(wDm.StoreDatabaseNameTag,
		func(data []byte) ([]byte, error) {
			responseChan <- data
			return nil, nil
		})

	// Send encryption status to main thread
	m.mh.SendMessage(wDm.StoreDatabaseNameTag, []byte(databaseName))

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
	msg := &wDm.EncryptionStatusMessage{
		DatabaseName:     databaseName,
		EncryptionStatus: encryption,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return false, err
	}

	// Register response callback with channel that will wait for the response
	responseChan := make(chan []byte)
	m.mh.RegisterCallback(wDm.EncryptionStatusTag,
		func(data []byte) ([]byte, error) {
			responseChan <- data
			return nil, nil
		})

	// Send encryption status to main thread
	m.mh.SendMessage(wDm.EncryptionStatusTag, data)

	// Wait for response
	var response wDm.EncryptionStatusReply
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

// receiveCB is the callback for wasmModel.Receive. Returns a UUID of 0 on error
// or the JSON marshalled UUID (uint64) on success.
func (m *manager) receiveCB(data []byte) ([]byte, error) {
	var msg wDm.TransferMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return zeroUUID, errors.Errorf(
			"failed to JSON unmarshal %T from main thread: %+v", msg, err)
	}

	uuid := m.model.Receive(
		msg.MessageID, msg.Nickname, msg.Text, msg.PartnerKey, msg.SenderKey, msg.DmToken,
		msg.Codeset, msg.Timestamp, msg.Round, msg.MType, msg.Status)

	uuidData, err := json.Marshal(uuid)
	if err != nil {
		return zeroUUID, errors.Errorf("failed to JSON marshal UUID : %+v", err)
	}
	return uuidData, nil
}

// receiveTextCB is the callback for wasmModel.ReceiveText. Returns a UUID of 0
// on error or the JSON marshalled UUID (uint64) on success.
func (m *manager) receiveTextCB(data []byte) ([]byte, error) {
	var msg wDm.TransferMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return []byte{}, errors.Errorf(
			"failed to JSON unmarshal %T from main thread: %+v", msg, err)
	}

	uuid := m.model.ReceiveText(
		msg.MessageID, msg.Nickname, string(msg.Text), msg.PartnerKey, msg.SenderKey, msg.DmToken,
		msg.Codeset, msg.Timestamp, msg.Round, msg.Status)

	uuidData, err := json.Marshal(uuid)
	if err != nil {
		return []byte{}, errors.Errorf("failed to JSON marshal UUID : %+v", err)
	}

	return uuidData, nil
}

// receiveReplyCB is the callback for wasmModel.ReceiveReply. Returns a UUID of
// 0 on error or the JSON marshalled UUID (uint64) on success.
func (m *manager) receiveReplyCB(data []byte) ([]byte, error) {
	var msg wDm.TransferMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return zeroUUID, errors.Errorf(
			"failed to JSON unmarshal %T from main thread: %+v", msg, err)
	}

	uuid := m.model.ReceiveReply(msg.MessageID, msg.ReactionTo, msg.Nickname,
		string(msg.Text), msg.PartnerKey, msg.SenderKey, msg.DmToken, msg.Codeset, msg.Timestamp,
		msg.Round, msg.Status)

	uuidData, err := json.Marshal(uuid)
	if err != nil {
		return zeroUUID, errors.Errorf("failed to JSON marshal UUID : %+v", err)
	}

	return uuidData, nil
}

// receiveReactionCB is the callback for wasmModel.ReceiveReaction. Returns a
// UUID of 0 on error or the JSON marshalled UUID (uint64) on success.
func (m *manager) receiveReactionCB(data []byte) ([]byte, error) {
	var msg wDm.TransferMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return zeroUUID, errors.Errorf(
			"failed to JSON unmarshal %T from main thread: %+v", msg, err)
	}

	uuid := m.model.ReceiveReaction(msg.MessageID, msg.ReactionTo, msg.Nickname,
		string(msg.Text), msg.PartnerKey, msg.SenderKey, msg.DmToken, msg.Codeset, msg.Timestamp,
		msg.Round, msg.Status)

	uuidData, err := json.Marshal(uuid)
	if err != nil {
		return zeroUUID, errors.Errorf("failed to JSON marshal UUID : %+v", err)
	}

	return uuidData, nil
}

// updateSentStatusCB is the callback for wasmModel.UpdateSentStatus. Always
// returns nil; meaning, no response is supplied (or expected).
func (m *manager) updateSentStatusCB(data []byte) ([]byte, error) {
	var msg wDm.TransferMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, errors.Errorf(
			"failed to JSON unmarshal %T from main thread: %+v", msg, err)
	}

	m.model.UpdateSentStatus(
		msg.UUID, msg.MessageID, msg.Timestamp, msg.Round, msg.Status)

	return nil, nil
}

// blockSenderCB is the callback for wasmModel.BlockSender. Always
// returns nil; meaning, no response is supplied (or expected).
func (m *manager) blockSenderCB(data []byte) ([]byte, error) {
	m.model.BlockSender(data)
	return nil, nil
}

// unblockSenderCB is the callback for wasmModel.UnblockSender. Always
// returns nil; meaning, no response is supplied (or expected).
func (m *manager) unblockSenderCB(data []byte) ([]byte, error) {
	m.model.UnblockSender(data)
	return nil, nil
}

// getConversationCB is the callback for wasmModel.GetConversation.
// Returns nil on error or the JSON marshalled Conversation on success.
func (m *manager) getConversationCB(data []byte) ([]byte, error) {
	result := m.model.GetConversation(data)
	return json.Marshal(result)
}

// getConversationsCB is the callback for wasmModel.GetConversations.
// Returns nil on error or the JSON marshalled list of Conversation on success.
func (m *manager) getConversationsCB(_ []byte) ([]byte, error) {
	result := m.model.GetConversations()
	return json.Marshal(result)
}
