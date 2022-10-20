////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package utils

import (
	"encoding/base64"
	"os"
	"syscall/js"
)

// LocalStorage contains the js.Value representation of localStorage.
type LocalStorage struct {
	// The Javascript value containing the localStorage object
	v js.Value
}

// jsStorage is the global that stores Javascript as window.localStorage.
//
//  - Specification:
//    https://html.spec.whatwg.org/multipage/webstorage.html#dom-localstorage-dev
//  - Documentation:
//    https://developer.mozilla.org/en-US/docs/Web/API/Window/localStorage
var jsStorage = LocalStorage{js.Global().Get("localStorage")}

// GetLocalStorage returns Javascript's local storage.
func GetLocalStorage() *LocalStorage {
	return &jsStorage
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
	keyValue := ls.getItem(keyName)
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
	ls.setItem(keyName, encodedKeyValue)
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
	ls.removeItem(keyName)
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

	return keyName.String(), nil
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
