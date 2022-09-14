////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package wasm

import (
	"gitlab.com/elixxir/client/bindings"
	"gitlab.com/elixxir/xxdk-wasm/utils"
	"syscall/js"
)

////////////////////////////////////////////////////////////////////////////////
// Basic Channel API                                                          //
////////////////////////////////////////////////////////////////////////////////

// ChannelsManager wraps the [bindings.ChannelsManager] object so its methods
// can be wrapped to be Javascript compatible.
type ChannelsManager struct {
	api *bindings.ChannelsManager
}

// newChannelsManagerJS creates a new Javascript compatible object
// (map[string]interface{}) that matches the ChannelsManager structure.
func newChannelsManagerJS(api *bindings.ChannelsManager) map[string]interface{} {
	cm := ChannelsManager{api}
	channelsManagerMap := map[string]interface{}{
		// Basic Channel API
		"GetID":         js.FuncOf(cm.GetID),
		"JoinChannel":   js.FuncOf(cm.JoinChannel),
		"GetChannels":   js.FuncOf(cm.GetChannels),
		"GetChannelId":  js.FuncOf(cm.GetChannelId),
		"GetChannel":    js.FuncOf(cm.GetChannel),
		"LeaveChannel":  js.FuncOf(cm.LeaveChannel),
		"ReplayChannel": js.FuncOf(cm.ReplayChannel),

		// Channel Sending Methods and Reports
		"SendGeneric":      js.FuncOf(cm.SendGeneric),
		"SendAdminGeneric": js.FuncOf(cm.SendAdminGeneric),
		"SendMessage":      js.FuncOf(cm.SendMessage),
		"SendReply":        js.FuncOf(cm.SendReply),
		"SendReaction":     js.FuncOf(cm.SendReaction),

		// Channel Receiving Logic and Callback Registration
		"RegisterReceiveHandler": js.FuncOf(cm.RegisterReceiveHandler),
	}

	return channelsManagerMap
}

// GetID returns the ID for this ChannelsManager in the ChannelsManager tracker.
//
// Returns:
//  - int
func (ch *ChannelsManager) GetID(js.Value, []js.Value) interface{} {
	return ch.api.GetID()
}

// NewChannelsManager constructs a ChannelsManager.
// FIXME: This is a work in progress and should not be used an event model is
//  implemented in the style of the bindings layer's AuthCallbacks. Remove this
//  note when that has been done.
//
// Parameters:
//  - args[0] - ID of ChannelsManager object in tracker (int). This can be
//    retrieved using [E2e.GetID].
//  - args[1] - ID of UserDiscovery object in tracker (int). This can be
//    retrieved using [UserDiscovery.GetID].
//
// Returns:
//  - Javascript representation of the ChannelsManager object.
//  - Throws a TypeError if logging in fails.
func NewChannelsManager(_ js.Value, args []js.Value) interface{} {
	cm, err := bindings.NewChannelsManager(args[0].Int(), args[1].Int())
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return newChannelsManagerJS(cm)
}

// JoinChannel joins the given channel. It will fail if the channel has already
// been joined.
//
// Parameters:
//  - args[0] - JSON of [bindings.ChannelDef] (Uint8Array).
//
// Returns:
//  - Throws a TypeError if joining the channel fails.
func (ch *ChannelsManager) JoinChannel(_ js.Value, args []js.Value) interface{} {
	channelJson := utils.CopyBytesToGo(args[0])

	err := ch.api.JoinChannel(channelJson)
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return nil
}

// GetChannels returns the IDs of all channels that have been joined.
//
// Returns:
//  - JSON of an array of marshalled [id.ID] (Uint8Array).
//  - Throws a TypeError if getting the channels fails.
func (ch *ChannelsManager) GetChannels(js.Value, []js.Value) interface{} {
	channelList, err := ch.api.GetChannels()
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return utils.CopyBytesToJS(channelList)
}

// GetChannelId returns the ID of the channel given the channel's cryptographic
// information.
//
// Parameters:
//  - args[0] - JSON of [bindings.ChannelDef] (Uint8Array). This can be
//    retrieved using [Channel.Get].
//
// Returns:
//  - JSON of the channel [id.ID] (Uint8Array).
//  - Throws a TypeError if getting the channel's ID fails.
func (ch *ChannelsManager) GetChannelId(_ js.Value, args []js.Value) interface{} {
	channelJson := utils.CopyBytesToGo(args[0])

	chanID, err := ch.api.GetChannelId(channelJson)
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return utils.CopyBytesToJS(chanID)
}

