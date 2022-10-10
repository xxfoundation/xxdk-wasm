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
func (s *LocalStorage) GetItem(keyName string) ([]byte, error) {
	keyValue := s.v.Call("getItem", keyName)
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
func (s *LocalStorage) SetItem(keyName string, keyValue []byte) {
	encodedKeyValue := base64.StdEncoding.EncodeToString(keyValue)
	s.v.Call("setItem", keyName, encodedKeyValue)
}

// RemoveItem removes a key's value from local storage given its name. If there
// is no item with the given key, this function does nothing. Underneath, it
// calls localStorage.RemoveItem().
//
//  - Specification:
//    https://html.spec.whatwg.org/multipage/webstorage.html#dom-storage-removeitem-dev
//  - Documentation:
//    https://developer.mozilla.org/en-US/docs/Web/API/Storage/removeItem
func (s *LocalStorage) RemoveItem(keyName string) {
	s.v.Call("removeItem", keyName)
}

// Clear clears all the keys in storage. Underneath, it calls
// localStorage.clear().
//
//  - Specification:
//    https://html.spec.whatwg.org/multipage/webstorage.html#dom-storage-clear-dev
//  - Documentation:
//    https://developer.mozilla.org/en-US/docs/Web/API/Storage/clear
func (s *LocalStorage) Clear() {
	s.v.Call("clear")
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
func (s *LocalStorage) Key(n int) (string, error) {
	keyName := s.v.Call("key", n)
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
func (s *LocalStorage) Length() int {
	return s.v.Get("length").Int()
}
