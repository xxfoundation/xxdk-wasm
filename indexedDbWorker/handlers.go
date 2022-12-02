////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package main

import (
	"encoding/json"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/v4/channels"
	"gitlab.com/elixxir/client/v4/cmix/rounds"
	cryptoBroadcast "gitlab.com/elixxir/crypto/broadcast"
	cryptoChannel "gitlab.com/elixxir/crypto/channel"
	"gitlab.com/elixxir/crypto/fastRNG"
	"gitlab.com/elixxir/xxdk-wasm/indexedDb"
	"gitlab.com/xx_network/crypto/csprng"
	"gitlab.com/xx_network/primitives/id"
	"time"
)

// registerHandlers registers all the reception handlers to manage messages from
// the main thread for the channels.EventModel.
func (mh *messageHandler) registerHandlers() {
	// Handler for NewWASMEventModel
	mh.registerHandler(indexedDb.NewWASMEventModelTag, func(data []byte) []byte {
		var message indexedDb.NewWASMEventModelMessage
		err := json.Unmarshal(data, &message)
		if err != nil {
			jww.ERROR.Printf("Could not JSON unmarshal "+
				"NewWASMEventModelMessage from main thread: %+v", err)
			return []byte{}
		}

		// Create new encryption cipher
		rng := fastRNG.NewStreamGenerator(12, 1024, csprng.NewSystemRNG)
		encryption, err := cryptoChannel.NewCipherFromJSON(
			[]byte(message.EncryptionJSON), rng.GetStream())
		if err != nil {
			jww.ERROR.Printf("Could not JSON unmarshal channel cipher from "+
				"main thread: %+v", err)
			return []byte{}
		}

		mh.model, err = NewWASMEventModel(message.Path, encryption,
			mh.messageReceivedCallback, mh.storeEncryptionStatus)
		if err != nil {
			return []byte(err.Error())
		}
		return []byte{}
	})

	// Handler for wasmModel.JoinChannel
	mh.registerHandler(indexedDb.JoinChannelTag, func(data []byte) []byte {
		var channel cryptoBroadcast.Channel
		err := json.Unmarshal(data, &channel)
		if err != nil {
			jww.ERROR.Printf("Could not JSON unmarshal broadcast.Channel from "+
				"main thread: %+v", err)
			return nil
		}

		mh.model.JoinChannel(&channel)
		return nil
	})

	// Handler for wasmModel.LeaveChannel
	mh.registerHandler(indexedDb.LeaveChannelTag, func(data []byte) []byte {
		channelID, err := id.Unmarshal(data)
		if err != nil {
			jww.ERROR.Printf(
				"Could not unmarshal channel ID from main thread: %+v", err)
			return nil
		}

		mh.model.LeaveChannel(channelID)
		return nil
	})

	// Handler for wasmModel.ReceiveMessage
	mh.registerHandler(indexedDb.ReceiveMessageTag, func(data []byte) []byte {
		var msg channels.ModelMessage
		err := json.Unmarshal(data, &msg)
		if err != nil {
			jww.ERROR.Printf("Could not JSON unmarshal channels.ModelMessage "+
				"from main thread: %+v", err)
			return nil
		}

		uuid := mh.model.ReceiveMessage(msg.ChannelID, msg.MessageID,
			msg.Nickname, string(msg.Content), msg.PubKey, msg.CodesetVersion,
			msg.Timestamp, msg.Lease, rounds.Round{ID: msg.Round}, msg.Type,
			msg.Status, msg.Hidden)

		uuidData, err := json.Marshal(uuid)
		if err != nil {
			jww.ERROR.Printf(
				"Could not JSON marshal UUID from ReceiveMessage: %+v", err)
			return nil
		}
		return uuidData
	})

	// Handler for wasmModel.ReceiveReply
	mh.registerHandler(indexedDb.ReceiveReplyTag, func(data []byte) []byte {
		var msg indexedDb.ReceiveReplyMessage
		err := json.Unmarshal(data, &msg)
		if err != nil {
			jww.ERROR.Printf("Could not JSON unmarshal ReceiveReplyMessage "+
				"from main thread: %+v", err)
			return nil
		}

		uuid := mh.model.ReceiveReply(msg.Message.ChannelID,
			msg.Message.MessageID, msg.ReactionTo, msg.Message.Nickname,
			string(msg.Message.Content), msg.Message.PubKey,
			msg.Message.CodesetVersion, msg.Message.Timestamp,
			msg.Message.Lease, rounds.Round{ID: msg.Message.Round},
			msg.Message.Type, msg.Message.Status, msg.Message.Hidden)

		uuidData, err := json.Marshal(uuid)
		if err != nil {
			jww.ERROR.Printf(
				"Could not JSON marshal UUID from ReceiveReply: %+v", err)
			return nil
		}
		return uuidData
	})

	// Handler for wasmModel.ReceiveReaction
	mh.registerHandler(indexedDb.ReceiveReactionTag, func(data []byte) []byte {
		var msg indexedDb.ReceiveReplyMessage
		err := json.Unmarshal(data, &msg)
		if err != nil {
			jww.ERROR.Printf("Could not JSON unmarshal ReceiveReplyMessage "+
				"from main thread: %+v", err)
			return nil
		}

		uuid := mh.model.ReceiveReaction(msg.Message.ChannelID,
			msg.Message.MessageID, msg.ReactionTo, msg.Message.Nickname,
			string(msg.Message.Content), msg.Message.PubKey,
			msg.Message.CodesetVersion, msg.Message.Timestamp,
			msg.Message.Lease, rounds.Round{ID: msg.Message.Round},
			msg.Message.Type, msg.Message.Status, msg.Message.Hidden)

		uuidData, err := json.Marshal(uuid)
		if err != nil {
			jww.ERROR.Printf(
				"Could not JSON marshal UUID from ReceiveReaction: %+v", err)
			return nil
		}
		return uuidData
	})

	// Handler for wasmModel.UpdateFromMessageID
	mh.registerHandler(indexedDb.UpdateFromMessageIDTag, func(data []byte) []byte {
		var msg indexedDb.MessageUpdateInfo
		err := json.Unmarshal(data, &msg)
		if err != nil {
			jww.ERROR.Printf("Could not JSON unmarshal MessageUpdateInfo "+
				"from main thread: %+v", err)
			return nil
		}
		var timestamp *time.Time
		var round *rounds.Round
		var pinned, hidden *bool
		var status *channels.SentStatus
		if msg.TimestampSet {
			timestamp = &msg.Timestamp
		}
		if msg.RoundIDSet {
			round = &rounds.Round{ID: msg.RoundID}
		}
		if msg.PinnedSet {
			pinned = &msg.Pinned
		}
		if msg.HiddenSet {
			hidden = &msg.Hidden
		}
		if msg.StatusSet {
			status = &msg.Status
		}

		uuid := mh.model.UpdateFromMessageID(
			msg.MessageID, timestamp, round, pinned, hidden, status)

		uuidData, err := json.Marshal(uuid)
		if err != nil {
			jww.ERROR.Printf(
				"Could not JSON marshal UUID from UpdateFromMessageID: %+v", err)
			return nil
		}
		return uuidData
	})

	// Handler for wasmModel.UpdateFromUUID
	mh.registerHandler(indexedDb.UpdateFromUUIDTag, func(data []byte) []byte {
		var msg indexedDb.MessageUpdateInfo
		err := json.Unmarshal(data, &msg)
		if err != nil {
			jww.ERROR.Printf("Could not JSON unmarshal MessageUpdateInfo "+
				"from main thread: %+v", err)
			return nil
		}
		var messageID *cryptoChannel.MessageID
		var timestamp *time.Time
		var round *rounds.Round
		var pinned, hidden *bool
		var status *channels.SentStatus
		if msg.MessageIDSet {
			messageID = &msg.MessageID
		}
		if msg.TimestampSet {
			timestamp = &msg.Timestamp
		}
		if msg.RoundIDSet {
			round = &rounds.Round{ID: msg.RoundID}
		}
		if msg.PinnedSet {
			pinned = &msg.Pinned
		}
		if msg.HiddenSet {
			hidden = &msg.Hidden
		}
		if msg.StatusSet {
			status = &msg.Status
		}

		mh.model.UpdateFromUUID(
			msg.UUID, messageID, timestamp, round, pinned, hidden, status)
		return nil
	})

	// Handler for wasmModel.GetMessage
	mh.registerHandler(indexedDb.GetMessageTag, func(data []byte) []byte {
		messageID, err := cryptoChannel.UnmarshalMessageID(data)
		if err != nil {
			jww.ERROR.Printf("Could not JSON unmarshal channel.MessageID "+
				"from main thread: %+v", err)
			return nil
		}

		reply := indexedDb.GetMessageMessage{}

		message, err := mh.model.GetMessage(messageID)
		if err != nil {
			reply.Error = err.Error()
		} else {
			reply.Message = message
		}

		messageData, err := json.Marshal(message)
		if err != nil {
			jww.ERROR.Printf(
				"Could not JSON marshal UUID from ReceiveReaction: %+v", err)
			return nil
		}
		return messageData
	})
}

