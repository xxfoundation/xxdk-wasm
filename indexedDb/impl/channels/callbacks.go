////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package main

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"

	"gitlab.com/elixxir/client/v4/channels"
	"gitlab.com/elixxir/client/v4/cmix/rounds"
	cryptoBroadcast "gitlab.com/elixxir/crypto/broadcast"
	"gitlab.com/elixxir/crypto/fastRNG"
	idbCrypto "gitlab.com/elixxir/crypto/indexedDb"
	"gitlab.com/elixxir/crypto/message"
	"gitlab.com/elixxir/wasm-utils/exception"
	wChannels "gitlab.com/elixxir/xxdk-wasm/indexedDb/worker/channels"
	"gitlab.com/elixxir/xxdk-wasm/worker"
	"gitlab.com/xx_network/crypto/csprng"
	"gitlab.com/xx_network/primitives/id"
)

var zeroUUID = []byte{0, 0, 0, 0, 0, 0, 0, 0}

// manager handles the event model and the message callbacks, which is used to
// send information between the event model and the main thread.
type manager struct {
	wtm   *worker.ThreadManager
	model channels.EventModel
}

// registerCallbacks registers all the reception callbacks to manage messages
// from the main thread for the channels.EventModel.
func (m *manager) registerCallbacks() {
	m.wtm.RegisterCallback(wChannels.NewWASMEventModelTag, m.newWASMEventModelCB)
	m.wtm.RegisterCallback(wChannels.JoinChannelTag, m.joinChannelCB)
	m.wtm.RegisterCallback(wChannels.LeaveChannelTag, m.leaveChannelCB)
	m.wtm.RegisterCallback(wChannels.ReceiveMessageTag, m.receiveMessageCB)
	m.wtm.RegisterCallback(wChannels.ReceiveReplyTag, m.receiveReplyCB)
	m.wtm.RegisterCallback(wChannels.ReceiveReactionTag, m.receiveReactionCB)
	m.wtm.RegisterCallback(wChannels.UpdateFromUUIDTag, m.updateFromUuidCB)
	m.wtm.RegisterCallback(wChannels.UpdateFromMessageIDTag, m.updateFromMessageIdCB)
	m.wtm.RegisterCallback(wChannels.GetMessageTag, m.getMessageCB)
	m.wtm.RegisterCallback(wChannels.DeleteMessageTag, m.deleteMessageCB)
	m.wtm.RegisterCallback(wChannels.MuteUserTag, m.muteUserCB)
}

// newWASMEventModelCB is the callback for NewWASMEventModel. Returns an empty
// slice on success or an error message on failure.
func (m *manager) newWASMEventModelCB(message []byte, reply func(message []byte)) {
	var msg wChannels.NewWASMEventModelMessage
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

	m.model, err = NewWASMEventModel(msg.DatabaseName, encryption, m)
	if err != nil {
		reply([]byte(err.Error()))
		return
	}

	reply(nil)
}

// EventUpdate implements [bindings.ChannelUICallbacks.EventUpdate].
func (m *manager) EventUpdate(eventType int64, jsonData []byte) {
	// Package parameters for sending
	msg := &wChannels.EventUpdateCallbackMessage{
		EventType: eventType,
		JsonData:  jsonData,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		exception.Throwf("[CH] Could not JSON marshal %T for EventUpdate "+
			"callback: %+v", msg, err)
	}

	// Send it to the main thread
	err = m.wtm.SendNoResponse(wChannels.EventUpdateCallbackTag, data)
	if err != nil {
		exception.Throwf("[CH] Could not send message for EventUpdate "+
			"callback: %+v", msg, err)
	}
}

// joinChannelCB is the callback for wasmModel.JoinChannel. Always returns nil;
// meaning, no response is supplied (or expected).
func (m *manager) joinChannelCB(message []byte, _ func([]byte)) {
	var channel cryptoBroadcast.Channel
	err := json.Unmarshal(message, &channel)
	if err != nil {
		jww.ERROR.Printf("[CH] Could not JSON unmarshal %T from main thread: "+
			"%+v", channel, err)
		return
	}

	m.model.JoinChannel(&channel)
}

