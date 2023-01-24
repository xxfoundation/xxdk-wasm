////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file                                                               //
////////////////////////////////////////////////////////////////////////////////

package main

import (
	"encoding/json"
	"github.com/forPelevin/gomoji"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	cEmoji "gitlab.com/elixxir/client/v4/emoji"
	"strings"
)

// Set contains the set of emoji's that the backend supports. This object will
// be used to remove all unsupported emojis in the given emoji-mart JSON set.
type Set struct {
	// replacementMap contains a list of emoji code-points in the emoji-mart set
	// that must be replaced to adhere to backend recognized code-points.
	replacementMap map[codepoint]skin

	// supportedEmojis contains a list of all Unicode codepoints for the emojis
	// that are supported. This allows for quick lookup when comparing against
	// the emoji-mart of emojis.
	supportedEmojis map[codepoint]struct{}
}

// NewSet initialises a new Emoji Set with all supported backend emojis.
func NewSet() *Set {
	return &Set{
		replacementMap: map[codepoint]skin{
			"2764-fe0f": {
				Unified: "2764", // Has codepoint "2764-fe0f" in front-end
				Native:  "❤",
			},
		},
		supportedEmojis: emojiListToMap(cEmoji.SupportedEmojis()),
	}
}

// SanitizeEmojiMartSet removes all unsupported emojis from the emoji-mart set
// JSON. It also replaces any mismatched codepoints (where the same Emoji has
// two different codepoints).
func (s *Set) SanitizeEmojiMartSet(frontendEmojiSetJson []byte) ([]byte, error) {

	// Unmarshal the emoji-mart set JSON
	var frontEndEmojiSet emojiMartSet
	err := json.Unmarshal(frontendEmojiSetJson, &frontEndEmojiSet)
	if err != nil {
		return nil, errors.Errorf(
			"failed to unmarshal emoji-mart set JSON: %+v", err)
	}

	jww.DEBUG.Printf(
		"Finding incompatible emojis and replacing mismatched codepoints.")

	// Find all incompatible emojis in the front end set
	emojisToRemove := s.findIncompatibleEmojis(&frontEndEmojiSet)

	jww.DEBUG.Printf("Removing incompatible emojis.")

	// Remove all incompatible emojis from the set
	removeIncompatibleEmojis(&frontEndEmojiSet, emojisToRemove)

	jww.INFO.Printf("Removed %d incompatible codepoints.", len(emojisToRemove))

	return json.Marshal(frontEndEmojiSet)
}

// findIncompatibleEmojis returns a list of emojis in the emojiMartSet that are
// not supported by the Set. Also, any emojiMartSet emoji codepoints that are
// incompatible and have replacements (as defined in Set) are replaced.
func (s *Set) findIncompatibleEmojis(set *emojiMartSet) (emojisToRemove []emojiID) {
	// Iterate over all emojis in the emojiMartSet.Emojis list
	for id, Emoji := range set.Emojis {
		var newSkins []skin
		for _, Skin := range Emoji.Skins {
			// Determine if the emoji's codepoint should be replaced or removed
			replacement, replace := s.replace(Skin.Unified)
			if replace {
				jww.TRACE.Printf("Replaced codepoint %q with %q for emoji %q",
					Skin.Unified, replacement, id)
				newSkins = append(newSkins, replacement)
			} else if !s.remove(Skin.Unified) {
				newSkins = append(newSkins, Skin)
			} else {
				jww.TRACE.Printf("Removed codepoint %q from emoji %q",
					Skin.Unified, id)
			}
		}

		if len(newSkins) > 0 {
			// Write to the set the possible edits (if emojis were replaced
			// or removed)
			Emoji.Skins = newSkins
			set.Emojis[id] = Emoji
		} else {
			// If all skins have been removed, then mark the emoji for removal
			emojisToRemove = append(emojisToRemove, id)
			jww.DEBUG.Printf("All skins removed for emoji %q", id)
		}
	}

	return emojisToRemove
}

// removeIncompatibleEmojis removes all the emojis in emojisToRemove from the
// emojiMartSet set.
func removeIncompatibleEmojis(set *emojiMartSet, emojisToRemove []emojiID) {
	jww.DEBUG.Printf(
		"Removing %d emojis: %s", len(emojisToRemove), emojisToRemove)

	// Remove all incompatible emojis from the emojiMartSet.Emojis list
	for _, char := range emojisToRemove {
		jww.TRACE.Printf("Removing %q from emojiMartSet.Emojis", char)
		delete(set.Emojis, char)
	}

	// Remove all incompatible emojis from the emojiMartSet.Categories list
	for _, cat := range set.Categories {
		// Iterate over the emoji list backwards to make removal of elements
		// from the slice easier
		for i := len(cat.Emojis) - 1; i >= 0; i-- {
			for _, char := range emojisToRemove {
				if cat.Emojis[i] == char {
					cat.Emojis = append(cat.Emojis[:i], cat.Emojis[i+1:]...)
					jww.TRACE.Printf(
						"Removing %q from emojiMartSet.Categories", char)
				}
			}
		}
	}

	// Remove all incompatible emojis from the emojiMartSet.Aliases list
	for alias, id := range set.Aliases {
		for _, removedId := range emojisToRemove {
			if id == removedId {
				delete(set.Aliases, alias)
				jww.TRACE.Printf(
					"Removing %q from emojiMartSet.Aliases", alias)
			}
		}
	}
}

// replace returns whether the front end Unicode codepoint must be replaced.
// It will return a boolean on whether this codepoint needs to be replaced
// and what the codepoint must be replaced with.
func (s *Set) replace(code codepoint) (replacement skin, replace bool) {
	replacement, replace = s.replacementMap[code]
	return replacement, replace
}

// remove returns true if the code point should be removed from the parent list.
func (s *Set) remove(code codepoint) bool {
	_, exists := s.supportedEmojis[code]
	return !exists
}

// emojiListToMap constructs a map for simple lookup for gomoji.Emoji's
// Unicode codepoint.
func emojiListToMap(list []gomoji.Emoji) map[codepoint]struct{} {
	emojiMap := make(map[codepoint]struct{}, len(list))
	for _, e := range list {
		emojiMap[backToFrontCodePoint(e.CodePoint)] = struct{}{}
	}
	return emojiMap
}

// backToFrontCodePoint converts Unicode codepoint format found in gomoji.Emoji
// to the one found in the emoji-mart JSON. The specific conversion is making it
// lowercase and replacing " " with "-".
func backToFrontCodePoint(code string) codepoint {
	return codepoint(strings.ToLower(strings.ReplaceAll(code, " ", "-")))
}
