////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package indexedDb

import (
	"time"
)

const (
	// Text representation of primary key value (keyPath).
	pkeyName = "id"

	// Text representation of the names of the various [idb.ObjectStore].
	messageStoreName  = "messages"
	channelsStoreName = "channels"

	// Message index names.
	messageStoreMessageIndex   = "message_id_index"
	messageStoreChannelIndex   = "channel_id_index"
	messageStoreParentIndex    = "parent_message_id_index"
	messageStoreTimestampIndex = "timestamp_index"
	messageStorePinnedIndex    = "pinned_index"

	// Message keyPath names (must match json struct tags).
	messageStoreMessage   = "message_id"
	messageStoreChannel   = "channel_id"
	messageStoreParent    = "parent_message_id"
	messageStoreTimestamp = "timestamp"
	messageStorePinned    = "pinned"
)

// Message defines the IndexedDb representation of a single Message.
//
// A Message belongs to one Channel.
//
// A Message may belong to one Message (Parent).
//
// The user's nickname can change each message, but the rest does not. We
// still duplicate all of it for each entry to simplify code for now.
type Message struct {
	ID              uint64        `json:"id"` // Matches pkeyName
	Nickname        string        `json:"nickname"`
	MessageID       []byte        `json:"message_id"`        // Index
	ChannelID       []byte        `json:"channel_id"`        // Index
	ParentMessageID []byte        `json:"parent_message_id"` // Index
	Timestamp       time.Time     `json:"timestamp"`         // Index
	Lease           time.Duration `json:"lease"`
	Status          uint8         `json:"status"`
	Hidden          bool          `json:"hidden"`
	Pinned          bool          `json:"pinned"` // Index
	Text            string        `json:"text"`
	Type            uint16        `json:"type"`
	Round           uint64        `json:"round"`

	// User cryptographic Identity struct -- could be pulled out
	Pubkey         []byte `json:"pubkey"` // Index
	CodesetVersion uint8  `json:"codeset_version"`
}

// Channel defines the IndexedDb representation of a single Channel.
//
// A Channel has many Message.
type Channel struct {
	ID          []byte `json:"id"` // Matches pkeyName
	Name        string `json:"name"`
	Description string `json:"description"`
}
