////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package main

// emojiID represents the alias for an emoji in emoji-mart. For example, the
// alias for the emoji 💯 would be "100". This adheres strictly to how Emoji
// Mart categorizes their emojis within the categories section of their JSON
// file.
type emojiID string

// codepoint represents the Unicode codepoint or codepoints for an emoji. They
// are in lowercase and if there are multiple codepoints, they are seperated by
// a dash ("-"). For example, the emoji 💯 would have the codepoint "1f4af".
type codepoint string

// emojiMartSet is a representation of the JSON file format containing the emoji
// list in emoji-mart.
//
// Doc: https://github.com/missive/emoji-mart/
// JSON example: https://github.com/missive/emoji-mart/blob/main/packages/emoji-mart-data/sets/14/native.json
type emojiMartSet struct {
	Categories []category             `json:"categories"`
	Emojis     map[emojiID]emoji      `json:"emojis"`
	Aliases    map[string]emojiID     `json:"aliases"`
	Sheet      map[string]interface{} `json:"sheet"`
}

// category adheres to the category field of the emoji-mart JSON file.
type category struct {
	Emojis []emojiID `json:"emojis"`
	ID     string    `json:"id"`
}

// emoji adheres to the emoji field of the emoji-mart JSON file.
type emoji struct {
	Emoticons []string `json:"emoticons,omitempty"`
	ID        emojiID  `json:"id"`
	Name      string   `json:"name"`
	Keywords  []string `json:"keywords"`
	Skins     []skin   `json:"skins"`
	Version   float32  `json:"version"`
}

// skin adheres to the skin field within the emoji field of the emoji-mart JSON
// file.
type skin struct {
	Unified codepoint `json:"unified"`
	Native  string    `json:"native"`
}
