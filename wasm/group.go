////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"gitlab.com/elixxir/client/v4/bindings"
	"gitlab.com/elixxir/xxdk-wasm/utils"
	"syscall/js"
)

////////////////////////////////////////////////////////////////////////////////
// Group Chat                                                                 //
////////////////////////////////////////////////////////////////////////////////

// GroupChat wraps the [bindings.GroupChat] object so its methods can be wrapped
// to be Javascript compatible.
type GroupChat struct {
	api *bindings.GroupChat
}

// newGroupChatJS creates a new Javascript compatible object (map[string]any)
// that matches the [GroupChat] structure.
func newGroupChatJS(api *bindings.GroupChat) map[string]any {
	gc := GroupChat{api}
	gcMap := map[string]any{
		"MakeGroup":     js.FuncOf(gc.MakeGroup),
		"ResendRequest": js.FuncOf(gc.ResendRequest),
		"JoinGroup":     js.FuncOf(gc.JoinGroup),
		"LeaveGroup":    js.FuncOf(gc.LeaveGroup),
		"Send":          js.FuncOf(gc.Send),
		"GetGroups":     js.FuncOf(gc.GetGroups),
		"GetGroup":      js.FuncOf(gc.GetGroup),
		"NumGroups":     js.FuncOf(gc.NumGroups),
	}

	return gcMap
}

// NewGroupChat creates a bindings-layer group chat manager.
//
// Parameters:
//   - args[0] - ID of [E2e] object in tracker (int).
//   - args[1] - Javascript object that has functions that implement the
//     [bindings.GroupRequest] interface.
//   - args[2] - Javascript object that has functions that implement the
//     [bindings.GroupChatProcessor] interface.
//
// Returns:
//   - Javascript representation of the [GroupChat] object.
//   - Throws a TypeError if creating the [GroupChat] fails.
func NewGroupChat(_ js.Value, args []js.Value) any {
	requestFunc := &groupRequest{utils.WrapCB(args[1], "Callback")}
	p := &groupChatProcessor{
		utils.WrapCB(args[2], "Process"), utils.WrapCB(args[2], "String")}

	api, err := bindings.NewGroupChat(args[0].Int(), requestFunc, p)
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return newGroupChatJS(api)
}

