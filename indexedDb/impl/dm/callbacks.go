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
	"gitlab.com/elixxir/crypto/fastRNG"
	idbCrypto "gitlab.com/elixxir/crypto/indexedDb"
	"gitlab.com/elixxir/wasm-utils/exception"
	wDm "gitlab.com/elixxir/xxdk-wasm/indexedDb/worker/dm"
	"gitlab.com/elixxir/xxdk-wasm/worker"
	"gitlab.com/xx_network/crypto/csprng"
)

var zeroUUID = []byte{0, 0, 0, 0, 0, 0, 0, 0}

// manager handles the event model and the message callbacks, which is used to
// send information between the event model and the main thread.
type manager struct {
	wtm   *worker.ThreadManager
	model dm.EventModel
}

// registerCallbacks registers all the reception callbacks to manage messages
// from the main thread for the channels.EventModel.
func (m *manager) registerCallbacks() {
	m.wtm.RegisterCallback(wDm.NewWASMEventModelTag, m.newWASMEventModelCB)
	m.wtm.RegisterCallback(wDm.ReceiveTag, m.receiveCB)
	m.wtm.RegisterCallback(wDm.ReceiveTextTag, m.receiveTextCB)
	m.wtm.RegisterCallback(wDm.ReceiveReplyTag, m.receiveReplyCB)
	m.wtm.RegisterCallback(wDm.ReceiveReactionTag, m.receiveReactionCB)
	m.wtm.RegisterCallback(wDm.UpdateSentStatusTag, m.updateSentStatusCB)
	m.wtm.RegisterCallback(wDm.DeleteMessageTag, m.deleteMessageCB)
	m.wtm.RegisterCallback(wDm.GetConversationTag, m.getConversationCB)
	m.wtm.RegisterCallback(wDm.GetConversationsTag, m.getConversationsCB)
}

// newWASMEventModelCB is the callback for NewWASMEventModel. Returns an empty
// slice on success or an error message on failure.
func (m *manager) newWASMEventModelCB(message []byte, reply func(message []byte)) {
	var msg wDm.NewWASMEventModelMessage
	err := json.Unmarshal(message, &msg)
	if err != nil {
		reply([]byte(errors.Wrapf(err,
			"failed to JSON unmarshal %T from main thread", msg).Error()))
		return
	}

	// Create new encryption cipher
	rng := fastRNG.NewStreamGenerator(12, 1024, csprng.NewSystemRNG)
	encryption, err := idbCrypto.NewCipherFromJSON(
		[]byte(msg.EncryptionJSON), rng.GetStream())
	if err != nil {
		reply([]byte(errors.Wrap(err,
			"failed to JSON unmarshal Cipher from main thread").Error()))
		return
	}

	m.model, err = NewWASMEventModel(
		msg.DatabaseName, encryption, m.messageReceivedCallback)
	if err != nil {
		reply([]byte(err.Error()))
		return
	}

	reply(nil)
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
	err = m.wtm.SendNoResponse(wDm.MessageReceivedCallbackTag, data)
	if err != nil {
		exception.Throwf("[DM] Could not send message for "+
			"MessageReceivedCallback: %+v", msg, err)
	}
}

// receiveCB is the callback for wasmModel.Receive. Returns a UUID of 0 on error
// or the JSON marshalled UUID (uint64) on success.
func (m *manager) receiveCB(message []byte, reply func(message []byte)) {
	var msg wDm.TransferMessage
	err := json.Unmarshal(message, &msg)
	if err != nil {
		jww.ERROR.Printf("[DM] Could not JSON unmarshal payload for Receive "+
			"from main thread: %+v", err)
		reply(zeroUUID)
		return
	}

	uuid := m.model.Receive(
		msg.MessageID, msg.Nickname, msg.Text, msg.PartnerKey, msg.SenderKey,
		msg.DmToken, msg.Codeset, msg.Timestamp, msg.Round, msg.MType,
		msg.Status)

	replyMsg, err := json.Marshal(uuid)
	if err != nil {
		exception.Throwf(
			"[DM] Could not JSON marshal UUID for Receive: %+v", err)
	}

	reply(replyMsg)
}

// receiveTextCB is the callback for wasmModel.ReceiveText. Returns a UUID of 0
// on error or the JSON marshalled UUID (uint64) on success.
func (m *manager) receiveTextCB(message []byte, reply func(message []byte)) {
	var msg wDm.TransferMessage
	err := json.Unmarshal(message, &msg)
	if err != nil {
		jww.ERROR.Printf("[DM] Could not JSON unmarshal payload for "+
			"ReceiveText from main thread: %+v", err)
		reply(zeroUUID)
		return
	}

	uuid := m.model.ReceiveText(
		msg.MessageID, msg.Nickname, string(msg.Text), msg.PartnerKey,
		msg.SenderKey, msg.DmToken, msg.Codeset, msg.Timestamp, msg.Round,
		msg.Status)

	replyMsg, err := json.Marshal(uuid)
	if err != nil {
		exception.Throwf(
			"[DM] Could not JSON marshal UUID for ReceiveText: %+v", err)
	}

	reply(replyMsg)
}

