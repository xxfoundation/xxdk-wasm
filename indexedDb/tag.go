////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package indexedDb

// Tag describes how a message sent to or from the worker should be handled.
type Tag string

// List of tags that can be used when sending a message or registering a handler
// to receive a message.
const (
	ReadyTag Tag = "Ready"

	NewWASMEventModelTag Tag = "NewWASMEventModel"
	EncryptionStatusTag  Tag = "EncryptionStatus"
	StoreDatabaseNameTag Tag = "StoreDatabaseName"

	JoinChannelTag         Tag = "JoinChannel"
	LeaveChannelTag        Tag = "LeaveChannel"
	ReceiveMessageTag      Tag = "ReceiveMessage"
	ReceiveReplyTag        Tag = "ReceiveReply"
	ReceiveReactionTag     Tag = "ReceiveReaction"
	UpdateFromUUIDTag      Tag = "UpdateFromUUID"
	UpdateFromMessageIDTag Tag = "UpdateFromMessageID"
	GetMessageTag          Tag = "GetMessage"
	DeleteMessageTag       Tag = "DeleteMessage"
)

// deleteAfterReceiving is a list of Tags that will have their handler deleted
// after a message is received. This is mainly used for responses where the
// handler will only handle it once and never again.
var deleteAfterReceiving = map[Tag]struct{}{
	ReadyTag:               {},
	NewWASMEventModelTag:   {},
	EncryptionStatusTag:    {},
	JoinChannelTag:         {},
	LeaveChannelTag:        {},
	ReceiveMessageTag:      {},
	ReceiveReplyTag:        {},
	ReceiveReactionTag:     {},
	UpdateFromUUIDTag:      {},
	UpdateFromMessageIDTag: {},
	DeleteMessageTag:       {},
}
