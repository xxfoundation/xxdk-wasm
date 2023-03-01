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
	"strings"
	"sync"
	"syscall/js"
	"time"

	"github.com/hack-pad/go-indexeddb/idb"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"

	"gitlab.com/elixxir/client/v4/cmix/rounds"
	"gitlab.com/elixxir/client/v4/dm"
	cryptoChannel "gitlab.com/elixxir/crypto/channel"
	"gitlab.com/elixxir/crypto/message"
	"gitlab.com/elixxir/xxdk-wasm/indexedDb/impl"
	"gitlab.com/elixxir/xxdk-wasm/utils"
	"gitlab.com/xx_network/primitives/id"
)

// wasmModel implements dm.EventModel interface, which uses the channels system
// passed an object that adheres to in order to get events on the channel.
type wasmModel struct {
	db                *idb.Database
	cipher            cryptoChannel.Cipher
	receivedMessageCB MessageReceivedCallback
	updateMux         sync.Mutex
}

// joinConversation is used for joining new conversations.
func (w *wasmModel) joinConversation(nickname string,
	pubKey ed25519.PublicKey, dmToken uint32, codeset uint8) error {
	parentErr := errors.New("failed to joinConversation")

	// Build object
	newConvo := Conversation{
		Pubkey:         pubKey,
		Nickname:       nickname,
		Token:          dmToken,
		CodesetVersion: codeset,
		Blocked:        false,
	}

	// Convert to jsObject
	newConvoJson, err := json.Marshal(&newConvo)
	if err != nil {
		return errors.WithMessagef(parentErr,
			"Unable to marshal Conversation: %+v", err)
	}
	convoObj, err := utils.JsonToJS(newConvoJson)
	if err != nil {
		return errors.WithMessagef(parentErr,
			"Unable to marshal Conversation: %+v", err)
	}

	_, err = impl.Put(w.db, conversationStoreName, convoObj)
	if err != nil {
		return errors.WithMessagef(parentErr,
			"Unable to put Conversation: %+v", err)
	}
	return nil
}

// buildMessage is a private helper that converts typical dm.EventModel inputs
// into a basic Message structure for insertion into storage.
//
// NOTE: ID is not set inside this function because we want to use the
// autoincrement key by default. If you are trying to overwrite an existing
// message, then you need to set it manually yourself.
func buildMessage(messageID, parentID, text []byte, partnerKey,
	senderKey ed25519.PublicKey, timestamp time.Time, round id.Round,
	mType dm.MessageType, codeset uint8, status dm.Status) *Message {
	return &Message{
		MessageID:          messageID,
		ConversationPubKey: partnerKey[:],
		ParentMessageID:    parentID,
		Timestamp:          timestamp,
		SenderPubKey:       senderKey[:],
		Status:             uint8(status),
		CodesetVersion:     codeset,
		Text:               text,
		Type:               uint16(mType),
		Round:              uint64(round),
	}
}

func (w *wasmModel) Receive(messageID message.ID, nickname string, text []byte,
	partnerKey, senderKey ed25519.PublicKey, dmToken uint32, codeset uint8, timestamp time.Time,
	round rounds.Round, mType dm.MessageType, status dm.Status) uint64 {
	parentErr := "[DM indexedDB] failed to Receive"
	jww.TRACE.Printf("[DM indexedDB] Receive(%s)", messageID)

	uuid, err := w.receiveWrapper(messageID, nil, nickname, string(text),
		partnerKey, senderKey, dmToken, codeset, timestamp, round, mType, status)
	if err != nil {
		jww.ERROR.Printf("%+v", errors.WithMessagef(err, parentErr))
		return 0
	}
	return uuid
}

