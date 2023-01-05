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
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/v4/channels"
	"gitlab.com/elixxir/client/v4/cmix/rounds"
	cryptoBroadcast "gitlab.com/elixxir/crypto/broadcast"
	cryptoChannel "gitlab.com/elixxir/crypto/channel"
	"gitlab.com/elixxir/crypto/fastRNG"
	"gitlab.com/elixxir/crypto/message"
	"gitlab.com/elixxir/xxdk-wasm/indexedDb"
	"gitlab.com/elixxir/xxdk-wasm/indexedDbWorker"
	mChannels "gitlab.com/elixxir/xxdk-wasm/indexedDbWorker/channels"
	"gitlab.com/xx_network/crypto/csprng"
	"gitlab.com/xx_network/primitives/id"
	"time"
)

// manager handles the event model and the message handler, which is used to
// send information between the event model and the main thread.
type manager struct {
	mh    *indexedDb2.MessageHandler
	model channels.EventModel
}

// RegisterHandlers registers all the reception handlers to manage messages from
// the main thread for the channels.EventModel.
func (m *manager) RegisterHandlers() {

	m.mh.RegisterHandler(indexedDb.NewWASMEventModelTag, m.newWASMEventModelHandler)
	m.mh.RegisterHandler(indexedDb.JoinChannelTag, m.joinChannelHandler)
	m.mh.RegisterHandler(indexedDb.LeaveChannelTag, m.leaveChannelHandler)
	m.mh.RegisterHandler(indexedDb.ReceiveMessageTag, m.receiveMessageHandler)
	m.mh.RegisterHandler(indexedDb.ReceiveReplyTag, m.receiveReplyHandler)
	m.mh.RegisterHandler(indexedDb.ReceiveReactionTag, m.receiveReactionHandler)
	m.mh.RegisterHandler(indexedDb.UpdateFromUUIDTag, m.updateFromUUIDHandler)
	m.mh.RegisterHandler(indexedDb.UpdateFromMessageIDTag, m.updateFromMessageIDHandler)
	m.mh.RegisterHandler(indexedDb.GetMessageTag, m.getMessageHandler)
	m.mh.RegisterHandler(indexedDb.DeleteMessageTag, m.deleteMessageHandler)
}

// newWASMEventModelHandler is the handler for NewWASMEventModel.
func (m *manager) newWASMEventModelHandler(data []byte) []byte {
	var msg mChannels.NewWASMEventModelMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		jww.ERROR.Printf("Could not JSON unmarshal "+
			"NewWASMEventModelMessage from main thread: %+v", err)
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
// storeEncryptionStatus adhere to the indexedDbWorker.MessageReceivedCallback type.
func (m *manager) messageReceivedCallback(
	uuid uint64, channelID *id.ID, update bool) {
	// Package parameters for sending
	msg := &mChannels.MessageReceivedCallbackMessage{
		UUID:      uuid,
		ChannelID: channelID,
		Update:    update,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		jww.ERROR.Printf(
			"Could not JSON marshal MessageReceivedCallbackMessage: %+v", err)
		return
	}

	// Send it to the main thread
	m.mh.SendResponse(indexedDb.GetMessageTag, indexedDb.InitID, data)
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
	m.mh.RegisterHandler(indexedDb.EncryptionStatusTag,
		func(data []byte) []byte {
			responseChan <- data
			return nil
		})

	// Send encryption status to main thread
	m.mh.SendResponse(indexedDb.EncryptionStatusTag, indexedDb.InitID, data)

	// Wait for response
	var response mChannels.EncryptionStatusReply
	select {
	case responseData := <-responseChan:
		if err = json.Unmarshal(responseData, &response); err != nil {
			return false, err
		}
	case <-time.After(indexedDb.ResponseTimeout):
		return false, errors.Errorf("timed out after %s waiting for "+
			"response about the database encryption status from local "+
			"storage in the main thread", indexedDb.ResponseTimeout)
	}

	// If the response contain an error, return it
	if response.Error != "" {
		return false, errors.New(response.Error)
	}

	// Return the encryption status
	return response.EncryptionStatus, nil
}

// joinChannelHandler is the handler for wasmModel.JoinChannel.
func (m *manager) joinChannelHandler(data []byte) []byte {
	var channel cryptoBroadcast.Channel
	err := json.Unmarshal(data, &channel)
	if err != nil {
		jww.ERROR.Printf("Could not JSON unmarshal broadcast.Channel from "+
			"main thread: %+v", err)
		return nil
	}

	m.model.JoinChannel(&channel)
	return nil
}

// leaveChannelHandler is the handler for wasmModel.LeaveChannel.
func (m *manager) leaveChannelHandler(data []byte) []byte {
	channelID, err := id.Unmarshal(data)
	if err != nil {
		jww.ERROR.Printf(
			"Could not unmarshal channel ID from main thread: %+v", err)
		return nil
	}

	m.model.LeaveChannel(channelID)
	return nil
}

// receiveMessageHandler is the handler for wasmModel.ReceiveMessage.
func (m *manager) receiveMessageHandler(data []byte) []byte {
	var msg channels.ModelMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		jww.ERROR.Printf("Could not JSON unmarshal channels.ModelMessage "+
			"from main thread: %+v", err)
		return nil
	}

	uuid := m.model.ReceiveMessage(msg.ChannelID, msg.MessageID, msg.Nickname,
		string(msg.Content), msg.PubKey, msg.DmToken, msg.CodesetVersion,
		msg.Timestamp, msg.Lease, rounds.Round{ID: msg.Round}, msg.Type,
		msg.Status, msg.Hidden)

	uuidData, err := json.Marshal(uuid)
	if err != nil {
		jww.ERROR.Printf(
			"Could not JSON marshal UUID from ReceiveMessage: %+v", err)
		return nil
	}
	return uuidData
}

