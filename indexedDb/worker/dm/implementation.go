////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package dm

import (
	"crypto/ed25519"
	"encoding/json"
	"time"

	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/v4/cmix/rounds"
	"gitlab.com/elixxir/client/v4/dm"
	"gitlab.com/elixxir/crypto/message"
	"gitlab.com/elixxir/xxdk-wasm/worker"
)

// wasmModel implements dm.EventModel interface, which uses the channels system
// passed an object that adheres to in order to get events on the channel.
type wasmModel struct {
	wh *worker.Manager
}

// TransferMessage is JSON marshalled and sent to the worker.
type TransferMessage struct {
	UUID       uint64            `json:"uuid,omitempty"`
	MessageID  message.ID        `json:"messageID,omitempty"`
	ReactionTo message.ID        `json:"reactionTo,omitempty"`
	Nickname   string            `json:"nickname,omitempty"`
	Text       []byte            `json:"text,omitempty"`
	PartnerKey ed25519.PublicKey `json:"partnerKey,omitempty"`
	SenderKey  ed25519.PublicKey `json:"senderKey,omitempty"`
	DmToken    uint32            `json:"dmToken,omitempty"`
	Codeset    uint8             `json:"codeset,omitempty"`
	Timestamp  time.Time         `json:"timestamp"`
	Round      rounds.Round      `json:"round"`
	MType      dm.MessageType    `json:"mType,omitempty"`
	Status     dm.Status         `json:"status,omitempty"`
}

func (w *wasmModel) Receive(messageID message.ID, nickname string, text []byte,
	partnerKey, senderKey ed25519.PublicKey, dmToken uint32, codeset uint8, timestamp time.Time,
	round rounds.Round, mType dm.MessageType, status dm.Status) uint64 {
	msg := TransferMessage{
		MessageID:  messageID,
		Nickname:   nickname,
		Text:       text,
		PartnerKey: partnerKey,
		SenderKey:  senderKey,
		DmToken:    dmToken,
		Codeset:    codeset,
		Timestamp:  timestamp,
		Round:      round,
		MType:      mType,
		Status:     status,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		jww.ERROR.Printf(
			"[DM] Could not JSON marshal payload for Receive: %+v", err)
		return 0
	}

	response, err := w.wh.SendMessage(ReceiveTag, data)
	if err != nil {
		jww.FATAL.Panicf("[DM] Failed to send to %q: %+v", ReceiveTag, err)
	}

	var uuid uint64
	if err = json.Unmarshal(response, &uuid); err != nil {
		jww.ERROR.Printf("[DM] Failed to JSON unmarshal UUID from worker for "+
			"%q: %+v", ReceiveTag, err)
		return 0
	}

	return uuid
}

func (w *wasmModel) ReceiveText(messageID message.ID, nickname, text string,
	partnerKey, senderKey ed25519.PublicKey, dmToken uint32, codeset uint8,
	timestamp time.Time, round rounds.Round, status dm.Status) uint64 {
	msg := TransferMessage{
		MessageID:  messageID,
		Nickname:   nickname,
		Text:       []byte(text),
		PartnerKey: partnerKey,
		SenderKey:  senderKey,
		DmToken:    dmToken,
		Codeset:    codeset,
		Timestamp:  timestamp,
		Round:      round,
		Status:     status,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		jww.ERROR.Printf(
			"Could not JSON marshal payload for TransferMessage: %+v", err)
		return 0
	}

	response, err := w.wh.SendMessage(ReceiveTextTag, data)
	if err != nil {
		jww.FATAL.Panicf("[DM] Failed to send to %q: %+v", ReceiveTextTag, err)
	}

	var uuid uint64
	if err = json.Unmarshal(response, &uuid); err != nil {
		jww.ERROR.Printf("[DM] Failed to JSON unmarshal UUID from worker for "+
			"%q: %+v", ReceiveTextTag, err)
		return 0
	}

	return uuid
}

