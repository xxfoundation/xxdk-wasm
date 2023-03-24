////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"encoding/json"
	"gitlab.com/elixxir/client/v4/bindings"
	"gitlab.com/elixxir/client/v4/emoji"
	"gitlab.com/elixxir/xxdk-wasm/src/api/utils"
	"syscall/js"
)

// SupportedEmojis returns a list of emojis that are supported by the backend.
//
// Returns:
//   - JSON of an array of gomoji.Emoji (Uint8Array).
//   - Throws a TypeError if marshalling the JSON fails.
//
// Example JSON:
//
//	[
//	  {
//	    "slug": "smiling-face",
//	    "character": "☺️",
//	    "unicode_name": "E0.6 smiling face",
//	    "code_point": "263A FE0F",
//	    "group": "Smileys \u0026 Emotion",
//	    "sub_group": "face-affection"
//	  },
//	  {
//	    "slug": "frowning-face",
//	    "character": "☹️",
//	    "unicode_name": "E0.7 frowning face",
//	    "code_point": "2639 FE0F",
//	    "group": "Smileys \u0026 Emotion",
//	    "sub_group": "face-concerned"
//	  },
//	  {
//	    "slug": "banana",
//	    "character": "�",
//	    "unicode_name": "E0.6 banana",
//	    "code_point": "1F34C",
//	    "group": "Food \u0026 Drink",
//	    "sub_group": "food-fruit"
//	  }
//	]
func SupportedEmojis(js.Value, []js.Value) any {
	data, err := json.Marshal(emoji.SupportedEmojis())
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return utils.CopyBytesToJS(data)
}

// ValidateReaction checks that the reaction only contains a single emoji.
//
// Parameters:
//   - args[0] - The reaction emoji to validate (string).
//
// Returns:
//   - If the reaction is valid, returns null.
//   - If the reaction is invalid, returns a Javascript Error object containing
//     emoji.InvalidReaction.
func ValidateReaction(_ js.Value, args []js.Value) any {
	err := bindings.ValidateReaction(args[0].String())
	if err != nil {
		return utils.JsError(err)
	}

	return nil
}
