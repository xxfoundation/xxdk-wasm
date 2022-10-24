////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package storage

import (
	"encoding/json"
	"github.com/pkg/errors"
	"os"
)

const indexedDbListKey = "xxDkWasmIndexedDbList"

// GetIndexedDbList returns the list of stored indexedDb databases.
func GetIndexedDbList() (map[string]struct{}, error) {
	list := make(map[string]struct{})
	listBytes, err := GetLocalStorage().GetItem(indexedDbListKey)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	} else if err == nil {
		err = json.Unmarshal(listBytes, &list)
		if err != nil {
			return nil, err
		}
	}

	return list, nil
}

// StoreIndexedDb saved the indexedDb database name to storage.
func StoreIndexedDb(databaseName string) error {
	list, err := GetIndexedDbList()
	if err != nil {
		return err
	}

	list[databaseName] = struct{}{}

	listBytes, err := json.Marshal(list)
	if err != nil {
		return err
	}

	GetLocalStorage().SetItem(indexedDbListKey, listBytes)

	return nil
}
