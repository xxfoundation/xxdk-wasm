////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package channels

import (
	"crypto/ed25519"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"

	"gitlab.com/elixxir/client/v4/channels"
	"gitlab.com/elixxir/client/v4/cmix/rounds"
	cryptoBroadcast "gitlab.com/elixxir/crypto/broadcast"
	"gitlab.com/elixxir/crypto/message"
	"gitlab.com/elixxir/xxdk-wasm/worker"
	"gitlab.com/xx_network/primitives/id"
)

// wasmModel implements [channels.EventModel] interface, which uses the channels
// system passed an object that adheres to in order to get events on the
// channel.
type wasmModel struct {
	wm *worker.Manager
}

// JoinChannel is called whenever a channel is joined locally.
func (w *wasmModel) JoinChannel(channel *cryptoBroadcast.Channel) {
	data, err := json.Marshal(channel)
	if err != nil {
		jww.ERROR.Printf(
			"[CH] Could not JSON marshal broadcast.Channel: %+v", err)
		return
	}

	if err = w.wm.SendNoResponse(JoinChannelTag, data); err != nil {
		jww.FATAL.Panicf("[CH] Failed to send to %q: %+v", JoinChannelTag, err)
	}
}

// LeaveChannel is called whenever a channel is left locally.
func (w *wasmModel) LeaveChannel(channelID *id.ID) {
	err := w.wm.SendNoResponse(LeaveChannelTag, channelID.Marshal())
	if err != nil {
		jww.FATAL.Panicf("[CH] Failed to send to %q: %+v", LeaveChannelTag, err)
	}
}

// ReceiveMessage is called whenever a message is received on a given channel.
//
// It may be called multiple times on the same message; it is incumbent on the
// user of the API to filter such called by message ID.
func (w *wasmModel) ReceiveMessage(channelID *id.ID, messageID message.ID,
	nickname, text string, pubKey ed25519.PublicKey, dmToken uint32,
	codeset uint8, timestamp time.Time, lease time.Duration, round rounds.Round,
	mType channels.MessageType, status channels.SentStatus, hidden bool) uint64 {
	msg := channels.ModelMessage{
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
		DmToken:        dmToken,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		jww.ERROR.Printf(
			"[CH] Could not JSON marshal payload for ReceiveMessage: %+v", err)
		return 0
	}

	response, err := w.wm.SendMessage(ReceiveMessageTag, data)
	if err != nil {
		jww.FATAL.Panicf(
			"[CH] Failed to send to %q: %+v", ReceiveMessageTag, err)
	}

	var uuid uint64
	if err = json.Unmarshal(response, &uuid); err != nil {
		jww.ERROR.Printf("[CH] Failed to JSON unmarshal UUID from worker for "+
			"%q: %+v", ReceiveMessageTag, err)
		return 0
	}

	return uuid
}

// ReceiveReplyMessage is JSON marshalled and sent to the worker for
// [wasmModel.ReceiveReply].
type ReceiveReplyMessage struct {
	ReactionTo            message.ID `json:"replyTo"`
	channels.ModelMessage `json:"message"`
}

