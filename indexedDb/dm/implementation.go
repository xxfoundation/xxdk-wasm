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
	"strings"
	"sync"
	"syscall/js"
	"time"

	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/v4/cmix/rounds"
	"gitlab.com/elixxir/client/v4/dm"
	"gitlab.com/elixxir/xxdk-wasm/indexedDb"
	"gitlab.com/elixxir/xxdk-wasm/utils"
	"gitlab.com/xx_network/primitives/id"

	"github.com/hack-pad/go-indexeddb/idb"
	cryptoChannel "gitlab.com/elixxir/crypto/channel"
	"gitlab.com/elixxir/crypto/message"
)

// wasmModel implements [dm.Receiver] interface, which uses the channels
// system passed an object that adheres to in order to get events on the
// channel.
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

	_, err = indexedDb.Put(w.db, conversationStoreName, convoObj)
	if err != nil {
		return errors.WithMessagef(parentErr,
			"Unable to put Conversation: %+v", err)
	}
	return nil
}

// buildMessage is a private helper that converts typical [dm.Receiver]
// inputs into a basic Message structure for insertion into storage.
//
// NOTE: ID is not set inside this function because we want to use the
// autoincrement key by default. If you are trying to overwrite an existing
// message, then you need to set it manually yourself.
func buildMessage(messageID, parentID []byte, text []byte,
	pubKey ed25519.PublicKey, timestamp time.Time, round id.Round,
	mType dm.MessageType, status dm.Status) *Message {
	return &Message{
		MessageID:          messageID,
		ConversationPubKey: pubKey,
		ParentMessageID:    parentID,
		Timestamp:          timestamp,
		Status:             uint8(status),
		Text:               text,
		Type:               uint16(mType),
		Round:              uint64(round),
	}
}

func (w *wasmModel) Receive(messageID message.ID, nickname string, text []byte,
	pubKey ed25519.PublicKey, dmToken uint32, codeset uint8,
	timestamp time.Time, round rounds.Round, mType dm.MessageType,
	status dm.Status) uint64 {
	parentErr := errors.New("failed to Receive")

	// If there is no extant Conversation, create one.
	_, err := indexedDb.Get(
		w.db, conversationStoreName, utils.CopyBytesToJS(pubKey))
	if err != nil {
		if strings.Contains(err.Error(), indexedDb.ErrDoesNotExist) {
			err = w.joinConversation(nickname, pubKey, dmToken, codeset)
			if err != nil {
				jww.ERROR.Printf("%+v", err)
			}
		} else {
			jww.ERROR.Printf("%+v", errors.WithMessagef(parentErr,
				"Unable to get Conversation: %+v", err))
		}
		return 0
	} else {
		jww.DEBUG.Printf("Conversation with %s already joined", nickname)
	}

	// Handle encryption, if it is present
	if w.cipher != nil {
		text, err = w.cipher.Encrypt(text)
		if err != nil {
			jww.ERROR.Printf("Failed to encrypt Message: %+v", err)
			return 0
		}
	}

	msgToInsert := buildMessage(messageID.Bytes(), nil, text,
		pubKey, timestamp, round.ID, mType, status)
	uuid, err := w.receiveHelper(msgToInsert, false)
	if err != nil {
		jww.ERROR.Printf("Failed to receive Message: %+v", err)
	}

	go w.receivedMessageCB(uuid, pubKey, false)
	return uuid
}

func (w *wasmModel) ReceiveText(messageID message.ID, nickname, text string,
	pubKey ed25519.PublicKey, dmToken uint32, codeset uint8,
	timestamp time.Time, round rounds.Round, status dm.Status) uint64 {
	parentErr := errors.New("failed to ReceiveText")

	// If there is no extant Conversation, create one.
	_, err := indexedDb.Get(
		w.db, conversationStoreName, utils.CopyBytesToJS(pubKey))
	if err != nil {
		if strings.Contains(err.Error(), indexedDb.ErrDoesNotExist) {
			err = w.joinConversation(nickname, pubKey, dmToken, codeset)
			if err != nil {
				jww.ERROR.Printf("%+v", err)
			}
		} else {
			jww.ERROR.Printf("%+v", errors.WithMessagef(parentErr,
				"Unable to get Conversation: %+v", err))
		}
		return 0
	} else {
		jww.DEBUG.Printf("Conversation with %s already joined", nickname)
	}

	// Handle encryption, if it is present
	textBytes := []byte(text)
	if w.cipher != nil {
		textBytes, err = w.cipher.Encrypt(textBytes)
		if err != nil {
			jww.ERROR.Printf("Failed to encrypt Message: %+v", err)
			return 0
		}
	}

	msgToInsert := buildMessage(messageID.Bytes(), nil, textBytes,
		pubKey, timestamp, round.ID, dm.TextType, status)

	uuid, err := w.receiveHelper(msgToInsert, false)
	if err != nil {
		jww.ERROR.Printf("Failed to receive Message: %+v", err)
	}

	go w.receivedMessageCB(uuid, pubKey, false)
	return uuid
}

