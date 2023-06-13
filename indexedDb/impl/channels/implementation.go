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
	"strconv"
	"strings"
	"syscall/js"
	"time"

	"github.com/hack-pad/go-indexeddb/idb"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"

	"gitlab.com/elixxir/client/v4/bindings"
	"gitlab.com/elixxir/client/v4/channels"
	"gitlab.com/elixxir/client/v4/cmix/rounds"
	cryptoBroadcast "gitlab.com/elixxir/crypto/broadcast"
	idbCrypto "gitlab.com/elixxir/crypto/indexedDb"
	"gitlab.com/elixxir/crypto/message"
	"gitlab.com/elixxir/wasm-utils/utils"
	"gitlab.com/elixxir/xxdk-wasm/indexedDb/impl"
	"gitlab.com/xx_network/primitives/id"
)

// wasmModel implements [channels.EventModel] interface backed by IndexedDb.
// NOTE: This model is NOT thread safe - it is the responsibility of the
// caller to ensure that its methods are called sequentially.
type wasmModel struct {
	db          *idb.Database
	cipher      idbCrypto.Cipher
	eventUpdate func(eventType int64, jsonMarshallable any)
}

// JoinChannel is called whenever a channel is joined locally.
func (w *wasmModel) JoinChannel(channel *cryptoBroadcast.Channel) {
	parentErr := errors.New("failed to JoinChannel")

	// Build object
	newChannel := Channel{
		ID:          channel.ReceptionID.Marshal(),
		Name:        channel.Name,
		Description: channel.Description,
	}

	// Convert to jsObject
	newChannelJson, err := json.Marshal(&newChannel)
	if err != nil {
		jww.ERROR.Printf("%+v", errors.WithMessagef(parentErr,
			"Unable to marshal Channel: %+v", err))
		return
	}
	channelObj, err := utils.JsonToJS(newChannelJson)
	if err != nil {
		jww.ERROR.Printf("%+v", errors.WithMessagef(parentErr,
			"Unable to marshal Channel: %+v", err))
		return
	}

	_, err = impl.Put(w.db, channelStoreName, channelObj)
	if err != nil {
		jww.ERROR.Printf("%+v", errors.WithMessagef(parentErr,
			"Unable to put Channel: %+v", err))
	}
}

// LeaveChannel is called whenever a channel is left locally.
func (w *wasmModel) LeaveChannel(channelID *id.ID) {
	parentErr := errors.New("failed to LeaveChannel")

	// Delete the channel from storage
	err := impl.Delete(w.db, channelStoreName, js.ValueOf(channelID.String()))
	if err != nil {
		jww.ERROR.Printf("%+v", errors.WithMessagef(parentErr,
			"Unable to delete Channel: %+v", err))
		return
	}

	// Clean up lingering data
	err = w.deleteMsgByChannel(channelID)
	if err != nil {
		jww.ERROR.Printf("%+v", errors.WithMessagef(parentErr,
			"Deleting Channel's Message data failed: %+v", err))
		return
	}
	jww.DEBUG.Printf("Successfully deleted channel: %s", channelID)
}

// deleteMsgByChannel is a private helper that uses messageStoreChannelIndex
// to delete all Message with the given Channel ID.
func (w *wasmModel) deleteMsgByChannel(channelID *id.ID) error {
	parentErr := errors.New("failed to deleteMsgByChannel")

	// Prepare the Transaction
	txn, err := w.db.Transaction(idb.TransactionReadWrite, messageStoreName)
	if err != nil {
		return errors.WithMessagef(parentErr,
			"Unable to create Transaction: %+v", err)
	}
	store, err := txn.ObjectStore(messageStoreName)
	if err != nil {
		return errors.WithMessagef(parentErr,
			"Unable to get ObjectStore: %+v", err)
	}
	index, err := store.Index(messageStoreChannelIndex)
	if err != nil {
		return errors.WithMessagef(parentErr,
			"Unable to get Index: %+v", err)
	}

	// Set up the operation
	keyRange, err := idb.NewKeyRangeOnly(impl.EncodeBytes(channelID.Marshal()))
	cursorRequest, err := index.OpenCursorRange(keyRange, idb.CursorNext)
	if err != nil {
		return errors.WithMessagef(parentErr, "Unable to open Cursor: %+v", err)
	}

	// Perform the operation
	err = impl.SendCursorRequest(cursorRequest,
		func(cursor *idb.CursorWithValue) error {
			_, err := cursor.Delete()
			return err
		})
	if err != nil {
		return errors.WithMessagef(parentErr,
			"Unable to delete Message data: %+v", err)
	}
	return nil
}

