////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package main

import (
	"bytes"
	"crypto/ed25519"
	"encoding/json"
	"strings"
	"syscall/js"
	"time"

	"github.com/hack-pad/go-indexeddb/idb"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"

	"gitlab.com/elixxir/client/v4/bindings"
	"gitlab.com/elixxir/client/v4/cmix/rounds"
	"gitlab.com/elixxir/client/v4/dm"
	idbCrypto "gitlab.com/elixxir/crypto/indexedDb"
	"gitlab.com/elixxir/crypto/message"
	"gitlab.com/elixxir/wasm-utils/utils"
	"gitlab.com/elixxir/xxdk-wasm/indexedDb/impl"
	"gitlab.com/xx_network/primitives/id"
	"gitlab.com/xx_network/primitives/netTime"
)

// wasmModel implements dm.EventModel interface backed by IndexedDb.
// NOTE: This model is NOT thread safe - it is the responsibility of the
// caller to ensure that its methods are called sequentially.
type wasmModel struct {
	db            *idb.Database
	cipher        idbCrypto.Cipher
	eventCallback eventUpdate
}

// upsertConversation is used for joining or updating a Conversation.
func (w *wasmModel) upsertConversation(nickname string,
	pubKey ed25519.PublicKey, partnerToken uint32, codeset uint8,
	blockedTimestamp *time.Time) error {
	parentErr := errors.New("[DM indexedDB] failed to upsertConversation")

	// Build object
	newConvo := Conversation{
		Pubkey:           pubKey,
		Nickname:         nickname,
		Token:            partnerToken,
		CodesetVersion:   codeset,
		BlockedTimestamp: blockedTimestamp,
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
func buildMessage(messageID, parentID []byte, text string, partnerKey []byte,
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
		dm.ReplyType, status)
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
	parentErr := errors.New("[DM indexedDB] failed to UpdateSentStatus")
	jww.TRACE.Printf(
		"[DM indexedDB] UpdateSentStatus(%d, %s, ...)", uuid, messageID)

	// Convert messageID to the key generated by json.Marshal
	key := js.ValueOf(uuid)

	// Use the key to get the existing Message
	currentMsg, err := impl.Get(w.db, messageStoreName, key)
	if err != nil {
		jww.ERROR.Printf("%+v", errors.WithMessagef(parentErr,
			"Unable to get message: %+v", err))
		return
	}

	// Extract the existing Message and update the Status
	newMessage, err := valueToMessage(currentMsg)
	if err != nil {
		jww.ERROR.Printf("%+v", err)
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
	_, err = w.upsertMessage(newMessage)
	if err != nil {
		jww.ERROR.Printf("%+v", errors.Wrap(parentErr, err.Error()))
		return
	}

	jww.TRACE.Printf("[DM indexedDB] Calling ReceiveMessageCB(%v, %v, t, f)",
		uuid, newMessage.ConversationPubKey)
	go w.eventCallback(bindings.DmMessageReceived, bindings.DmMessageReceivedJSON{
		UUID:               uuid,
		PubKey:             newMessage.ConversationPubKey,
		MessageUpdate:      true,
		ConversationUpdate: false,
	})
}

// receiveWrapper is a higher-level wrapper of upsertMessage.
func (w *wasmModel) receiveWrapper(messageID message.ID, parentID *message.ID, nickname,
	data string, partnerKey, senderKey ed25519.PublicKey, partnerToken uint32, codeset uint8,
	timestamp time.Time, round rounds.Round, mType dm.MessageType, status dm.Status) (uint64, error) {

	// Keep track of whether a Conversation was altered
	var convoToUpdate *Conversation

	// Determine whether Conversation needs to be created
	result, err := w.getConversation(partnerKey)
	if err != nil {
		if !strings.Contains(err.Error(), impl.ErrDoesNotExist) {
			return 0, err
		} else {
			// If sender key is not equal to partner key, then this is a convo being
			// created on a self sent message. This means that we shouldn't use the nickname
			// as it would be our own nickname.
			nick := nickname
			if !bytes.Equal(partnerKey, senderKey) {
				nick = ""
			}
			// If there is no extant Conversation, create one.
			jww.DEBUG.Printf(
				"[DM indexedDB] Joining conversation with %s", nickname)

			convoToUpdate = &Conversation{
				Pubkey:           partnerKey,
				Nickname:         nick,
				Token:            partnerToken,
				CodesetVersion:   codeset,
				BlockedTimestamp: nil,
			}
		}
	} else {
		jww.DEBUG.Printf(
			"[DM indexedDB] Conversation with %s already joined", nickname)

		// Update Conversation if nickname was altered
		isFromPartner := bytes.Equal(result.Pubkey, senderKey)
		nicknameChanged := result.Nickname != nickname
		if isFromPartner && nicknameChanged {
			jww.DEBUG.Printf(
				"[DM indexedDB] Updating from nickname %s to %s",
				result.Nickname, nickname)
			convoToUpdate = result
			convoToUpdate.Nickname = nickname
		}

		// Fix conversation if dmToken is altered
		dmTokenChanged := result.Token != partnerToken
		if isFromPartner && dmTokenChanged {
			jww.WARN.Printf(
				"[DM indexedDB] Updating from dmToken %d to %d",
				result.Token, partnerToken)
			convoToUpdate = result
			convoToUpdate.Token = partnerToken
		}
	}

	// Update the conversation in storage, if needed
	conversationUpdated := convoToUpdate != nil
	if conversationUpdated {
		err = w.upsertConversation(convoToUpdate.Nickname, convoToUpdate.Pubkey,
			convoToUpdate.Token, convoToUpdate.CodesetVersion, convoToUpdate.BlockedTimestamp)
		if err != nil {
			return 0, err
		}
	}

	// Handle encryption, if it is present
	if w.cipher != nil {
		data, err = w.cipher.Encrypt([]byte(data))
		if err != nil {
			return 0, err
		}
	}

	var parentIdBytes []byte
	if parentID != nil {
		parentIdBytes = parentID.Marshal()
	}

	msgToInsert := buildMessage(messageID.Bytes(), parentIdBytes, data,
		partnerKey, senderKey, timestamp, round.ID, mType, codeset, status)

	uuid, err := w.upsertMessage(msgToInsert)
	if err != nil {
		return 0, err
	}

	jww.TRACE.Printf("[DM indexedDB] Calling ReceiveMessageCB(%v, %v, f, %t)",
		uuid, partnerKey, conversationUpdated)
	go w.eventCallback(bindings.DmMessageReceived, bindings.DmMessageReceivedJSON{
		UUID:               uuid,
		PubKey:             partnerKey,
		MessageUpdate:      false,
		ConversationUpdate: conversationUpdated,
	})
	return uuid, nil
}

// upsertMessage is a helper function that will update an existing record
// if Message.ID is specified. Otherwise, it will perform an insert.
func (w *wasmModel) upsertMessage(msg *Message) (uint64, error) {
	// Convert to jsObject
	newMessageJson, err := json.Marshal(msg)
	if err != nil {
		return 0, errors.Errorf("Unable to marshal Message: %+v", err)
	}
	messageObj, err := utils.JsonToJS(newMessageJson)
	if err != nil {
		return 0, errors.Errorf("Unable to marshal Message: %+v", err)
	}

	// Store message to database
	msgIdObj, err := impl.Put(w.db, messageStoreName, messageObj)
	if err != nil {
		return 0, errors.Errorf("Unable to put Message: %+v\n%s",
			err, newMessageJson)
	}

	uuid := msgIdObj.Int()
	jww.DEBUG.Printf("[DM indexedDB] Successfully stored message %d", uuid)
	return uint64(uuid), nil
}

// BlockSender silences messages sent by the indicated sender
// public key.
func (w *wasmModel) BlockSender(senderPubKey ed25519.PublicKey) {
	parentErr := "failed to BlockSender"
	err := w.setBlocked(senderPubKey, true)
	if err != nil {
		jww.ERROR.Printf("%+v", errors.WithMessage(err, parentErr))
	}
}

// UnblockSender allows messages sent by the indicated sender
// public key.
func (w *wasmModel) UnblockSender(senderPubKey ed25519.PublicKey) {
	parentErr := "failed to UnblockSender"
	err := w.setBlocked(senderPubKey, false)
	if err != nil {
		jww.ERROR.Printf("%+v", errors.WithMessage(err, parentErr))
	}
}

// setBlocked is a helper for blocking/unblocking a given Conversation.
func (w *wasmModel) setBlocked(senderPubKey ed25519.PublicKey, isBlocked bool) error {
	// Get current Conversation and set blocked accordingly
	resultConvo, err := w.getConversation(senderPubKey)
	if err != nil {
		return err
	}

	var timeBlocked *time.Time
	if isBlocked {
		blockUser := netTime.Now()
		timeBlocked = &blockUser
	}

	return w.upsertConversation(resultConvo.Nickname, resultConvo.Pubkey,
		resultConvo.Token, resultConvo.CodesetVersion, timeBlocked)
}

// DeleteMessage deletes the message with the given message.ID belonging to
// the sender. If the message exists and belongs to the sender, then it is
// deleted and DeleteMessage returns true. If it does not exist, it returns
// false.
func (w *wasmModel) DeleteMessage(messageID message.ID, senderPubKey ed25519.PublicKey) bool {
	parentErr := "failed to DeleteMessage"
	msgId := impl.EncodeBytes(messageID.Marshal())

	// Use the key to get the existing Message
	currentMsg, err := impl.GetIndex(w.db, messageStoreName,
		messageStoreMessageIndex, msgId)
	if err != nil {
		jww.ERROR.Printf("%s: %+v", parentErr, err)
		return false
	}

	// Convert the js.Value to a proper object
	msgObj, err := valueToMessage(currentMsg)
	if err != nil {
		jww.ERROR.Printf("%s: %+v", parentErr, err)
		return false
	}

	// Ensure the public keys match
	if !bytes.Equal(msgObj.SenderPubKey, senderPubKey) {
		jww.ERROR.Printf("%s: %s", parentErr, "Public keys do not match")
		return false
	}

	// Perform the delete
	err = impl.DeleteIndex(w.db, messageStoreName, messageStoreMessageIndex,
		msgPkeyName, msgId)
	if err != nil {
		jww.ERROR.Printf("%s: %+v", parentErr, err)
		return false
	}

	go w.eventCallback(bindings.DmMessageDeleted, bindings.DmMessageDeletedJSON{
		MessageID: messageID,
	})
	return true
}

// GetConversation returns the conversation held by the model (receiver).
func (w *wasmModel) GetConversation(senderPubKey ed25519.PublicKey) *dm.ModelConversation {
	parentErr := "failed to GetConversation"
	resultConvo, err := w.getConversation(senderPubKey)
	if err != nil {
		jww.ERROR.Printf("%+v", errors.WithMessage(err, parentErr))
		return nil
	}

	return &dm.ModelConversation{
		Pubkey:           resultConvo.Pubkey,
		Nickname:         resultConvo.Nickname,
		Token:            resultConvo.Token,
		CodesetVersion:   resultConvo.CodesetVersion,
		BlockedTimestamp: resultConvo.BlockedTimestamp,
	}
}

// getConversation is a helper that returns the Conversation with the given senderPubKey.
func (w *wasmModel) getConversation(senderPubKey ed25519.PublicKey) (*Conversation, error) {
	resultObj, err := impl.Get(w.db, conversationStoreName, impl.EncodeBytes(senderPubKey))
	if err != nil {
		return nil, err
	}

	resultConvo := &Conversation{}
	err = json.Unmarshal([]byte(utils.JsToJson(resultObj)), resultConvo)
	if err != nil {
		return nil, err
	}
	return resultConvo, nil
}

// GetConversations returns any conversations held by the model (receiver).
func (w *wasmModel) GetConversations() []dm.ModelConversation {
	parentErr := "failed to GetConversations"

	results, err := impl.GetAll(w.db, conversationStoreName)
	if err != nil {
		jww.ERROR.Printf("%+v", errors.WithMessage(err, parentErr))
		return nil
	}

	conversations := make([]dm.ModelConversation, len(results))
	for i := range results {
		resultConvo := &Conversation{}
		err = json.Unmarshal([]byte(utils.JsToJson(results[i])), resultConvo)
		if err != nil {
			jww.ERROR.Printf("%+v", errors.WithMessage(err, parentErr))
			return nil
		}
		conversations[i] = dm.ModelConversation{
			Pubkey:           resultConvo.Pubkey,
			Nickname:         resultConvo.Nickname,
			Token:            resultConvo.Token,
			CodesetVersion:   resultConvo.CodesetVersion,
			BlockedTimestamp: resultConvo.BlockedTimestamp,
		}
	}
	return conversations
}

// valueToMessage is a helper for converting js.Value to Message.
func valueToMessage(msgObj js.Value) (*Message, error) {
	resultMsg := &Message{}
	return resultMsg, json.Unmarshal([]byte(utils.JsToJson(msgObj)), resultMsg)
}
