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

// jsObject is the Golang type translation for a JavaScript object
type jsObject map[string]interface{}

// wasmModel implements [channels.EventModel] interface which uses the channels
// system passed an object which adheres to in order to get events on the channel.
type wasmModel struct {
	db *idb.Database
}

// JoinChannel is called whenever a channel is joined locally.
func (w *wasmModel) JoinChannel(channel *cryptoBroadcast.Channel) {
	parentErr := errors.New("failed to JoinChannel")

	// Build Channel object
	newChannel := Channel{
		Id:          channel.ReceptionID.Marshal(),
		Name:        channel.Name,
		Description: channel.Description,
	}

	// Convert Channel to jsObject
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
	ctx := context.Background()
	txn, err := w.db.Transaction(idb.TransactionReadWrite, messageStoreName)
	if err != nil {
		jww.ERROR.Printf("%+v", errors.Wrapf(parentErr,
			"Unable to create Transaction: %+v", err))
		return
	}
	store, err := txn.ObjectStore(messageStoreName)
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
	ctx := context.Background()
	txn, err := w.db.Transaction(idb.TransactionReadWrite, messageStoreName)
	if err != nil {
		jww.ERROR.Printf("%+v", errors.Wrapf(parentErr,
			"Unable to create Transaction: %+v", err))
		return
	}
	store, err := txn.ObjectStore(messageStoreName)
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
	senderUsername string, text string,
	timestamp time.Time, lease time.Duration, round rounds.Round) {
}

// ReceiveReply is called whenever a message is received which is a reply
// on a given channel. It may be called multiple times on the same message,
// it is incumbent on the user of the API to filter such called by message ID
// Messages may arrive our of order, so a reply in theory can arrive before
// the initial message, as a result it may be important to buffer replies.
func (w *wasmModel) ReceiveReply(ChannelID *id.ID, messageID cryptoChannel.MessageID,
	SenderUsername string, text string, timestamp time.Time,
	lease time.Duration, round rounds.Round) {

}

// ReceiveReaction is called whenever a reaction to a message is received
// on a given channel. It may be called multiple times on the same reaction,
// it is incumbent on the user of the API to filter such called by message ID
// Messages may arrive our of order, so a reply in theory can arrive before
// the initial message, as a result it may be important to buffer reactions.
func (w *wasmModel) ReceiveReaction(channelID *id.ID, messageID cryptoChannel.MessageID,
	reactionTo cryptoChannel.MessageID, senderUsername string,
	reaction string, timestamp time.Time, lease time.Duration,
	round rounds.Round) {
}

// MessageSent is called whenever the user sends a message. It should be
//designated as "sent" and that delivery is unknown.
func (w *wasmModel) MessageSent(channelID *id.ID, messageID cryptoChannel.MessageID,
	myUsername string, text string, timestamp time.Time,
	lease time.Duration, round rounds.Round) {

}

// ReplySent is called whenever the user sends a reply. It should be
// designated as "sent" and that delivery is unknown.
func (w *wasmModel) ReplySent(channelID *id.ID, messageID cryptoChannel.MessageID,
	replyTo cryptoChannel.MessageID, myUsername string, text string,
	timestamp time.Time, lease time.Duration, round rounds.Round) {

}

// ReactionSent is called whenever the user sends a reply. It should be
// designated as "sent" and that delivery is unknown.
func (w *wasmModel) ReactionSent(channelID *id.ID, messageID cryptoChannel.MessageID,
	reactionTo cryptoChannel.MessageID, senderUsername string,
	reaction string, timestamp time.Time, lease time.Duration,
	round rounds.Round) {

}

// UpdateSentStatus is called whenever the sent status of a message
// has changed
func (w *wasmModel) UpdateSentStatus(messageID cryptoChannel.MessageID,
	status channels.SentStatus) {

}
