////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package utils

import (
	"encoding/base64"
	"encoding/json"
	"sort"
	"syscall/js"
	"testing"
)

import (
	"bytes"
	"fmt"
	"strings"
)

// Tests that CopyBytesToGo returns a byte slice that matches the Uint8Array.
func TestCopyBytesToGo(t *testing.T) {
	for i, val := range testBytes {
		// Create Uint8Array and set each element individually
		jsBytes := Uint8Array.New(len(val))
		for j, v := range val {
			jsBytes.SetIndex(j, v)
		}

		goBytes := CopyBytesToGo(jsBytes)

		if !bytes.Equal(val, goBytes) {
			t.Errorf("Failed to recevie expected bytes from Uint8Array (%d)."+
				"\nexpected: %d\nreceived: %d",
				i, val, goBytes)
		}
	}
}

// Tests that CopyBytesToJS returns a Javascript Uint8Array with values matching
// the original byte slice.
func TestCopyBytesToJS(t *testing.T) {
	for i, val := range testBytes {
		jsBytes := CopyBytesToJS(val)

		// Generate the expected string to match the output of toString() on a
		// Uint8Array
		expected := strings.ReplaceAll(fmt.Sprintf("%d", val), " ", ",")[1:]
		expected = expected[:len(expected)-1]

		// Get the string value of the Uint8Array
		jsString := jsBytes.Call("toString").String()

		if expected != jsString {
			t.Errorf("Failed to recevie expected string representation of "+
				"the Uint8Array (%d).\nexpected: %s\nreceived: %s",
				i, expected, jsString)
		}
	}
}

// Tests that a byte slice converted to Javascript via CopyBytesToJS and
// converted back to Go via CopyBytesToGo matches the original.
func TestCopyBytesToJSCopyBytesToGo(t *testing.T) {
	for i, val := range testBytes {
		jsBytes := CopyBytesToJS(val)
		goBytes := CopyBytesToGo(jsBytes)

		if !bytes.Equal(val, goBytes) {
			t.Errorf("Failed to recevie expected bytes from Uint8Array (%d)."+
				"\nexpected: %d\nreceived: %d",
				i, val, goBytes)
		}
	}

}

// Tests that JsToJson can convert a Javascript object to JSON that matches the
// output of json.Marshal on the Go version of the same object.
func TestJsToJson(t *testing.T) {
	testObj := map[string]interface{}{
		"nil":    nil,
		"bool":   true,
		"int":    1,
		"float":  1.5,
		"string": "I am string",
		"array":  []interface{}{1, 2, 3},
		"object": map[string]interface{}{"int": 5},
	}

	expected, err := json.Marshal(testObj)
	if err != nil {
		t.Errorf("Failed to JSON marshal test object: %+v", err)
	}

	jsJson := JsToJson(js.ValueOf(testObj))

	// Javascript does not return the JSON object fields sorted so the letters
	// of each Javascript string are sorted and compared
	er := []rune(string(expected))
	sort.SliceStable(er, func(i, j int) bool { return er[i] < er[j] })
	jj := []rune(jsJson)
	sort.SliceStable(jj, func(i, j int) bool { return jj[i] < jj[j] })

	if string(er) != string(jj) {
		t.Errorf("Recieved incorrect JSON from Javascript object."+
			"\nexpected: %s\nreceived: %s", expected, jsJson)
	}
}

// Tests that JsonToJS can convert a JSON object with multiple types to a
// Javascript object and that all values match.
func TestJsonToJS(t *testing.T) {
	testObj := map[string]interface{}{
		"nil":    nil,
		"bool":   true,
		"int":    1,
		"float":  1.5,
		"string": "I am string",
		"bytes":  []byte{1, 2, 3},
		"array":  []interface{}{1, 2, 3},
		"object": map[string]interface{}{"int": 5},
	}
	jsonData, err := json.Marshal(testObj)
	if err != nil {
		t.Errorf("Failed to JSON marshal test object: %+v", err)
	}

	jsObj, err := JsonToJS(jsonData)
	if err != nil {
		t.Errorf("Failed to convert JSON to Javascript object: %+v", err)
	}

	for key, val := range testObj {
		jsVal := jsObj.Get(key)
		switch key {
		case "nil":
			if !jsVal.IsNull() {
				t.Errorf("Key %s is not null.", key)
			}
		case "bool":
			if jsVal.Bool() != val {
				t.Errorf("Incorrect value for key %s."+
					"\nexpected: %t\nreceived: %t", key, val, jsVal.Bool())
			}
		case "int":
			if jsVal.Int() != val {
				t.Errorf("Incorrect value for key %s."+
					"\nexpected: %d\nreceived: %d", key, val, jsVal.Int())
			}
		case "float":
			if jsVal.Float() != val {
				t.Errorf("Incorrect value for key %s."+
					"\nexpected: %f\nreceived: %f", key, val, jsVal.Float())
			}
		case "string":
			if jsVal.String() != val {
				t.Errorf("Incorrect value for key %s."+
					"\nexpected: %s\nreceived: %s", key, val, jsVal.String())
			}
		case "bytes":
			if jsVal.String() != base64.StdEncoding.EncodeToString(val.([]byte)) {
				t.Errorf("Incorrect value for key %s."+
					"\nexpected: %s\nreceived: %s", key,
					base64.StdEncoding.EncodeToString(val.([]byte)),
					jsVal.String())
			}
		case "array":
			for i, v := range val.([]interface{}) {
				if jsVal.Index(i).Int() != v {
					t.Errorf("Incorrect value for key %s index %d."+
						"\nexpected: %d\nreceived: %d",
						key, i, v, jsVal.Index(i).Int())
				}
			}
		case "object":
			if jsVal.Get("int").Int() != val.(map[string]interface{})["int"] {
				t.Errorf("Incorrect value for key %s."+
					"\nexpected: %d\nreceived: %d", key,
					val.(map[string]interface{})["int"], jsVal.Get("int").Int())
			}
		}
	}
}

// Tests that JSON can be converted to a Javascript object via JsonToJS and back
// to JSON using JsToJson and matches the original.
func TestJsonToJSJsToJson(t *testing.T) {
	testObj := map[string]interface{}{
		"nil":    nil,
		"bool":   true,
		"int":    1,
		"float":  1.5,
		"string": "I am string",
		"bytes":  []byte{1, 2, 3},
		"array":  []interface{}{1, 2, 3},
		"object": map[string]interface{}{"int": 5},
	}
	jsonData, err := json.Marshal(testObj)
	if err != nil {
		t.Errorf("Failed to JSON marshal test object: %+v", err)
	}

	jsObj, err := JsonToJS(jsonData)
	if err != nil {
		t.Errorf("Failed to convert the Javascript object to JSON: %+v", err)
	}

	jsJson := JsToJson(jsObj)

	// Javascript does not return the JSON object fields sorted so the letters
	// of each Javascript string are sorted and compared
	er := []rune(string(jsonData))
	sort.SliceStable(er, func(i, j int) bool { return er[i] < er[j] })
	jj := []rune(jsJson)
	sort.SliceStable(jj, func(i, j int) bool { return jj[i] < jj[j] })

	if string(er) != string(jj) {
		t.Errorf("JSON from Javascript does not match original."+
			"\nexpected: %s\nreceived: %s", jsonData, jsJson)
	}
}
