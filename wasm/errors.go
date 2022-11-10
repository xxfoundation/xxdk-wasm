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

// CreateUserFriendlyErrorMessage will convert the passed in error string to an
// error string that is user-friendly if a substring match is found to a
// common error. Common errors is a map that can be updated using
// UpdateCommonErrors. If the error is not common, some simple parsing is done
// on the error message to make it more user-accessible, removing backend
// specific jargon.
//
// Parameters:
//   - args[0] - an error returned from the backend (string).
//
// Returns
//  - A user-friendly error message. This should be devoid of technical speak
//    but still be meaningful for front-end or back-end teams (string).
func CreateUserFriendlyErrorMessage(_ js.Value, args []js.Value) interface{} {
	return bindings.CreateUserFriendlyErrorMessage(args[0].String())
}

// UpdateCommonErrors updates the internal error mapping database. This internal
// database maps errors returned from the backend to user-friendly error
// messages.
//
// Parameters:
//  - args[0] - Contents of a JSON file whose format conforms to the example
//    below (string).
//
// Example Input:
//  {
//    "Failed to Unmarshal Conversation": "Could not retrieve conversation",
//    "Failed to unmarshal SentRequestMap": "Failed to pull up friend requests",
//    "cannot create username when network is not health": "Cannot create username, unable to connect to network",
//  }
//
// Returns:
//  - Throws a TypeError if the JSON cannot be unmarshalled.
func UpdateCommonErrors(_ js.Value, args []js.Value) interface{} {
	err := bindings.UpdateCommonErrors(args[0].String())
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return nil
}
