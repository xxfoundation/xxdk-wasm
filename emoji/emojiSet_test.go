////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file                                                               //
////////////////////////////////////////////////////////////////////////////////

package main

import (
	"encoding/json"
	"reflect"
	"testing"
)

// Benchmarks the entire code path for sanitizing front end's emoji file.
// This includes loading the file, parsing the data and sanitizing.
// NOTE: This does not include writing the file due to limitations of the
// embed.FS interface.
func BenchmarkSet_SanitizeFrontendEmojiList(b *testing.B) {
	backendSet := NewSet()

	for i := 0; i < b.N; i++ {

		// Sanitize front end example
		_, err := backendSet.SanitizeEmojiMartSet(emojiMartJson)
		if err != nil {
			b.Fatalf("Failed to Sanitize front end emojis: %+v", err)
		}
	}
}

// Comprehensive test of Set.SanitizeEmojiMartSet using the front end example
// JSON file (emojiMart.json).
func TestSet_SanitizeFrontEndEmojis_FrontEndExample(t *testing.T) {

	backendSet := NewSet()

	// Sanitize front end example
	sanitizedSetJson, err := backendSet.SanitizeEmojiMartSet(emojiMartJson)
	if err != nil {
		t.Fatalf("Failed to Sanitize front end emojis: %+v", err)
	}

	// Unmarshal front end example
	unsanitizedSet := &emojiMartSet{}
	err = json.Unmarshal(emojiMartJson, unsanitizedSet)
	if err != nil {
		t.Fatalf("Failed to unmarshal unsanitized set: %+v", err)
	}

	// Unmarshal sanitization of front end example
	sanitizedSet := &emojiMartSet{}
	err = json.Unmarshal(sanitizedSetJson, sanitizedSet)
	if err != nil {
		t.Fatalf("Failed to unmarshal sanitized set: +%v", err)
	}

	// The front end example has known unsanitary emojis.
	// Therefore, sanitization of the front end example
	// should not contain the exact same data.
	if reflect.DeepEqual(sanitizedSet, unsanitizedSet) {
		t.Fatalf("No evidence of sanitization performed")
	}

	// Check of replacement of the heart emoji (❤️).
	// This is the known unsanitary emoji (it is not supported
	// by backend's library).
	heart, exists := sanitizedSet.Emojis["heart"]
	if !exists {
		t.Fatalf("Heart emoji was removed when it should have been replaced")
	}

	expectedHeartReplacement := skin{
		Unified: "2764",
		Native:  "❤",
	}

	if !reflect.DeepEqual(expectedHeartReplacement, heart.Skins[0]) {
		t.Fatalf("Heart emoji was not replaced as expected."+
			"\nExpected: %+v"+
			"\nReceived: %+v", expectedHeartReplacement, heart.Skins[0])
	}
}

// Test of Set.findIncompatibleEmojis using the front end example
// JSON file (emojiMart.json).
func TestSet_findIncompatibleEmojis_FrontEndExample(t *testing.T) {
	backendSet := NewSet()

	// Unmarshal front end example
	unsanitizedSet := &emojiMartSet{}
	err := json.Unmarshal(emojiMartJson, unsanitizedSet)
	if err != nil {
		t.Fatalf("Failed to unmarshal unsanitized set: %+v", err)
	}

	emojisToRemove := backendSet.findIncompatibleEmojis(unsanitizedSet)
	if len(emojisToRemove) != 0 {
		t.Fatalf("Front end example should not contain any removable emojis.")
	}
}

// Tests Set.findIncompatibleEmojis using a custom emojiMartSet object.
// There is one ID  with skins that should be marked as removable as all skins
// are invalid when compared against backend. There is another ID that should
// not be removed, and should not be returned by set.findIncompatibleEmojis.
func TestSet_findIncompatibleEmojis_RemovableExample(t *testing.T) {
	backendSet := NewSet()

	// Construct custom emojiMartSet.
	toBeRemovedId := emojiID("shouldBeRemoved")
	emd := constructRemovableEmojiSet(toBeRemovedId)

	// Construct a removable list
	removedEmojis := backendSet.findIncompatibleEmojis(emd)

	// Ensure only oen emoji is removed
	if len(removedEmojis) != 1 {
		t.Fatalf("findIncompatibleEmojis should only mark one emoji for removal.")
	}

	// Ensure the single removed emoji is the expected emoji
	if removedEmojis[0] != toBeRemovedId {
		t.Fatalf("findIncompatibleEmojis should have found %s to be removable",
			toBeRemovedId)
	}
}