// GetChannel returns the underlying cryptographic structure for a given
// channel.
//
// Parameters:
//  - args[0] - JSON of the channel [id.ID] (Uint8Array). This can be retrieved
//    using [ChannelsManager.GetChannelId].
//
// Returns:
//  - JSON of [bindings.ChannelDef] (Uint8Array).
//  - Throws a TypeError if getting the channel fails.
func (ch *ChannelsManager) GetChannel(_ js.Value, args []js.Value) interface{} {
	marshalledChanId := utils.CopyBytesToGo(args[0])

	def, err := ch.api.GetChannel(marshalledChanId)
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return utils.CopyBytesToJS(def)
}

// LeaveChannel leaves the given channel. It will return an error if the channel
// was not previously joined.
//
// Parameters:
//  - args[0] - JSON of the channel [id.ID] (Uint8Array). This can be retrieved
//    using [ChannelsManager.GetChannelId].
//
// Returns:
//  - Throws a TypeError if the channel does not exist.
func (ch *ChannelsManager) LeaveChannel(_ js.Value, args []js.Value) interface{} {
	marshalledChanId := utils.CopyBytesToGo(args[0])

	err := ch.api.LeaveChannel(marshalledChanId)
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return nil
}

// ReplayChannel replays all messages from the channel within the network's
// memory (~3 weeks) over the event model.
//
// Parameters:
//  - args[0] - JSON of the channel [id.ID] (Uint8Array). This can be retrieved
//    using [ChannelsManager.GetChannelId].
//
// Returns:
//  - Throws a TypeError if the replay fails.
func (ch *ChannelsManager) ReplayChannel(_ js.Value, args []js.Value) interface{} {
	marshalledChanId := utils.CopyBytesToGo(args[0])

	err := ch.api.ReplayChannel(marshalledChanId)
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Channel Sending Methods and Reports                                        //
////////////////////////////////////////////////////////////////////////////////

// SendGeneric is used to send a raw message over a channel. In general, it
// should be wrapped in a function which defines the wire protocol. If the final
// message, before being sent over the wire, is too long, this will return an
// error. Due to the underlying encoding using compression, it isn't possible to
// define the largest payload that can be sent, but it will always be possible
// to send a payload of 802 bytes at minimum. The meaning of validUntil depends
// on the use case.
//
// Parameters:
//  - args[0] - JSON of the channel [id.ID] (Uint8Array). This can be retrieved
//    using [ChannelsManager.GetChannelId].
//  - args[1] - The message type of the message. This will be a valid
//    [channels.MessageType] (int).
//  - args[2] - The contents of the message (Uint8Array).
//  - args[3] - The lease of the message. This will be how long the message is
//    valid until, in milliseconds. As per the [channels.Manager] documentation,
//    this has different meanings depending on the use case. These use cases may
//    be generic enough that they will not be enumerated here (int).
//  - args[4] - JSON of [xxdk.CMIXParams]. If left empty
//    [bindings.GetDefaultCMixParams] will be used internally (Uint8Array).
//
// Returns a promise:
//  - Resolves to the JSON of [bindings.ChannelSendReport] (Uint8Array).
//  - Rejected with an error if sending fails.
func (ch *ChannelsManager) SendGeneric(_ js.Value, args []js.Value) interface{} {
	marshalledChanId := utils.CopyBytesToGo(args[0])
	messageType := args[1].Int()
	message := utils.CopyBytesToGo(args[2])
	leaseTimeMS := int64(args[3].Int())
	cmixParamsJSON := utils.CopyBytesToGo(args[4])

	promiseFn := func(resolve, reject func(args ...interface{}) js.Value) {
		sendReport, err := ch.api.SendGeneric(
			marshalledChanId, messageType, message, leaseTimeMS, cmixParamsJSON)
		if err != nil {
			reject(utils.JsTrace(err))
		} else {
			resolve(utils.CopyBytesToJS(sendReport))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// SendAdminGeneric is used to send a raw message over a channel encrypted with
// admin keys, identifying it as sent by the admin. In general, it should be
// wrapped in a function that defines the wire protocol. If the final message,
// before being sent over the wire, is too long, this will return an error. The
// message must be at most 510 bytes long.
//
// Parameters:
//  - args[0] - The PEM-encode admin RSA private key (Uint8Array).
//  - args[1] - JSON of the channel [id.ID] (Uint8Array). This can be retrieved
//    using [ChannelsManager.GetChannelId].
//  - args[2] - The message type of the message. This will be a valid
//    [channels.MessageType] (int).
//  - args[3] - The contents of the message (Uint8Array).
//  - args[4] - The lease of the message. This will be how long the message is
//    valid until, in milliseconds. As per the [channels.Manager] documentation,
//    this has different meanings depending on the use case. These use cases may
//    be generic enough that they will not be enumerated here (int).
//  - args[5] - JSON of [xxdk.CMIXParams]. If left empty
//    [bindings.GetDefaultCMixParams] will be used internally (Uint8Array).
//
// Returns a promise:
//  - Resolves to the JSON of [bindings.ChannelSendReport] (Uint8Array).
//  - Rejected with an error if sending fails.
func (ch *ChannelsManager) SendAdminGeneric(_ js.Value, args []js.Value) interface{} {
	adminPrivateKey := utils.CopyBytesToGo(args[0])
	marshalledChanId := utils.CopyBytesToGo(args[1])
	messageType := args[2].Int()
	message := utils.CopyBytesToGo(args[3])
	leaseTimeMS := int64(args[4].Int())
	cmixParamsJSON := utils.CopyBytesToGo(args[5])

	promiseFn := func(resolve, reject func(args ...interface{}) js.Value) {
		sendReport, err := ch.api.SendAdminGeneric(adminPrivateKey,
			marshalledChanId, messageType, message, leaseTimeMS, cmixParamsJSON)
		if err != nil {
			reject(utils.JsTrace(err))
		} else {
			resolve(utils.CopyBytesToJS(sendReport))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// SendMessage is used to send a formatted message over a channel.
// Due to the underlying encoding using compression, it isn't possible to define
// the largest payload that can be sent, but it will always be possible to send
// a payload of 798 bytes at minimum.
//
// The message will auto delete validUntil after the round it is sent in,
// lasting forever if [channels.ValidForever] is used.
//
// Parameters:
//  - args[0] - JSON of the channel [id.ID] (Uint8Array). This can be retrieved
//    using [ChannelsManager.GetChannelId].
//  - args[1] - The contents of the message (string).
//  - args[2] - The lease of the message. This will be how long the message is
//    valid until, in milliseconds. As per the [channels.Manager] documentation,
//    this has different meanings depending on the use case. These use cases may
//    be generic enough that they will not be enumerated here (int).
//  - args[3] - JSON of [xxdk.CMIXParams]. If left empty
//    [bindings.GetDefaultCMixParams] will be used internally (Uint8Array).
//
// Returns a promise:
//  - Resolves to the JSON of [bindings.ChannelSendReport] (Uint8Array).
//  - Rejected with an error if sending fails.
func (ch *ChannelsManager) SendMessage(_ js.Value, args []js.Value) interface{} {
	marshalledChanId := utils.CopyBytesToGo(args[0])
	message := args[1].String()
	leaseTimeMS := int64(args[2].Int())
	cmixParamsJSON := utils.CopyBytesToGo(args[3])

	promiseFn := func(resolve, reject func(args ...interface{}) js.Value) {
		sendReport, err := ch.api.SendMessage(
			marshalledChanId, message, leaseTimeMS, cmixParamsJSON)
		if err != nil {
			reject(utils.JsTrace(err))
		} else {
			resolve(utils.CopyBytesToJS(sendReport))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// SendReply is used to send a formatted message over a channel.
// Due to the underlying encoding using compression, it isn't possible to define
// the largest payload that can be sent, but it will always be possible to send
// a payload of 766 bytes at minimum.
//
// If the message ID the reply is sent to does not exist, then the other side
// will post the message as a normal message and not a reply.
// The message will auto delete validUntil after the round it is sent in,
// lasting forever if ValidForever is used.
//
// Parameters:
//  - args[0] - JSON of the channel [id.ID] (Uint8Array). This can be retrieved
//    using [ChannelsManager.GetChannelId].
//  - args[1] - The contents of the message. The message should be at most 510
//    bytes. This is expected to be Unicode, and thus a string data type is
//    expected (string).
//  - args[2] - JSON of [channel.MessageID] of the message you wish to reply to.
//    This may be found in the [bindings.ChannelSendReport] if replying to your
//    own. Alternatively, if reacting to another user's message, you may
//    retrieve it via the [bindings.ChannelMessageReceptionCallback] registered
//    using  RegisterReceiveHandler (Uint8Array).
//  - args[3] - The lease of the message. This will be how long the message is
//    valid until, in milliseconds. As per the [channels.Manager] documentation,
//    this has different meanings depending on the use case. These use cases may
//    be generic enough that they will not be enumerated here (int).
//  - args[4] - JSON of [xxdk.CMIXParams]. If left empty
//    [bindings.GetDefaultCMixParams] will be used internally (Uint8Array).
//
// Returns a promise:
//  - Resolves to the JSON of [bindings.ChannelSendReport] (Uint8Array).
//  - Rejected with an error if sending fails.
func (ch *ChannelsManager) SendReply(_ js.Value, args []js.Value) interface{} {
	marshalledChanId := utils.CopyBytesToGo(args[0])
	message := args[1].String()
	messageToReactTo := utils.CopyBytesToGo(args[2])
	leaseTimeMS := int64(args[3].Int())
	cmixParamsJSON := utils.CopyBytesToGo(args[4])

	promiseFn := func(resolve, reject func(args ...interface{}) js.Value) {
		sendReport, err := ch.api.SendReply(marshalledChanId, message,
			messageToReactTo, leaseTimeMS, cmixParamsJSON)
		if err != nil {
			reject(utils.JsTrace(err))
		} else {
			resolve(utils.CopyBytesToJS(sendReport))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// SendReaction is used to send a reaction to a message over a channel.
// The reaction must be a single emoji with no other characters, and will
// be rejected otherwise.
// Users will drop the reaction if they do not recognize the reactTo message.
//
// Parameters:
//  - args[0] - JSON of the channel [id.ID] (Uint8Array). This can be retrieved
//    using [ChannelsManager.GetChannelId].
//  - args[1] - The user's reaction. This should be a single emoji with no
//    other characters. As such, a Unicode string is expected (string).
//  - args[2] - JSON of [channel.MessageID] of the message you wish to reply to.
//    This may be found in the [bindings.ChannelSendReport] if replying to your
//    own. Alternatively, if reacting to another user's message, you may
//    retrieve it via the ChannelMessageReceptionCallback registered using
//    RegisterReceiveHandler (Uint8Array).
//  - args[3] - JSON of [xxdk.CMIXParams]. If left empty
//    [bindings.GetDefaultCMixParams] will be used internally (Uint8Array).
//
// Returns a promise:
//  - Resolves to the JSON of [bindings.ChannelSendReport] (Uint8Array).
//  - Rejected with an error if sending fails.
func (ch *ChannelsManager) SendReaction(_ js.Value, args []js.Value) interface{} {
	marshalledChanId := utils.CopyBytesToGo(args[0])
	reaction := args[1].String()
	messageToReactTo := utils.CopyBytesToGo(args[2])
	cmixParamsJSON := utils.CopyBytesToGo(args[3])

	promiseFn := func(resolve, reject func(args ...interface{}) js.Value) {
		sendReport, err := ch.api.SendReaction(
			marshalledChanId, reaction, messageToReactTo, cmixParamsJSON)
		if err != nil {
			reject(utils.JsTrace(err))
		} else {
			resolve(utils.CopyBytesToJS(sendReport))
		}
	}

	return utils.CreatePromise(promiseFn)
}

////////////////////////////////////////////////////////////////////////////////
// Channel Receiving Logic and Callback Registration                          //
////////////////////////////////////////////////////////////////////////////////

// channelMessageReceptionCallback wraps Javascript callbacks to adhere to the
// [bindings.ChannelMessageReceptionCallback] interface.
type channelMessageReceptionCallback struct {
	callback func(args ...interface{}) js.Value
}

func (cmrCB *channelMessageReceptionCallback) Callback(
	receivedChannelMessageReport []byte, err error) {
	cmrCB.callback(utils.CopyBytesToJS(receivedChannelMessageReport),
		utils.JsTrace(err))
}

// RegisterReceiveHandler is used to register handlers for non-default message
// types. They can be processed by modules. It is important that such modules
// sync up with the event model implementation.
//
// There can only be one handler per [channels.MessageType], and this will
// return an error on any re-registration.
//
// Parameters:
//  - args[0] - The message type of the message. This will be a valid
//    [channels.MessageType] (int).
//  - args[1] - Javascript object that has functions that implement the
//    [bindings.ChannelMessageReceptionCallback] interface. This callback will
//    be executed when a channel message of the messageType is received.
//
// Returns:
//  - Throws a TypeError if registering the handler fails.
func (ch *ChannelsManager) RegisterReceiveHandler(_ js.Value, args []js.Value) interface{} {
	messageType := args[0].Int()
	listenerCb := &channelMessageReceptionCallback{
		utils.WrapCB(args[1], "Callback")}

	err := ch.api.RegisterReceiveHandler(messageType, listenerCb)
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return nil
}