// receiveReplyHandler is the handler for wasmModel.ReceiveReply.
func (m *manager) receiveReplyHandler(data []byte) []byte {
	var msg mChannels.ReceiveReplyMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		jww.ERROR.Printf("Could not JSON unmarshal ReceiveReplyMessage "+
			"from main thread: %+v", err)
		return nil
	}

	uuid := m.model.ReceiveReply(msg.ChannelID, msg.MessageID, msg.ReactionTo,
		msg.Nickname, string(msg.Content), msg.PubKey, msg.DmToken,
		msg.CodesetVersion, msg.Timestamp, msg.Lease,
		rounds.Round{ID: msg.Round}, msg.Type, msg.Status, msg.Hidden)

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
	var msg mChannels.ReceiveReplyMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		jww.ERROR.Printf("Could not JSON unmarshal ReceiveReplyMessage "+
			"from main thread: %+v", err)
		return nil
	}

	uuid := m.model.ReceiveReaction(msg.ChannelID, msg.MessageID,
		msg.ReactionTo, msg.Nickname, string(msg.Content), msg.PubKey,
		msg.DmToken, msg.CodesetVersion, msg.Timestamp, msg.Lease,
		rounds.Round{ID: msg.Round}, msg.Type, msg.Status, msg.Hidden)

	uuidData, err := json.Marshal(uuid)
	if err != nil {
		jww.ERROR.Printf(
			"Could not JSON marshal UUID from ReceiveReaction: %+v", err)
		return nil
	}
	return uuidData
}

// updateFromUUIDHandler is the handler for wasmModel.UpdateFromUUID.
func (m *manager) updateFromUUIDHandler(data []byte) []byte {
	var msg mChannels.MessageUpdateInfo
	err := json.Unmarshal(data, &msg)
	if err != nil {
		jww.ERROR.Printf("Could not JSON unmarshal MessageUpdateInfo "+
			"from main thread: %+v", err)
		return nil
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

	m.model.UpdateFromUUID(
		msg.UUID, messageID, timestamp, round, pinned, hidden, status)
	return nil
}

// updateFromMessageIDHandler is the handler for wasmModel.UpdateFromMessageID.
func (m *manager) updateFromMessageIDHandler(data []byte) []byte {
	var msg mChannels.MessageUpdateInfo
	err := json.Unmarshal(data, &msg)
	if err != nil {
		jww.ERROR.Printf("Could not JSON unmarshal MessageUpdateInfo "+
			"from main thread: %+v", err)
		return nil
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

	uuid := m.model.UpdateFromMessageID(
		msg.MessageID, timestamp, round, pinned, hidden, status)

	uuidData, err := json.Marshal(uuid)
	if err != nil {
		jww.ERROR.Printf(
			"Could not JSON marshal UUID from UpdateFromMessageID: %+v", err)
		return nil
	}
	return uuidData
}

// getMessageHandler is the handler for wasmModel.GetMessage.
func (m *manager) getMessageHandler(data []byte) []byte {
	messageID, err := message.UnmarshalID(data)
	if err != nil {
		jww.ERROR.Printf("Could not JSON unmarshal message ID from main "+
			"thread: %+v", err)
		return nil
	}

	reply := mChannels.GetMessageMessage{}

	msg, err := m.model.GetMessage(messageID)
	if err != nil {
		reply.Error = err.Error()
	} else {
		reply.Message = msg
	}

	messageData, err := json.Marshal(msg)
	if err != nil {
		jww.ERROR.Printf(
			"Could not JSON marshal UUID from ReceiveReaction: %+v", err)
		return nil
	}
	return messageData
}

// deleteMessageHandler is the handler for wasmModel.DeleteMessage.
func (m *manager) deleteMessageHandler(data []byte) []byte {
	messageID, err := message.UnmarshalID(data)
	if err != nil {
		jww.ERROR.Printf("Could not JSON unmarshal message ID from main "+
			"thread: %+v", err)
		return nil
	}

	err = m.model.DeleteMessage(messageID)
	if err != nil {
		return []byte(err.Error())
	}

	return nil
}
