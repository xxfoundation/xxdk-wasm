////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 Privategrity Corporation                                   /
//                                                                             /
// All rights reserved.                                                        /
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
	"gitlab.com/elixxir/client/channels"
	"syscall/js"
	"time"

	"gitlab.com/elixxir/client/cmix/rounds"
	cryptoBroadcast "gitlab.com/elixxir/crypto/broadcast"
	cryptoChannel "gitlab.com/elixxir/crypto/channel"
	"gitlab.com/xx_network/primitives/id"
)

// dbTimeout is the global timeout for operations with the storage context.Contact
const dbTimeout = time.Second

// jsObject is the Golang type translation for a JavaScript object
type jsObject map[string]interface{}

// wasmModel implements [channels.EventModel] interface which uses the channels
// system passed an object which adheres to in order to get events on the channel.
type wasmModel struct {
	db *idb.Database
}

// newContext builds a context for database operations
func newContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	return ctx, cancel
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
		jww.ERROR.Printf("%+v", errors.Wrapf(parentErr,
			"Unable to marshal Channel: %+v", err))
		return
	}
	var channelObj *jsObject
	err = json.Unmarshal(newChannelJson, channelObj)
	if err != nil {
		jww.ERROR.Printf("%+v", errors.Wrapf(parentErr,
			"Unable to unmarshal Channel: %+v", err))
		return
	}

	// Prepare the Transaction
	ctx, cancel := newContext()
	defer cancel()
	txn, err := w.db.Transaction(idb.TransactionReadWrite, channelsStoreName)
	if err != nil {
		jww.ERROR.Printf("%+v", errors.Wrapf(parentErr,
			"Unable to create Transaction: %+v", err))
		return
	}
	store, err := txn.ObjectStore(channelsStoreName)
	if err != nil {
		jww.ERROR.Printf("%+v", errors.Wrapf(parentErr,
			"Unable to get ObjectStore: %+v", err))
		return
	}

	// Perform the operation
	_, err = store.Add(js.ValueOf(*channelObj))
	if err != nil {
		jww.ERROR.Printf("%+v", errors.Wrapf(parentErr,
			"Unable to Add Channel: %+v", err))
		return
	}

	// Wait for the operation to return
	err = txn.Await(ctx)
	if err != nil {
		jww.ERROR.Printf("%+v", errors.Wrapf(parentErr,
			"Adding Channel failed: %+v", err))
		return
	}
}

// LeaveChannel is called whenever a channel is left locally.
func (w *wasmModel) LeaveChannel(channelID *id.ID) {
	parentErr := errors.New("failed to LeaveChannel")

	// Prepare the Transaction
	ctx, cancel := newContext()
	defer cancel()
	txn, err := w.db.Transaction(idb.TransactionReadWrite, channelsStoreName)
	if err != nil {
		jww.ERROR.Printf("%+v", errors.Wrapf(parentErr,
			"Unable to create Transaction: %+v", err))
		return
	}
	store, err := txn.ObjectStore(channelsStoreName)
	if err != nil {
		jww.ERROR.Printf("%+v", errors.Wrapf(parentErr,
			"Unable to get ObjectStore: %+v", err))
		return
	}

	// Perform the operation
	_, err = store.Delete(js.ValueOf(channelID.String()))
	if err != nil {
		jww.ERROR.Printf("%+v", errors.Wrapf(parentErr,
			"Unable to Delete Channel: %+v", err))
		return
	}

	// Wait for the operation to return
	err = txn.Await(ctx)
	if err != nil {
		jww.ERROR.Printf("%+v", errors.Wrapf(parentErr,
			"Deleting Channel failed: %+v", err))
		return
	}
}