func (w *wasmModel) ReceiveText(messageID message.ID, nickname, text string,
	partnerKey, senderKey ed25519.PublicKey, dmToken uint32, codeset uint8,
	timestamp time.Time, round rounds.Round, status dm.Status) uint64 {
	parentErr := "[DM indexedDB] failed to ReceiveText"
	jww.TRACE.Printf("[DM indexedDB] ReceiveText(%s)", messageID)

	uuid, err := w.receiveWrapper(messageID, nil, nickname, text,
		partnerKey, senderKey, dmToken, codeset, timestamp, round,
		dm.TextType, status)
	if err != nil {
		jww.ERROR.Printf("%+v", errors.WithMessagef(err, parentErr))
		return 0
	}
	return uuid
}

func (w *wasmModel) ReceiveReply(messageID, reactionTo message.ID, nickname,
	text string, partnerKey, senderKey ed25519.PublicKey, dmToken uint32, codeset uint8,
	timestamp time.Time, round rounds.Round, status dm.Status) uint64 {
	parentErr := "[DM indexedDB] failed to ReceiveReply"
	jww.TRACE.Printf("[DM indexedDB] ReceiveReply(%s)", messageID)

	uuid, err := w.receiveWrapper(messageID, &reactionTo, nickname, text,
		partnerKey, senderKey, dmToken, codeset, timestamp, round,
		dm.TextType, status)
	if err != nil {
		jww.ERROR.Printf("%+v", errors.WithMessagef(err, parentErr))
		return 0
	}
	return uuid
}

func (w *wasmModel) ReceiveReaction(messageID, reactionTo message.ID, nickname,
	reaction string, partnerKey, senderKey ed25519.PublicKey, dmToken uint32, codeset uint8,
	timestamp time.Time, round rounds.Round, status dm.Status) uint64 {
	parentErr := "[DM indexedDB] failed to ReceiveReaction"
	jww.TRACE.Printf("[DM indexedDB] ReceiveReaction(%s)", messageID)

	uuid, err := w.receiveWrapper(messageID, &reactionTo, nickname, reaction,
		partnerKey, senderKey, dmToken, codeset, timestamp, round,
		dm.ReactionType, status)
	if err != nil {
		jww.ERROR.Printf("%+v", errors.WithMessagef(err, parentErr))
		return 0
	}
	return uuid
}

func (w *wasmModel) UpdateSentStatus(uuid uint64, messageID message.ID,
	timestamp time.Time, round rounds.Round, status dm.Status) {
	parentErr := errors.New("failed to UpdateSentStatus")

	// FIXME: this is a bit of race condition without the mux.
	//        This should be done via the transactions (i.e., make a
	//        special version of receiveHelper)
	w.updateMux.Lock()
	defer w.updateMux.Unlock()
	jww.TRACE.Printf(
		"[DM indexedDB] UpdateSentStatus(%d, %s, ...)", uuid, messageID)

	// Convert messageID to the key generated by json.Marshal
	key := js.ValueOf(uuid)

	// Use the key to get the existing Message
	currentMsg, err := impl.Get(w.db, messageStoreName, key)
	if err != nil {
		jww.ERROR.Printf("[DM indexedDB] %+v", errors.WithMessagef(parentErr,
			"Unable to get message: %+v", err))
		return
	}

	// Extract the existing Message and update the Status
	newMessage := &Message{}
	err = json.Unmarshal([]byte(utils.JsToJson(currentMsg)), newMessage)
	if err != nil {
		jww.ERROR.Printf("[DM indexedDB] %+v", errors.WithMessagef(parentErr,
			"Could not JSON unmarshal message: %+v", err))
		return
	}

	newMessage.Status = uint8(status)
	if !messageID.Equals(message.ID{}) {
		newMessage.MessageID = messageID.Bytes()
	}

	if round.ID != 0 {
		newMessage.Round = uint64(round.ID)
	}

	if !timestamp.Equal(time.Time{}) {
		newMessage.Timestamp = timestamp
	}

	// Store the updated Message
	_, err = w.receiveHelper(newMessage, true)
	if err != nil {
		jww.ERROR.Printf("[DM indexedDB] %+v",
			errors.Wrap(parentErr, err.Error()))
		return
	}

	jww.TRACE.Printf("[DM indexedDB] Calling ReceiveMessageCB(%v, %v, t)",
		uuid, newMessage.ConversationPubKey)
	go w.receivedMessageCB(uuid, newMessage.ConversationPubKey, true)
}

