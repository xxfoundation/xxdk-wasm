////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"gitlab.com/xx_network/primitives/id"
	"syscall/js"

	"gitlab.com/elixxir/client/bindings"
	"gitlab.com/elixxir/xxdk-wasm/indexedDb"
	"gitlab.com/elixxir/xxdk-wasm/utils"
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
// (map[string]interface{}) that matches the [ChannelsManager] structure.
func newChannelsManagerJS(api *bindings.ChannelsManager) map[string]interface{} {
	cm := ChannelsManager{api}
	channelsManagerMap := map[string]interface{}{
		// Basic Channel API
		"GetID":         js.FuncOf(cm.GetID),
		"JoinChannel":   js.FuncOf(cm.JoinChannel),
		"GetChannels":   js.FuncOf(cm.GetChannels),
		"LeaveChannel":  js.FuncOf(cm.LeaveChannel),
		"ReplayChannel": js.FuncOf(cm.ReplayChannel),

		// Channel Sending Methods and Reports
		"SendGeneric":      js.FuncOf(cm.SendGeneric),
		"SendAdminGeneric": js.FuncOf(cm.SendAdminGeneric),
		"SendMessage":      js.FuncOf(cm.SendMessage),
		"SendReply":        js.FuncOf(cm.SendReply),
		"SendReaction":     js.FuncOf(cm.SendReaction),
		"GetIdentity":      js.FuncOf(cm.GetIdentity),
		"GetStorageTag":    js.FuncOf(cm.GetStorageTag),
		"SetNickname":      js.FuncOf(cm.SetNickname),
		"DeleteNickname":   js.FuncOf(cm.DeleteNickname),
		"GetNickname":      js.FuncOf(cm.GetNickname),

		// Channel Receiving Logic and Callback Registration
		"RegisterReceiveHandler": js.FuncOf(cm.RegisterReceiveHandler),
	}

	return channelsManagerMap
}

// GetID returns the ID for this [ChannelsManager] in the [ChannelsManager]
// tracker.
//
// Returns:
//  - Tracker ID (int).
func (ch *ChannelsManager) GetID(js.Value, []js.Value) interface{} {
	return ch.api.GetID()
}

// GenerateChannelIdentity creates a new private channel identity
// ([channel.PrivateIdentity]). The public component can be retrieved as JSON
// via [GetPublicChannelIdentityFromPrivate].
//
// Parameters:
//  - args[0] - ID of [Cmix] object in tracker (int).
//
// Returns:
//  - JSON of [channel.PrivateIdentity] (Uint8Array).
//  - Throws a TypeError if generating the identity fails.
func GenerateChannelIdentity(_ js.Value, args []js.Value) interface{} {
	pi, err := bindings.GenerateChannelIdentity(args[0].Int())
	if err != nil {
		utils.Throw(utils.TypeError, err)
	}

	return utils.CopyBytesToJS(pi)
}

// GetPublicChannelIdentity constructs a public identity ([channel.Identity])
// from a bytes version and returns it JSON marshaled.
//
// Parameters:
//  - args[0] - Bytes of the public identity ([channel.Identity]) (Uint8Array).
//
// Returns:
//  - JSON of the constructed [channel.Identity] (Uint8Array).
//  - Throws a TypeError if unmarshalling the bytes or marshalling the identity
//    fails.
func GetPublicChannelIdentity(_ js.Value, args []js.Value) interface{} {
	marshaledPublic := utils.CopyBytesToGo(args[0])
	pi, err := bindings.GetPublicChannelIdentity(marshaledPublic)
	if err != nil {
		utils.Throw(utils.TypeError, err)
	}

	return utils.CopyBytesToJS(pi)
}