// ReceiveMessage is called whenever a message is received on a given channel.
//
// It may be called multiple times on the same message; it is incumbent on the
// user of the API to filter such called by message ID.
func (w *wasmModel) ReceiveMessage(channelID *id.ID, messageID message.ID,
	nickname, text string, pubKey ed25519.PublicKey, dmToken uint32,
	codeset uint8, timestamp time.Time, lease time.Duration, round rounds.Round,
	mType channels.MessageType, status channels.SentStatus, hidden bool) uint64 {
	var err error

	// Handle encryption, if it is present
	if w.cipher != nil {
		text, err = w.cipher.Encrypt([]byte(text))
		if err != nil {
			jww.ERROR.Printf("Failed to encrypt Message: %+v", err)
			return 0
		}
	}

	channelIDBytes := channelID.Marshal()

	msgToInsert := buildMessage(
		channelIDBytes, messageID.Bytes(), nil, nickname,
		text, pubKey, dmToken, codeset, timestamp, lease, round.ID, mType,
		false, hidden, status)

	uuid, err := w.upsertMessage(msgToInsert)
	if err != nil {
		jww.ERROR.Printf("Failed to receive Message: %+v", err)
		return 0
	}

	go w.eventUpdate(bindings.MessageReceived, bindings.MessageReceivedJson{
		Uuid:      int64(uuid),
		ChannelID: channelID,
		Update:    false,
	})
	return uuid
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
	var err error

	// Handle encryption, if it is present
	if w.cipher != nil {
		text, err = w.cipher.Encrypt([]byte(text))
		if err != nil {
			jww.ERROR.Printf("Failed to encrypt Message: %+v", err)
			return 0
		}
	}

	channelIDBytes := channelID.Marshal()

	msgToInsert := buildMessage(channelIDBytes, messageID.Bytes(),
		replyTo.Bytes(), nickname, text, pubKey, dmToken, codeset,
		timestamp, lease, round.ID, mType, hidden, false, status)

	uuid, err := w.upsertMessage(msgToInsert)
	if err != nil {
		jww.ERROR.Printf("Failed to receive reply: %+v", err)
		return 0
	}

	go w.eventUpdate(bindings.MessageReceived, bindings.MessageReceivedJson{
		Uuid:      int64(uuid),
		ChannelID: channelID,
		Update:    false,
	})
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
	var err error

	// Handle encryption, if it is present
	if w.cipher != nil {
		reaction, err = w.cipher.Encrypt([]byte(reaction))
		if err != nil {
			jww.ERROR.Printf("Failed to encrypt Message: %+v", err)
			return 0
		}
	}

	channelIDBytes := channelID.Marshal()
	msgToInsert := buildMessage(
		channelIDBytes, messageID.Bytes(), reactionTo.Bytes(), nickname,
		reaction, pubKey, dmToken, codeset, timestamp, lease, round.ID, mType,
		false, hidden, status)

	uuid, err := w.upsertMessage(msgToInsert)
	if err != nil {
		jww.ERROR.Printf("Failed to receive reaction: %+v", err)
		return 0
	}

	go w.eventUpdate(bindings.MessageReceived, bindings.MessageReceivedJson{
		Uuid:      int64(uuid),
		ChannelID: channelID,
		Update:    false,
	})
	return uuid
}

