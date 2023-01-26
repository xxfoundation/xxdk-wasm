////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package main

import (
	"time"
)

const (
	// Text representation of primary key value (keyPath).
	msgPkeyName   = "id"
	convoPkeyName = "pub_key"

	// Text representation of the names of the various [idb.ObjectStore].
	messageStoreName      = "messages"
	conversationStoreName = "conversations"

	// Message index names.
	messageStoreMessageIndex      = "message_id_index"
	messageStoreConversationIndex = "conversation_id_index"
	messageStoreParentIndex       = "parent_message_id_index"
	messageStoreTimestampIndex    = "timestamp_index"

	// Message keyPath names (must match json struct tags).
	messageStoreMessage      = "message_id"
	messageStoreConversation = "conversation_id"
	messageStoreParent       = "parent_message_id"
	messageStoreTimestamp    = "timestamp"
)

// Message defines the IndexedDb representation of a single Message.
//
// A Message belongs to one Conversation.
// A Message may belong to one Message (Parent).
type Message struct {
	ID                 uint64    `json:"id"`                   // Matches msgPkeyName
	MessageID          []byte    `json:"message_id"`           // Index
	ConversationPubKey []byte    `json:"conversation_pub_key"` // Index
	ParentMessageID    []byte    `json:"parent_message_id"`    // Index
	Timestamp          time.Time `json:"timestamp"`            // Index
	Status             uint8     `json:"status"`
	Text               []byte    `json:"text"`
	Type               uint16    `json:"type"`
	Round              uint64    `json:"round"`
}

// Conversation defines the IndexedDb representation of a single
// message exchange between two recipients.
// A Conversation has many Message.
type Conversation struct {
	Pubkey         []byte `json:"pub_key"` // Matches convoPkeyName
	Nickname       string `json:"nickname"`
	Token          uint32 `json:"token"`
	CodesetVersion uint8  `json:"codeset_version"`
	Blocked        bool   `json:"blocked"`
}
