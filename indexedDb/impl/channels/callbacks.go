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
	"gitlab.com/elixxir/client/v4/channels"
	"gitlab.com/elixxir/client/v4/cmix/rounds"
	cryptoBroadcast "gitlab.com/elixxir/crypto/broadcast"
	cryptoChannel "gitlab.com/elixxir/crypto/channel"
	"gitlab.com/elixxir/crypto/fastRNG"
	"gitlab.com/elixxir/crypto/message"
	wChannels "gitlab.com/elixxir/xxdk-wasm/indexedDb/worker/channels"
	"gitlab.com/elixxir/xxdk-wasm/worker"
	"gitlab.com/xx_network/crypto/csprng"
	"gitlab.com/xx_network/primitives/id"
	"time"
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
	m.wtm.RegisterCallback(wChannels.UpdateFromUUIDTag, m.updateFromUUIDCB)
	m.wtm.RegisterCallback(wChannels.UpdateFromMessageIDTag, m.updateFromMessageIDCB)
	m.wtm.RegisterCallback(wChannels.GetMessageTag, m.getMessageCB)
	m.wtm.RegisterCallback(wChannels.DeleteMessageTag, m.deleteMessageCB)
	m.wtm.RegisterCallback(wChannels.MuteUserTag, m.muteUserCB)
}

// newWASMEventModelCB is the callback for NewWASMEventModel. Returns an empty
// slice on success or an error message on failure.
func (m *manager) newWASMEventModelCB(data []byte) ([]byte, error) {
	var msg wChannels.NewWASMEventModelMessage
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
		return []byte{}, errors.Errorf(
			"failed to JSON unmarshal Cipher from main thread: %+v", err)
	}

	m.model, err = NewWASMEventModel(msg.DatabaseName, encryption,
		m.messageReceivedCallback, m.deletedMessageCallback, m.mutedUserCallback)
	if err != nil {
		return []byte(err.Error()), nil
	}

	return []byte{}, nil
}

// messageReceivedCallback sends calls to the channels.MessageReceivedCallback
// in the main thread.
//
// storeEncryptionStatus adhere to the channels.MessageReceivedCallback type.
func (m *manager) messageReceivedCallback(
	uuid uint64, channelID *id.ID, update bool) {
	// Package parameters for sending
	msg := &wChannels.MessageReceivedCallbackMessage{
		UUID:      uuid,
		ChannelID: channelID,
		Update:    update,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		jww.ERROR.Printf("Could not JSON marshal %T: %+v", msg, err)
		return
	}

	// Send it to the main thread
	m.wtm.SendMessage(wChannels.MessageReceivedCallbackTag, data)
}

// deletedMessageCallback sends calls to the channels.DeletedMessageCallback in
// the main thread.
//
// storeEncryptionStatus adhere to the channels.MessageReceivedCallback type.
func (m *manager) deletedMessageCallback(messageID message.ID) {
	m.wtm.SendMessage(wChannels.DeletedMessageCallbackTag, messageID.Marshal())
}

// mutedUserCallback sends calls to the channels.MutedUserCallback in the main
// thread.
//
// storeEncryptionStatus adhere to the channels.MessageReceivedCallback type.
func (m *manager) mutedUserCallback(
	channelID *id.ID, pubKey ed25519.PublicKey, unmute bool) {
	// Package parameters for sending
	msg := &wChannels.MuteUserMessage{
		ChannelID: channelID,
		PubKey:    pubKey,
		Unmute:    unmute,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		jww.ERROR.Printf("Could not JSON marshal %T: %+v", msg, err)
		return
	}

	// Send it to the main thread
	m.wtm.SendMessage(wChannels.MutedUserCallbackTag, data)
}

// joinChannelCB is the callback for wasmModel.JoinChannel. Always returns nil;
// meaning, no response is supplied (or expected).
func (m *manager) joinChannelCB(data []byte) ([]byte, error) {
	var channel cryptoBroadcast.Channel
	err := json.Unmarshal(data, &channel)
	if err != nil {
		return nil, errors.Errorf(
			"failed to JSON unmarshal %T from main thread: %+v", channel, err)
	}

	m.model.JoinChannel(&channel)
	return nil, nil
}

// leaveChannelCB is the callback for wasmModel.LeaveChannel. Always returns
// nil; meaning, no response is supplied (or expected).
func (m *manager) leaveChannelCB(data []byte) ([]byte, error) {
	channelID, err := id.Unmarshal(data)
	if err != nil {
		return nil, errors.Errorf(
			"failed to JSON unmarshal %T from main thread: %+v", channelID, err)
	}

	m.model.LeaveChannel(channelID)
	return nil, nil
}

// receiveMessageCB is the callback for wasmModel.ReceiveMessage. Returns a UUID
// of 0 on error or the JSON marshalled UUID (uint64) on success.
func (m *manager) receiveMessageCB(data []byte) ([]byte, error) {
	var msg channels.ModelMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return zeroUUID, errors.Errorf(
			"failed to JSON unmarshal %T from main thread: %+v", msg, err)
	}

	uuid := m.model.ReceiveMessage(msg.ChannelID, msg.MessageID, msg.Nickname,
		string(msg.Content), msg.PubKey, msg.DmToken, msg.CodesetVersion,
		msg.Timestamp, msg.Lease, rounds.Round{ID: msg.Round}, msg.Type,
		msg.Status, msg.Hidden)

	uuidData, err := json.Marshal(uuid)
	if err != nil {
		return zeroUUID, errors.Errorf("failed to JSON marshal UUID: %+v", err)
	}
	return uuidData, nil
}

