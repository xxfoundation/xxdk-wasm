////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package indexedDb

import (
	"crypto/ed25519"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"

	"gitlab.com/elixxir/client/v4/channels"
	"gitlab.com/elixxir/client/v4/cmix/rounds"
	cryptoBroadcast "gitlab.com/elixxir/crypto/broadcast"
	cryptoChannel "gitlab.com/elixxir/crypto/channel"
	"gitlab.com/xx_network/primitives/id"
)

// wasmModel implements [channels.EventModel] interface, which uses the channels
// system passed an object that adheres to in order to get events on the
// channel.
type wasmModel struct {
	wh *workerHandler
}

// JoinChannel is called whenever a channel is joined locally.
func (w *wasmModel) JoinChannel(channel *cryptoBroadcast.Channel) {
	data, err := json.Marshal(channel)
	if err != nil {
		jww.ERROR.Printf("Could not JSON marshal broadcast.Channel: %+v", err)
		return
	}

	w.wh.sendMessage(JoinChannelTag, data, nil)
}

// LeaveChannel is called whenever a channel is left locally.
func (w *wasmModel) LeaveChannel(channelID *id.ID) {
	w.wh.sendMessage(LeaveChannelTag, channelID.Marshal(), nil)
}

// ReceiveMessage is called whenever a message is received on a given channel.
//
// It may be called multiple times on the same message; it is incumbent on the
// user of the API to filter such called by message ID.
func (w *wasmModel) ReceiveMessage(channelID *id.ID,
	messageID cryptoChannel.MessageID, nickname, text string,
	pubKey ed25519.PublicKey, codeset uint8,
	timestamp time.Time, lease time.Duration, round rounds.Round,
	mType channels.MessageType, status channels.SentStatus, hidden bool) uint64 {
	message := channels.ModelMessage{
		Nickname:       nickname,
		MessageID:      messageID,
		ChannelID:      channelID,
		Timestamp:      timestamp,
		Lease:          lease,
		Status:         status,
		Hidden:         hidden,
		Pinned:         false,
		Content:        []byte(text),
		Type:           mType,
		Round:          round.ID,
		PubKey:         pubKey,
		CodesetVersion: codeset,
	}

	data, err := json.Marshal(message)
	if err != nil {
		jww.ERROR.Printf(
			"Could not JSON marshal payload for ReceiveMessage: %+v", err)
		return 0
	}

	uuidChan := make(chan uint64)
	w.wh.sendMessage(ReceiveMessageTag, data, func(data []byte) {
		var uuid uint64
		err = json.Unmarshal(data, &uuid)
		if err != nil {
			jww.ERROR.Printf(
				"Could not JSON unmarshal response to ReceiveMessage: %+v", err)
			uuidChan <- 0
		}
		uuidChan <- uuid
	})

	return <-uuidChan
}

// ReceiveReplyMessage is JSON marshalled and sent to the worker for
// [wasmModel.ReceiveReply].
type ReceiveReplyMessage struct {
	ReactionTo cryptoChannel.MessageID `json:"replyTo"`
	Message    channels.ModelMessage   `json:"message"`
}

