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
	userStoreName     = "users"

	// Message index names.
	messageStoreChannelIndex   = "channel_id_index"
	messageStoreUserIndex      = "user_pubkey_index"
	messageStoreParentIndex    = "parent_message_id_index"
	messageStoreTimestampIndex = "timestamp_index"
	messageStorePinnedIndex    = "pinned_index"

	// UserIdentity index names.
	userStorePubkeyIndex = "pubkey_index"

	// Message keyPath names (must match json struct tags).
	messageStoreChannel   = "channel_id"
	messageStoreUser      = "user_pubkey"
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
// A Message belongs to one User (cryptographic identity).
// The user's nickname can change each message.
type Message struct {
	Id              []byte        `json:"id"` // Matches pkeyName
	Nickname        string        `json:"nickname"`
	ChannelId       []byte        `json:"channel_id"`        // Index
	UserPubkey      []byte        `json:"user_pubkey"`       // Index
	ParentMessageId []byte        `json:"parent_message_id"` // Index
	Timestamp       time.Time     `json:"timestamp"`         // Index
	Lease           time.Duration `json:"lease"`
	Status          uint8         `json:"status"`
	Hidden          bool          `json:"hidden"`
	Pinned          bool          `json:"pinned"` // Index
	Text            string        `json:"text"`
}

// Channel defines the IndexedDb representation of a single Channel.
//
// A Channel has many Message.
type Channel struct {
	Id          []byte `json:"id"` // Matches pkeyName
	Name        string `json:"name"`
	Description string `json:"description"`
}

// User defines the IndexedDb representation of a single user's
// cryptographic identity
//
// A User has many Message.
type User struct {
	Pubkey         []byte `json:"pubkey"` // Matches pkeyName
	Honorific      string `json:"honorific"`
	Adjective      string `json:"adjective"`
	Noun           string `json:"noun"`
	Codename       string `json:"codename"`
	Color          string `json:"color"`
	Extension      string `json:"extension"`
	CodesetVersion uint8  `json:"codeset_version"`
}