// receiveReplyCB is the callback for wasmModel.ReceiveReply. Returns a UUID of
// 0 on error or the JSON marshalled UUID (uint64) on success.
func (m *manager) receiveReplyCB(message []byte, reply func(message []byte)) {
	var msg wDm.TransferMessage
	err := json.Unmarshal(message, &msg)
	if err != nil {
		jww.ERROR.Printf("[DM] Could not JSON unmarshal payload for "+
			"ReceiveReply from main thread: %+v", err)
		reply(zeroUUID)
		return
	}

	uuid := m.model.ReceiveReply(msg.MessageID, msg.ReactionTo, msg.Nickname,
		string(msg.Text), msg.PartnerKey, msg.SenderKey, msg.DmToken, msg.Codeset, msg.Timestamp,
		msg.Round, msg.Status)

	replyMsg, err := json.Marshal(uuid)
	if err != nil {
		exception.Throwf(
			"[DM] Could not JSON marshal UUID for ReceiveReply: %+v", err)
	}

	reply(replyMsg)
}

// receiveReactionCB is the callback for wasmModel.ReceiveReaction. Returns a
// UUID of 0 on error or the JSON marshalled UUID (uint64) on success.
func (m *manager) receiveReactionCB(message []byte, reply func(message []byte)) {
	var msg wDm.TransferMessage
	err := json.Unmarshal(message, &msg)
	if err != nil {
		jww.ERROR.Printf("[DM] Could not JSON unmarshal payload for "+
			"ReceiveReaction from main thread: %+v", err)
		reply(zeroUUID)
		return
	}

	uuid := m.model.ReceiveReaction(msg.MessageID, msg.ReactionTo, msg.Nickname,
		string(msg.Text), msg.PartnerKey, msg.SenderKey, msg.DmToken, msg.Codeset, msg.Timestamp,
		msg.Round, msg.Status)

	replyMsg, err := json.Marshal(uuid)
	if err != nil {
		exception.Throwf(
			"[DM] Could not JSON marshal UUID for ReceiveReaction: %+v", err)
	}

	reply(replyMsg)
}

// updateSentStatusCB is the callback for wasmModel.UpdateSentStatus. Always
// returns nil; meaning, no response is supplied (or expected).
func (m *manager) updateSentStatusCB(message []byte, _ func([]byte)) {
	var msg wDm.TransferMessage
	err := json.Unmarshal(message, &msg)
	if err != nil {
		jww.ERROR.Printf("[DM] Could not JSON unmarshal %T for "+
			"UpdateSentStatus from main thread: %+v", msg, err)
		return
	}

	m.model.UpdateSentStatus(
		msg.UUID, msg.MessageID, msg.Timestamp, msg.Round, msg.Status)
}

// deleteMessageCB is the callback for wasmModel.DeleteMessage. Returns a JSON
// marshalled bool.
func (m *manager) deleteMessageCB(message []byte, reply func(message []byte)) {
	var msg wDm.TransferMessage
	err := json.Unmarshal(message, &msg)
	if err != nil {
		jww.ERROR.Printf("[DM] Could not JSON unmarshal %T for "+
			"UpdateSentStatus from main thread: %+v", msg, err)
		reply([]byte{0})
		return
	}

	if m.model.DeleteMessage(msg.MessageID, msg.SenderKey) {
		reply([]byte{1})
		return
	}
	reply([]byte{0})
}

// getConversationCB is the callback for wasmModel.GetConversation.
// Returns nil on error or the JSON marshalled Conversation on success.
func (m *manager) getConversationCB(message []byte, reply func(message []byte)) {
	result := m.model.GetConversation(message)
	replyMessage, err := json.Marshal(result)
	if err != nil {
		exception.Throwf("[DM] Could not JSON marshal %T for "+
			"GetConversation: %+v", result, err)
	}
	reply(replyMessage)
}

// getConversationsCB is the callback for wasmModel.GetConversations.
// Returns nil on error or the JSON marshalled list of Conversation on success.
func (m *manager) getConversationsCB(_ []byte, reply func(message []byte)) {
	result := m.model.GetConversations()
	replyMessage, err := json.Marshal(result)
	if err != nil {
		exception.Throwf("[DM] Could not JSON marshal %T for "+
			"GetConversations: %+v", result, err)
	}
	reply(replyMessage)
}
