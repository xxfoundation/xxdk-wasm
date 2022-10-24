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
)

// Key to store if the database is encrypted or not
const databaseEncryptionToggleKey = "xxdkWasmDatabaseEncryptionToggle/"

// StoreIndexedDbEncryptionStatus stores the encryption status if it has not
// been previously saved. If it has, it returns its value.
func StoreIndexedDbEncryptionStatus(
	databaseName string, encryption bool) (bool, error) {
	data, err := GetLocalStorage().GetItem(
		databaseEncryptionToggleKey + databaseName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			GetLocalStorage().SetItem(
				databaseEncryptionToggleKey+databaseName, []byte{1})
			return encryption, nil
		} else {
			return false, err
		}
	}

	return data[0] == 1, nil
}
