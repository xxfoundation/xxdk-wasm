////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package channelEventModel

import "gitlab.com/elixxir/xxdk-wasm/worker"

// List of tags that can be used when sending a message or registering a handler
// to receive a message.
const (
	NewWASMEventModelTag       worker.Tag = "NewWASMEventModel"
	MessageReceivedCallbackTag worker.Tag = "MessageReceivedCallback"
	EncryptionStatusTag        worker.Tag = "EncryptionStatus"
	StoreDatabaseNameTag       worker.Tag = "StoreDatabaseName"

	ReceiveReplyTag     worker.Tag = "ReceiveReply"
	ReceiveReactionTag  worker.Tag = "ReceiveReaction"
	ReceiveTag          worker.Tag = "Receive"
	ReceiveTextTag      worker.Tag = "ReceiveText"
	UpdateSentStatusTag worker.Tag = "UpdateSentStatusTag"
)
