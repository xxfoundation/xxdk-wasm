////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"syscall/js"

	jww "github.com/spf13/jwalterweatherman"

	"gitlab.com/elixxir/client/v4/bindings"
	"gitlab.com/elixxir/client/v4/dm"
	"gitlab.com/elixxir/crypto/codename"
	"gitlab.com/elixxir/wasm-utils/exception"
	"gitlab.com/elixxir/wasm-utils/utils"
	indexDB "gitlab.com/elixxir/xxdk-wasm/indexedDb/worker/dm"
)

////////////////////////////////////////////////////////////////////////////////
// Basic Channel API                                                          //
////////////////////////////////////////////////////////////////////////////////

// DMClient wraps the [bindings.DMClient] object so its methods can be wrapped
// to be Javascript compatible.
type DMClient struct {
	api *bindings.DMClient
}

// newDMClientJS creates a new Javascript compatible object (map[string]any)
// that matches the [DMClient] structure.
func newDMClientJS(api *bindings.DMClient) map[string]any {
	cm := DMClient{api}
	dmClientMap := map[string]any{
		// Basic Channel API
		"GetID": js.FuncOf(cm.GetID),

		// Identity and Nickname Controls
		"GetPublicKey":          js.FuncOf(cm.GetPublicKey),
		"GetToken":              js.FuncOf(cm.GetToken),
		"GetIdentity":           js.FuncOf(cm.GetIdentity),
		"ExportPrivateIdentity": js.FuncOf(cm.ExportPrivateIdentity),
		"GetNickname":           js.FuncOf(cm.GetNickname),
		"SetNickname":           js.FuncOf(cm.SetNickname),
		"BlockPartner":          js.FuncOf(cm.BlockPartner),
		"UnblockPartner":        js.FuncOf(cm.UnblockPartner),
		"IsBlocked":             js.FuncOf(cm.IsBlocked),
		"GetBlockedPartners":    js.FuncOf(cm.GetBlockedPartners),
		"GetDatabaseName":       js.FuncOf(cm.GetDatabaseName),

		// Share URL
		"GetShareURL": js.FuncOf(cm.GetShareURL),

		// DM Sending Methods and Reports
		"SendText":     js.FuncOf(cm.SendText),
		"SendReply":    js.FuncOf(cm.SendReply),
		"SendReaction": js.FuncOf(cm.SendReaction),
		"SendInvite":   js.FuncOf(cm.SendInvite),
		"SendSilent":   js.FuncOf(cm.SendSilent),
		"Send":         js.FuncOf(cm.Send),

		// Notifications
		"GetNotificationLevel": js.FuncOf(cm.GetNotificationLevel),
		"SetMobileNotificationsLevel": js.FuncOf(
			cm.SetMobileNotificationsLevel),
	}

	return dmClientMap
}

// newDmNotificationUpdate adds the callbacks from the Javascript object.
func newDmNotificationUpdate(value js.Value) *dmNotificationUpdate {
	return &dmNotificationUpdate{callback: utils.WrapCB(value, "Callback")}
}

// dmNotificationUpdate wraps Javascript callbacks to adhere to the
// [bindings.DmNotificationUpdate] interface.
type dmNotificationUpdate struct {
	callback func(args ...any) js.Value
}

// Callback is called everytime there is an update to the notification filter
// or notification status for DMs.
//
// Parameters:
//   - nfJSON - JSON of [dm.NotificationFilter], which is passed into
//     [GetDmNotificationReportsForMe] to filter DM notifications for the user.
//   - changedStateListJSON - JSON of a slice of [dm.NotificationState]. It
//     includes all added or changed notification states for DM conversations.
//   - deletedListJSON - JSON of a slice of [ed25519.PublicKey]. It includes
//     conversation that were deleted.
//
// Example nfJSON:
//
//	{
//	  "identifier": "MWL6mvtZ9UUm7jP3ainyI4erbRl+wyVaO5MOWboP0rA=",
//	  "myID": "AqDqg6Tcs359dBNRBCX7XHaotRDhz1ZRQNXIsGaubvID",
//	  "tags": [
//	    "61334HtH85DPIifvrM+JzRmLqfV5R4AMEmcPelTmFX0=",
//	    "zc/EPwtx5OKTVdwLcI15bghjJ7suNhu59PcarXE+m9o=",
//	    "FvArzVJ/082UEpMDCWJsopCLeLnxJV6NXINNkJTk3k8="
//	  ],
//	  "PublicKeys": {
//	    "61334HtH85DPIifvrM+JzRmLqfV5R4AMEmcPelTmFX0=": "b3HygDv8gjteune9wgBm3YtVuAo2foOusRmj0m5nl6E=",
//	    "FvArzVJ/082UEpMDCWJsopCLeLnxJV6NXINNkJTk3k8=": "uOLitBZcCh2TEW406jXHJ+Rsi6LybsH8R1u4Mxv/7hA=",
//	    "zc/EPwtx5OKTVdwLcI15bghjJ7suNhu59PcarXE+m9o=": "lqLD1EzZBxB8PbILUJIfFq4JI0RKThpUQuNlTNgZAWk="
//	  },
//	  "allowedTypes": {"1": {}, "2": {}}
//	}
//
// Example changedStateListJSON:
//
//	[
//	  {"pubKey": "lqLD1EzZBxB8PbILUJIfFq4JI0RKThpUQuNlTNgZAWk=", "level": 40},
//	  {"pubKey": "uOLitBZcCh2TEW406jXHJ+Rsi6LybsH8R1u4Mxv/7hA=", "level": 10},
//	  {"pubKey": "b3HygDv8gjteune9wgBm3YtVuAo2foOusRmj0m5nl6E=", "level": 10}
//	]
//
// Example deletedListJSON:
//
//	[
//	  "lqLD1EzZBxB8PbILUJIfFq4JI0RKThpUQuNlTNgZAWk=",
//	  "lqLD1EzZBxB8PbILUJIfFq4JI0RKThpUQuNlTNgZAWk="
//	]
func (dmNU *dmNotificationUpdate) Callback(
	nfJSON, changedStateListJSON, deletedListJSON []byte) {
	dmNU.callback(utils.CopyBytesToJS(nfJSON),
		utils.CopyBytesToJS(changedStateListJSON),
		utils.CopyBytesToJS(deletedListJSON))
}