// UpdateFromUUID is called whenever a message at the UUID is modified.
//
// messageID, timestamp, round, pinned, and hidden are all nillable and may be
// updated based upon the UUID at a later date. If a nil value is passed, then
// make no update.
//
// Returns an error if the message cannot be updated. It must return
// channels.NoMessageErr if the message does not exist.
func (w *wasmModel) UpdateFromUUID(uuid uint64, messageID *message.ID,
	timestamp *time.Time, round *rounds.Round, pinned, hidden *bool,
	status *channels.SentStatus) error {
	parentErr := "failed to UpdateFromUUID"

	// Convert messageID to the key generated by json.Marshal
	key := js.ValueOf(uuid)

	// Use the key to get the existing Message
	msgObj, err := impl.Get(w.db, messageStoreName, key)
	if err != nil {
		if strings.Contains(err.Error(), impl.ErrDoesNotExist) {
			return errors.WithMessage(channels.NoMessageErr, parentErr)
		}
		return errors.WithMessage(err, parentErr)
	}

	currentMsg, err := valueToMessage(msgObj)
	if err != nil {
		return errors.WithMessagef(err,
			"%s Failed to marshal Message", parentErr)
	}

	_, err = w.updateMessage(currentMsg, messageID, timestamp,
		round, pinned, hidden, status)
	if err != nil {
		return errors.WithMessage(err, parentErr)
	}
	return nil
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
//
// Returns an error if the message cannot be updated. It must return
// channels.NoMessageErr if the message does not exist.
func (w *wasmModel) UpdateFromMessageID(messageID message.ID,
	timestamp *time.Time, round *rounds.Round, pinned, hidden *bool,
	status *channels.SentStatus) (uint64, error) {
	parentErr := "failed to UpdateFromMessageID"

	msgObj, err := impl.GetIndex(w.db, messageStoreName,
		messageStoreMessageIndex, impl.EncodeBytes(messageID.Marshal()))
	if err != nil {
		if strings.Contains(err.Error(), impl.ErrDoesNotExist) {
			return 0, errors.WithMessage(channels.NoMessageErr, parentErr)
		}
		return 0, errors.WithMessage(err, parentErr)
	}

	currentMsg, err := valueToMessage(msgObj)
	if err != nil {
		return 0, errors.WithMessagef(err,
			"%s Failed to marshal Message", parentErr)
	}

	uuid, err := w.updateMessage(currentMsg, &messageID, timestamp,
		round, pinned, hidden, status)
	if err != nil {
		return 0, errors.WithMessage(err, parentErr)
	}
	return uuid, nil
}

// buildMessage is a private helper that converts typical [channels.EventModel]
// inputs into a basic Message structure for insertion into storage.
//
// NOTE: ID is not set inside this function because we want to use the
// autoincrement key by default. If you are trying to overwrite an existing
// message, then you need to set it manually yourself.
func buildMessage(channelID, messageID, parentID []byte, nickname,
	text string, pubKey ed25519.PublicKey, dmToken uint32, codeset uint8,
	timestamp time.Time, lease time.Duration, round id.Round,
	mType channels.MessageType, pinned, hidden bool,
	status channels.SentStatus) *Message {
	return &Message{
		MessageID:       messageID,
		Nickname:        nickname,
		ChannelID:       channelID,
		ParentMessageID: parentID,
		Timestamp:       timestamp,
		Lease:           strconv.FormatInt(int64(lease), 10),
		Status:          uint8(status),
		Hidden:          hidden,
		Pinned:          pinned,
		Text:            text,
		Type:            uint16(mType),
		Round:           uint64(round),
		// User Identity Info
		Pubkey:         pubKey,
		DmToken:        dmToken,
		CodesetVersion: codeset,
	}
}

// updateMessage is a helper for updating a stored message.
func (w *wasmModel) updateMessage(currentMsg *Message, messageID *message.ID,
	timestamp *time.Time, round *rounds.Round, pinned, hidden *bool,
	status *channels.SentStatus) (uint64, error) {

	if status != nil {
		currentMsg.Status = uint8(*status)
	}
	if messageID != nil {
		currentMsg.MessageID = messageID.Bytes()
	}

	if round != nil {
		currentMsg.Round = uint64(round.ID)
	}

	if timestamp != nil {
		currentMsg.Timestamp = *timestamp
	}

	if pinned != nil {
		currentMsg.Pinned = *pinned
	}

	if hidden != nil {
		currentMsg.Hidden = *hidden
	}

	// Store the updated Message
	uuid, err := w.upsertMessage(currentMsg)
	if err != nil {
		return 0, err
	}

	channelID, err := id.Unmarshal(currentMsg.ChannelID)
	if err != nil {
		return 0, err
	}

	go w.eventUpdate(bindings.MessageReceived, bindings.MessageReceivedJson{
		Uuid:      int64(uuid),
		ChannelID: channelID,
		Update:    true,
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
		// Do not error out when this message already exists inside
		// the DB. Instead, set the ID and re-attempt as an update.
		if msg.ID == 0 { // always error out when not an insert attempt
			msgID, inErr := message.UnmarshalID(msg.MessageID)
			if inErr == nil {
				jww.WARN.Printf("upsertMessage duplicate: %+v",
					err)
				rnd := &rounds.Round{ID: id.Round(msg.Round)}
				status := (*channels.SentStatus)(&msg.Status)
				return w.UpdateFromMessageID(msgID,
					&msg.Timestamp,
					rnd,
					&msg.Pinned,
					&msg.Hidden,
					status)
			}
		}
		return 0, errors.Errorf("Unable to put Message: %+v\n%s",
			err, newMessageJson)
	}

	uuid := msgIdObj.Int()
	jww.DEBUG.Printf("Successfully stored message %d", uuid)
	return uint64(uuid), nil
}

// GetMessage returns the message with the given [channel.MessageID].
func (w *wasmModel) GetMessage(
	messageID message.ID) (channels.ModelMessage, error) {
	msgIDStr := impl.EncodeBytes(messageID.Marshal())

	resultObj, err := impl.GetIndex(w.db, messageStoreName,
		messageStoreMessageIndex, msgIDStr)
	if err != nil {
		return channels.ModelMessage{}, err
	}

	lookupResult, err := valueToMessage(resultObj)
	if err != nil {
		return channels.ModelMessage{}, err
	}

	var channelId *id.ID
	if lookupResult.ChannelID != nil {
		channelId, err = id.Unmarshal(lookupResult.ChannelID)
		if err != nil {
			return channels.ModelMessage{}, err
		}
	}

	var parentMsgId message.ID
	if lookupResult.ParentMessageID != nil {
		parentMsgId, err = message.UnmarshalID(lookupResult.ParentMessageID)
		if err != nil {
			return channels.ModelMessage{}, err
		}
	}

	lease := time.Duration(0)
	if len(lookupResult.Lease) > 0 {
		leaseInt, err := strconv.ParseInt(lookupResult.Lease, 10, 64)
		if err != nil {
			return channels.ModelMessage{}, err
		}
		lease = time.Duration(leaseInt)
	}

	return channels.ModelMessage{
		UUID:            lookupResult.ID,
		Nickname:        lookupResult.Nickname,
		MessageID:       messageID,
		ChannelID:       channelId,
		ParentMessageID: parentMsgId,
		Timestamp:       lookupResult.Timestamp,
		Lease:           lease,
		Status:          channels.SentStatus(lookupResult.Status),
		Hidden:          lookupResult.Hidden,
		Pinned:          lookupResult.Pinned,
		Content:         []byte(lookupResult.Text),
		Type:            channels.MessageType(lookupResult.Type),
		Round:           id.Round(lookupResult.Round),
		PubKey:          lookupResult.Pubkey,
		CodesetVersion:  lookupResult.CodesetVersion,
	}, nil
}

// DeleteMessage removes a message with the given messageID from storage.
func (w *wasmModel) DeleteMessage(messageID message.ID) error {
	err := impl.DeleteIndex(w.db, messageStoreName,
		messageStoreMessageIndex, pkeyName, impl.EncodeBytes(messageID.Marshal()))
	if err != nil {
		return err
	}

	go w.eventUpdate(bindings.MessageDeleted,
		bindings.MessageDeletedJson{MessageID: messageID})

	return nil
}

// MuteUser is called whenever a user is muted or unmuted.
func (w *wasmModel) MuteUser(
	channelID *id.ID, pubKey ed25519.PublicKey, unmute bool) {

	go w.eventUpdate(bindings.UserMuted, bindings.UserMutedJson{
		ChannelID: channelID,
		PubKey:    pubKey,
		Unmute:    unmute,
	})
}

// valueToMessage is a helper for converting js.Value to Message.
func valueToMessage(msgObj js.Value) (*Message, error) {
	resultMsg := &Message{}
	return resultMsg, json.Unmarshal([]byte(utils.JsToJson(msgObj)), resultMsg)
}

// valueToFile is a helper for converting js.Value to File.
func valueToFile(fileObj js.Value) (*File, error) {
	resultFile := &File{}
	return resultFile, json.Unmarshal([]byte(utils.JsToJson(fileObj)), resultFile)
}