// receiveReplyCB is the callback for wasmModel.ReceiveReply. Returns a UUID of
// 0 on error or the JSON marshalled UUID (uint64) on success.
func (m *manager) receiveReplyCB(data []byte) ([]byte, error) {
	var msg wChannels.ReceiveReplyMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return zeroUUID, errors.Errorf(
			"failed to JSON unmarshal %T from main thread: %+v", msg, err)
	}

	uuid := m.model.ReceiveReply(msg.ChannelID, msg.MessageID, msg.ReactionTo,
		msg.Nickname, string(msg.Content), msg.PubKey, msg.DmToken,
		msg.CodesetVersion, msg.Timestamp, msg.Lease,
		rounds.Round{ID: msg.Round}, msg.Type, msg.Status, msg.Hidden)

	uuidData, err := json.Marshal(uuid)
	if err != nil {
		return zeroUUID, errors.Errorf("failed to JSON marshal UUID: %+v", err)
	}
	return uuidData, nil
}

// receiveReactionCB is the callback for wasmModel.ReceiveReaction. Returns a
// UUID of 0 on error or the JSON marshalled UUID (uint64) on success.
func (m *manager) receiveReactionCB(data []byte) ([]byte, error) {
	var msg wChannels.ReceiveReplyMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return zeroUUID, errors.Errorf(
			"failed to JSON unmarshal %T from main thread: %+v", msg, err)
	}

	uuid := m.model.ReceiveReaction(msg.ChannelID, msg.MessageID,
		msg.ReactionTo, msg.Nickname, string(msg.Content), msg.PubKey,
		msg.DmToken, msg.CodesetVersion, msg.Timestamp, msg.Lease,
		rounds.Round{ID: msg.Round}, msg.Type, msg.Status, msg.Hidden)

	uuidData, err := json.Marshal(uuid)
	if err != nil {
		return zeroUUID, errors.Errorf("failed to JSON marshal UUID: %+v", err)
	}
	return uuidData, nil
}

// updateFromUUIDCB is the callback for wasmModel.UpdateFromUUID. Always returns
// nil; meaning, no response is supplied (or expected).
func (m *manager) updateFromUUIDCB(data []byte) ([]byte, error) {
	var msg wChannels.MessageUpdateInfo
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, errors.Errorf(
			"failed to JSON unmarshal %T from main thread: %+v", msg, err)
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
		return []byte(err.Error()), nil
	}

	return nil, nil
}

// updateFromMessageIDCB is the callback for wasmModel.UpdateFromMessageID.
// Always returns nil; meaning, no response is supplied (or expected).
func (m *manager) updateFromMessageIDCB(data []byte) ([]byte, error) {
	var msg wChannels.MessageUpdateInfo
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, errors.Errorf(
			"failed to JSON unmarshal %T from main thread: %+v", msg, err)
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

	var ue wChannels.UuidError
	uuid, err := m.model.UpdateFromMessageID(
		msg.MessageID, timestamp, round, pinned, hidden, status)
	if err != nil {
		ue.Error = []byte(err.Error())
	} else {
		ue.UUID = uuid
	}

	data, err = json.Marshal(ue)
	if err != nil {
		return nil, errors.Errorf("failed to JSON marshal %T: %+v", ue, err)
	}

	return data, nil
}

// getMessageCB is the callback for wasmModel.GetMessage. Returns JSON
// marshalled channels.GetMessageMessage. If an error occurs, then Error will
// be set with the error message. Otherwise, Message will be set. Only one field
// will be set.
func (m *manager) getMessageCB(data []byte) ([]byte, error) {
	messageID, err := message.UnmarshalID(data)
	if err != nil {
		return nil, errors.Errorf(
			"failed to JSON unmarshal %T from main thread: %+v", messageID, err)
	}

	reply := wChannels.GetMessageMessage{}

	msg, err := m.model.GetMessage(messageID)
	if err != nil {
		reply.Error = err.Error()
	} else {
		reply.Message = msg
	}

	messageData, err := json.Marshal(reply)
	if err != nil {
		return nil, errors.Errorf("failed to JSON marshal %T from main thread "+
			"for GetMessage reply: %+v", reply, err)
	}
	return messageData, nil
}

// deleteMessageCB is the callback for wasmModel.DeleteMessage. Always returns
// nil; meaning, no response is supplied (or expected).
func (m *manager) deleteMessageCB(data []byte) ([]byte, error) {
	messageID, err := message.UnmarshalID(data)
	if err != nil {
		return nil, errors.Errorf(
			"failed to JSON unmarshal %T from main thread: %+v", messageID, err)
	}

	err = m.model.DeleteMessage(messageID)
	if err != nil {
		return []byte(err.Error()), nil
	}

	return nil, nil
}

// muteUserCB is the callback for wasmModel.MuteUser. Always returns nil;
// meaning, no response is supplied (or expected).
func (m *manager) muteUserCB(data []byte) ([]byte, error) {
	var msg wChannels.MuteUserMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, errors.Errorf(
			"failed to JSON unmarshal %T from main thread: %+v", msg, err)
	}

	m.model.MuteUser(msg.ChannelID, msg.PubKey, msg.Unmute)

	return nil, nil
}