// NewDMClient creates a new [DMClient] from a private identity
// ([codename.PrivateIdentity]), used for direct messaging.
//
// This is for instantiating a manager for an identity. For generating
// a new identity, use [codename.GenerateIdentity]. You should instantiate
// every load as there is no load function and associated state in
// this module.
//
// Parameters:
//   - args[0] - ID of [Cmix] object in tracker (int). This can be retrieved
//     using [Cmix.GetID].
//   - args[1] - ID of [Notifications] object in tracker. This can be retrieved
//     using [Notifications.GetID] (int).
//   - args[2] - Bytes of a private identity ([codename.PrivateIdentity]) that
//     is generated by [codename.GenerateIdentity] (Uint8Array).
//   - args[3] - A function that initialises and returns a Javascript object
//     that matches the [bindings.EventModel] interface. The function must match
//     the Build function in [bindings.EventModelBuilder].
//   - args[4] - A callback that is triggered everytime there is a change to the
//     notification status of a DM conversation It must be a Javascript object
//     that implements the callback in [bindings.DmNotificationUpdate].
//
// Returns:
//   - Javascript representation of the [DMClient] object.
//   - Throws an error if creating the manager fails.
func NewDMClient(_ js.Value, args []js.Value) any {
	cmixID := args[0].Int()
	notificationsID := args[1].Int()
	privateIdentity := utils.CopyBytesToGo(args[2])
	em := newDMReceiverBuilder(args[3])
	nu := newDmNotificationUpdate(args[4])

	cm, err :=
		bindings.NewDMClient(cmixID, notificationsID, privateIdentity, em, nu)
	if err != nil {
		exception.ThrowTrace(err)
		return nil
	}

	return newDMClientJS(cm)
}

// NewDMClientWithIndexedDb creates a new [DMClient] from a private identity
// ([codename.PrivateIdentity]) and an indexedDbWorker as a backend
// to manage the event model.
//
// This is for instantiating a manager for an identity. For generating
// a new identity, use [codename.GenerateIdentity]. You should instantiate
// every load as there is no load function and associated state in
// this module.
//
// This function initialises an indexedDbWorker database.
//
// Parameters:
//   - args[0] - ID of [Cmix] object in tracker (int). This can be retrieved
//     using [Cmix.GetID].
//   - args[1] - ID of [Notifications] object in tracker. This can be retrieved
//     using [Notifications.GetID] (int).
//   - args[2] - ID of [DbCipher] object in tracker (int). Create this object
//     with [NewDatabaseCipher] and get its id with [DbCipher.GetID].
//   - args[3] - Path to Javascript file that starts the worker (string).
//   - args[4] - Bytes of a private identity ([codename.PrivateIdentity]) that
//     is generated by [codename.GenerateIdentity] (Uint8Array).
//   - args[5] - The message receive callback. It is a function that takes in
//     the same parameters as [dm.MessageReceivedCallback]. On the Javascript
//     side, the UUID is returned as an int and the channelID as a Uint8Array.
//     The row in the database that was updated can be found using the UUID.
//     messageUpdate is true if the message already exists and was edited.
//     conversationUpdate is true if the Conversation was created or modified.
//   - args[6] - A callback that is triggered everytime there is a change to the
//     notification status of a DM conversation It must be a Javascript object
//     that implements the callback in [bindings.DmNotificationUpdate].
//
// Returns:
//   - Resolves to a Javascript representation of the [DMClient] object.
//   - Rejected with an error if loading indexedDbWorker or the manager fails.
//   - Throws an error if the cipher ID does not correspond to a cipher.
func NewDMClientWithIndexedDb(_ js.Value, args []js.Value) any {
	cmixID := args[0].Int()
	notificationsID := args[1].Int()
	cipherID := args[2].Int()
	wasmJsPath := args[3].String()
	privateIdentity := utils.CopyBytesToGo(args[4])
	messageReceivedCB := args[5]
	nu := newDmNotificationUpdate(args[6])

	cipher, err := dbCipherTrackerSingleton.get(cipherID)
	if err != nil {
		exception.ThrowTrace(err)
	}

	return newDMClientWithIndexedDb(cmixID, notificationsID, wasmJsPath,
		privateIdentity, messageReceivedCB, cipher, nu)
}

// NewDMClientWithIndexedDbUnsafe creates a new [DMClient] from a private
// identity ([codename.PrivateIdentity]) and an indexedDbWorker as a backend
// to manage the event model. However, the data is written in plain text and not
// encrypted. It is recommended that you do not use this in production.
//
// This is for instantiating a manager for an identity. For generating
// a new identity, use [codename.GenerateIdentity]. You should instantiate
// every load as there is no load function and associated state in
// this module.
//
// This function initialises an indexedDbWorker database.
//
// Parameters:
//   - args[0] - ID of [Cmix] object in tracker (int). This can be retrieved
//     using [Cmix.GetID].
//   - args[1] - ID of [Notifications] object in tracker. This can be retrieved
//     using [Notifications.GetID] (int).
//   - args[2] - Path to Javascript file that starts the worker (string).
//   - args[3] - Bytes of a private identity ([codename.PrivateIdentity]) that
//     is generated by [codename.GenerateIdentity] (Uint8Array).
//   - args[4] - The message receive callback. It is a function that takes in
//     the same parameters as [dm.MessageReceivedCallback]. On the Javascript
//     side, the UUID is returned as an int and the channelID as a Uint8Array.
//     The row in the database that was updated can be found using the UUID.
//     messageUpdate is true if the message already exists and was edited.
//     conversationUpdate is true if the Conversation was created or modified.
//   - args[5] - A callback that is triggered everytime there is a change to the
//     notification status of a DM conversation It must be a Javascript object
//     that implements the callback in [bindings.DmNotificationUpdate].
//
// Returns a promise:
//   - Resolves to a Javascript representation of the [DMClient] object.
//   - Rejected with an error if loading indexedDbWorker or the manager fails.
func NewDMClientWithIndexedDbUnsafe(_ js.Value, args []js.Value) any {
	cmixID := args[0].Int()
	notificationsID := args[1].Int()
	wasmJsPath := args[2].String()
	privateIdentity := utils.CopyBytesToGo(args[3])
	messageReceivedCB := args[4]
	nu := newDmNotificationUpdate(args[5])

	return newDMClientWithIndexedDb(cmixID, notificationsID, wasmJsPath,
		privateIdentity, messageReceivedCB, nil, nu)
}