// GetPublicChannelIdentityFromPrivate returns the public identity
// ([channel.Identity]) contained in the given private identity
// ([channel.PrivateIdentity]).
//
// Parameters:
//  - args[0] - Bytes of the private identity
//    (channel.PrivateIdentity]) (Uint8Array).
//
// Returns:
//  - JSON of the public identity ([channel.Identity]) (Uint8Array).
//  - Throws a TypeError if unmarshalling the bytes or marshalling the identity
//    fails.
func GetPublicChannelIdentityFromPrivate(_ js.Value, args []js.Value) interface{} {
	marshaledPrivate := utils.CopyBytesToGo(args[0])
	identity, err := bindings.GetPublicChannelIdentityFromPrivate(
		marshaledPrivate)
	if err != nil {
		utils.Throw(utils.TypeError, err)
	}

	return utils.CopyBytesToJS(identity)
}

// eventModelBuilder adheres to the [bindings.EventModelBuilder] interface.
type eventModelBuilder struct {
	build func(args ...interface{}) js.Value
}

// Build initializes and returns the event model.  It wraps a Javascript object
// that has all the methods in [bindings.EventModel] to make it adhere to the Go
// interface [bindings.EventModel].
func (emb *eventModelBuilder) Build(path string) bindings.EventModel {
	emJs := emb.build(path)
	return &eventModel{
		joinChannel:      utils.WrapCB(emJs, "JoinChannel"),
		leaveChannel:     utils.WrapCB(emJs, "LeaveChannel"),
		receiveMessage:   utils.WrapCB(emJs, "ReceiveMessage"),
		receiveReply:     utils.WrapCB(emJs, "ReceiveReply"),
		receiveReaction:  utils.WrapCB(emJs, "ReceiveReaction"),
		updateSentStatus: utils.WrapCB(emJs, "UpdateSentStatus"),
	}
}

// NewChannelsManager creates a new [ChannelsManager] from a new private
// identity ([channel.PrivateIdentity]).
//
// This is for creating a manager for an identity for the first time. For
// generating a new one channel identity, use [GenerateChannelIdentity]. To
// reload this channel manager, use [LoadChannelsManager], passing in the
// storage tag retrieved by [ChannelsManager.GetStorageTag].
//
// Parameters:
//  - args[0] - ID of [Cmix] object in tracker (int). This can be retrieved
//    using [Cmix.GetID].
//  - args[1] - Bytes of a private identity ([channel.PrivateIdentity]) that is
//    generated by [GenerateChannelIdentity] (Uint8Array).
//  - args[2] - A function that initialises and returns a Javascript object that
//    matches the [bindings.EventModel] interface. The function must match the
//    Build function in [bindings.EventModelBuilder].
//
// Returns:
//  - Javascript representation of the [ChannelsManager] object.
//  - Throws a TypeError if creating the manager fails.
func NewChannelsManager(_ js.Value, args []js.Value) interface{} {
	privateIdentity := utils.CopyBytesToGo(args[1])

	em := &eventModelBuilder{args[2].Invoke}

	cm, err := bindings.NewChannelsManager(args[0].Int(), privateIdentity, em)
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return newChannelsManagerJS(cm)
}

// LoadChannelsManager loads an existing [ChannelsManager].
//
// This is for loading a manager for an identity that has already been created.
// The channel manager should have previously been created with
// [NewChannelsManager] and the storage is retrievable with
// [ChannelsManager.GetStorageTag].
//
// Parameters:
//  - args[0] - ID of [Cmix] object in tracker (int). This can be retrieved
//    using [Cmix.GetID].
//  - args[1] - The storage tag associated with the previously created channel
//    manager and retrieved with [ChannelsManager.GetStorageTag] (string).
//  - args[2] - A function that initialises and returns a Javascript object that
//    matches the [bindings.EventModel] interface. The function must match the
//    Build function in [bindings.EventModelBuilder].
//
// Returns:
//  - Javascript representation of the [ChannelsManager] object.
//  - Throws a TypeError if loading the manager fails.
func LoadChannelsManager(_ js.Value, args []js.Value) interface{} {
	em := &eventModelBuilder{args[2].Invoke}
	cm, err := bindings.LoadChannelsManager(args[0].Int(), args[1].String(), em)
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return newChannelsManagerJS(cm)
}

