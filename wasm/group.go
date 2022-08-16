////////////////////////////////////////////////////////////////////////////////
// Copyright © 2020 xx network SEZC                                           //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file                                                               //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"gitlab.com/elixxir/client/bindings"
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

// newGroupChatJS creates a new Javascript compatible object
// (map[string]interface{}) that matches the GroupChat structure.
func newGroupChatJS(api *bindings.GroupChat) map[string]interface{} {
	gc := GroupChat{api}
	gcMap := map[string]interface{}{
		"MakeGroup":     js.FuncOf(gc.MakeGroup),
		"ResendRequest": js.FuncOf(gc.ResendRequest),
		"JoinGroup":     js.FuncOf(gc.JoinGroup),
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
//  - args[0] - ID of E2e object in tracker (int).
//  - args[1] - Javascript object that has functions that implement the
//    [bindings.GroupRequest] interface.
//  - args[2] - Javascript object that has functions that implement the
//    [bindings.GroupChatProcessor] interface.
//
// Returns:
//  - Javascript representation of the GroupChat object.
//  - Throws a TypeError if creating the GroupChat fails.
func NewGroupChat(_ js.Value, args []js.Value) interface{} {
	requestFunc := &groupRequest{args[1].Get("Callback").Invoke}
	p := &groupChatProcessor{
		args[2].Get("Process").Invoke, args[2].Get("String").Invoke}

	api, err := bindings.NewGroupChat(args[0].Int(), requestFunc, p)
	if err != nil {
		Throw(TypeError, err.Error())
		return nil
	}

	return newGroupChatJS(api)
}

// MakeGroup creates a new Group and sends a group request to all members in the
// group.
//
// Parameters:
//  - args[0] - JSON of array of [id.ID]; it contains the IDs of members the
//    user wants to add to the group (Uint8Array).
//  - args[1] - the initial message sent to all members in the group. This is an
//    optional parameter and may be nil (Uint8Array).
//  - args[2] - the name of the group decided by the creator. This is an
//    optional  parameter and may be nil. If nil the group will be assigned the
//    default name (Uint8Array).
//
// Returns:
//  - JSON of [bindings.GroupReport], which can be passed into
//    Cmix.WaitForRoundResult to see if the group request message send
//    succeeded.
//  - Throws a TypeError if making the group fails.
func (g *GroupChat) MakeGroup(_ js.Value, args []js.Value) interface{} {
	// (membershipBytes, message, name []byte) ([]byte, error)
	membershipBytes := CopyBytesToGo(args[0])
	message := CopyBytesToGo(args[1])
	name := CopyBytesToGo(args[2])

	report, err := g.api.MakeGroup(membershipBytes, message, name)
	if err != nil {
		Throw(TypeError, err.Error())
		return nil
	}

	return CopyBytesToJS(report)
}

// ResendRequest resends a group request to all members in the group.
//
// Parameters:
//  - args[0] - group's ID (Uint8Array). This can be found in the report
//  returned by GroupChat.MakeGroup.
//
// Returns:
//  - JSON of [bindings.GroupReport] (Uint8Array), which can be passed into
//    Cmix.WaitForRoundResult to see if the group request message send
//    succeeded.
//  - Throws a TypeError if resending the request fails.
func (g *GroupChat) ResendRequest(_ js.Value, args []js.Value) interface{} {
	report, err := g.api.ResendRequest(CopyBytesToGo(args[0]))
	if err != nil {
		Throw(TypeError, err.Error())
		return nil
	}

	return CopyBytesToJS(report)
}

// JoinGroup allows a user to join a group when a request is received.
// If an error is returned, handle it properly first; you may then retry later
// with the same trackedGroupId.
//
// Parameters:
//  - args[0] - ID of the Group object in tracker (int). This is received by
//    [bindings.GroupRequest.Callback].
//
// Returns:
//  - Throws a TypeError if joining the group fails.
func (g *GroupChat) JoinGroup(_ js.Value, args []js.Value) interface{} {
	err := g.api.JoinGroup(args[0].Int())
	if err != nil {
		Throw(TypeError, err.Error())
		return nil
	}

	return nil
}

// LeaveGroup deletes a group so a user no longer has access.
//
// Parameters:
//  - args[0] - group's ID (Uint8Array). This can be found in the report
//    returned by GroupChat.MakeGroup.
//
// Returns:
//  - Throws a TypeError if leaving the group fails.
func (g *GroupChat) LeaveGroup(_ js.Value, args []js.Value) interface{} {
	err := g.api.LeaveGroup(CopyBytesToGo(args[0]))
	if err != nil {
		Throw(TypeError, err.Error())
		return nil
	}

	return nil
}

// Send is the bindings-level function for sending to a group.
//
// Parameters:
//  - args[0] - group's ID (Uint8Array). This can be found in the report
//    returned by GroupChat.MakeGroup.
//  - args[1] - the message that the user wishes to send to the group
//    (Uint8Array).
//  - args[2] - the tag associated with the message (string). This tag may be
//    empty.
//
// Returns:
//  - JSON of [bindings.GroupSendReport] (Uint8Array), which can be passed into
//    Cmix.WaitForRoundResult to see if the group message send succeeded.
func (g *GroupChat) Send(_ js.Value, args []js.Value) interface{} {
	groupId := CopyBytesToGo(args[0])
	message := CopyBytesToGo(args[1])

	report, err := g.api.Send(groupId, message, args[2].String())
	if err != nil {
		Throw(TypeError, err.Error())
		return nil
	}

	return CopyBytesToJS(report)
}

// GetGroups returns a list of group IDs that the user is a member of.
//
// Returns:
//  - JSON of array of [id.ID] representing all group ID's (Uint8Array).
//  - Throws a TypeError if getting the groups fails.
func (g *GroupChat) GetGroups(js.Value, []js.Value) interface{} {
	// () ([]byte, error)
	groups, err := g.api.GetGroups()
	if err != nil {
		Throw(TypeError, err.Error())
		return nil
	}

	return CopyBytesToJS(groups)
}

// GetGroup returns the group with the group ID. If no group exists, then the
// error "failed to find group" is returned.
//
// Parameters:
//  - args[0] - group's ID (Uint8Array). This can be found in the report
//    returned by GroupChat.MakeGroup.
//
// Returns:
//  - Javascript representation of the Group object.
//  - Throws a TypeError if getting the group fails
func (g *GroupChat) GetGroup(_ js.Value, args []js.Value) interface{} {
	grp, err := g.api.GetGroup(CopyBytesToGo(args[0]))
	if err != nil {
		Throw(TypeError, err.Error())
		return nil
	}

	return newGroupJS(grp)
}

// NumGroups returns the number of groups the user is a part of.
//
// Returns:
//  - int
func (g *GroupChat) NumGroups(js.Value, []js.Value) interface{} {
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

// newGroupJS creates a new Javascript compatible object
// (map[string]interface{}) that matches the Group structure.
func newGroupJS(api *bindings.Group) map[string]interface{} {
	g := Group{api}
	gMap := map[string]interface{}{
		"GetName":        js.FuncOf(g.GetName),
		"GetID":          js.FuncOf(g.GetID),
		"GetTrackedID":   js.FuncOf(g.GetTrackedID),
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
//  - Uint8Array
func (g *Group) GetName(js.Value, []js.Value) interface{} {
	return CopyBytesToJS(g.api.GetName())
}

// GetID return the 33-byte unique group ID. This represents the id.ID object.
//
// Returns:
//  - Uint8Array
func (g *Group) GetID(js.Value, []js.Value) interface{} {
	return CopyBytesToJS(g.api.GetID())
}

// GetTrackedID returns the tracked ID of the Group object. This is used by the
// backend tracker.
//
// Returns:
//  - int
func (g *Group) GetTrackedID(js.Value, []js.Value) interface{} {
	return g.api.GetTrackedID()
}

// GetInitMessage returns initial message sent with the group request.
//
// Returns:
//  - Uint8Array
func (g *Group) GetInitMessage(js.Value, []js.Value) interface{} {
	return CopyBytesToJS(g.api.GetInitMessage())
}

// GetCreatedNano returns the time the group was created in nanoseconds. This is
// also the time the group requests were sent.
//
// Returns:
//  - int
func (g *Group) GetCreatedNano(js.Value, []js.Value) interface{} {
	return g.api.GetCreatedNano()
}

// GetCreatedMS returns the time the group was created in milliseconds. This is
// also the time the group requests were sent.
//
// Returns:
//  - int
func (g *Group) GetCreatedMS(js.Value, []js.Value) interface{} {
	return g.api.GetCreatedMS()
}

// GetMembership retrieves a list of group members. The list is in order;
// the first contact is the leader/creator of the group.
// All subsequent members are ordered by their ID.
//
// Returns:
//  - JSON of [group.Membership] (Uint8Array).
//  - Throws a TypeError if marshalling fails.
func (g *Group) GetMembership(js.Value, []js.Value) interface{} {
	membership, err := g.api.GetMembership()
	if err != nil {
		Throw(TypeError, err.Error())
		return nil
	}

	return CopyBytesToJS(membership)
}

// Serialize serializes the Group.
//
// Returns:
//  - Byte representation of the Group (Uint8Array).
func (g *Group) Serialize(js.Value, []js.Value) interface{} {
	return CopyBytesToJS(g.api.Serialize())
}

////////////////////////////////////////////////////////////////////////////////
// Callbacks                                                                  //
////////////////////////////////////////////////////////////////////////////////

// groupRequest wraps Javascript callbacks to adhere to the
// [bindings.GroupRequest] interface.
type groupRequest struct {
	callback func(args ...interface{}) js.Value
}

func (gr *groupRequest) Callback(g *bindings.Group) {
	gr.callback(newGroupJS(g))
}

// groupChatProcessor wraps Javascript callbacks to adhere to the
// [bindings.GroupChatProcessor] interface.
type groupChatProcessor struct {
	callback func(args ...interface{}) js.Value
	string   func(args ...interface{}) js.Value
}

func (gcp *groupChatProcessor) Process(decryptedMessage, msg,
	receptionId []byte, ephemeralId, roundId int64, err error) {
	gcp.callback(CopyBytesToJS(decryptedMessage), CopyBytesToJS(msg),
		CopyBytesToJS(receptionId), ephemeralId, roundId, err.Error())
}

func (gcp *groupChatProcessor) String() string {
	return gcp.string().String()
}