// leaveChannelCB is the callback for wasmModel.LeaveChannel. Always returns
// nil; meaning, no response is supplied (or expected).
func (m *manager) leaveChannelCB(message []byte, _ func([]byte)) {
	channelID, err := id.Unmarshal(message)
	if err != nil {
		jww.ERROR.Printf("[CH] Could not JSON unmarshal %T from main thread: "+
			"%+v", channelID, err)
		return
	}

	m.model.LeaveChannel(channelID)
}

// receiveMessageCB is the callback for wasmModel.ReceiveMessage. Returns a UUID
// of 0 on error or the JSON marshalled UUID (uint64) on success.
func (m *manager) receiveMessageCB(message []byte, reply func(message []byte)) {
	var msg channels.ModelMessage
	err := json.Unmarshal(message, &msg)
	if err != nil {
		jww.ERROR.Printf("[CH] Could not JSON unmarshal payload for "+
			"ReceiveMessage from main thread: %+v", err)
		reply(zeroUUID)
		return
	}

	uuid := m.model.ReceiveMessage(msg.ChannelID, msg.MessageID, msg.Nickname,
		string(msg.Content), msg.PubKey, msg.DmToken, msg.CodesetVersion,
		msg.Timestamp, msg.Lease, rounds.Round{ID: msg.Round}, msg.Type,
		msg.Status, msg.Hidden)

	replyMsg, err := json.Marshal(uuid)
	if err != nil {
		exception.Throwf(
			"[CH] Could not JSON marshal UUID for ReceiveMessage: %+v", err)
	}

	reply(replyMsg)
}

// receiveReplyCB is the callback for wasmModel.ReceiveReply. Returns a UUID of
// 0 on error or the JSON marshalled UUID (uint64) on success.
func (m *manager) receiveReplyCB(message []byte, reply func(message []byte)) {
	var msg wChannels.ReceiveReplyMessage
	err := json.Unmarshal(message, &msg)
	if err != nil {
		jww.ERROR.Printf("[CH] Could not JSON unmarshal payload for "+
			"ReceiveReply from main thread: %+v", err)
		reply(zeroUUID)
		return
	}

	uuid := m.model.ReceiveReply(msg.ChannelID, msg.MessageID, msg.ReactionTo,
		msg.Nickname, string(msg.Content), msg.PubKey, msg.DmToken,
		msg.CodesetVersion, msg.Timestamp, msg.Lease,
		rounds.Round{ID: msg.Round}, msg.Type, msg.Status, msg.Hidden)

	replyMsg, err := json.Marshal(uuid)
	if err != nil {
		exception.Throwf(
			"[CH] Could not JSON marshal UUID for ReceiveReply: %+v", err)
	}

	reply(replyMsg)
}

// receiveReactionCB is the callback for wasmModel.ReceiveReaction. Returns a
// UUID of 0 on error or the JSON marshalled UUID (uint64) on success.
func (m *manager) receiveReactionCB(message []byte, reply func(message []byte)) {
	var msg wChannels.ReceiveReplyMessage
	err := json.Unmarshal(message, &msg)
	if err != nil {
		jww.ERROR.Printf("[CH] Could not JSON unmarshal payload for "+
			"ReceiveReaction from main thread: %+v", err)
		reply(zeroUUID)
		return
	}

	uuid := m.model.ReceiveReaction(msg.ChannelID, msg.MessageID,
		msg.ReactionTo, msg.Nickname, string(msg.Content), msg.PubKey,
		msg.DmToken, msg.CodesetVersion, msg.Timestamp, msg.Lease,
		rounds.Round{ID: msg.Round}, msg.Type, msg.Status, msg.Hidden)

	replyMsg, err := json.Marshal(uuid)
	if err != nil {
		exception.Throwf(
			"[CH] Could not JSON marshal UUID for ReceiveReaction: %+v", err)
	}

	reply(replyMsg)
}

// updateFromUuidCB is the callback for wasmModel.UpdateFromUUID. Always returns
// nil; meaning, no response is supplied (or expected).
func (m *manager) updateFromUuidCB(messageData []byte, reply func(message []byte)) {
	var msg wChannels.MessageUpdateInfo
	err := json.Unmarshal(messageData, &msg)
	if err != nil {
		reply([]byte(errors.Errorf("failed to JSON unmarshal %T from main "+
			"thread: %+v", msg, err).Error()))
		return
	}

	var messageID *message.ID
	var timestamp *time.Time
	var round *rounds.Round
	var pinned, hidden *bool
	var status *channels.SentStatus
	if msg.MessageIDSet {
		messageID = &msg.MessageID
	}
	if msg.TimestampSet {
		timestamp = &msg.Timestamp
	}
	if msg.RoundIDSet {
		round = &rounds.Round{ID: msg.RoundID}
	}
	if msg.PinnedSet {
		pinned = &msg.Pinned
	}
	if msg.HiddenSet {
		hidden = &msg.Hidden
	}
	if msg.StatusSet {
		status = &msg.Status
	}

	err = m.model.UpdateFromUUID(
		msg.UUID, messageID, timestamp, round, pinned, hidden, status)
	if err != nil {
		reply([]byte(err.Error()))
	}

	reply(nil)
}