// NewChannelsManagerWithIndexedDb creates a new [ChannelsManager] from a new
// private identity ([channel.PrivateIdentity]) and using indexedDb as a backend
// to manage the event model.
//
// This is for creating a manager for an identity for the first time. For
// generating a new one channel identity, use [GenerateChannelIdentity]. To
// reload this channel manager, use [LoadChannelsManagerWithIndexedDb], passing
// in the storage tag retrieved by [ChannelsManager.GetStorageTag].
//
// This function initialises an indexedDb database.
//
// Parameters:
//  - args[0] - ID of [Cmix] object in tracker (int). This can be retrieved
//    using [Cmix.GetID].
//  - args[1] - Bytes of a private identity ([channel.PrivateIdentity]) that is
//    generated by [GenerateChannelIdentity] (Uint8Array).
//  - args[2] - Function that takes in the same parameters as
//    [indexedDb.MessageReceivedCallback]. On the Javascript side, the uuid is
//    returned as an int and the channelID as a Uint8Array. The row in the
//    database that was updated can be found using the UUID. The channel ID is
//    provided so that the recipient can filter if they want to the processes
//    the update now or not. An "update" bool is present which tells you if
//	  the row is new or if it is an edited old row
//
// Returns a promise:
//  - Resolves to a Javascript representation of the [ChannelsManager] object.
//  - Rejected with an error if loading indexedDb or the manager fails.
func NewChannelsManagerWithIndexedDb(_ js.Value, args []js.Value) interface{} {
	cmixID := args[0].Int()
	privateIdentity := utils.CopyBytesToGo(args[1])

	fn := func(uuid uint64, channelID *id.ID, update bool) {
		args[2].Invoke(uuid, utils.CopyBytesToJS(channelID.Marshal()), update)
	}

	model := indexedDb.NewWASMEventModelBuilder(fn)

	promiseFn := func(resolve, reject func(args ...interface{}) js.Value) {
		cm, err := bindings.NewChannelsManagerGoEventModel(
			cmixID, privateIdentity, model)
		if err != nil {
			reject(utils.JsTrace(err))
		} else {
			resolve(newChannelsManagerJS(cm))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// LoadChannelsManagerWithIndexedDb loads an existing [ChannelsManager] using
// an existing indexedDb database as a backend to manage the event model.
//
// This is for loading a manager for an identity that has already been created.
// The channel manager should have previously been created with
// [NewChannelsManagerWithIndexedDb] and the storage is retrievable with
// [ChannelsManager.GetStorageTag].
//
// Parameters:
//  - args[0] - ID of [Cmix] object in tracker (int). This can be retrieved
//    using [Cmix.GetID].
//  - args[1] - The storage tag associated with the previously created channel
//    manager and retrieved with [ChannelsManager.GetStorageTag] (string).
//  - args[2] - Function that takes in the same parameters as
//    [indexedDb.MessageReceivedCallback]. On the Javascript side, the uuid is
//    returned as an int and the channelID as a Uint8Array. The row in the
//    database that was updated can be found using the UUID. The channel ID is
//    provided so that the recipient can filter if they want to the processes
//    the update now or not. An "update" bool is present which tells you if
//	  the row is new or if it is an edited old row
//
// Returns a promise:
//  - Resolves to a Javascript representation of the [ChannelsManager] object.
//  - Rejected with an error if loading indexedDb or the manager fails.
func LoadChannelsManagerWithIndexedDb(_ js.Value, args []js.Value) interface{} {
	cmixID := args[0].Int()
	storageTag := args[1].String()

	fn := func(uuid uint64, channelID *id.ID, updated bool) {
		args[2].Invoke(uuid, utils.CopyBytesToJS(channelID.Marshal()), updated)
	}

	model := indexedDb.NewWASMEventModelBuilder(fn)

	promiseFn := func(resolve, reject func(args ...interface{}) js.Value) {
		cm, err := bindings.LoadChannelsManagerGoEventModel(
			cmixID, storageTag, model)
		if err != nil {
			reject(utils.JsTrace(err))
		} else {
			resolve(newChannelsManagerJS(cm))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// GenerateChannel is used to create a channel a new channel of which you are
// the admin. It is only for making new channels, not joining existing ones.
//
// It returns a pretty print of the channel and the private key.
//
// The name cannot be longer that __ characters. The description cannot be
// longer than __ and can only use ______ characters.
//
// Parameters:
//  - args[0] - ID of [Cmix] object in tracker (int).
//  - args[1] - The name of the new channel. The name cannot be longer than __
//    characters and must contain only _____ characters. It cannot be changed
//    once a channel is created. (string).
//  - args[2] - The description of a channel. The description cannot be longer
//    than __ characters and must contain only __ characters. It cannot be
//    changed once a channel is created (string).
//
// Returns:
//  - JSON of [bindings.ChannelGeneration], which describes a generated channel.
//    It contains both the public channel info and the private key for the
//    channel in PEM format (Uint8Array).
//  - Throws a TypeError if generating the channel fails.
func GenerateChannel(_ js.Value, args []js.Value) interface{} {
	gen, err := bindings.GenerateChannel(
		args[0].Int(), args[1].String(), args[2].String())
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return utils.CopyBytesToJS(gen)
}

// GetChannelInfo returns the info about a channel from its public description.
//
// Parameters:
//  - args[0] - The pretty print of the channel (string).
//
// The pretty print will be of the format:
//  <XXChannel-v1:Test Channel,description:This is a test channel,secrets:pn0kIs6P1pHvAe7u8kUyf33GYVKmkoCX9LhCtvKJZQI=,3A5eB5pzSHyxN09w1kOVrTIEr5UyBbzmmd9Ga5Dx0XA=,0,0,/zChIlLr2p3Vsm2X4+3TiFapoapaTi8EJIisJSqwfGc=>
//
// Returns:
//  - JSON of [bindings.ChannelInfo], which describes all relevant channel info
//    (Uint8Array).
//  - Throws a TypeError if getting the channel info fails.
func GetChannelInfo(_ js.Value, args []js.Value) interface{} {
	ci, err := bindings.GetChannelInfo(args[0].String())
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return utils.CopyBytesToJS(ci)
}

// JoinChannel joins the given channel. It will fail if the channel has already
// been joined.
//
// Parameters:
//  - args[0] - A portable channel string. Should be received from another user
//    or generated via GenerateChannel (string).
//
// The pretty print will be of the format:
//  <XXChannel-v1:Test Channel,description:This is a test channel,secrets:pn0kIs6P1pHvAe7u8kUyf33GYVKmkoCX9LhCtvKJZQI=,3A5eB5pzSHyxN09w1kOVrTIEr5UyBbzmmd9Ga5Dx0XA=,0,0,/zChIlLr2p3Vsm2X4+3TiFapoapaTi8EJIisJSqwfGc=>"
//
// Returns:
//  - JSON of [bindings.ChannelInfo], which describes all relevant channel info
//    (Uint8Array).
//  - Throws a TypeError if joining the channel fails.
func (ch *ChannelsManager) JoinChannel(_ js.Value, args []js.Value) interface{} {
	ci, err := ch.api.JoinChannel(args[0].String())
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return utils.CopyBytesToJS(ci)
}

// GetChannels returns the IDs of all channels that have been joined.
//
// Returns:
//  - JSON of an array of marshalled [id.ID] (Uint8Array).
//  - Throws a TypeError if getting the channels fails.
//
// JSON Example:
//  {
//    "U4x/lrFkvxuXu59LtHLon1sUhPJSCcnZND6SugndnVID",
//    "15tNdkKbYXoMn58NO6VbDMDWFEyIhTWEGsvgcJsHWAgD"
//  }
func (ch *ChannelsManager) GetChannels(js.Value, []js.Value) interface{} {
	channelList, err := ch.api.GetChannels()
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return utils.CopyBytesToJS(channelList)
}

// LeaveChannel leaves the given channel. It will return an error if the channel
// was not previously joined.
//
// Parameters:
//  - args[0] - Marshalled bytes of the channel [id.ID] (Uint8Array).
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
//  - args[0] - Marshalled bytes of the channel [id.ID] (Uint8Array).
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
//  - args[0] - Marshalled bytes of the channel [id.ID] (Uint8Array).
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
//  - args[1] - Marshalled bytes of the channel [id.ID] (Uint8Array).
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
//  - args[0] - Marshalled bytes of the channel [id.ID] (Uint8Array).
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

// SendReply is used to send a formatted message over a channel. Due to the
// underlying encoding using compression, it isn't possible to define the
// largest payload that can be sent, but it will always be possible to send a
// payload of 766 bytes at minimum.
//
// If the message ID the reply is sent to is nonexistent, the other side will
// post the message as a normal message and not a reply. The message will auto
// delete validUntil after the round it is sent in, lasting forever if
// [channels.ValidForever] is used.
//
// Parameters:
//  - args[0] - Marshalled bytes of the channel [id.ID] (Uint8Array).
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
//  - args[0] - Marshalled bytes of the channel [id.ID] (Uint8Array).
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

// GetIdentity returns the marshaled public identity ([channel.Identity]) that
// the channel is using.
//
// Returns:
//  - JSON of the [channel.Identity] (Uint8Array).
//  - Throws TypeError if marshalling the identity fails.
func (ch *ChannelsManager) GetIdentity(js.Value, []js.Value) interface{} {
	i, err := ch.api.GetIdentity()
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return utils.CopyBytesToJS(i)
}

// GetStorageTag returns the storage tag needed to reload the manager.
//
// Returns:
//  - Storage tag (string).
func (ch *ChannelsManager) GetStorageTag(js.Value, []js.Value) interface{} {
	return ch.api.GetStorageTag()
}

// SetNickname sets the nickname for a given channel. The nickname must be valid
// according to [IsNicknameValid].
//
// Parameters:
//  - args[0] - The nickname to set (string).
//  - args[1] - Marshalled bytes if the channel's [id.ID] (Uint8Array).
//
// Returns:
//  - Throws TypeError if unmarshalling the ID fails or the nickname is invalid.
func (ch *ChannelsManager) SetNickname(_ js.Value, args []js.Value) interface{} {
	err := ch.api.SetNickname(args[0].String(), utils.CopyBytesToGo(args[1]))
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return nil
}

// DeleteNickname deletes the nickname for a given channel.
//
// Parameters:
//  - args[0] - Marshalled bytes if the channel's [id.ID] (Uint8Array).
//
// Returns:
//  - Throws TypeError if deleting the nickname fails.
func (ch *ChannelsManager) DeleteNickname(_ js.Value, args []js.Value) interface{} {
	err := ch.api.DeleteNickname(utils.CopyBytesToGo(args[0]))
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return nil
}

// GetNickname returns the nickname set for a given channel. Returns an error if
// there is no nickname set.
//
// Parameters:
//  - args[0] - Marshalled bytes if the channel's [id.ID] (Uint8Array).
//
// Returns:
//  - The nickname (string).
//  - Throws TypeError if the channel has no nickname set.
func (ch *ChannelsManager) GetNickname(_ js.Value, args []js.Value) interface{} {
	nickname, err := ch.api.GetNickname(utils.CopyBytesToGo(args[0]))
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return nickname
}

// IsNicknameValid checks if a nickname is valid.
//
// Rules:
//  1. A nickname must not be longer than 24 characters.
//  2. A nickname must not be shorter than 1 character.
//
// Parameters:
//  - args[0] - Nickname to check (string).
//
// Returns:
//  - A Javascript Error object if the nickname is invalid with the reason why.
//  - Null if the nickname is valid.
func IsNicknameValid(_ js.Value, args []js.Value) interface{} {
	err := bindings.IsNicknameValid(args[0].String())
	if err != nil {
		return utils.JsError(err)
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Channel Receiving Logic and Callback Registration                          //
////////////////////////////////////////////////////////////////////////////////

// channelMessageReceptionCallback wraps Javascript callbacks to adhere to the
// [bindings.ChannelMessageReceptionCallback] interface.
type channelMessageReceptionCallback struct {
	callback func(args ...interface{}) js.Value
}

// Callback returns the context for a channel message.
//
// Parameters:
//  - receivedChannelMessageReport - Returns the JSON of
//   [bindings.ReceivedChannelMessageReport] (Uint8Array).
//  - err - Returns an error on failure (Error).
//
// Returns:
//  - It must return a unique UUID for the message that it can be referenced by
//    later (int).
func (cmrCB *channelMessageReceptionCallback) Callback(
	receivedChannelMessageReport []byte, err error) int {
	uuid := cmrCB.callback(
		utils.CopyBytesToJS(receivedChannelMessageReport), utils.JsTrace(err))

	return uuid.Int()
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

////////////////////////////////////////////////////////////////////////////////
// Event Model Logic                                                          //
////////////////////////////////////////////////////////////////////////////////

// eventModel wraps Javascript callbacks to adhere to the [bindings.EventModel]
// interface.
type eventModel struct {
	joinChannel      func(args ...interface{}) js.Value
	leaveChannel     func(args ...interface{}) js.Value
	receiveMessage   func(args ...interface{}) js.Value
	receiveReply     func(args ...interface{}) js.Value
	receiveReaction  func(args ...interface{}) js.Value
	updateSentStatus func(args ...interface{}) js.Value
}

// JoinChannel is called whenever a channel is joined locally.
//
// Parameters:
//  - channel - Returns the pretty print representation of a channel (string).
func (em *eventModel) JoinChannel(channel string) {
	em.joinChannel(channel)
}

// LeaveChannel is called whenever a channel is left locally.
//
// Parameters:
//  - ChannelId - Marshalled bytes of the channel [id.ID] (Uint8Array).
func (em *eventModel) LeaveChannel(channelID []byte) {
	em.leaveChannel(utils.CopyBytesToJS(channelID))
}

// ReceiveMessage is called whenever a message is received on a given channel.
// It may be called multiple times on the same message. It is incumbent on the
// user of the API to filter such called by message ID.
//
// Parameters:
//  - channelID - Marshalled bytes of the channel [id.ID] (Uint8Array).
//  - messageID - The bytes of the [channel.MessageID] of the received message
//    (Uint8Array).
//  - nickname - The nickname of the sender of the message (string).
//  - text - The content of the message (string).
//  - identity - JSON of the sender's public ([channel.Identity]) (Uint8Array).
//  - timestamp - Time the message was received; represented as nanoseconds
//    since unix epoch (int).
//  - lease - The number of nanoseconds that the message is valid for (int).
//  - roundId - The ID of the round that the message was received on (int).
//  - msgType - The type of message ([channels.MessageType]) to send (int).
//  - status - The [channels.SentStatus] of the message (int).
//
// Statuses will be enumerated as such:
//  Sent      =  0
//  Delivered =  1
//  Failed    =  2
//
// Returns:
//  - A non-negative unique UUID for the message that it can be referenced by
//    later with [eventModel.UpdateSentStatus].
func (em *eventModel) ReceiveMessage(channelID, messageID []byte, nickname,
	text string, identity []byte, timestamp, lease, roundId, msgType,
	status int64) int64 {
	uuid := em.receiveMessage(utils.CopyBytesToJS(channelID),
		utils.CopyBytesToJS(messageID), nickname, text,
		utils.CopyBytesToJS(identity),
		timestamp, lease, roundId, msgType, status)

	return int64(uuid.Int())
}

// ReceiveReply is called whenever a message is received that is a reply on a
// given channel. It may be called multiple times on the same message. It is
// incumbent on the user of the API to filter such called by message ID.
//
// Messages may arrive our of order, so a reply in theory can arrive before the
// initial message. As a result, it may be important to buffer replies.
//
// Parameters:
//  - channelID - Marshalled bytes of the channel [id.ID] (Uint8Array).
//  - messageID - The bytes of the [channel.MessageID] of the received message
//    (Uint8Array).
//  - reactionTo - The [channel.MessageID] for the message that received a reply
//    (Uint8Array).
//  - senderUsername - The username of the sender of the message (string).
//  - text - The content of the message (string).
//  - identity - JSON of the sender's public ([channel.Identity]) (Uint8Array).
//  - timestamp - Time the message was received; represented as nanoseconds
//    since unix epoch (int).
//  - lease - The number of nanoseconds that the message is valid for (int).
//  - roundId - The ID of the round that the message was received on (int).
//  - msgType - The type of message ([channels.MessageType]) to send (int).
//  - status - The [channels.SentStatus] of the message (int).
//
// Statuses will be enumerated as such:
//  Sent      =  0
//  Delivered =  1
//  Failed    =  2
//
// Returns:
//  - A non-negative unique UUID for the message that it can be referenced by
//    later with [eventModel.UpdateSentStatus].
func (em *eventModel) ReceiveReply(channelID, messageID, reactionTo []byte,
	senderUsername, text string, identity []byte, timestamp, lease, roundId,
	msgType, status int64) int64 {
	uuid := em.receiveReply(utils.CopyBytesToJS(channelID),
		utils.CopyBytesToJS(messageID), utils.CopyBytesToJS(reactionTo),
		senderUsername, text, utils.CopyBytesToJS(identity),
		timestamp, lease, roundId, msgType, status)

	return int64(uuid.Int())
}

// ReceiveReaction is called whenever a reaction to a message is received on a
// given channel. It may be called multiple times on the same reaction. It is
// incumbent on the user of the API to filter such called by message ID.
//
// Messages may arrive our of order, so a reply in theory can arrive before the
// initial message. As a result, it may be important to buffer reactions.
//
// Parameters:
//  - channelID - Marshalled bytes of the channel [id.ID] (Uint8Array).
//  - messageID - The bytes of the [channel.MessageID] of the received message
//    (Uint8Array).
//  - reactionTo - The [channel.MessageID] for the message that received a reply
//    (Uint8Array).
//  - senderUsername - The username of the sender of the message (string).
//  - reaction - The contents of the reaction message (string).
//  - identity - JSON of the sender's public ([channel.Identity]) (Uint8Array).
//  - timestamp - Time the message was received; represented as nanoseconds
//    since unix epoch (int).
//  - lease - The number of nanoseconds that the message is valid for (int).
//  - roundId - The ID of the round that the message was received on (int).
//  - msgType - The type of message ([channels.MessageType]) to send (int).
//  - status - The [channels.SentStatus] of the message (int).
//
// Statuses will be enumerated as such:
//  Sent      =  0
//  Delivered =  1
//  Failed    =  2
//
// Returns:
//  - A non-negative unique UUID for the message that it can be referenced by
//    later with [eventModel.UpdateSentStatus].
func (em *eventModel) ReceiveReaction(channelID, messageID, reactionTo []byte,
	senderUsername, reaction string, identity []byte, timestamp, lease, roundId,
	msgType, status int64) int64 {
	uuid := em.receiveReaction(utils.CopyBytesToJS(channelID),
		utils.CopyBytesToJS(messageID), utils.CopyBytesToJS(reactionTo),
		senderUsername, reaction, utils.CopyBytesToJS(identity),
		timestamp, lease, roundId, msgType, status)

	return int64(uuid.Int())
}

// UpdateSentStatus is called whenever the sent status of a message has
// changed.
//
// Parameters:
//  - uuid - The unique identifier for the message (int).
//  - messageID - The bytes of the [channel.MessageID] of the received message
//    (Uint8Array).
//  - timestamp - Time the message was received; represented as nanoseconds
//    since unix epoch (int).
//  - roundId - The ID of the round that the message was received on (int).
//  - status - The [channels.SentStatus] of the message (int).
//
// Statuses will be enumerated as such:
//  Sent      =  0
//  Delivered =  1
//  Failed    =  2
func (em *eventModel) UpdateSentStatus(
	uuid int64, messageID []byte, timestamp, roundID, status int64) {
	em.updateSentStatus(
		uuid, utils.CopyBytesToJS(messageID), timestamp, roundID, status)
}
