////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package storage

import (
	"encoding/base64"
	"encoding/json"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/xxdk-wasm/utils"
	"os"
	"strings"
	"syscall/js"
)

// localStorageWasmPrefix is prefixed to every keyName saved to local storage by
// LocalStorage. It allows the identifications and deletion of keys only created
// by this WASM binary while ignoring keys made by other scripts on the same
// page.
const localStorageWasmPrefix = "xxdkWasmStorage/"

// LocalStorage contains the js.Value representation of localStorage.
type LocalStorage struct {
	// The Javascript value containing the localStorage object
	v js.Value

	// The prefix appended to each key name. This is so that all keys created by
	// this structure can be deleted without affecting other keys in local
	// storage.
	prefix string
}

// jsStorage is the global that stores Javascript as window.localStorage.
//
//  - Specification:
//    https://html.spec.whatwg.org/multipage/webstorage.html#dom-localstorage-dev
//  - Documentation:
//    https://developer.mozilla.org/en-US/docs/Web/API/Window/localStorage
var jsStorage = newLocalStorage(localStorageWasmPrefix)

// newLocalStorage creates a new LocalStorage object with the specified prefix.
func newLocalStorage(prefix string) *LocalStorage {
	return &LocalStorage{
		v:      js.Global().Get("localStorage"),
		prefix: prefix,
	}
}

// GetLocalStorage returns Javascript's local storage.
func GetLocalStorage() *LocalStorage {
	return jsStorage
}

// GetItem returns a key's value from the local storage given its name. Returns
// os.ErrNotExist if the key does not exist. Underneath, it calls
// localStorage.GetItem().
//
//  - Specification:
//    https://html.spec.whatwg.org/multipage/webstorage.html#dom-storage-getitem-dev
//  - Documentation:
//    https://developer.mozilla.org/en-US/docs/Web/API/Storage/getItem
func (ls *LocalStorage) GetItem(keyName string) ([]byte, error) {
	keyValue := ls.getItem(ls.prefix + keyName)
	if keyValue.IsNull() {
		return nil, os.ErrNotExist
	}

	decodedKeyValue, err := base64.StdEncoding.DecodeString(keyValue.String())
	if err != nil {
		return nil, err
	}

	return decodedKeyValue, nil
}

// SetItem adds a key's value to local storage given its name. Underneath, it
// calls localStorage.SetItem().
//
//  - Specification:
//    https://html.spec.whatwg.org/multipage/webstorage.html#dom-storage-setitem-dev
//  - Documentation:
//    https://developer.mozilla.org/en-US/docs/Web/API/Storage/setItem
func (ls *LocalStorage) SetItem(keyName string, keyValue []byte) {
	encodedKeyValue := base64.StdEncoding.EncodeToString(keyValue)
	ls.setItem(ls.prefix+keyName, encodedKeyValue)
}

// RemoveItem removes a key's value from local storage given its name. If there
// is no item with the given key, this function does nothing. Underneath, it
// calls localStorage.RemoveItem().
//
//  - Specification:
//    https://html.spec.whatwg.org/multipage/webstorage.html#dom-storage-removeitem-dev
//  - Documentation:
//    https://developer.mozilla.org/en-US/docs/Web/API/Storage/removeItem
func (ls *LocalStorage) RemoveItem(keyName string) {
	ls.removeItem(ls.prefix + keyName)
}

// Clear clears all the keys in storage. Underneath, it calls
// localStorage.clear().
//
//  - Specification:
//    https://html.spec.whatwg.org/multipage/webstorage.html#dom-storage-clear-dev
//  - Documentation:
//    https://developer.mozilla.org/en-US/docs/Web/API/Storage/clear
func (ls *LocalStorage) Clear() {
	ls.clear()
}

// ClearPrefix clears all keys with the given prefix.
func (ls *LocalStorage) ClearPrefix(prefix string) {
	// Get a copy of all key names at once
	keys := ls.keys()

	// Loop through each key
	for i := 0; i < keys.Length(); i++ {
		if v := keys.Index(i); !v.IsNull() {
			keyName := strings.TrimPrefix(v.String(), ls.prefix)
			if strings.HasPrefix(keyName, prefix) {
				ls.removeItem(v.String())
			}
		}
	}
}

// ClearWASM clears all the keys in storage created by WASM.
func (ls *LocalStorage) ClearWASM() {
	// Get a copy of all key names at once
	keys := ls.keys()

	// Loop through each key
	for i := 0; i < keys.Length(); i++ {
		if v := keys.Index(i); !v.IsNull() {
			keyName := v.String()
			if strings.HasPrefix(keyName, ls.prefix) {
				ls.RemoveItem(strings.TrimPrefix(keyName, ls.prefix))
			}
		}
	}
}

// Key returns the name of the nth key in localStorage. Return os.ErrNotExist if
// the key does not exist. The order of keys is not defined. If there is no item
// with the given key, this function does nothing. Underneath, it calls
// localStorage.key().
//
//  - Specification:
//    https://html.spec.whatwg.org/multipage/webstorage.html#dom-storage-key-dev
//  - Documentation:
//    https://developer.mozilla.org/en-US/docs/Web/API/Storage/key
func (ls *LocalStorage) Key(n int) (string, error) {
	keyName := ls.key(n)
	if keyName.IsNull() {
		return "", os.ErrNotExist
	}

	return strings.TrimPrefix(keyName.String(), ls.prefix), nil
}

// Keys returns a list of all key names in local storage.
func (ls *LocalStorage) Keys() []string {
	keyNamesJson := utils.JSON.Call("stringify", ls.keys())

	var keyNames []string
	err := json.Unmarshal([]byte(keyNamesJson.String()), &keyNames)
	if err != nil {
		jww.FATAL.Panicf(
			"Failed to JSON unmarshal localStorage key name list: %+v", err)
	}

	return keyNames
}

// Length returns the number of keys in localStorage. Underneath, it accesses
// the property localStorage.length.
//
//  - Specification:
//    https://html.spec.whatwg.org/multipage/webstorage.html#dom-storage-key-dev
//  - Documentation:
//    https://developer.mozilla.org/en-US/docs/Web/API/Storage/length
func (ls *LocalStorage) Length() int {
	return ls.length().Int()
}

// Wrappers for Javascript Storage methods and properties.
func (ls *LocalStorage) getItem(keyName string) js.Value  { return ls.v.Call("getItem", keyName) }
func (ls *LocalStorage) setItem(keyName, keyValue string) { ls.v.Call("setItem", keyName, keyValue) }
func (ls *LocalStorage) removeItem(keyName string)        { ls.v.Call("removeItem", keyName) }
func (ls *LocalStorage) clear()                           { ls.v.Call("clear") }
func (ls *LocalStorage) key(n int) js.Value               { return ls.v.Call("key", n) }
func (ls *LocalStorage) length() js.Value                 { return ls.v.Get("length") }
func (ls *LocalStorage) keys() js.Value                   { return utils.Object.Call("keys", ls.v) }