// ReceiveReply is called whenever a message is received that is a reply on a
// given channel. It may be called multiple times on the same message; it is
// incumbent on the user of the API to filter such called by message ID.
//
// Messages may arrive our of order, so a reply, in theory, can arrive before
// the initial message. As a result, it may be important to buffer replies.
func (w *wasmModel) ReceiveReply(channelID *id.ID,
	messageID cryptoChannel.MessageID, replyTo cryptoChannel.MessageID,
	nickname, text string, pubKey ed25519.PublicKey, codeset uint8,
	timestamp time.Time, lease time.Duration, round rounds.Round,
	mType channels.MessageType, status channels.SentStatus, hidden bool) uint64 {
	message := ReceiveReplyMessage{
		ReactionTo: replyTo,
		Message: channels.ModelMessage{
			Nickname:       nickname,
			MessageID:      messageID,
			ChannelID:      channelID,
			Timestamp:      timestamp,
			Lease:          lease,
			Status:         status,
			Hidden:         hidden,
			Pinned:         false,
			Content:        []byte(text),
			Type:           mType,
			Round:          round.ID,
			PubKey:         pubKey,
			CodesetVersion: codeset,
		},
	}

	data, err := json.Marshal(message)
	if err != nil {
		jww.ERROR.Printf(
			"Could not JSON marshal payload for ReceiveReply: %+v", err)
		return 0
	}

	uuidChan := make(chan uint64)
	w.wh.sendMessage(ReceiveReplyTag, data, func(data []byte) {
		var uuid uint64
		err = json.Unmarshal(data, &uuid)
		if err != nil {
			jww.ERROR.Printf(
				"Could not JSON unmarshal response to ReceiveReply: %+v", err)
			uuidChan <- 0
		}
		uuidChan <- uuid
	})

	return <-uuidChan
}

// ReceiveReaction is called whenever a reaction to a message is received on a
// given channel. It may be called multiple times on the same reaction; it is
// incumbent on the user of the API to filter such called by message ID.
//
// Messages may arrive our of order, so a reply, in theory, can arrive before
// the initial message. As a result, it may be important to buffer reactions.
func (w *wasmModel) ReceiveReaction(channelID *id.ID,
	messageID cryptoChannel.MessageID, reactionTo cryptoChannel.MessageID,
	nickname, reaction string, pubKey ed25519.PublicKey, codeset uint8,
	timestamp time.Time, lease time.Duration, round rounds.Round,
	mType channels.MessageType, status channels.SentStatus, hidden bool) uint64 {

	message := ReceiveReplyMessage{
		ReactionTo: reactionTo,
		Message: channels.ModelMessage{
			Nickname:       nickname,
			MessageID:      messageID,
			ChannelID:      channelID,
			Timestamp:      timestamp,
			Lease:          lease,
			Status:         status,
			Hidden:         hidden,
			Pinned:         false,
			Content:        []byte(reaction),
			Type:           mType,
			Round:          round.ID,
			PubKey:         pubKey,
			CodesetVersion: codeset,
		},
	}

	data, err := json.Marshal(message)
	if err != nil {
		jww.ERROR.Printf(
			"Could not JSON marshal payload for ReceiveReaction: %+v", err)
		return 0
	}

	uuidChan := make(chan uint64)
	w.wh.sendMessage(ReceiveReactionTag, data, func(data []byte) {
		var uuid uint64
		err = json.Unmarshal(data, &uuid)
		if err != nil {
			jww.ERROR.Printf(
				"Could not JSON unmarshal response to ReceiveReaction: %+v", err)
			uuidChan <- 0
		}
		uuidChan <- uuid
	})

	return <-uuidChan
}

// MessageUpdateInfo is JSON marshalled and sent to the worker for
// [wasmModel.UpdateFromMessageID] and [wasmModel.UpdateFromUUID].
type MessageUpdateInfo struct {
	UUID uint64 `json:"uuid"`

	MessageID    cryptoChannel.MessageID `json:"messageID"`
	MessageIDSet bool                    `json:"messageIDSet"`

	Timestamp    time.Time `json:"timestamp"`
	TimestampSet bool      `json:"timestampSet"`

	RoundID    id.Round `json:"round"`
	RoundIDSet bool     `json:"roundIDSet"`

	Pinned    bool `json:"pinned"`
	PinnedSet bool `json:"pinnedSet"`

	Hidden    bool `json:"hidden"`
	HiddenSet bool `json:"hiddenSet"`

	Status    channels.SentStatus `json:"status"`
	StatusSet bool                `json:"statusSet"`
}