// Tests that Set.removeIncompatibleEmojis will modify the passed in
// emojiMartSet according to the list of emojis to remove.
func TestSet_removeIncompatibleEmojis(t *testing.T) {
	backendSet := NewSet()

	// Construct custom emojiMartSet.
	toBeRemovedId := emojiID("shouldBeRemoved")
	emd := constructRemovableEmojiSet(toBeRemovedId)

	// Construct a removable list
	removedEmojis := backendSet.findIncompatibleEmojis(emd)

	removeIncompatibleEmojis(emd, removedEmojis)

	// Test that categories has been modified
	for _, cat := range emd.Categories {
		for _, e := range cat.Emojis {
			if e == toBeRemovedId {
				t.Fatalf("EmojiID %q was never removed from "+
					"emojiMartSet.Categories", toBeRemovedId)
			}
		}
	}

	// Test that the emoji map has been modified
	if _, exists := emd.Emojis[toBeRemovedId]; exists {
		t.Fatalf("EmojiId %s was not removed from emojiMartSet.Emojis",
			toBeRemovedId)
	}

	// Test that the alias list has been modified
	for _, id := range emd.Aliases {
		if id == toBeRemovedId {
			t.Fatalf("EmojiId %s twas not removed from emojiMartSet.Aliases",
				toBeRemovedId)
		}
	}
}

// Tests backToFrontCodePoint converts backend Unicode codepoints to their
// front end equivalent.
func Test_backToFrontCodePoint(t *testing.T) {
	// Input for backend codepoints and their front end formatted pairings
	tests := []struct {
		input  string
		output codepoint
	}{
		{"0023 FE0F 20E3", "0023-fe0f-20e3"}, // Single-rune emoji (\u1F44B)
		{"002A FE0F 20E3", "002a-fe0f-20e3"}, // Duel-rune emoji with race modification (\u1F44B\u1F3FF)
		{"00A9 FE0F", "00a9-fe0f"},
		{"1F481 1F3FC 200D 2640 FE0F", "1f481-1f3fc-200d-2640-fe0f"},
		{"1F481 1F3FE 200D 2642 FE0F", "1f481-1f3fe-200d-2642-fe0f"},
		{"1F481 1F3FF", "1f481-1f3ff"},
		{"1F481 1F3FF 200D 2642 FE0F", "1f481-1f3ff-200d-2642-fe0f"},
		{"1F469 1F3FB 200D 1F9B0", "1f469-1f3fb-200d-1f9b0"},
		{"1F378", "1f378"},
		{"1F377", "1f377"},
		{"1F376", "1f376"},
	}

	// Test that for all examples, all backend codepoints are converted
	// to front end codepoints
	for _, test := range tests {
		received := backToFrontCodePoint(test.input)
		if received != test.output {
			t.Fatalf("Incorrect codepoint for %q.\nexpected: %s\nreceived: %s",
				test.input, test.output, received)
		}
	}

}

// constructRemovableEmojiSet returns an emojiMartSet object used for testing.
// This object will contain data that should be marked as removable by
// Set.findIncompatibleEmojis. This removable data from
// Set.findIncompatibleEmojis can be passed to removeIncompatibleEmojis.
func constructRemovableEmojiSet(toBeRemovedId emojiID) *emojiMartSet {
	return &emojiMartSet{
		Categories: []category{
			{Emojis: []emojiID{
				toBeRemovedId,
				"1",
				"two",
				"tree",
				"for",
				"golden rings",
			}},
		},
		Emojis: map[emojiID]emoji{
			toBeRemovedId: {
				Skins: []skin{{
					Unified: "00A9 FE0F 20F3",
					Native:  "",
				}, {
					Unified: "AAAA FE0F 20F3",
					Native:  "",
				}},
			},
			"shouldNotBeRemoved": {
				Skins: []skin{{
					Unified: "1f9e1",
					Native:  "🧡",
				}},
			},
		},
		Aliases: map[string]emojiID{
			"test":   toBeRemovedId,
			"tester": "safe",
			"alias":  "will_not",
			"secret": "beRemoved",
		},
	}
}
