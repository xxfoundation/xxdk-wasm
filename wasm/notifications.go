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
	utils "gitlab.com/elixxir/xxdk-wasm/jsutil"
)

type Notifications struct {
	api *bindings.Notifications
}

// newNotificationsJS wrapts the bindings Noticiation object and implements
// wrappers in JS for all it's functionality.
func newNotificationsJS(api *bindings.Notifications) map[string]any {
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
	// ✅ Parse ALL args BEFORE CreatePromise to avoid race conditions
	cMixID := args[0].Int()

	return utils.CreatePromise(func(resolve, reject func(...any) js.Value) {
		api, err := bindings.LoadNotifications(cMixID)
		if err != nil {
			errorConstructor := js.Global().Get("Error")
			errorObject := errorConstructor.New(err.Error())
			reject(errorObject)
			return
		}

		resolve(newNotificationsJS(api))
	})
}

// LoadNotificationsDummy returns a JS wrapped implementation of
// [bindings.Notifications] with a dummy notifications implementation.
//
// Parameters:
//   - args[0] - the cMixID integer
//
// Returns a notifications object or throws an error
func LoadNotificationsDummy(_ js.Value, args []js.Value) any {
	// ✅ Parse ALL args BEFORE CreatePromise to avoid race conditions
	cMixID := args[0].Int()

	return utils.CreatePromise(func(resolve, reject func(...any) js.Value) {
		api, err := bindings.LoadNotificationsDummy(cMixID)
		if err != nil {
			errorConstructor := js.Global().Get("Error")
			errorObject := errorConstructor.New(err.Error())
			reject(errorObject)
			return
		}

		resolve(newNotificationsJS(api))
	})
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
	// ✅ Parse ALL args BEFORE CreatePromise to avoid race conditions
	newToken := args[0].String()
	app := args[1].String()

	return utils.CreatePromise(func(resolve, reject func(...any) js.Value) {
		err := n.api.AddToken(newToken, app)
		if err != nil {
			errorConstructor := js.Global().Get("Error")
			errorObject := errorConstructor.New(err.Error())
			reject(errorObject)
			return
		}

		resolve(js.Undefined())
	})
}

// RemoveToken implements [bindings.Notifications.RemoveToken].
//
// Returns nothing or throws an error.
func (n *Notifications) RemoveToken(_ js.Value, args []js.Value) any {
	return utils.CreatePromise(func(resolve, reject func(...any) js.Value) {
		// No args to parse
		err := n.api.RemoveToken()
		if err != nil {
			errorConstructor := js.Global().Get("Error")
			errorObject := errorConstructor.New(err.Error())
			reject(errorObject)
			return
		}
		resolve(js.Undefined())
	})
}

// SetMaxState implements [bindings.Notifications.SetMaxState]
//
// Parameters:
//   - args[0] - maxState integer
//
// Returns nothing or throws an error
func (n *Notifications) SetMaxState(_ js.Value, args []js.Value) any {
	// ✅ Parse ALL args BEFORE CreatePromise to avoid race conditions
	maxState := int64(args[0].Int())

	return utils.CreatePromise(func(resolve, reject func(...any) js.Value) {
		err := n.api.SetMaxState(maxState)
		if err != nil {
			errorConstructor := js.Global().Get("Error")
			errorObject := errorConstructor.New(err.Error())
			reject(errorObject)
			return
		}

		resolve(js.Undefined())
	})
}

// GetMaxState implements [bindings.Notifications.GetMaxState]
//
// Returns the current maxState integer
func (n *Notifications) GetMaxState(_ js.Value, args []js.Value) any {
	return int64(n.api.GetMaxState())
}