// MakeGroup creates a new group and sends a group request to all members in the
// group.
//
// Parameters:
//   - args[0] - JSON of array of [id.ID]; it contains the IDs of members the
//     user wants to add to the group (Uint8Array).
//   - args[1] - The initial message sent to all members in the group. This is
//     an optional parameter and may be nil (Uint8Array).
//   - args[2] - The name of the group decided by the creator. This is an
//     optional  parameter and may be nil. If nil the group will be assigned the
//     default name (Uint8Array).
//
// Returns a promise:
//   - Resolves to the JSON of the [bindings.GroupReport], which can be passed
//     into [Cmix.WaitForRoundResult] to see if the send succeeded (Uint8Array).
//   - Rejected with an error if making the group fails.
func (g *GroupChat) MakeGroup(_ js.Value, args []js.Value) any {
	membershipBytes := utils.CopyBytesToGo(args[0])
	message := utils.CopyBytesToGo(args[1])
	name := utils.CopyBytesToGo(args[2])

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		sendReport, err := g.api.MakeGroup(membershipBytes, message, name)
		if err != nil {
			reject(utils.JsTrace(err))
		} else {
			resolve(utils.CopyBytesToJS(sendReport))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// ResendRequest resends a group request to all members in the group.
//
// Parameters:
//   - args[0] - The marshalled bytes of the group [id.ID] (Uint8Array). This
//     can be found in the report returned by [GroupChat.MakeGroup].
//
// Returns a promise:
//   - Resolves to the JSON of the [bindings.GroupReport], which can be passed
//     into [Cmix.WaitForRoundResult] to see if the send succeeded (Uint8Array).
//   - Rejected with an error if resending the request fails.
func (g *GroupChat) ResendRequest(_ js.Value, args []js.Value) any {
	groupId := utils.CopyBytesToGo(args[0])
	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		sendReport, err := g.api.ResendRequest(groupId)
		if err != nil {
			reject(utils.JsTrace(err))
		} else {
			resolve(utils.CopyBytesToJS(sendReport))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// JoinGroup allows a user to join a group when a request is received.
// If an error is returned, handle it properly first; you may then retry later
// with the same trackedGroupId.
//
// Parameters:
//   - args[0] - The result of calling [Group.Serialize] on any [bindings.Group]
//     object returned over the bindings (Uint8Array).
//
// Returns:
//   - Throws a TypeError if joining the group fails.
func (g *GroupChat) JoinGroup(_ js.Value, args []js.Value) any {
	err := g.api.JoinGroup(utils.CopyBytesToGo(args[0]))
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return nil
}

// LeaveGroup deletes a group so a user no longer has access.
//
// Parameters:
//   - args[0] - The marshalled bytes of the group [id.ID] (Uint8Array). This
//     can be found in the report returned by [GroupChat.MakeGroup].
//
// Returns:
//   - Throws a TypeError if leaving the group fails.
func (g *GroupChat) LeaveGroup(_ js.Value, args []js.Value) any {
	err := g.api.LeaveGroup(utils.CopyBytesToGo(args[0]))
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return nil
}

// Send is the bindings-level function for sending to a group.
//
// Parameters:
//   - args[0] - The marshalled bytes of the group [id.ID] (Uint8Array). This
//     can be found in the report returned by [GroupChat.MakeGroup].
//   - args[1] - The message that the user wishes to send to the group
//     (Uint8Array).
//   - args[2] - The tag associated with the message (string). This tag may be
//     empty.
//
// Returns a promise:
//   - Resolves to the JSON of the [bindings.GroupSendReport], which can be
//     passed into [Cmix.WaitForRoundResult] to see if the send succeeded
//     (Uint8Array).
//   - Rejected with an error if sending the message to the group fails.
func (g *GroupChat) Send(_ js.Value, args []js.Value) any {
	groupId := utils.CopyBytesToGo(args[0])
	message := utils.CopyBytesToGo(args[1])
	tag := args[2].String()

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		sendReport, err := g.api.Send(groupId, message, tag)
		if err != nil {
			reject(utils.JsTrace(err))
		} else {
			resolve(utils.CopyBytesToJS(sendReport))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// GetGroups returns a list of group IDs that the user is a member of.
//
// Returns:
//   - JSON of array of [id.ID] representing all group ID's (Uint8Array).
//   - Throws a TypeError if getting the groups fails.
func (g *GroupChat) GetGroups(js.Value, []js.Value) any {
	groups, err := g.api.GetGroups()
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return utils.CopyBytesToJS(groups)
}

// GetGroup returns the group with the group ID. If no group exists, then the
// error "failed to find group" is returned.
//
// Parameters:
//   - args[0] - The marshalled bytes of the group [id.ID] (Uint8Array). This
//     can be found in the report returned by [GroupChat.MakeGroup].
//
// Returns:
//   - Javascript representation of the [GroupChat] object.
//   - Throws a TypeError if getting the group fails.
func (g *GroupChat) GetGroup(_ js.Value, args []js.Value) any {
	grp, err := g.api.GetGroup(utils.CopyBytesToGo(args[0]))
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return newGroupJS(grp)
}

// NumGroups returns the number of groups the user is a part of.
//
// Returns:
//   - Number of groups (int).
func (g *GroupChat) NumGroups(js.Value, []js.Value) any {
	return g.api.NumGroups()
}

////////////////////////////////////////////////////////////////////////////////
// Group Structure                                                            //
////////////////////////////////////////////////////////////////////////////////

// Group wraps the [bindings.Group] object so its methods can be wrapped to be
// Javascript compatible.
type Group struct {
	api *bindings.Group
}

// newGroupJS creates a new Javascript compatible object (map[string]any) that
// matches the [Group] structure.
func newGroupJS(api *bindings.Group) map[string]any {
	g := Group{api}
	gMap := map[string]any{
		"GetName":        js.FuncOf(g.GetName),
		"GetID":          js.FuncOf(g.GetID),
		"GetInitMessage": js.FuncOf(g.GetInitMessage),
		"GetCreatedNano": js.FuncOf(g.GetCreatedNano),
		"GetCreatedMS":   js.FuncOf(g.GetCreatedMS),
		"GetMembership":  js.FuncOf(g.GetMembership),
		"Serialize":      js.FuncOf(g.Serialize),
	}

	return gMap
}

// GetName returns the name set by the user for the group.
//
// Returns:
//   - Group name (Uint8Array).
func (g *Group) GetName(js.Value, []js.Value) any {
	return utils.CopyBytesToJS(g.api.GetName())
}

// GetID return the 33-byte unique group ID. This represents the [id.ID] object.
//
// Returns:
//   - Marshalled bytes of the group [id.ID] (Uint8Array).
func (g *Group) GetID(js.Value, []js.Value) any {
	return utils.CopyBytesToJS(g.api.GetID())
}

// GetInitMessage returns initial message sent with the group request.
//
// Returns:
//   - Initial group message contents (Uint8Array).
func (g *Group) GetInitMessage(js.Value, []js.Value) any {
	return utils.CopyBytesToJS(g.api.GetInitMessage())
}

// GetCreatedNano returns the time the group was created in nanoseconds. This is
// also the time the group requests were sent.
//
// Returns:
//   - The time the group was created, in nanoseconds (int).
func (g *Group) GetCreatedNano(js.Value, []js.Value) any {
	return g.api.GetCreatedNano()
}

// GetCreatedMS returns the time the group was created in milliseconds. This is
// also the time the group requests were sent.
//
// Returns:
//   - The time the group was created, in milliseconds (int).
func (g *Group) GetCreatedMS(js.Value, []js.Value) any {
	return g.api.GetCreatedMS()
}

// GetMembership retrieves a list of group members. The list is in order;
// the first contact is the leader/creator of the group.
// All subsequent members are ordered by their ID.
//
// Returns:
//   - JSON of [group.Membership] (Uint8Array).
//   - Throws a TypeError if marshalling fails.
func (g *Group) GetMembership(js.Value, []js.Value) any {
	membership, err := g.api.GetMembership()
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return utils.CopyBytesToJS(membership)
}

// Serialize serializes the [Group].
//
// Returns:
//   - Byte representation of the [Group] (Uint8Array).
func (g *Group) Serialize(js.Value, []js.Value) any {
	return utils.CopyBytesToJS(g.api.Serialize())
}

// DeserializeGroup converts the results of [Group.Serialize] into a
// [bindings.Group] so that its methods can be called.
//
// Parameters:
//   - args[0] - Byte representation of the [bindings.Group] (Uint8Array).
//
// Returns:
//   - Javascript representation of the [GroupChat] object.
//   - Throws a TypeError if getting the group fails.
func DeserializeGroup(_ js.Value, args []js.Value) any {
	grp, err := bindings.DeserializeGroup(utils.CopyBytesToGo(args[0]))
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return newGroupJS(grp)
}

////////////////////////////////////////////////////////////////////////////////
// Callbacks                                                                  //
////////////////////////////////////////////////////////////////////////////////

// groupRequest wraps Javascript callbacks to adhere to the
// [bindings.GroupRequest] interface.
type groupRequest struct {
	callback func(args ...any) js.Value
}

// Callback is called when a group request is received.
//
// Parameters:
//   - g - Returns the JSON of [bindings.Group] (Uint8Array).
func (gr *groupRequest) Callback(g *bindings.Group) {
	gr.callback(newGroupJS(g))
}

// groupChatProcessor wraps Javascript callbacks to adhere to the
// [bindings.GroupChatProcessor] interface.
type groupChatProcessor struct {
	process func(args ...any) js.Value
	string  func(args ...any) js.Value
}

// Process decrypts and hands off the message to its internal down stream
// message processing system.
//
// Parameters:
//   - decryptedMessage - Returns the JSON of [bindings.GroupChatMessage]
//     (Uint8Array).
//   - msg - Returns the marshalled bytes of [format.Message] (Uint8Array).
//   - receptionId - Returns the marshalled bytes of the sender's [id.ID]
//     (Uint8Array).
//   - ephemeralId - Returns the [ephemeral.Id] of the sender (int).
//   - roundId - Returns the ID of the round sent on (int).
//   - err - Returns an error on failure (Error).
func (gcp *groupChatProcessor) Process(decryptedMessage, msg,
	receptionId []byte, ephemeralId, roundId int64, roundURL string, err error) {
	gcp.process(utils.CopyBytesToJS(decryptedMessage),
		utils.CopyBytesToJS(msg), utils.CopyBytesToJS(receptionId), ephemeralId,
		roundId, roundURL, utils.JsTrace(err))
}

// String returns a name identifying this processor. Used for debugging.
func (gcp *groupChatProcessor) String() string {
	return gcp.string().String()
}
