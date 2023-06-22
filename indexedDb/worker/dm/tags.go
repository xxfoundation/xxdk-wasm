////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package dm

import "gitlab.com/elixxir/xxdk-wasm/worker"

// List of tags that can be used when sending a message or registering a handler
// to receive a message.
const (
	NewWASMEventModelTag       worker.Tag = "NewWASMEventModel"
	MessageReceivedCallbackTag worker.Tag = "MessageReceivedCallback"

	ReceiveReplyTag     worker.Tag = "ReceiveReply"
	ReceiveReactionTag  worker.Tag = "ReceiveReaction"
	ReceiveTag          worker.Tag = "Receive"
	ReceiveTextTag      worker.Tag = "ReceiveText"
	UpdateSentStatusTag worker.Tag = "UpdateSentStatus"
	DeleteMessageTag    worker.Tag = "DeleteMessage"

	GetConversationTag  worker.Tag = "GetConversation"
	GetConversationsTag worker.Tag = "GetConversations"
)
