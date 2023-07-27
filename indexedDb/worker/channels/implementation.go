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

	w.wm.SendMessage(JoinChannelTag, data, nil)
}

// LeaveChannel is called whenever a channel is left locally.
func (w *wasmModel) LeaveChannel(channelID *id.ID) {
	w.wm.SendMessage(LeaveChannelTag, channelID.Marshal(), nil)
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

	uuidChan := make(chan uint64)
	w.wm.SendMessage(ReceiveMessageTag, data, func(data []byte) {
		var uuid uint64
		err = json.Unmarshal(data, &uuid)
		if err != nil {
			jww.ERROR.Printf("[CH] Could not JSON unmarshal response to "+
				"ReceiveMessage: %+v", err)
			uuidChan <- 0
		}
		uuidChan <- uuid
	})

	select {
	case uuid := <-uuidChan:
		return uuid
	case <-time.After(worker.ResponseTimeout):
		jww.ERROR.Printf("[CH] Timed out after %s waiting for response from "+
			"the worker about ReceiveMessage", worker.ResponseTimeout)
	}

	return 0
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

	uuidChan := make(chan uint64)
	w.wm.SendMessage(ReceiveReplyTag, data, func(data []byte) {
		var uuid uint64
		err = json.Unmarshal(data, &uuid)
		if err != nil {
			jww.ERROR.Printf("[CH] Could not JSON unmarshal response to "+
				"ReceiveReply: %+v", err)
			uuidChan <- 0
		}
		uuidChan <- uuid
	})

	select {
	case uuid := <-uuidChan:
		return uuid
	case <-time.After(worker.ResponseTimeout):
		jww.ERROR.Printf("[CH] Timed out after %s waiting for response from "+
			"the worker about ReceiveReply", worker.ResponseTimeout)
	}

	return 0
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

	uuidChan := make(chan uint64)
	w.wm.SendMessage(ReceiveReactionTag, data, func(data []byte) {
		var uuid uint64
		err = json.Unmarshal(data, &uuid)
		if err != nil {
			jww.ERROR.Printf("[CH] Could not JSON unmarshal response to "+
				"ReceiveReaction: %+v", err)
			uuidChan <- 0
		}
		uuidChan <- uuid
	})

	select {
	case uuid := <-uuidChan:
		return uuid
	case <-time.After(worker.ResponseTimeout):
		jww.ERROR.Printf("[CH] Timed out after %s waiting for response from "+
			"the worker about ReceiveReply", worker.ResponseTimeout)
	}

	return 0
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

	errChan := make(chan error)
	w.wm.SendMessage(UpdateFromUUIDTag, data, func(data []byte) {
		if data != nil {
			errChan <- errors.New(string(data))
		} else {
			errChan <- nil
		}
	})

	select {
	case err = <-errChan:
		return err
	case <-time.After(worker.ResponseTimeout):
		return errors.Errorf("timed out after %s waiting for response from "+
			"the worker about UpdateFromUUID", worker.ResponseTimeout)
	}
}

// UuidError is JSON marshalled and sent to the worker for
// [wasmModel.UpdateFromMessageID].
type UuidError struct {
	UUID  uint64 `json:"uuid"`
	Error []byte `json:"error"`
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

	uuidChan := make(chan uint64)
	errChan := make(chan error)
	w.wm.SendMessage(UpdateFromMessageIDTag, data,
		func(data []byte) {
			var ue UuidError
			if err = json.Unmarshal(data, &ue); err != nil {
				errChan <- errors.Errorf("could not JSON unmarshal response "+
					"to UpdateFromMessageID: %+v", err)
			} else if ue.Error != nil {
				errChan <- errors.New(string(ue.Error))
			} else {
				uuidChan <- ue.UUID
			}
		})

	select {
	case uuid := <-uuidChan:
		return uuid, nil
	case err = <-errChan:
		return 0, err
	case <-time.After(worker.ResponseTimeout):
		return 0, errors.Errorf("timed out after %s waiting for response from "+
			"the worker about UpdateFromMessageID", worker.ResponseTimeout)
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
	msgChan := make(chan GetMessageMessage)
	w.wm.SendMessage(GetMessageTag, messageID.Marshal(),
		func(data []byte) {
			var msg GetMessageMessage
			err := json.Unmarshal(data, &msg)
			if err != nil {
				jww.ERROR.Printf("[CH] Could not JSON unmarshal response to "+
					"GetMessage: %+v", err)
			}
			msgChan <- msg
		})

	select {
	case msg := <-msgChan:
		if msg.Error != "" {
			return channels.ModelMessage{}, errors.New(msg.Error)
		}
		return msg.Message, nil
	case <-time.After(worker.ResponseTimeout):
		return channels.ModelMessage{}, errors.Errorf("timed out after %s "+
			"waiting for response from the worker about GetMessage",
			worker.ResponseTimeout)
	}
}

// DeleteMessage removes a message with the given messageID from storage.
func (w *wasmModel) DeleteMessage(messageID message.ID) error {
	errChan := make(chan error)
	w.wm.SendMessage(DeleteMessageTag, messageID.Marshal(),
		func(data []byte) {
			if data != nil {
				errChan <- errors.New(string(data))
			} else {
				errChan <- nil
			}
		})

	select {
	case err := <-errChan:
		return err
	case <-time.After(worker.ResponseTimeout):
		return errors.Errorf("timed out after %s waiting for response from "+
			"the worker about DeleteMessage", worker.ResponseTimeout)
	}
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

	w.wm.SendMessage(MuteUserTag, data, nil)
}