func newDMClientWithIndexedDb(cmixID, notificationsID int, wasmJsPath string,
	privateIdentity []byte, cb js.Value, cipher *DbCipher, nuCB bindings.DmNotificationUpdate) any {

	messageReceivedCB := func(uuid uint64, pubKey ed25519.PublicKey,
		messageUpdate, conversationUpdate bool) {
		cb.Invoke(uuid, utils.CopyBytesToJS(pubKey[:]),
			messageUpdate, conversationUpdate)
	}

	promiseFn := func(resolve, reject func(args ...any) js.Value) {

		pi, err := codename.UnmarshalPrivateIdentity(privateIdentity)
		if err != nil {
			reject(exception.NewTrace(err))
		}
		dmPath := base64.RawStdEncoding.EncodeToString(pi.PubKey[:])
		model, err := indexDB.NewWASMEventModel(
			dmPath, wasmJsPath, cipher.api, messageReceivedCB)
		if err != nil {
			reject(exception.NewTrace(err))
		}

		cm, err := bindings.NewDMClientWithGoEventModel(
			cmixID, notificationsID, privateIdentity, model, nuCB)
		if err != nil {
			reject(exception.NewTrace(err))
		} else {
			resolve(newDMClientJS(cm))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// GetID returns the ECDH Public Key for this [DMClient] in the [DMClient]
// tracker.
//
// Returns:
//   - Tracker ID (int).
func (dmc *DMClient) GetID(js.Value, []js.Value) any {
	return dmc.api.GetID()
}

// GetPublicKey returns the bytes of the public key for this client.
//
// Returns:
//   - Public key (Uint8Array).
func (dmc *DMClient) GetPublicKey(js.Value, []js.Value) any {
	return dmc.api.GetPublicKey()
}

// GetToken returns the DM token of this client.
func (dmc *DMClient) GetToken(js.Value, []js.Value) any {
	return dmc.api.GetToken()
}

// GetIdentity returns the public identity associated with this client.
//
// Returns:
//   - JSON [codename.Identity] (Uint8Array).
func (dmc *DMClient) GetIdentity(js.Value, []js.Value) any {
	return utils.CopyBytesToJS(dmc.api.GetIdentity())
}

// ExportPrivateIdentity encrypts and exports the private identity to a portable
// string.
//
// Parameters:
//   - args[0] - Password to encrypt the identity with (string).
//
// Returns:
//   - Encrypted private identity bytes (Uint8Array).
//   - Throws TypeError if exporting the identity fails.
func (dmc *DMClient) ExportPrivateIdentity(_ js.Value, args []js.Value) any {
	i, err := dmc.api.ExportPrivateIdentity(args[0].String())
	if err != nil {
		exception.ThrowTrace(err)
		return nil
	}

	return utils.CopyBytesToJS(i)
}

// GetNickname gets the nickname associated with this DM user. Throws an error
// if no nickname is set.
//
// Returns:
//   - The nickname (string).
//   - Throws an error if the channel has no nickname set.
func (dmc *DMClient) GetNickname(_ js.Value, _ []js.Value) any {
	nickname, err := dmc.api.GetNickname()
	if err != nil {
		exception.ThrowTrace(err)
		return nil
	}

	return nickname
}

// SetNickname sets the nickname to use for this user.
//
// Parameters:
//   - args[0] - The nickname to set (string).
//
// Returns:
//   - Throws an error if setting the nickname fails.
func (dmc *DMClient) SetNickname(_ js.Value, args []js.Value) any {
	err := dmc.api.SetNickname(args[0].String())
	if err != nil {
		exception.ThrowTrace(err)
		return nil
	}

	return nil
}

// BlockPartner prevents receiving messages and notifications from the partner.
//
// Parameters:
//   - args[0] - The partner's [ed25519.PublicKey] key to block (Uint8Array).
//
// Returns a promise that exits upon completion.
func (dmc *DMClient) BlockPartner(_ js.Value, args []js.Value) any {
	partnerPubKey := utils.CopyBytesToGo(args[0])
	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		dmc.api.BlockPartner(partnerPubKey)
		resolve()
	}
	return utils.CreatePromise(promiseFn)
}

// UnblockPartner unblocks a blocked partner to allow DM messages.
//
// Parameters:
//   - args[0] - The partner's [ed25519.PublicKey] to unblock (Uint8Array).
func (dmc *DMClient) UnblockPartner(_ js.Value, args []js.Value) any {
	partnerPubKey := utils.CopyBytesToGo(args[0])
	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		dmc.api.UnblockPartner(partnerPubKey)
		resolve()
	}
	return utils.CreatePromise(promiseFn)
}

// IsBlocked indicates if the given partner is blocked.
//
// Parameters:
//   - args[0] - The partner's [ed25519.PublicKey] public key to check
//     (Uint8Array).
//
// Returns:
//   - boolean
func (dmc *DMClient) IsBlocked(_ js.Value, args []js.Value) any {
	partnerPubKey := utils.CopyBytesToGo(args[0])
	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		isBlocked := dmc.api.IsBlocked(partnerPubKey)
		resolve(isBlocked)
	}
	return utils.CreatePromise(promiseFn)
}

// GetBlockedPartners returns all partners who are blocked by this user.
//
// Returns:
//   - JSON of an array of [ed25519.PublicKey] (Uint8Array)
//
// Example return:
//
//	[
//	  "TYWuCfyGBjNWDtl/Roa6f/o206yYPpuB6sX2kJZTe98=",
//	  "4JLRzgtW1SZ9c5pE+v0WwrGPj1t19AuU6Gg5IND5ymA=",
//	  "CWDqF1bnhulW2pko+zgmbDZNaKkmNtFdUgY4bTm2DhA="
//	]
func (dmc *DMClient) GetBlockedPartners(js.Value, []js.Value) any {
	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		blocked := utils.CopyBytesToJS(dmc.api.GetBlockedPartners())
		resolve(blocked)
	}
	return utils.CreatePromise(promiseFn)
}

