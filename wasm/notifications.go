////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"syscall/js"

	"gitlab.com/elixxir/client/v4/bindings"
	"gitlab.com/elixxir/client/v4/notifications"
	"gitlab.com/elixxir/wasm-utils/utils"
)

type Notifications struct {
	api bindings.Notifications
}

// newNotificationsJS wrapts the bindings Noticiation object and implements
// wrappers in JS for all it's functionality.
func newNotificationsJS(api bindings.Notifications) map[string]any {
	n := Notifications{api}
	notificationsImplJS := map[string]any{
		"AddToken":    js.FuncOf(n.AddToken),
		"RemoveToken": js.FuncOf(n.RemoveToken),
		"SetMaxState": js.FuncOf(n.SetMaxState),
		"GetMaxState": js.FuncOf(n.GetMaxState),
		"GetID":       js.FuncOf(n.GetID),
	}
	return notificationsImplJS
}

// LoadNotifications returns a JS wrapped implementation of
// [bindings.Notifications].
//
// Parameters:
//   - args[0] - the cMixID integer
//
// Returns a notifications object or throws an error
func LoadNotifications(_ js.Value, args []js.Value) any {
	cMixID := args[0].Int()
	api, err := bindings.LoadNotifications(cMixID)
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return newNotificationsJS(api)
}

// LoadNotificationsDummy returns a JS wrapped implementation of
// [bindings.Notifications] with a dummy notifications implementation.
//
// Parameters:
//   - args[0] - the cMixID integer
//
// Returns a notifications object or throws an error
func LoadNotificationsDummy(_ js.Value, args []js.Value) any {
	cMixID := args[0].Int()
	api, err := bindings.LoadNotificationsDummy(cMixID)
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return newNotificationsJS(api)
}

// GetID returns the bindings ID for the [bindings.Notifications] object
func (n *Notifications) GetID(js.Value, []js.Value) any {
	return n.api.GetID()
}

// AddToken implements [bindings.Notifications.AddToken].
//
// Parameters:
//   - args[0] - newToken string
//   - args[1] - app string
//
// Returns nothing or an error (throwable)
func (n *Notifications) AddToken(_ js.Value, args []js.Value) any {
	newToken := args[0].String()
	app := args[1].String()

	err := n.api.AddToken(newToken, app)
	if err != nil {
		utils.Throw(utils.TypeError, err)
	}

	return nil
}

// RemoveToken implements [bindings.Notifications.RemoveToken].
//
// Returns nothing or throws an error.
func (n *Notifications) RemoveToken(_ js.Value, args []js.Value) any {
	err := n.api.RemoveToken()
	if err != nil {
		utils.Throw(utils.TypeError, err)
	}
	return nil
}

// SetMaxState implements [bindings.Notifications.SetMaxState]
//
// Parameters:
//   - args[0] - maxState integer
//
// Returns nothing or throws an error
func (n *Notifications) SetMaxState(_ js.Value, args []js.Value) any {
	maxState := int64(args[0].Int())

	err := n.api.SetMaxState(notifications.NotificationState(maxState))
	if err != nil {
		utils.Throw(utils.TypeError, err)
	}

	return nil
}

// GetMaxState implements [bindings.Notifications.GetMaxState]
//
// Returns the current maxState integer
func (n *Notifications) GetMaxState(_ js.Value, args []js.Value) any {
	return int64(n.api.GetMaxState())
}
