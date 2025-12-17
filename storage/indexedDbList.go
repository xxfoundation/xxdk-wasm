////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package storage

import (
	json "github.com/goccy/go-json"
	"io/fs"

	"github.com/pkg/errors"

	"gitlab.com/elixxir/xxdk-wasm/indexedDb/worker/kv"
)

const indexedDbListKey = "xxDkWasmIndexedDbList"

// GetIndexedDbList returns the list of stored indexedDb databases.
func GetIndexedDbList() (map[string]struct{}, error) {
	store := kv.GetStore()
	if store == nil {
		return nil, errors.New("KV store not available")
	}

	list := make(map[string]struct{})
	listBytes, err := store.Get(indexedDbListKey)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
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
	store := kv.GetStore()
	if store == nil {
		return errors.New("KV store not available")
	}

	list, err := GetIndexedDbList()
	if err != nil {
		return err
	}

	list[databaseName] = struct{}{}

	listBytes, err := json.Marshal(list)
	if err != nil {
		return err
	}

	err = store.Set(indexedDbListKey, listBytes)
	if err != nil {
		return errors.Wrapf(err,
			"kv: failed to set %q", indexedDbListKey)
	}

	return nil
}