////////////////////////////////////////////////////////////////////////////////
// Channel Sending Methods and Reports                                        //
////////////////////////////////////////////////////////////////////////////////

// SendText is used to send a formatted direct message to a user.
//
// Parameters:
//   - args[0] - The bytes of the public key of the partner's ED25519 signing
//     key (Uint8Array).
//   - args[1] - The token used to derive the reception ID for the partner (int).
//   - args[2] - The contents of the message. The message should be at most 510
//     bytes. This is expected to be Unicode, and thus a string data type is
//     expected (string).
//   - args[3] - The lease of the message. This will be how long the message is
//     valid until, in milliseconds. As per the [channels.Manager]
//     documentation, this has different meanings depending on the use case.
//     These use cases may be generic enough that they will not be enumerated
//     here (int).
//   - args[3] - JSON of [xxdk.CMIXParams.] If left empty, then
//     [GetDefaultCMixParams] will be used internally (Uint8Array).
//
// Returns a promise:
//   - Resolves to the JSON of [bindings.ChannelSendReport] (Uint8Array).
//   - Rejected with an error if sending fails.
func (dmc *DMClient) SendText(_ js.Value, args []js.Value) any {
	partnerPubKeyBytes := utils.CopyBytesToGo(args[0])
	partnerToken := int32(args[1].Int())
	message := args[2].String()
	leaseTimeMS := int64(args[3].Int())
	cmixParamsJSON := utils.CopyBytesToGo(args[4])

	jww.DEBUG.Printf("SendText(%s, %d, %s...)",
		base64.RawStdEncoding.EncodeToString(partnerPubKeyBytes)[:8],
		partnerToken, truncate(message, 10))

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		sendReport, err := dmc.api.SendText(partnerPubKeyBytes, partnerToken,
			message, leaseTimeMS, cmixParamsJSON)
		if err != nil {
			reject(exception.NewTrace(err))
		} else {
			resolve(utils.CopyBytesToJS(sendReport))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// SendReply is used to send a formatted direct message reply.
//
// If the message ID that the reply is sent to does not exist, then the other
// side will post the message as a normal message and not as a reply.
//
// The message will auto delete leaseTime after the round it is sent in, lasting
// forever if [bindings.ValidForever] is used.
//
// Parameters:
//   - args[0] - The bytes of the public key of the partner's ED25519 signing
//     key (Uint8Array).
//   - args[1] - The token used to derive the reception ID for the partner (int).
//   - args[2] - The contents of the reply message. The message should be at
//     most 510 bytes. This is expected to be Unicode, and thus a string data
//     type is expected (string).
//   - args[3] - The bytes of the [message.ID] of the message you wish to reply
//     to. This may be found in the [bindings.ChannelSendReport] if replying to
//     your own. Alternatively, if reacting to another user's message, you may
//     retrieve it via the [bindings.ChannelMessageReceptionCallback] registered
//     using [ChannelsManager.RegisterReceiveHandler] (Uint8Array).
//   - args[4] - The lease of the message. This will be how long the message is
//     valid until, in milliseconds. As per the [channels.Manager]
//     documentation, this has different meanings depending on the use case.
//     These use cases may be generic enough that they will not be enumerated
//     here (int).
//   - args[5] - JSON of [xxdk.CMIXParams.] If left empty, then
//     [GetDefaultCMixParams] will be used internally (Uint8Array).
//
// Returns a promise:
//   - Resolves to the JSON of [bindings.ChannelSendReport] (Uint8Array).
//   - Rejected with an error if sending fails.
func (dmc *DMClient) SendReply(_ js.Value, args []js.Value) any {
	partnerPubKeyBytes := utils.CopyBytesToGo(args[0])
	partnerToken := int32(args[1].Int())
	replyMessage := args[2].String()
	replyToBytes := utils.CopyBytesToGo(args[3])
	leaseTimeMS := int64(args[4].Int())
	cmixParamsJSON := utils.CopyBytesToGo(args[5])

	jww.DEBUG.Printf("SendReply(%s, %d, %s: %s...)",
		base64.RawStdEncoding.EncodeToString(partnerPubKeyBytes)[:8],
		partnerToken,
		base64.RawStdEncoding.EncodeToString(replyToBytes),
		truncate(replyMessage, 10))

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		sendReport, err := dmc.api.SendReply(partnerPubKeyBytes, partnerToken,
			replyMessage, replyToBytes, leaseTimeMS, cmixParamsJSON)
		if err != nil {
			reject(exception.NewTrace(err))
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
//   - args[0] - The bytes of the public key of the partner's ED25519 signing
//     key (Uint8Array).
//   - args[1] - The token used to derive the reception ID for the partner (int).
//   - args[2] - The user's reaction. This should be a single emoji with no
//     other characters. As such, a Unicode string is expected (string).
//   - args[3] - The bytes of the [message.ID] of the message you wish to react
//     to. This may be found in the [bindings.ChannelSendReport] if replying to
//     your own. Alternatively, if reacting to another user's message, you may
//     retrieve it via the [bindings.ChannelMessageReceptionCallback] registered
//     using [ChannelsManager.RegisterReceiveHandler] (Uint8Array).
//   - args[3] - JSON of [xxdk.CMIXParams]. If left empty
//     [bindings.GetDefaultCMixParams] will be used internally (Uint8Array).
//
// Returns a promise:
//   - Resolves to the JSON of [bindings.ChannelSendReport] (Uint8Array).
//   - Rejected with an error if sending fails.
func (dmc *DMClient) SendReaction(_ js.Value, args []js.Value) any {
	partnerPubKeyBytes := utils.CopyBytesToGo(args[0])
	partnerToken := int32(args[1].Int())
	reaction := args[2].String()
	reactToBytes := utils.CopyBytesToGo(args[3])
	cmixParamsJSON := utils.CopyBytesToGo(args[4])

	jww.DEBUG.Printf("SendReaction(%s, %d, %s: %s...)",
		base64.RawStdEncoding.EncodeToString(partnerPubKeyBytes)[:8],
		partnerToken,
		base64.RawStdEncoding.EncodeToString(reactToBytes),
		truncate(reaction, 10))

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		sendReport, err := dmc.api.SendReaction(partnerPubKeyBytes,
			partnerToken, reaction, reactToBytes, cmixParamsJSON)
		if err != nil {
			reject(exception.NewTrace(err))
		} else {
			resolve(utils.CopyBytesToJS(sendReport))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// SendSilent is used to send to a channel a message with no notifications.
// Its primary purpose is to communicate new nicknames without calling [Send].
//
// It takes no payload intentionally as the message should be very lightweight.
//
// Parameters:
//   - args[0] - The bytes of the public key of the partner's ED25519
//     signing key (Uint8Array).
//   - args[1] - The token used to derive the reception ID for the partner
//     (int).
//   - args[2] - JSON of [xxdk.CMIXParams]. If left empty
//     [bindings.GetDefaultCMixParams] will be used internally (Uint8Array).
//
// Returns a promise:
//   - Resolves to the JSON of [bindings.ChannelSendReport] (Uint8Array).
//   - Rejected with an error if sending fails.
func (dmc *DMClient) SendSilent(_ js.Value, args []js.Value) any {
	var (
		partnerPubKeyBytes = utils.CopyBytesToGo(args[0])
		partnerToken       = int32(args[1].Int())
		cmixParamsJSON     = utils.CopyBytesToGo(args[2])
	)

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		sendReport, err := dmc.api.SendSilent(
			partnerPubKeyBytes, partnerToken, cmixParamsJSON)
		if err != nil {
			reject(exception.NewTrace(err))
		} else {
			resolve(utils.CopyBytesToJS(sendReport))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// SendInvite is used to send to a DM partner an invitation to another
// channel.
//
// If the channel ID for the invitee channel is not recognized by the Manager,
// then an error will be returned.
//
// Parameters:
//   - args[0] - The bytes of the public key of the partner's ED25519 signing
//     key (Uint8Array).
//   - args[1] - The token used to derive the reception ID for the partner (int).
//   - args[2] - JSON of the invitee channel [id.ID].
//     This can be retrieved from [GetChannelJSON]. (Uint8Array).
//   - args[3] - The contents of the message. The message should be at most 510
//     bytes. This is expected to be Unicode, and thus a string data type is
//     expected.
//   - args[4] - The URL to append the channel info to.
//   - args[5] - A JSON marshalled [xxdk.CMIXParams]. This may be empty,
//     and GetDefaultCMixParams will be used internally.
//
// Returns a promise:
//   - Resolves to the JSON of [bindings.ChannelSendReport] (Uint8Array).
//   - Rejected with an error if sending fails.
func (dmc *DMClient) SendInvite(_ js.Value, args []js.Value) any {
	var (
		partnerPubKeyBytes = utils.CopyBytesToGo(args[0])
		partnerToken       = int32(args[1].Int())
		inviteToJSON       = utils.CopyBytesToGo(args[2])
		msg                = args[3].String()
		host               = args[4].String()
		cmixParamsJSON     = utils.CopyBytesToGo(args[5])
	)

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		sendReport, err := dmc.api.SendInvite(
			partnerPubKeyBytes, partnerToken, inviteToJSON, msg, host,
			cmixParamsJSON)
		if err != nil {
			reject(exception.NewTrace(err))
		} else {
			resolve(utils.CopyBytesToJS(sendReport))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// Send is used to send a raw message. In general, it
// should be wrapped in a function that defines the wire protocol.
//
// If the final message, before being sent over the wire, is too long, this will
// return an error. Due to the underlying encoding using compression, it is not
// possible to define the largest payload that can be sent, but it will always
// be possible to send a payload of 802 bytes at minimum.
//
// The meaning of leaseTimeMS depends on the use case.
//
// Parameters:
//   - args[0] - Marshalled bytes of the partner pubkyey  (Uint8Array).
//   - args[1] - The token used to derive the reception ID for the partner
//     (int).
//   - args[2] - The message type of the message. This will be a valid
//     [dm.MessageType] (int)
//   - args[3] - The contents of the message. This need not be of data type
//     string, as the message could be a specified format that the channel may
//     recognize (Uint8Array)
//   - args[4] - The bytes of the [message.ID] of the message you wish to react
//     to. This may be found in the [bindings.ChannelSendReport] if replying to
//     your own. Alternatively, if reacting to another user's message, you may
//     retrieve it via the [bindings.ChannelMessageReceptionCallback] registered
//     using [ChannelsManager.RegisterReceiveHandler] (Uint8Array).
//   - args[5] - JSON of [xxdk.CMIXParams]. If left empty
//     [bindings.GetDefaultCMixParams] will be used internally (Uint8Array).
//
// Returns a promise:
//   - Resolves to the JSON of [bindings.ChannelSendReport] (Uint8Array).
//   - Rejected with an error if sending fails.
func (dmc *DMClient) Send(_ js.Value, args []js.Value) any {
	partnerPubKeyBytes := utils.CopyBytesToGo(args[0])
	partnerToken := int32(args[1].Int())
	messageType := args[2].Int()
	plaintext := utils.CopyBytesToGo(args[3])
	leaseTimeMS := int64(args[4].Int())
	cmixParamsJSON := utils.CopyBytesToGo(args[5])

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		sendReport, err := dmc.api.Send(partnerPubKeyBytes, partnerToken,
			messageType, plaintext, leaseTimeMS, cmixParamsJSON)
		if err != nil {
			reject(exception.NewTrace(err))
		} else {
			resolve(utils.CopyBytesToJS(sendReport))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// GetDatabaseName returns the storage tag, so users listening to the database
// can separately listen and read updates there.
//
// Returns:
//   - The storage tag (string).
func (dmc *DMClient) GetDatabaseName(js.Value, []js.Value) any {
	return base64.RawStdEncoding.EncodeToString(dmc.api.GetPublicKey()) +
		"_speakeasy_dm"
}

////////////////////////////////////////////////////////////////////////////////
// DM Share URL                                                          //
////////////////////////////////////////////////////////////////////////////////

// DMShareURL is returned from [DMClient.GetShareURL]. It includes the
// user's share URL.
//
// JSON example for a user:
//
//	{
//	 "url": "https://internet.speakeasy.tech/?l=32&m=5&p=EfDzQDa4fQ5BoqNIMbECFDY9ckRr_fadd8F1jE49qJc%3D&t=4231817746&v=1",
//	 "password": "hunter2",
//	}
type DMShareURL struct {
	URL      string `json:"url"`
	Password string `json:"password"`
}

// DMUser is returned from [DecodeDMShareURL]. It includes the token
// and public key of the user who created the URL.
//
// JSON example for a user:
//
//	{
//	 "token": 4231817746,
//	 "publicKey": "EfDzQDa4fQ5BoqNIMbECFDY9ckRr/fadd8F1jE49qJc="
//	}
type DMUser struct {
	Token     int32  `json:"token"`
	PublicKey []byte `json:"publicKey"`
}

// GetShareURL generates a URL that can be used to share a URL to initiate a
// direct messages with this user.
//
// Parameters:
//   - args[0] - The URL to append the DM info to (string).
//
// Returns:
//   - JSON of [DMShareURL] (Uint8Array).
func (dmc *DMClient) GetShareURL(_ js.Value, args []js.Value) any {
	host := args[0].String()
	urlReport, err := dmc.api.GetShareURL(host)
	if err != nil {
		exception.ThrowTrace(err)
		return nil
	}

	return utils.CopyBytesToJS(urlReport)
}

// GetNotificationLevel Gets the notification level for a given dm pubkey
//
// Parameters:
//   - args[0] - partnerPublic key (Uint8Array)
//
// Returns:
//   - int of notification level
func (dmc *DMClient) GetNotificationLevel(_ js.Value, args []js.Value) any {
	partnerPubKey := utils.CopyBytesToGo(args[0])

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		level, err := dmc.api.GetNotificationLevel(partnerPubKey)
		if err != nil {
			reject(exception.NewTrace(err))
		} else {
			resolve(level)
		}
	}

	return utils.CreatePromise(promiseFn)
}

// SetMobileNotificationsLevel sets the notification level for a given pubkey.
//
// Parameters:
//   - args[0] - partnerPublicKey (Uint8Array)
//   - args[1] - the notification level (integer)
//
// Returns:
//   - error or nothing
func (dmc *DMClient) SetMobileNotificationsLevel(_ js.Value,
	args []js.Value) any {
	partnerPubKey := utils.CopyBytesToGo(args[0])
	level := args[1].Int()

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		err := dmc.api.SetMobileNotificationsLevel(partnerPubKey, level)
		if err != nil {
			reject(exception.NewTrace(err))
		} else {
			resolve()
		}
	}

	return utils.CreatePromise(promiseFn)
}

// DecodeDMShareURL decodes the user's URL into a [DMUser].
//
// Parameters:
//   - args[0] - The user's share URL. Should be received from another user or
//     generated via [DMClient.GetShareURL] (string).
//
// Returns:
//   - JSON of [DMUser] (Uint8Array).
func DecodeDMShareURL(_ js.Value, args []js.Value) any {
	url := args[0].String()
	report, err := bindings.DecodeDMShareURL(url)
	if err != nil {
		exception.ThrowTrace(err)
		return nil
	}

	return utils.CopyBytesToJS(report)
}

// GetDmNotificationReportsForMe checks the notification data against the filter
// list to determine which notifications belong to the user. A list of
// notification reports is returned detailing all notifications for the user.
//
// Parameters:
//   - args[0] - notificationFilterJSON - JSON (Uint8Array) of
//     [dm.NotificationFilter].
//   - args[1] - notificationDataCSV - CSV containing notification data.
//
// Example JSON of a slice of [dm.NotificationFilter]:
//
//	{
//	  "identifier": "MWL6mvtZ9UUm7jP3ainyI4erbRl+wyVaO5MOWboP0rA=",
//	  "myID": "AqDqg6Tcs359dBNRBCX7XHaotRDhz1ZRQNXIsGaubvID",
//	  "tags": [
//	    "61334HtH85DPIifvrM+JzRmLqfV5R4AMEmcPelTmFX0=",
//	    "zc/EPwtx5OKTVdwLcI15bghjJ7suNhu59PcarXE+m9o=",
//	    "FvArzVJ/082UEpMDCWJsopCLeLnxJV6NXINNkJTk3k8="
//	  ],
//	  "PublicKeys": {
//	    "61334HtH85DPIifvrM+JzRmLqfV5R4AMEmcPelTmFX0=": "b3HygDv8gjteune9wgBm3YtVuAo2foOusRmj0m5nl6E=",
//	    "FvArzVJ/082UEpMDCWJsopCLeLnxJV6NXINNkJTk3k8=": "uOLitBZcCh2TEW406jXHJ+Rsi6LybsH8R1u4Mxv/7hA=",
//	    "zc/EPwtx5OKTVdwLcI15bghjJ7suNhu59PcarXE+m9o=": "lqLD1EzZBxB8PbILUJIfFq4JI0RKThpUQuNlTNgZAWk="
//	  },
//	  "allowedTypes": {"1": {}, "2": {}}
//	}
//
// Returns:
//   - []byte - JSON of a slice of [dm.NotificationReport].
//
// Example return:
//
//	[
//	  {"partner": "WUSO3trAYeBf4UeJ5TEL+Q4usoyFf0shda0YUmZ3z8k=", "type": 1},
//	  {"partner": "5MY652JsVv5YLE6wGRHIFZBMvLklACnT5UtHxmEOJ4o=", "type": 2}
//	]
func GetDmNotificationReportsForMe(_ js.Value, args []js.Value) any {
	notificationFilterJson := utils.CopyBytesToGo(args[0])
	notificationDataCsv := args[1].String()

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		forme, err := bindings.GetDmNotificationReportsForMe(
			notificationFilterJson, notificationDataCsv)
		if err != nil {
			reject(exception.NewTrace(err))
		} else {
			resolve(utils.CopyBytesToJS(forme))
		}
	}

	return utils.CreatePromise(promiseFn)
}

////////////////////////////////////////////////////////////////////////////////
// Event Model Logic                                                          //
////////////////////////////////////////////////////////////////////////////////

// dmReceiverBuilder adheres to the [bindings.DMReceiverBuilder] interface.
type dmReceiverBuilder struct {
	build func(args ...any) js.Value
}

// newDMReceiverBuilder maps the methods on the Javascript object to a new
// dmReceiverBuilder.
func newDMReceiverBuilder(arg js.Value) *dmReceiverBuilder {
	return &dmReceiverBuilder{build: arg.Invoke}
}

// Build initializes and returns the event model. It wraps a Javascript object
// that has all the methods in [bindings.EventModel] to make it adhere to the Go
// interface [bindings.EventModel].
func (emb *dmReceiverBuilder) Build(path string) bindings.DMReceiver {
	emJs := emb.build(path)
	return &dmReceiver{
		receive:          utils.WrapCB(emJs, "ReceiveText"),
		receiveText:      utils.WrapCB(emJs, "ReceiveText"),
		receiveReply:     utils.WrapCB(emJs, "ReceiveReply"),
		receiveReaction:  utils.WrapCB(emJs, "ReceiveReaction"),
		updateSentStatus: utils.WrapCB(emJs, "UpdateSentStatus"),
		deleteMessage:    utils.WrapCB(emJs, "DeleteMessage"),
		getConversation:  utils.WrapCB(emJs, "GetConversation"),
		getConversations: utils.WrapCB(emJs, "GetConversations"),
	}
}

// dmReceiver wraps Javascript callbacks to adhere to the [dm.EventModel]
// interface.
type dmReceiver struct {
	receive          func(args ...any) js.Value
	receiveText      func(args ...any) js.Value
	receiveReply     func(args ...any) js.Value
	receiveReaction  func(args ...any) js.Value
	updateSentStatus func(args ...any) js.Value
	deleteMessage    func(args ...any) js.Value
	getConversation  func(args ...any) js.Value
	getConversations func(args ...any) js.Value
}

// Receive is called when a raw direct message is received with unknown type.
// It may be called multiple times on the same message. It is incumbent on the
// user of the API to filter such called by message ID.
//
// The user must interpret the message type and perform their own message
// parsing.
//
// Parameters:
//   - messageID - The bytes of the [dm.MessageID] of the received message
//     (Uint8Array).
//   - nickname - The nickname of the sender of the message (string).
//   - text - The bytes content of the message (Uint8Array).
//   - partnerKey - The partner's [ed25519.PublicKey]. This is required to
//     respond (Uint8Array).
//   - senderKey - The sender's [ed25519.PublicKey] (Uint8Array).
//   - dmToken - The senders direct messaging token. This is required to respond
//     (int).
//   - codeset - The codeset version (int)
//   - timestamp - Time the message was received; represented as nanoseconds
//     since unix epoch (int).
//   - roundID - The ID of the round that the message was received on (int).
//   - mType - The type of message ([channels.MessageType]) to send (int).
//   - status - the [dm.SentStatus] of the message (int).
//
// Statuses will be enumerated as such:
//
//	Sent      =  0
//	Delivered =  1
//	Failed    =  2
//
// Returns:
//   - A non-negative unique UUID for the message that it can be referenced by
//     later with [dmReceiver.UpdateSentStatus].
func (em *dmReceiver) Receive(messageID []byte, nickname string, text,
	partnerKey, senderKey []byte, dmToken int32, codeset int, timestamp,
	roundId, mType, status int64) int64 {
	uuid := em.receive(utils.CopyBytesToJS(messageID), nickname,
		utils.CopyBytesToJS(text), utils.CopyBytesToJS(partnerKey),
		utils.CopyBytesToJS(senderKey),
		dmToken, codeset, timestamp, roundId, mType, status)

	return int64(uuid.Int())
}

// ReceiveText is called whenever a direct message is received that is a text
// type. It may be called multiple times on the same message. It is incumbent on
// the user of the API to filter such called by message ID.
//
// Messages may arrive our of order, so a reply in theory can arrive before the
// initial message. As a result, it may be important to buffer replies.
//
// Parameters:
//   - messageID - The bytes of the [dm.MessageID] of the received message
//     (Uint8Array).
//   - nickname - The nickname of the sender of the message (string).
//   - text - The content of the message (string).
//   - partnerKey - The partner's [ed25519.PublicKey]. This is required to
//     respond (Uint8Array).
//   - senderKey - The sender's [ed25519.PublicKey] (Uint8Array).
//   - dmToken - The senders direct messaging token. This is required to respond
//     (int).
//   - codeset - The codeset version (int)
//   - timestamp - Time the message was received; represented as nanoseconds
//     since unix epoch (int).
//   - roundId - The ID of the round that the message was received on (int).
//   - msgType - The type of message ([channels.MessageType]) to send (int).
//   - status - The [channels.SentStatus] of the message (int).
//   - roundID - The ID of the round that the message was received on (int).
//   - status - the [dm.SentStatus] of the message (int).
//
// Statuses will be enumerated as such:
//
//	Sent      =  0
//	Delivered =  1
//	Failed    =  2
//
// Returns:
//   - A non-negative unique UUID for the message that it can be referenced by
//     later with [dmReceiver.UpdateSentStatus].
func (em *dmReceiver) ReceiveText(messageID []byte, nickname, text string,
	partnerKey, senderKey []byte, dmToken int32, codeset int, timestamp,
	roundId, status int64) int64 {

	uuid := em.receiveText(utils.CopyBytesToJS(messageID), nickname, text,
		utils.CopyBytesToJS(partnerKey), utils.CopyBytesToJS(senderKey),
		dmToken, codeset, timestamp, roundId, status)

	return int64(uuid.Int())
}

// ReceiveReply is called whenever a direct message is received that is a reply.
// It may be called multiple times on the same message. It is incumbent on the
// user of the API to filter such called by message ID.
//
// Messages may arrive our of order, so a reply in theory can arrive before the
// initial message. As a result, it may be important to buffer replies.
//
// Parameters:
//   - messageID - The bytes of the [dm.MessageID] of the received message
//     (Uint8Array).
//   - reactionTo - The [dm.MessageID] for the message that received a reply
//     (Uint8Array).
//   - nickname - The nickname of the sender of the message (string).
//   - text - The content of the message (string).
//   - partnerKey - The partner's [ed25519.PublicKey]. This is required to
//     respond (Uint8Array).
//   - senderKey - The sender's [ed25519.PublicKey] (Uint8Array).
//   - dmToken - The senders direct messaging token. This is required to respond
//     (int).
//   - codeset - The codeset version (int)
//   - timestamp - Time the message was received; represented as nanoseconds
//     since unix epoch (int).
//   - roundId - The ID of the round that the message was received on (int).
//   - msgType - The type of message ([channels.MessageType]) to send (int).
//   - status - The [channels.SentStatus] of the message (int).
//   - roundID - The ID of the round that the message was received on (int).
//   - status - the [dm.SentStatus] of the message (int).
//
// Statuses will be enumerated as such:
//
//	Sent      =  0
//	Delivered =  1
//	Failed    =  2
//
// Returns:
//   - A non-negative unique UUID for the message that it can be referenced by
//     later with [dmReceiver.UpdateSentStatus].
func (em *dmReceiver) ReceiveReply(messageID, reactionTo []byte, nickname,
	text string, partnerKey, senderKey []byte, dmToken int32, codeset int,
	timestamp, roundId, status int64) int64 {
	uuid := em.receiveReply(utils.CopyBytesToJS(messageID),
		utils.CopyBytesToJS(reactionTo), nickname, text,
		utils.CopyBytesToJS(partnerKey), utils.CopyBytesToJS(senderKey),
		dmToken, codeset, timestamp, roundId, status)

	return int64(uuid.Int())
}

// ReceiveReaction is called whenever a reaction to a direct message is
// received. It may be called multiple times on the same reaction. It is
// incumbent on the user of the API to filter such called by message ID.
//
// Messages may arrive our of order, so a reply in theory can arrive before the
// initial message. As a result, it may be important to buffer reactions.
//
// Parameters:
//   - messageID - The bytes of the [dm.MessageID] of the received message
//     (Uint8Array).
//   - reactionTo - The [dm.MessageID] for the message that received a reply
//     (Uint8Array).
//   - nickname - The nickname of the sender of the message (string).
//   - reaction - The content of the reaction message (string).
//   - partnerKey - The partner's [ed25519.PublicKey]. This is required to
//     respond (Uint8Array).
//   - senderKey - The sender's [ed25519.PublicKey] (Uint8Array).
//   - dmToken - The senders direct messaging token. This is required to respond
//     (int).
//   - codeset - The codeset version (int)
//   - timestamp - Time the message was received; represented as nanoseconds
//     since unix epoch (int).
//   - roundId - The ID of the round that the message was received on (int).
//   - msgType - The type of message ([channels.MessageType]) to send (int).
//   - status - The [channels.SentStatus] of the message (int).
//   - roundID - The ID of the round that the message was received on (int).
//   - status - the [dm.SentStatus] of the message (int).
//
// Statuses will be enumerated as such:
//
//	Sent      =  0
//	Delivered =  1
//	Failed    =  2
//
// Returns:
//   - A non-negative unique UUID for the message that it can be referenced by
//     later with [dmReceiver.UpdateSentStatus].
func (em *dmReceiver) ReceiveReaction(messageID, reactionTo []byte,
	nickname, reaction string, partnerKey, senderKey []byte, dmToken int32,
	codeset int, timestamp, roundId, status int64) int64 {
	uuid := em.receiveReaction(utils.CopyBytesToJS(messageID),
		utils.CopyBytesToJS(reactionTo), nickname, reaction,
		utils.CopyBytesToJS(partnerKey), utils.CopyBytesToJS(senderKey),
		dmToken, codeset, timestamp, roundId, status)

	return int64(uuid.Int())
}

// UpdateSentStatus is called whenever the sent status of a message has changed.
//
// Parameters:
//   - uuid - The unique identifier for the message (int).
//   - messageID - The bytes of the [channel.MessageID] of the received message
//     (Uint8Array).
//   - timestamp - Time the message was received; represented as nanoseconds
//     since unix epoch (int).
//   - roundId - The ID of the round that the message was received on (int).
//   - status - The [channels.SentStatus] of the message (int).
//
// Statuses will be enumerated as such:
//
//	Sent      =  0
//	Delivered =  1
//	Failed    =  2
func (em *dmReceiver) UpdateSentStatus(
	uuid int64, messageID []byte, timestamp, roundID, status int64) {
	em.updateSentStatus(
		uuid, utils.CopyBytesToJS(messageID), timestamp, roundID, status)
}

// DeleteMessage deletes the message with the given [message.ID] belonging to
// the sender. If the message exists and belongs to the sender, then it is
// deleted and [DeleteMessage] returns true. If it does not exist, it returns
// false.
//
// Parameters:
//   - messageID - The bytes of the [message.ID] of the message to delete
//     (Uint8Array).
//   - senderPubKey - The [ed25519.PublicKey] of the sender of the message
//     (Uint8Array).
func (em *dmReceiver) DeleteMessage(messageID, senderPubKey []byte) bool {
	return em.deleteMessage(
		utils.CopyBytesToJS(messageID), utils.CopyBytesToJS(senderPubKey)).Bool()
}

// GetConversation returns the conversation held by the model (receiver).
//
// Parameters:
//   - senderPubKey - The unique public key for the conversation (Uint8Array).
//
// Returns:
//   - JSON of [dm.ModelConversation] (Uint8Array).
func (em *dmReceiver) GetConversation(senderPubKey []byte) []byte {
	result := utils.CopyBytesToGo(
		em.getConversation(utils.CopyBytesToJS(senderPubKey)))

	var conversation dm.ModelConversation
	err := json.Unmarshal(result, &conversation)
	if err != nil {
		return nil
	}

	conversationsBytes, _ := json.Marshal(conversation)
	return conversationsBytes
}

// GetConversations returns all conversations held by the model (receiver).
//
// Returns:
//   - JSON of [][dm.ModelConversation] (Uint8Array).
func (em *dmReceiver) GetConversations() []byte {
	result := utils.CopyBytesToGo(em.getConversations())

	var conversations []dm.ModelConversation
	err := json.Unmarshal(result, &conversations)
	if err != nil {
		return nil
	}

	conversationsBytes, _ := json.Marshal(conversations)
	return conversationsBytes
}

// truncate truncates the string to length n. If the string is trimmed, then
// ellipses (...) are appended.
func truncate(s string, n int) string {
	if len(s)-3 <= n {
		return s
	}
	return s[:n] + "..."
}
