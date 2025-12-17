////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package storage

import (
	"io/fs"

	"github.com/pkg/errors"

	"gitlab.com/elixxir/xxdk-wasm/indexedDb/worker/kv"
)

// Key to store if the database is encrypted or not
const databaseEncryptionToggleKey = "xxdkWasmDatabaseEncryptionToggle/"

// StoreIndexedDbEncryptionStatus stores the encryption status if it has not
// been previously saved. If it has, then it returns its value.
func StoreIndexedDbEncryptionStatus(
	databaseName string, encryptionStatus bool) (
	loadedEncryptionStatus bool, err error) {
	store := kv.GetStore()
	if store == nil {
		return false, errors.New("KV store not available")
	}

	keyName := databaseEncryptionToggleKey + databaseName
	data, err := store.Get(keyName)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if err = store.Set(keyName, []byte{1}); err != nil {
				return false,
					errors.Wrapf(err, "kv: failed to set %q", keyName)
			}
			return encryptionStatus, nil
		} else {
			return false, err
		}
	}

	return data[0] == 1, nil
}
