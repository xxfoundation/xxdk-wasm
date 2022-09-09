////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file                                                               //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm
// +build js,wasm

package indexedDb

import (
	"context"
	"encoding/json"
	"github.com/hack-pad/go-indexeddb/idb"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/xxdk-wasm/utils"
	"syscall/js"
	"time"

	"gitlab.com/elixxir/client/channels"
	"gitlab.com/elixxir/client/cmix/rounds"
	cryptoBroadcast "gitlab.com/elixxir/crypto/broadcast"
	cryptoChannel "gitlab.com/elixxir/crypto/channel"
	"gitlab.com/xx_network/primitives/id"
)

// dbTimeout is the global timeout for operations with the storage context.Contact
const dbTimeout = time.Second

// wasmModel implements [channels.EventModel] interface which uses the channels
// system passed an object which adheres to in order to get events on the channel.
type wasmModel struct {
	db *idb.Database
}

// newContext builds a context for database operations
func newContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), dbTimeout)
}

// JoinChannel is called whenever a channel is joined locally.
func (w *wasmModel) JoinChannel(channel *cryptoBroadcast.Channel) {
	parentErr := errors.New("failed to JoinChannel")

	// Build object
	newChannel := Channel{
		Id:          channel.ReceptionID.Marshal(),
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

	// Prepare the Transaction
	txn, err := w.db.Transaction(idb.TransactionReadWrite, channelsStoreName)
	if err != nil {
		jww.ERROR.Printf("%+v", errors.WithMessagef(parentErr,
			"Unable to create Transaction: %+v", err))
		return
	}
	store, err := txn.ObjectStore(channelsStoreName)
	if err != nil {
		jww.ERROR.Printf("%+v", errors.WithMessagef(parentErr,
			"Unable to get ObjectStore: %+v", err))
		return
	}

	// Perform the operation
	_, err = store.Add(channelObj)
	if err != nil {
		jww.ERROR.Printf("%+v", errors.WithMessagef(parentErr,
			"Unable to Add Channel: %+v", err))
		return
	}

	// Wait for the operation to return
	ctx, cancel := newContext()
	err = txn.Await(ctx)
	cancel()
	if err != nil {
		jww.ERROR.Printf("%+v", errors.WithMessagef(parentErr,
			"Adding Channel failed: %+v", err))
		return
	}
	jww.DEBUG.Printf("Successfully added channel: %s",
		channel.ReceptionID.String())
}

// LeaveChannel is called whenever a channel is left locally.
func (w *wasmModel) LeaveChannel(channelID *id.ID) {
	parentErr := errors.New("failed to LeaveChannel")

	// Prepare the Transaction
	txn, err := w.db.Transaction(idb.TransactionReadWrite, channelsStoreName)
	if err != nil {
		jww.ERROR.Printf("%+v", errors.WithMessagef(parentErr,
			"Unable to create Transaction: %+v", err))
		return
	}
	store, err := txn.ObjectStore(channelsStoreName)
	if err != nil {
		jww.ERROR.Printf("%+v", errors.WithMessagef(parentErr,
			"Unable to get ObjectStore: %+v", err))
		return
	}

	// Perform the operation
	_, err = store.Delete(js.ValueOf(channelID.String()))
	if err != nil {
		jww.ERROR.Printf("%+v", errors.WithMessagef(parentErr,
			"Unable to Delete Channel: %+v", err))
		return
	}

	// Wait for the operation to return
	ctx, cancel := newContext()
	err = txn.Await(ctx)
	cancel()
	if err != nil {
		jww.ERROR.Printf("%+v", errors.WithMessagef(parentErr,
			"Deleting Channel failed: %+v", err))
		return
	}
	jww.DEBUG.Printf("Successfully deleted channel: %s", channelID.String())
}

// ReceiveMessage is called whenever a message is received on a given channel
// It may be called multiple times on the same message, it is incumbent on
// the user of the API to filter such called by message ID.
func (w *wasmModel) ReceiveMessage(channelID *id.ID, messageID cryptoChannel.MessageID,
	senderUsername string, text string, timestamp time.Time, lease time.Duration,
	_ rounds.Round, status channels.SentStatus) {
	parentErr := errors.New("failed to ReceiveMessage")

	err := w.receiveHelper(buildMessage(channelID.Marshal(), messageID.Bytes(),
		nil, senderUsername, text, timestamp, lease, status))
	if err != nil {
		jww.ERROR.Printf("%+v", errors.Wrap(parentErr, err.Error()))
	}
}

// ReceiveReply is called whenever a message is received which is a reply
// on a given channel. It may be called multiple times on the same message,
// it is incumbent on the user of the API to filter such called by message ID
// Messages may arrive our of order, so a reply in theory can arrive before
// the initial message, as a result it may be important to buffer replies.
func (w *wasmModel) ReceiveReply(channelID *id.ID, messageID cryptoChannel.MessageID,
	replyTo cryptoChannel.MessageID, senderUsername string, text string,
	timestamp time.Time, lease time.Duration, _ rounds.Round, status channels.SentStatus) {
	parentErr := errors.New("failed to ReceiveReply")

	err := w.receiveHelper(buildMessage(channelID.Marshal(), messageID.Bytes(),
		replyTo.Bytes(), senderUsername, text, timestamp, lease, status))
	if err != nil {
		jww.ERROR.Printf("%+v", errors.Wrap(parentErr, err.Error()))
	}
}

// ReceiveReaction is called whenever a reaction to a message is received
// on a given channel. It may be called multiple times on the same reaction,
// it is incumbent on the user of the API to filter such called by message ID
// Messages may arrive our of order, so a reply in theory can arrive before
// the initial message, as a result it may be important to buffer reactions.
func (w *wasmModel) ReceiveReaction(channelID *id.ID, messageID cryptoChannel.MessageID,
	reactionTo cryptoChannel.MessageID, senderUsername string, reaction string,
	timestamp time.Time, lease time.Duration, _ rounds.Round, status channels.SentStatus) {
	parentErr := errors.New("failed to ReceiveReaction")

	err := w.receiveHelper(buildMessage(channelID.Marshal(), messageID.Bytes(),
		reactionTo.Bytes(), senderUsername, reaction, timestamp, lease, status))
	if err != nil {
		jww.ERROR.Printf("%+v", errors.Wrap(parentErr, err.Error()))
	}
}

// UpdateSentStatus is called whenever the SentStatus of a message
// has changed
func (w *wasmModel) UpdateSentStatus(messageID cryptoChannel.MessageID,
	status channels.SentStatus) {
	parentErr := errors.New("failed to UpdateSentStatus")
	newMessage := &Message{
		Id:     messageID.Bytes(),
		Status: uint8(status),
	}

	err := w.receiveHelper(newMessage)
	if err != nil {
		jww.ERROR.Printf("%+v", errors.Wrap(parentErr, err.Error()))
	}
}

// buildMessage is a private helper that converts typical [channels.EventModel]
// inputs into a basic Message structure for insertion into storage
func buildMessage(channelID []byte, messageID []byte,
	parentId []byte, senderUsername string, text string,
	timestamp time.Time, lease time.Duration, status channels.SentStatus) *Message {
	return &Message{
		Id:              messageID,
		SenderUsername:  senderUsername,
		ChannelId:       channelID,
		ParentMessageId: parentId,
		Timestamp:       timestamp,
		Lease:           lease,
		Status:          uint8(status),
		Hidden:          false,
		Pinned:          false,
		Text:            text,
	}
}

// receiveHelper is a private helper for receiving any sort of message
func (w *wasmModel) receiveHelper(newMessage *Message) error {
	// Convert to jsObject
	newMessageJson, err := json.Marshal(newMessage)
	if err != nil {
		return errors.Errorf("Unable to marshal Message: %+v", err)
	}
	messageObj, err := utils.JsonToJS(newMessageJson)
	if err != nil {
		return errors.Errorf("Unable to marshal Message: %+v", err)
	}

	// Prepare the Transaction
	txn, err := w.db.Transaction(idb.TransactionReadWrite, messageStoreName)
	if err != nil {
		return errors.Errorf("Unable to create Transaction: %+v", err)
	}
	store, err := txn.ObjectStore(messageStoreName)
	if err != nil {
		return errors.Errorf("Unable to get ObjectStore: %+v", err)
	}

	// Perform the upsert (put) operation
	_, err = store.Put(messageObj)
	if err != nil {
		return errors.Errorf("Unable to upsert Message: %+v", err)
	}

	// Wait for the operation to return
	ctx, cancel := newContext()
	err = txn.Await(ctx)
	cancel()
	if err != nil {
		return errors.Errorf("Upserting Message failed: %+v", err)
	}
	jww.DEBUG.Printf("Successfully stored message from %s",
		newMessage.SenderUsername)
	return nil
}

// dump given [idb.ObjectStore] contents to string slice for debugging purposes
func (w *wasmModel) dump(objectStoreName string) ([]string, error) {
	parentErr := errors.Errorf("failed to dump %s", objectStoreName)

	txn, err := w.db.Transaction(idb.TransactionReadOnly, objectStoreName)
	if err != nil {
		return nil, errors.WithMessagef(parentErr,
			"Unable to create Transaction: %+v", err)
	}
	store, err := txn.ObjectStore(objectStoreName)
	if err != nil {
		return nil, errors.WithMessagef(parentErr,
			"Unable to get ObjectStore: %+v", err)
	}
	cursorRequest, err := store.OpenCursor(idb.CursorNext)
	if err != nil {
		return nil, errors.WithMessagef(parentErr,
			"Unable to open Cursor: %+v", err)
	}

	// Run the query
	jww.DEBUG.Printf("%s values:", objectStoreName)
	results := make([]string, 0)
	ctx, cancel := newContext()
	err = cursorRequest.Iter(ctx, func(cursor *idb.CursorWithValue) error {
		value, err := cursor.Value()
		if err != nil {
			return err
		}
		valueStr := utils.JsToJson(value)
		results = append(results, valueStr)
		jww.DEBUG.Printf("- %v", valueStr)
		return nil
	})
	cancel()
	if err != nil {
		return nil, errors.WithMessagef(parentErr,
			"Unable to dump ObjectStore: %+v", err)
	}
	return results, nil
}