func (w *wasmModel) ReceiveReply(messageID, reactionTo message.ID, nickname,
	text string, partnerKey, senderKey ed25519.PublicKey, dmToken uint32, codeset uint8,
	timestamp time.Time, round rounds.Round, status dm.Status) uint64 {
	msg := TransferMessage{
		MessageID:  messageID,
		ReactionTo: reactionTo,
		Nickname:   nickname,
		Text:       []byte(text),
		PartnerKey: partnerKey,
		SenderKey:  senderKey,
		DmToken:    dmToken,
		Codeset:    codeset,
		Timestamp:  timestamp,
		Round:      round,
		Status:     status,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		jww.ERROR.Printf(
			"Could not JSON marshal payload for TransferMessage: %+v", err)
		return 0
	}

	response, err := w.wh.SendMessage(ReceiveReplyTag, data)
	if err != nil {
		jww.FATAL.Panicf("[DM] Failed to send to %q: %+v", ReceiveReplyTag, err)
	}

	var uuid uint64
	if err = json.Unmarshal(response, &uuid); err != nil {
		jww.ERROR.Printf("[DM] Failed to JSON unmarshal UUID from worker for "+
			"%q: %+v", ReceiveReplyTag, err)
		return 0
	}

	return uuid
}

func (w *wasmModel) ReceiveReaction(messageID, reactionTo message.ID, nickname,
	reaction string, partnerKey, senderKey ed25519.PublicKey, dmToken uint32, codeset uint8,
	timestamp time.Time, round rounds.Round, status dm.Status) uint64 {
	msg := TransferMessage{
		MessageID:  messageID,
		ReactionTo: reactionTo,
		Nickname:   nickname,
		Text:       []byte(reaction),
		PartnerKey: partnerKey,
		SenderKey:  senderKey,
		DmToken:    dmToken,
		Codeset:    codeset,
		Timestamp:  timestamp,
		Round:      round,
		Status:     status,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		jww.ERROR.Printf(
			"Could not JSON marshal payload for TransferMessage: %+v", err)
		return 0
	}

	response, err := w.wh.SendMessage(ReceiveReactionTag, data)
	if err != nil {
		jww.FATAL.Panicf("[DM] Failed to send to %q: %+v", ReceiveReactionTag, err)
	}

	var uuid uint64
	if err = json.Unmarshal(response, &uuid); err != nil {
		jww.ERROR.Printf("[DM] Failed to JSON unmarshal UUID from worker for "+
			"%q: %+v", ReceiveReactionTag, err)
		return 0
	}

	return uuid
}

func (w *wasmModel) UpdateSentStatus(uuid uint64, messageID message.ID,
	timestamp time.Time, round rounds.Round, status dm.Status) {
	msg := TransferMessage{
		UUID:      uuid,
		MessageID: messageID,
		Timestamp: timestamp,
		Round:     round,
		Status:    status,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		jww.ERROR.Printf(
			"Could not JSON marshal payload for TransferMessage: %+v", err)
	}

	if err = w.wh.SendNoResponse(UpdateSentStatusTag, data); err != nil {
		jww.FATAL.Panicf("[DM] Failed to send to %q: %+v", UpdateSentStatusTag, err)
	}
}

func (w *wasmModel) DeleteMessage(
	messageID message.ID, senderPubKey ed25519.PublicKey) bool {
	msg := TransferMessage{
		MessageID: messageID,
		SenderKey: senderPubKey,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		jww.ERROR.Printf(
			"Could not JSON marshal payload for TransferMessage: %+v", err)
	}

	response, err := w.wh.SendMessage(DeleteMessageTag, data)
	if err != nil {
		jww.FATAL.Panicf("[DM] Failed to send to %q: %+v", DeleteMessageTag, err)
	} else if len(response) == 0 {
		jww.FATAL.Panicf(
			"[DM] Received empty response from %q", DeleteMessageTag)
	}

	return response[0] == 1
}

func (w *wasmModel) GetConversation(senderPubKey ed25519.PublicKey) *dm.ModelConversation {
	response, err := w.wh.SendMessage(GetConversationTag, senderPubKey)
	if err != nil {
		jww.FATAL.Panicf("[DM] Failed to send to %q: %+v", GetConversationTag, err)
	}

	var result dm.ModelConversation
	if err = json.Unmarshal(response, &result); err != nil {
		jww.ERROR.Printf("[DM] Failed to JSON unmarshal %T from worker for "+
			"%q: %+v", result, GetConversationTag, err)
		return nil
	}

	return &result
}

func (w *wasmModel) GetConversations() []dm.ModelConversation {
	response, err := w.wh.SendMessage(GetConversationsTag, nil)
	if err != nil {
		jww.FATAL.Panicf("[DM] Failed to send to %q: %+v", GetConversationsTag, err)
	}

	var result []dm.ModelConversation
	if err = json.Unmarshal(response, &result); err != nil {
		jww.ERROR.Printf("[DM] Failed to JSON unmarshal %T from worker for "+
			"%q: %+v", result, GetConversationsTag, err)
		return nil
	}

	return result
}
