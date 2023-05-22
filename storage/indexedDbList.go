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
	"os"

	"github.com/pkg/errors"

	"gitlab.com/elixxir/wasm-utils/storage"
)

const indexedDbListKey = "xxDkWasmIndexedDbList"

// GetIndexedDbList returns the list of stored indexedDb databases.
func GetIndexedDbList() (map[string]struct{}, error) {
	list := make(map[string]struct{})
	listBytes, err := storage.GetLocalStorage().Get(indexedDbListKey)
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

	err = storage.GetLocalStorage().Set(indexedDbListKey, listBytes)
	if err != nil {
		return errors.Wrapf(err,
			"localStorage: failed to set %q", indexedDbListKey)
	}

	return nil
}