// ReceiveReply is called whenever a message is received that is a reply on a
// given channel. It may be called multiple times on the same message; it is
// incumbent on the user of the API to filter such called by message ID.
//
// Messages may arrive our of order, so a reply, in theory, can arrive before
// the initial message. As a result, it may be important to buffer replies.
func (w *wasmModel) ReceiveReply(channelID *id.ID, messageID,
	replyTo message.ID, nickname, text string, pubKey ed25519.PublicKey,
	dmToken uint32, codeset uint8, timestamp time.Time, lease time.Duration,
	round rounds.Round, mType channels.MessageType, status channels.SentStatus,
	hidden bool) uint64 {
	msg := ReceiveReplyMessage{
		ReactionTo: replyTo,
		ModelMessage: channels.ModelMessage{
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
			DmToken:        dmToken,
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		jww.ERROR.Printf(
			"[CH] Could not JSON marshal payload for ReceiveReply: %+v", err)
		return 0
	}

	response, err := w.wm.SendMessage(ReceiveReplyTag, data)
	if err != nil {
		jww.FATAL.Panicf(
			"[CH] Failed to send to %q: %+v", ReceiveReplyTag, err)
	}

	var uuid uint64
	if err = json.Unmarshal(response, &uuid); err != nil {
		jww.ERROR.Printf("[CH] Failed to JSON unmarshal UUID from worker for "+
			"%q: %+v", ReceiveReplyTag, err)
		return 0
	}

	return uuid
}

// ReceiveReaction is called whenever a reaction to a message is received on a
// given channel. It may be called multiple times on the same reaction; it is
// incumbent on the user of the API to filter such called by message ID.
//
// Messages may arrive our of order, so a reply, in theory, can arrive before
// the initial message. As a result, it may be important to buffer reactions.
func (w *wasmModel) ReceiveReaction(channelID *id.ID, messageID,
	reactionTo message.ID, nickname, reaction string, pubKey ed25519.PublicKey,
	dmToken uint32, codeset uint8, timestamp time.Time, lease time.Duration,
	round rounds.Round, mType channels.MessageType, status channels.SentStatus,
	hidden bool) uint64 {

	msg := ReceiveReplyMessage{
		ReactionTo: reactionTo,
		ModelMessage: channels.ModelMessage{
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
			DmToken:        dmToken,
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		jww.ERROR.Printf(
			"[CH] Could not JSON marshal payload for ReceiveReaction: %+v", err)
		return 0
	}

	response, err := w.wm.SendMessage(ReceiveReactionTag, data)
	if err != nil {
		jww.FATAL.Panicf(
			"[CH] Failed to send to %q: %+v", ReceiveReactionTag, err)
	}

	var uuid uint64
	if err = json.Unmarshal(response, &uuid); err != nil {
		jww.ERROR.Printf("[CH] Failed to JSON unmarshal UUID from worker for "+
			"%q: %+v", ReceiveReactionTag, err)
		return 0
	}

	return uuid
}

// MessageUpdateInfo is JSON marshalled and sent to the worker for
// [wasmModel.UpdateFromMessageID] and [wasmModel.UpdateFromUUID].
type MessageUpdateInfo struct {
	UUID uint64 `json:"uuid"`

	MessageID    message.ID `json:"messageID"`
	MessageIDSet bool       `json:"messageIDSet"`

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

// UpdateFromUUID is called whenever a message at the UUID is modified.
//
// messageID, timestamp, round, pinned, and hidden are all nillable and may be
// updated based upon the UUID at a later date. If a nil value is passed, then
// make no update.
//
// Returns an error if the message cannot be updated. It must return
// [channels.NoMessageErr] if the message does not exist.
func (w *wasmModel) UpdateFromUUID(uuid uint64, messageID *message.ID,
	timestamp *time.Time, round *rounds.Round, pinned, hidden *bool,
	status *channels.SentStatus) error {
	msg := MessageUpdateInfo{UUID: uuid}
	if messageID != nil {
		msg.MessageID = *messageID
		msg.MessageIDSet = true
	}
	if timestamp != nil {
		msg.Timestamp = *timestamp
		msg.TimestampSet = true
	}
	if round != nil {
		msg.RoundID = round.ID
		msg.RoundIDSet = true
	}
	if pinned != nil {
		msg.Pinned = *pinned
		msg.PinnedSet = true
	}
	if hidden != nil {
		msg.Hidden = *hidden
		msg.HiddenSet = true
	}
	if status != nil {
		msg.Status = *status
		msg.StatusSet = true
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return errors.Errorf(
			"could not JSON marshal payload for UpdateFromUUID: %+v", err)
	}

	response, err := w.wm.SendMessage(UpdateFromUUIDTag, data)
	if err != nil {
		jww.FATAL.Panicf(
			"[CH] Failed to send to %q: %+v", UpdateFromUUIDTag, err)
	} else if len(response) > 0 {
		return errors.New(string(response))
	}

	return nil
}

// UuidError is JSON marshalled and sent to the worker for
// [wasmModel.UpdateFromMessageID].
type UuidError struct {
	UUID  uint64 `json:"uuid"`
	Error string `json:"error"`
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
func (w *wasmModel) UpdateFromMessageID(messageID message.ID,
	timestamp *time.Time, round *rounds.Round, pinned, hidden *bool,
	status *channels.SentStatus) (uint64, error) {

	msg := MessageUpdateInfo{MessageID: messageID, MessageIDSet: true}
	if timestamp != nil {
		msg.Timestamp = *timestamp
		msg.TimestampSet = true
	}
	if round != nil {
		msg.RoundID = round.ID
		msg.RoundIDSet = true
	}
	if pinned != nil {
		msg.Pinned = *pinned
		msg.PinnedSet = true
	}
	if hidden != nil {
		msg.Hidden = *hidden
		msg.HiddenSet = true
	}
	if status != nil {
		msg.Status = *status
		msg.StatusSet = true
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return 0, errors.Errorf("could not JSON marshal payload for "+
			"UpdateFromMessageID: %+v", err)
	}

	response, err := w.wm.SendMessage(UpdateFromMessageIDTag, data)
	if err != nil {
		jww.FATAL.Panicf(
			"[CH] Failed to send to %q: %+v", UpdateFromMessageIDTag, err)
	}

	var ue UuidError
	if err = json.Unmarshal(response, &ue); err != nil {
		return 0, errors.Errorf("could not JSON unmarshal response to %q: %+v",
			UpdateFromMessageIDTag, err)
	} else if len(ue.Error) > 0 {
		return 0, errors.New(ue.Error)
	} else {
		return ue.UUID, nil
	}
}

// GetMessageMessage is JSON marshalled and sent to the worker for
// [wasmModel.GetMessage].
type GetMessageMessage struct {
	Message channels.ModelMessage `json:"message"`
	Error   string                `json:"error"`
}

// GetMessage returns the message with the given [channel.MessageID].
func (w *wasmModel) GetMessage(
	messageID message.ID) (channels.ModelMessage, error) {

	response, err := w.wm.SendMessage(GetMessageTag, messageID.Marshal())
	if err != nil {
		jww.FATAL.Panicf(
			"[CH] Failed to send to %q: %+v", GetMessageTag, err)
	}

	var msg GetMessageMessage
	if err = json.Unmarshal(response, &msg); err != nil {
		return channels.ModelMessage{}, errors.Wrapf(err,
			"[CH] Could not JSON unmarshal response to %q", GetMessageTag)
	}

	if msg.Error != "" {
		return channels.ModelMessage{}, errors.New(msg.Error)
	}

	return msg.Message, nil
}

// DeleteMessage removes a message with the given messageID from storage.
func (w *wasmModel) DeleteMessage(messageID message.ID) error {
	response, err := w.wm.SendMessage(DeleteMessageTag, messageID.Marshal())
	if err != nil {
		jww.FATAL.Panicf(
			"[CH] Failed to send to %q: %+v", DeleteMessageTag, err)
	} else if len(response) > 0 {
		return errors.New(string(response))
	}

	return nil
}

// MuteUserMessage is JSON marshalled and sent to the worker for
// [wasmModel.MuteUser].
type MuteUserMessage struct {
	ChannelID *id.ID            `json:"channelID"`
	PubKey    ed25519.PublicKey `json:"pubKey"`
	Unmute    bool              `json:"unmute"`
}

// MuteUser is called whenever a user is muted or unmuted.
func (w *wasmModel) MuteUser(
	channelID *id.ID, pubKey ed25519.PublicKey, unmute bool) {
	msg := MuteUserMessage{
		ChannelID: channelID,
		PubKey:    pubKey,
		Unmute:    unmute,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		jww.ERROR.Printf("[CH] Could not marshal MuteUserMessage: %+v", err)
		return
	}

	err = w.wm.SendNoResponse(MuteUserTag, data)
	if err != nil {
		jww.FATAL.Panicf("[CH] Failed to send to %q: %+v", MuteUserTag, err)
	}
}
