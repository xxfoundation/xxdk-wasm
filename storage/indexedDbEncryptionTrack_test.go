////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package storage

import (
	"testing"

	"gitlab.com/elixxir/xxdk-wasm/indexedDb/worker/kv"
)

// Tests that StoreIndexedDbEncryptionStatus stores the initial encryption value
// and return that value on subsequent checks.
func TestStoreIndexedDbEncryptionStatus(t *testing.T) {
	store := kv.GetStore()
	if store == nil {
		t.Skip("KV store not available")
	}

	databaseName := "databaseA"

	encryptionStatus, err := StoreIndexedDbEncryptionStatus(databaseName, true)
	if err != nil {
		t.Errorf("Failed to store/get encryption status: %+v", err)
	}

	if encryptionStatus != true {
		t.Errorf("Incorrect encryption values.\nexpected: %t\nreceived: %t",
			true, encryptionStatus)
	}

	encryptionStatus, err = StoreIndexedDbEncryptionStatus(databaseName, false)
	if err != nil {
		t.Errorf("Failed to store/get encryption status: %+v", err)
	}

	if encryptionStatus != true {
		t.Errorf("Incorrect encryption values.\nexpected: %t\nreceived: %t",
			true, encryptionStatus)
	}
}