// UpdateFromMessageID is called whenever a message with the message ID is
// modified.
//
// The API needs to return the UUID of the modified message that can be
// referenced at a later time.
//
// timestamp, round, pinned, and hidden are all nillable and may be updated
// based upon the UUID at a later date. If a nil value is passed, then make
// no update.
func (w *wasmModel) UpdateFromMessageID(messageID cryptoChannel.MessageID,
	timestamp *time.Time, round *rounds.Round, pinned, hidden *bool,
	status *channels.SentStatus) uint64 {
	message := MessageUpdateInfo{MessageID: messageID, MessageIDSet: true}
	if timestamp != nil {
		message.Timestamp = *timestamp
		message.TimestampSet = true
	}
	if round != nil {
		message.RoundID = round.ID
		message.RoundIDSet = true
	}
	if pinned != nil {
		message.Pinned = *pinned
		message.PinnedSet = true
	}
	if hidden != nil {
		message.Hidden = *hidden
		message.HiddenSet = true
	}
	if status != nil {
		message.Status = *status
		message.StatusSet = true
	}

	data, err := json.Marshal(message)
	if err != nil {
		jww.ERROR.Printf(
			"Could not JSON marshal payload for UpdateFromMessageID: %+v", err)
		return 0
	}

	uuidChan := make(chan uint64)
	w.wh.sendMessage(UpdateFromMessageIDTag, data, func(data []byte) {
		var uuid uint64
		err = json.Unmarshal(data, &uuid)
		if err != nil {
			jww.ERROR.Printf(
				"Could not JSON unmarshal response to UpdateFromMessageID: %+v", err)
			uuidChan <- 0
		}
		uuidChan <- uuid
	})

	return <-uuidChan
}

// UpdateFromUUID is called whenever a message at the UUID is modified.
//
// messageID, timestamp, round, pinned, and hidden are all nillable and may
// be updated based upon the UUID at a later date. If a nil value is passed,
// then make no update.
func (w *wasmModel) UpdateFromUUID(uuid uint64,
	messageID *cryptoChannel.MessageID, timestamp *time.Time,
	round *rounds.Round, pinned, hidden *bool, status *channels.SentStatus) {
	message := MessageUpdateInfo{UUID: uuid}
	if messageID != nil {
		message.MessageID = *messageID
		message.MessageIDSet = true
	}
	if timestamp != nil {
		message.Timestamp = *timestamp
		message.TimestampSet = true
	}
	if round != nil {
		message.RoundID = round.ID
		message.RoundIDSet = true
	}
	if pinned != nil {
		message.Pinned = *pinned
		message.PinnedSet = true
	}
	if hidden != nil {
		message.Hidden = *hidden
		message.HiddenSet = true
	}
	if status != nil {
		message.Status = *status
		message.StatusSet = true
	}

	data, err := json.Marshal(message)
	if err != nil {
		jww.ERROR.Printf(
			"Could not JSON marshal payload for UpdateFromUUID: %+v", err)
		return
	}

	w.wh.sendMessage(UpdateFromUUIDTag, data, nil)
}

// GetMessageMessage is JSON marshalled and sent to the worker for
// [wasmModel.GetMessage].
type GetMessageMessage struct {
	Message channels.ModelMessage `json:"message"`
	Error   string                `json:"error"`
}

// GetMessage returns the message with the given [channel.MessageID].
func (w *wasmModel) GetMessage(
	messageID cryptoChannel.MessageID) (channels.ModelMessage, error) {
	msgChan := make(chan GetMessageMessage)
	w.wh.sendMessage(GetMessageTag, messageID.Marshal(), func(data []byte) {
		var msg GetMessageMessage
		err := json.Unmarshal(data, &msg)
		if err != nil {
			jww.ERROR.Printf(
				"Could not JSON unmarshal response to GetMessage: %+v", err)
		}
		msgChan <- msg
	})

	msg := <-msgChan
	if msg.Error != "" {
		return channels.ModelMessage{}, errors.New(msg.Error)
	}

	return msg.Message, nil
}