// updateFromMessageIdCB is the callback for wasmModel.UpdateFromMessageID.
// Always returns nil; meaning, no response is supplied (or expected).
func (m *manager) updateFromMessageIdCB(message []byte, reply func(message []byte)) {
	var ue wChannels.UuidError
	defer func() {
		if replyMessage, err := json.Marshal(ue); err != nil {
			exception.Throwf("[CH] Failed to JSON marshal %T for "+
				"UpdateFromMessageID: %+v", ue, err)
		} else {
			reply(replyMessage)
		}
	}()

	var msg wChannels.MessageUpdateInfo
	err := json.Unmarshal(message, &msg)
	if err != nil {
		ue.Error = errors.Errorf(
			"failed to JSON unmarshal %T from main thread: %+v", msg, err).Error()
		return
	}
	var timestamp *time.Time
	var round *rounds.Round
	var pinned, hidden *bool
	var status *channels.SentStatus
	if msg.TimestampSet {
		timestamp = &msg.Timestamp
	}
	if msg.RoundIDSet {
		round = &rounds.Round{ID: msg.RoundID}
	}
	if msg.PinnedSet {
		pinned = &msg.Pinned
	}
	if msg.HiddenSet {
		hidden = &msg.Hidden
	}
	if msg.StatusSet {
		status = &msg.Status
	}

	uuid, err := m.model.UpdateFromMessageID(
		msg.MessageID, timestamp, round, pinned, hidden, status)
	if err != nil {
		ue.Error = err.Error()
	} else {
		ue.UUID = uuid
	}
}

// getMessageCB is the callback for wasmModel.GetMessage. Returns JSON
// marshalled channels.GetMessageMessage. If an error occurs, then Error will
// be set with the error message. Otherwise, Message will be set. Only one field
// will be set.
func (m *manager) getMessageCB(messageData []byte, reply func(message []byte)) {
	var replyMsg wChannels.GetMessageMessage
	defer func() {
		if replyMessage, err := json.Marshal(replyMsg); err != nil {
			exception.Throwf("[CH] Failed to JSON marshal %T for "+
				"GetMessage: %+v", replyMsg, err)
		} else {
			reply(replyMessage)
		}
	}()

	messageID, err := message.UnmarshalID(messageData)
	if err != nil {
		replyMsg.Error = errors.Errorf("failed to JSON unmarshal %T from "+
			"main thread: %+v", messageID, err).Error()
		return
	}

	msg, err := m.model.GetMessage(messageID)
	if err != nil {
		replyMsg.Error = err.Error()
	} else {
		replyMsg.Message = msg
	}
}

// deleteMessageCB is the callback for wasmModel.DeleteMessage. Always returns
// nil; meaning, no response is supplied (or expected).
func (m *manager) deleteMessageCB(messageData []byte, reply func(message []byte)) {
	messageID, err := message.UnmarshalID(messageData)
	if err != nil {
		reply([]byte(errors.Errorf("failed to JSON unmarshal %T from main "+
			"thread: %+v", messageID, err).Error()))
		return
	}

	err = m.model.DeleteMessage(messageID)
	if err != nil {
		reply([]byte(err.Error()))
	}

	reply(nil)
}

// muteUserCB is the callback for wasmModel.MuteUser. Always returns nil;
// meaning, no response is supplied (or expected).
func (m *manager) muteUserCB(message []byte, _ func([]byte)) {
	var msg wChannels.MuteUserMessage
	err := json.Unmarshal(message, &msg)
	if err != nil {
		jww.ERROR.Printf("[CH] Could not JSON unmarshal %T for MuteUser from "+
			"main thread: %+v", msg, err)
		return
	}
	m.model.MuteUser(msg.ChannelID, msg.PubKey, msg.Unmute)
}