func (w *wasmModel) ReceiveReply(messageID, reactionTo message.ID, nickname,
	text string, pubKey ed25519.PublicKey, dmToken uint32, codeset uint8,
	timestamp time.Time, round rounds.Round, status dm.Status) uint64 {
	parentErr := errors.New("failed to ReceiveReply")

	// If there is no extant Conversation, create one.
	_, err := indexedDb.Get(
		w.db, conversationStoreName, utils.CopyBytesToJS(pubKey))
	if err != nil {
		if strings.Contains(err.Error(), indexedDb.ErrDoesNotExist) {
			err = w.joinConversation(nickname, pubKey, dmToken, codeset)
			if err != nil {
				jww.ERROR.Printf("%+v", err)
			}
		} else {
			jww.ERROR.Printf("%+v", errors.WithMessagef(parentErr,
				"Unable to get Conversation: %+v", err))
		}
		return 0
	} else {
		jww.DEBUG.Printf("Conversation with %s already joined", nickname)
	}

	// Handle encryption, if it is present
	textBytes := []byte(text)
	if w.cipher != nil {
		textBytes, err = w.cipher.Encrypt(textBytes)
		if err != nil {
			jww.ERROR.Printf("Failed to encrypt Message: %+v", err)
			return 0
		}
	}

	msgToInsert := buildMessage(messageID.Bytes(), reactionTo.Marshal(), textBytes,
		pubKey, timestamp, round.ID, dm.TextType, status)

	uuid, err := w.receiveHelper(msgToInsert, false)
	if err != nil {
		jww.ERROR.Printf("Failed to receive Message: %+v", err)
	}

	go w.receivedMessageCB(uuid, pubKey, false)
	return uuid
}

func (w *wasmModel) ReceiveReaction(messageID, _ message.ID, nickname,
	reaction string, pubKey ed25519.PublicKey, dmToken uint32, codeset uint8,
	timestamp time.Time, round rounds.Round, status dm.Status) uint64 {
	parentErr := errors.New("failed to ReceiveText")

	// If there is no extant Conversation, create one.
	_, err := indexedDb.Get(
		w.db, conversationStoreName, utils.CopyBytesToJS(pubKey))
	if err != nil {
		if strings.Contains(err.Error(), indexedDb.ErrDoesNotExist) {
			err = w.joinConversation(nickname, pubKey, dmToken, codeset)
			if err != nil {
				jww.ERROR.Printf("%+v", err)
			}
		} else {
			jww.ERROR.Printf("%+v", errors.WithMessagef(parentErr,
				"Unable to get Conversation: %+v", err))
		}
		return 0
	} else {
		jww.DEBUG.Printf("Conversation with %s already joined", nickname)
	}

	// Handle encryption, if it is present
	textBytes := []byte(reaction)
	if w.cipher != nil {
		textBytes, err = w.cipher.Encrypt(textBytes)
		if err != nil {
			jww.ERROR.Printf("Failed to encrypt Message: %+v", err)
			return 0
		}
	}

	msgToInsert := buildMessage(messageID.Bytes(), nil, textBytes,
		pubKey, timestamp, round.ID, dm.ReactionType, status)

	uuid, err := w.receiveHelper(msgToInsert, false)
	if err != nil {
		jww.ERROR.Printf("Failed to receive Message: %+v", err)
	}

	go w.receivedMessageCB(uuid, pubKey, false)
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

	// Convert messageID to the key generated by json.Marshal
	key := js.ValueOf(uuid)

	// Use the key to get the existing Message
	currentMsg, err := indexedDb.Get(w.db, messageStoreName, key)
	if err != nil {
		jww.ERROR.Printf("%+v", errors.WithMessagef(parentErr,
			"Unable to get message: %+v", err))
		return
	}

	// Extract the existing Message and update the Status
	newMessage := &Message{}
	err = json.Unmarshal([]byte(utils.JsToJson(currentMsg)), newMessage)
	if err != nil {
		jww.ERROR.Printf("%+v", errors.WithMessagef(parentErr,
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
		jww.ERROR.Printf("%+v", errors.Wrap(parentErr, err.Error()))
	}
	go w.receivedMessageCB(uuid, newMessage.ConversationPubKey, true)
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
	result, err := indexedDb.Put(w.db, messageStoreName, messageObj)
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
			return uuid, nil
		}
		return 0, errors.Errorf("uuid lookup failure: %+v", err)
	}
	uuid := uint64(result.Int())
	jww.DEBUG.Printf("Successfully stored message %d", uuid)

	return uuid, nil
}

// msgIDLookup gets the UUID of the Message with the given messageID.
func (w *wasmModel) msgIDLookup(messageID message.ID) (uint64, error) {
	resultObj, err := indexedDb.GetIndex(w.db, messageStoreName,
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
