////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file                                                               //
////////////////////////////////////////////////////////////////////////////////

package main

import (
	_ "embed"
	"encoding/json"
	"github.com/nsf/jsondiff"
	"reflect"
	"testing"
)

//go:embed emojiMart.json
var emojiMartJson []byte

// Tests that marshaling the emojiMartData object and unmarshalling that JSON
// data back into an object does not cause loss in data.
func Test_emojiMartData_JSON_Marshal_Unmarshal(t *testing.T) {
	exampleData := emojiMartSet{
		Categories: []category{
			{ID: "100", Emojis: []emojiID{"100"}},
			{ID: "21"},
			{ID: "20"},
		},
		Emojis: map[emojiID]emoji{
			"100": {
				ID:       "100",
				Name:     "Hundred Points",
				Keywords: []string{"hunna"},
				Skins:    nil,
				Version:  0,
			},
		},
		Aliases: map[string]emojiID{
			"lady_beetle": "ladybug",
		},
		Sheet: map[string]interface{}{
			"test": "data",
		},
	}

	marshaled, err := json.Marshal(&exampleData)
	if err != nil {
		t.Fatalf("Failed to marshal: %+v", err)
	}

	unmarshalData := emojiMartSet{}
	err = json.Unmarshal(marshaled, &unmarshalData)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %+v", err)
	}

	if reflect.DeepEqual(unmarshalData, marshaled) {
		t.Fatalf("Failed to unmarshal example and maintain original data."+
			"\nExpected: %+v"+
			"\nReceived: %+v", exampleData, unmarshalData)
	}
}

// Tests that the emoji-mart example JSON can be JSON unmarshalled and
// marshalled and that the result is semantically identical to the original.
func Test_emojiMartDataJSON_Example(t *testing.T) {
	emojiMart := &emojiMartSet{}
	err := json.Unmarshal(emojiMartJson, emojiMart)
	if err != nil {
		t.Fatalf("Failed to unamrshal: %+v", err)
	}

	marshalled, err := json.Marshal(emojiMart)
	if err != nil {
		t.Fatalf("Failed to marshal: %+v", err)
	}

	opts := jsondiff.DefaultConsoleOptions()
	d, s := jsondiff.Compare(emojiMartJson, marshalled, &opts)
	if d != jsondiff.FullMatch {
		t.Errorf("Diff failed for marshalled JSON: %s\nDifferences: %s", d, s)
	}
}
