////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package channelEventModel

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
	UUID       uint64            `json:"uuid"`
	MessageID  message.ID        `json:"messageID"`
	ReactionTo message.ID        `json:"reactionTo"`
	Nickname   string            `json:"nickname"`
	Text       []byte            `json:"text"`
	PubKey     ed25519.PublicKey `json:"pubKey"`
	DmToken    uint32            `json:"dmToken"`
	Codeset    uint8             `json:"codeset"`
	Timestamp  time.Time         `json:"timestamp"`
	Round      rounds.Round      `json:"round"`
	MType      dm.MessageType    `json:"mType"`
	Status     dm.Status         `json:"status"`
}

func (w *wasmModel) Receive(messageID message.ID, nickname string, text []byte,
	pubKey ed25519.PublicKey, dmToken uint32, codeset uint8, timestamp time.Time,
	round rounds.Round, mType dm.MessageType, status dm.Status) uint64 {
	msg := TransferMessage{
		MessageID: messageID,
		Nickname:  nickname,
		Text:      text,
		PubKey:    pubKey,
		DmToken:   dmToken,
		Codeset:   codeset,
		Timestamp: timestamp,
		Round:     round,
		MType:     mType,
		Status:    status,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		jww.ERROR.Printf(
			"Could not JSON marshal payload for TransferMessage: %+v", err)
		return 0
	}

	uuidChan := make(chan uint64)
	w.wh.SendMessage(ReceiveTag, data, func(data []byte) {
		var uuid uint64
		err = json.Unmarshal(data, &uuid)
		if err != nil {
			jww.ERROR.Printf(
				"Could not JSON unmarshal response to Receive: %+v", err)
			uuidChan <- 0
		}
		uuidChan <- uuid
	})

	select {
	case uuid := <-uuidChan:
		return uuid
	case <-time.After(worker.ResponseTimeout):
		jww.ERROR.Printf("Timed out after %s waiting for response from the "+
			"worker about Receive", worker.ResponseTimeout)
	}

	return 0
}

func (w *wasmModel) ReceiveText(messageID message.ID, nickname, text string,
	pubKey ed25519.PublicKey, dmToken uint32, codeset uint8,
	timestamp time.Time, round rounds.Round, status dm.Status) uint64 {
	msg := TransferMessage{
		MessageID: messageID,
		Nickname:  nickname,
		Text:      []byte(text),
		PubKey:    pubKey,
		DmToken:   dmToken,
		Codeset:   codeset,
		Timestamp: timestamp,
		Round:     round,
		Status:    status,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		jww.ERROR.Printf(
			"Could not JSON marshal payload for TransferMessage: %+v", err)
		return 0
	}

	uuidChan := make(chan uint64)
	w.wh.SendMessage(ReceiveTextTag, data, func(data []byte) {
		var uuid uint64
		err = json.Unmarshal(data, &uuid)
		if err != nil {
			jww.ERROR.Printf(
				"Could not JSON unmarshal response to ReceiveText: %+v", err)
			uuidChan <- 0
		}
		uuidChan <- uuid
	})

	select {
	case uuid := <-uuidChan:
		return uuid
	case <-time.After(worker.ResponseTimeout):
		jww.ERROR.Printf("Timed out after %s waiting for response from the "+
			"worker about ReceiveText", worker.ResponseTimeout)
	}

	return 0
}

func (w *wasmModel) ReceiveReply(messageID, reactionTo message.ID, nickname,
	text string, pubKey ed25519.PublicKey, dmToken uint32, codeset uint8,
	timestamp time.Time, round rounds.Round, status dm.Status) uint64 {
	msg := TransferMessage{
		MessageID:  messageID,
		ReactionTo: reactionTo,
		Nickname:   nickname,
		Text:       []byte(text),
		PubKey:     pubKey,
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

	uuidChan := make(chan uint64)
	w.wh.SendMessage(ReceiveReplyTag, data, func(data []byte) {
		var uuid uint64
		err = json.Unmarshal(data, &uuid)
		if err != nil {
			jww.ERROR.Printf(
				"Could not JSON unmarshal response to ReceiveReply: %+v", err)
			uuidChan <- 0
		}
		uuidChan <- uuid
	})

	select {
	case uuid := <-uuidChan:
		return uuid
	case <-time.After(worker.ResponseTimeout):
		jww.ERROR.Printf("Timed out after %s waiting for response from the "+
			"worker about ReceiveReply", worker.ResponseTimeout)
	}

	return 0
}

func (w *wasmModel) ReceiveReaction(messageID, reactionTo message.ID, nickname,
	reaction string, pubKey ed25519.PublicKey, dmToken uint32, codeset uint8,
	timestamp time.Time, round rounds.Round, status dm.Status) uint64 {
	msg := TransferMessage{
		MessageID:  messageID,
		ReactionTo: reactionTo,
		Nickname:   nickname,
		Text:       []byte(reaction),
		PubKey:     pubKey,
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

	uuidChan := make(chan uint64)
	w.wh.SendMessage(ReceiveReactionTag, data, func(data []byte) {
		var uuid uint64
		err = json.Unmarshal(data, &uuid)
		if err != nil {
			jww.ERROR.Printf(
				"Could not JSON unmarshal response to ReceiveReaction: %+v", err)
			uuidChan <- 0
		}
		uuidChan <- uuid
	})

	select {
	case uuid := <-uuidChan:
		return uuid
	case <-time.After(worker.ResponseTimeout):
		jww.ERROR.Printf("Timed out after %s waiting for response from the "+
			"worker about ReceiveReaction", worker.ResponseTimeout)
	}

	return 0
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

	w.wh.SendMessage(UpdateSentStatusTag, data, nil)
}

// BlockSender silences messages sent by the indicated sender
// public key.
func (w *wasmModel) BlockSender(senderPubKey ed25519.PublicKey) {
}

// UnblockSender allows messages sent by the indicated sender
// public key.
func (w *wasmModel) UnblockSender(senderPubKey ed25519.PublicKey) {
}

// GetCognversations returns any conversations held by the
// model (receiver)
func (w *wasmModel) GetConversations(
	senderPubKey ed25519.PublicKey) []dm.ModelConversation {
	return nil
}

// GetConversations returns any conversations held by the
// model (receiver)
func (w *wasmModel) GetAllConversations() []dm.ModelConversation {
	return nil
}