// receiveWrapper is a higher-level wrapper of receiveHelper.
func (w *wasmModel) receiveWrapper(messageID message.ID, parentID *message.ID, nickname,
	data string, partnerKey, senderKey ed25519.PublicKey, dmToken uint32, codeset uint8,
	timestamp time.Time, round rounds.Round, mType dm.MessageType, status dm.Status) (uint64, error) {

	// If there is no extant Conversation, create one.
	_, err := impl.Get(w.db, conversationStoreName, js.ValueOf(partnerKey))
	if err != nil {
		if strings.Contains(err.Error(), impl.ErrDoesNotExist) {
			err = w.joinConversation(nickname, partnerKey, dmToken,
				codeset)
			if err != nil {
				return 0, err
			}
		} else {
			return 0, err
		}
	} else {
		jww.DEBUG.Printf(
			"[DM indexedDB] Conversation with %s already joined", nickname)
	}

	// Handle encryption, if it is present
	textBytes := []byte(data)
	if w.cipher != nil {
		textBytes, err = w.cipher.Encrypt(textBytes)
		if err != nil {
			return 0, err
		}
	}

	var parentIdBytes []byte
	if parentID != nil {
		parentIdBytes = parentID.Marshal()
	}

	msgToInsert := buildMessage(messageID.Bytes(), parentIdBytes, textBytes,
		partnerKey, senderKey, timestamp, round.ID, mType, codeset, status)

	uuid, err := w.receiveHelper(msgToInsert, false)
	if err != nil {
		return 0, err
	}

	jww.TRACE.Printf("[DM indexedDB] Calling ReceiveMessageCB(%v, %v, f)",
		uuid, partnerKey)
	go w.receivedMessageCB(uuid, partnerKey, false)
	return uuid, nil
}

// receiveHelper is a private helper for receiving any sort of message.
func (w *wasmModel) receiveHelper(
	newMessage *Message, isUpdate bool) (uint64, error) {
	// Convert to jsObject
	newMessageJson, err := json.Marshal(newMessage)
	if err != nil {
		return 0, errors.Errorf("Unable to marshal Message: %+v", err)
	}
	messageObj, err := utils.JsonToJS(newMessageJson)
	if err != nil {
		return 0, errors.Errorf("Unable to marshal Message: %+v", err)
	}

	// Unset the primaryKey for inserts so that it can be auto-populated and
	// incremented
	if !isUpdate {
		messageObj.Delete("id")
	}

	// Store message to database
	result, err := impl.Put(w.db, messageStoreName, messageObj)
	if err != nil {
		return 0, errors.Errorf("Unable to put Message: %+v", err)
	}

	// NOTE: Sometimes the insert fails to return an error but hits a duplicate
	//  insert, so this fallthrough returns the UUID entry in that case.
	if result.IsUndefined() {
		msgID := message.ID{}
		copy(msgID[:], newMessage.MessageID)
		uuid, errLookup := w.msgIDLookup(msgID)
		if uuid != 0 && errLookup == nil {
			jww.WARN.Printf("[DM indexedDB] Result undefined, but found"+
				" duplicate? %d, %s", uuid, msgID)
			return uuid, nil
		}
		return 0, errors.Errorf("uuid lookup failure: %+v", err)
	}
	uuid := uint64(result.Int())
	jww.DEBUG.Printf("[DM indexedDB] Successfully stored message %d", uuid)

	return uuid, nil
}

// msgIDLookup gets the UUID of the Message with the given messageID.
func (w *wasmModel) msgIDLookup(messageID message.ID) (uint64, error) {
	resultObj, err := impl.GetIndex(w.db, messageStoreName,
		messageStoreMessageIndex, utils.CopyBytesToJS(messageID.Marshal()))
	if err != nil {
		return 0, err
	}

	uuid := uint64(0)
	if !resultObj.IsUndefined() {
		uuid = uint64(resultObj.Get("id").Int())
	}
	return uuid, nil
}
