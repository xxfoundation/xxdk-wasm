////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package storage

import (
	"github.com/pkg/errors"
	"os"

	"gitlab.com/elixxir/wasm-utils/storage"
)

// Key to store if the database is encrypted or not
const databaseEncryptionToggleKey = "xxdkWasmDatabaseEncryptionToggle/"

// StoreIndexedDbEncryptionStatus stores the encryption status if it has not
// been previously saved. If it has, then it returns its value.
func StoreIndexedDbEncryptionStatus(
	databaseName string, encryptionStatus bool) (
	loadedEncryptionStatus bool, err error) {
	ls := storage.GetLocalStorage()
	data, err := ls.Get(databaseEncryptionToggleKey + databaseName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			keyName := databaseEncryptionToggleKey + databaseName
			if err = ls.Set(keyName, []byte{1}); err != nil {
				return false,
					errors.Wrapf(err, "localStorage: failed to set %q", keyName)
			}
			return encryptionStatus, nil
		} else {
			return false, err
		}
	}

	return data[0] == 1, nil
}
