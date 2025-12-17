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

// SupportedEmojis returns a list of emojis that are supported by the backend.
// The list includes all emojis described in [UTS #51 section A.1: Data Files].
//
// Returns:
//   - JSON of an array of emoji.Emoji (Uint8Array).
//   - Throws an error if marshalling the JSON fails.
//
// Example JSON:
//
//		[
//		  {
//	     "character": "☹️",
//	     "name": "frowning face",
//	     "comment": "E0.7",
//	     "codePoint": "2639 FE0F",
//	     "group": "Smileys \u0026 Emotion",
//	     "subgroup": "face-concerned"
//		  },
//		  {
//	     "character": "☺️",
//	     "name": "smiling face",
//	     "comment": "E0.6",
//	     "codePoint": "263A FE0F",
//	     "group": "Smileys \u0026 Emotion",
//	     "subgroup": "face-affection"
//		  },
//		  {
//	     "character": "☢️",
//	     "name": "radioactive",
//	     "comment": "E1.0",
//	     "codePoint": "2622 FE0F",
//	     "group": "Symbols",
//	     "subgroup": "warning"
//		  }
//		]
//
// [UTS #51 section A.1: Data Files]: https://www.unicode.org/reports/tr51/#Data_Files
func SupportedEmojis(js.Value, []js.Value) any {
	return utils.CreatePromise(func(resolve, reject func(...any) js.Value) {
		data, err := bindings.SupportedEmojis()
		if err != nil {
			errorConstructor := js.Global().Get("Error")
			errorObject := errorConstructor.New(err.Error())
			reject(errorObject)
			return
		}

		resolve(utils.CopyBytesToJS(data))
	})
}

// SupportedEmojisMap returns a map of emojis that are supported by the backend
// as described by [SupportedEmojis].
//
// Returns:
//   - JSON of a map of emoji.Emoji (Uint8Array).
//   - Throws an error if marshalling the JSON fails.
//
// Example JSON:
//
//		[
//		  {
//	     "character": "☹️",
//	     "name": "frowning face",
//	     "comment": "E0.7",
//	     "codePoint": "2639 FE0F",
//	     "group": "Smileys \u0026 Emotion",
//	     "subgroup": "face-concerned"
//		  },
//		  {
//	     "character": "☺️",
//	     "name": "smiling face",
//	     "comment": "E0.6",
//	     "codePoint": "263A FE0F",
//	     "group": "Smileys \u0026 Emotion",
//	     "subgroup": "face-affection"
//		  },
//		  {
//	     "character": "☢️",
//	     "name": "radioactive",
//	     "comment": "E1.0",
//	     "codePoint": "2622 FE0F",
//	     "group": "Symbols",
//	     "subgroup": "warning"
//		  }
//		]
func SupportedEmojisMap(js.Value, []js.Value) any {
	return utils.CreatePromise(func(resolve, reject func(...any) js.Value) {
		data, err := bindings.SupportedEmojisMap()
		if err != nil {
			errorConstructor := js.Global().Get("Error")
			errorObject := errorConstructor.New(err.Error())
			reject(errorObject)
			return
		}

		resolve(utils.CopyBytesToJS(data))
	})
}

// ValidateReaction checks that the reaction only contains a single grapheme
// (one or more codepoints that appear as a single character to the user).
//
// Parameters:
//   - args[0] - The reaction to validate (string).
//
// Returns:
//   - If the reaction is valid, returns null.
//   - If the reaction is invalid, returns a Javascript Error object containing
//     emoji.InvalidReaction.
func ValidateReaction(_ js.Value, args []js.Value) any {
	err := bindings.ValidateReaction(args[0].String())
	if err != nil {
		return js.Global().Get("Error").New(err.Error())
	}

	return js.Null()
}