// ReceiveMessage is called whenever a message is received on a given channel
// It may be called multiple times on the same message, it is incumbent on
// the user of the API to filter such called by message ID.
func (w *wasmModel) ReceiveMessage(channelID *id.ID, messageID cryptoChannel.MessageID,
	senderUsername string, text string, timestamp time.Time, lease time.Duration,
	_ rounds.Round) {
	parentErr := errors.New("failed to ReceiveMessage")

	err := w.receiveHelper(buildMessage(channelID.Marshal(), messageID.Bytes(), nil,
		senderUsername, channels.Delivered, text, timestamp, lease))
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
	replyTo cryptoChannel.MessageID, senderUsername string,
	text string, timestamp time.Time, lease time.Duration, _ rounds.Round) {
	parentErr := errors.New("failed to ReceiveReply")

	err := w.receiveHelper(buildMessage(channelID.Marshal(), messageID.Bytes(),
		replyTo.Bytes(), senderUsername, channels.Delivered, text, timestamp, lease))
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
	reactionTo cryptoChannel.MessageID, senderUsername string,
	reaction string, timestamp time.Time, lease time.Duration, _ rounds.Round) {
	parentErr := errors.New("failed to ReceiveReaction")

	err := w.receiveHelper(buildMessage(channelID.Marshal(), messageID.Bytes(),
		reactionTo.Bytes(), senderUsername, channels.Delivered, reaction, timestamp, lease))
	if err != nil {
		jww.ERROR.Printf("%+v", errors.Wrap(parentErr, err.Error()))
	}
}

// MessageSent is called whenever the user sends a message. It should be
//designated as "sent" and that delivery is unknown.
func (w *wasmModel) MessageSent(channelID *id.ID, messageID cryptoChannel.MessageID,
	myUsername string, text string, timestamp time.Time,
	lease time.Duration, _ rounds.Round) {
	parentErr := errors.New("failed to MessageSent")

	err := w.receiveHelper(buildMessage(channelID.Marshal(), messageID.Bytes(),
		nil, myUsername, channels.Sent, text, timestamp, lease))
	if err != nil {
		jww.ERROR.Printf("%+v", errors.Wrap(parentErr, err.Error()))
	}
}

// ReplySent is called whenever the user sends a reply. It should be
// designated as "sent" and that delivery is unknown.
func (w *wasmModel) ReplySent(channelID *id.ID, messageID cryptoChannel.MessageID,
	replyTo cryptoChannel.MessageID, myUsername string, text string,
	timestamp time.Time, lease time.Duration, _ rounds.Round) {
	parentErr := errors.New("failed to ReplySent")

	err := w.receiveHelper(buildMessage(channelID.Marshal(), messageID.Bytes(),
		replyTo.Bytes(), myUsername, channels.Sent, text, timestamp, lease))
	if err != nil {
		jww.ERROR.Printf("%+v", errors.Wrap(parentErr, err.Error()))
	}
}

// ReactionSent is called whenever the user sends a reply. It should be
// designated as "sent" and that delivery is unknown.
func (w *wasmModel) ReactionSent(channelID *id.ID, messageID cryptoChannel.MessageID,
	reactionTo cryptoChannel.MessageID, myUsername string,
	reaction string, timestamp time.Time, lease time.Duration, _ rounds.Round) {
	parentErr := errors.New("failed to ReactionSent")

	err := w.receiveHelper(buildMessage(channelID.Marshal(), messageID.Bytes(),
		reactionTo.Bytes(), myUsername, channels.Sent, reaction, timestamp, lease))
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
	parentId []byte, senderUsername string, status channels.SentStatus,
	text string, timestamp time.Time, lease time.Duration) *Message {
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
	var messageObj *jsObject
	err = json.Unmarshal(newMessageJson, messageObj)
	if err != nil {
		return errors.Errorf("Unable to unmarshal Message: %+v", err)
	}

	// Prepare the Transaction
	ctx, cancel := newContext()
	defer cancel()
	txn, err := w.db.Transaction(idb.TransactionReadWrite, messageStoreName)
	if err != nil {
		return errors.Errorf("Unable to create Transaction: %+v", err)
	}
	store, err := txn.ObjectStore(messageStoreName)
	if err != nil {
		return errors.Errorf("Unable to get ObjectStore: %+v", err)
	}

	// Perform the upsert (put) operation
	_, err = store.Put(js.ValueOf(*messageObj))
	if err != nil {
		return errors.Errorf("Unable to upsert Message: %+v", err)
	}

	// Wait for the operation to return
	err = txn.Await(ctx)
	if err != nil {
		return errors.Errorf("Upserting Message failed: %+v", err)
	}
	return nil
}