// messageReceivedCallback sends calls to the indexedDb.MessageReceivedCallback
// in the main thread.
//
// storeEncryptionStatus adhere to the indexedDb.MessageReceivedCallback type.
func (mh *messageHandler) messageReceivedCallback(
	uuid uint64, channelID *id.ID, update bool) {
	// Package parameters for sending
	msg := &indexedDb.MessageReceivedCallbackMessage{
		UUID:      uuid,
		ChannelID: channelID,
		Update:    update,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		jww.ERROR.Printf(
			"Could not JSON marshal MessageReceivedCallbackMessage: %+v", err)
		return
	}

	// Send it to the main thread
	mh.sendResponse(indexedDb.GetMessageTag, indexedDb.InitID, data)
}

// storeEncryptionStatus augments the functionality of
// storage.StoreIndexedDbEncryptionStatus. It takes the database name and
// encryption status
//
// storeEncryptionStatus adhere to the storeEncryptionStatusFn type.
func (mh *messageHandler) storeEncryptionStatus(
	databaseName string, encryption bool) (bool, error) {
	// Package parameters for sending
	msg := &indexedDb.EncryptionStatusMessage{
		DatabaseName:     databaseName,
		EncryptionStatus: encryption,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return false, err
	}

	// Register response handler with channel that will wait for the response
	responseChan := make(chan []byte)
	mh.registerHandler(indexedDb.EncryptionStatusTag,
		func(data []byte) []byte {
			responseChan <- data
			return nil
		})

	// Send encryption status to main thread
	mh.sendResponse(indexedDb.EncryptionStatusTag, indexedDb.InitID, data)

	// Wait for response
	var response indexedDb.EncryptionStatusReply
	select {
	case responseData := <-responseChan:
		if err = json.Unmarshal(responseData, &response); err != nil {
			return false, err
		}
	case <-time.After(indexedDb.ResponseTimeout):
		return false, errors.Errorf("timed out after %s waiting for "+
			"response about the database encryption status from local "+
			"storage in the main thread", indexedDb.ResponseTimeout)
	}

	// If the response contain an error, return it
	if response.Error != "" {
		return false, errors.New(response.Error)
	}

	// Return the encryption status
	return response.EncryptionStatus, nil
}
