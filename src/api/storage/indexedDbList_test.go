////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package storage

import (
	"reflect"
	"testing"
)

// Tests that three indexedDb database names stored with StoreIndexedDb are
// retrieved with GetIndexedDbList.
func TestStoreIndexedDb_GetIndexedDbList(t *testing.T) {
	expected := map[string]struct{}{"db1": {}, "db2": {}, "db3": {}}

	for name := range expected {
		err := StoreIndexedDb(name)
		if err != nil {
			t.Errorf("Failed to store database name %q: %+v", name, err)
		}
	}

	list, err := GetIndexedDbList()
	if err != nil {
		t.Errorf("Failed to get database list: %+v", err)
	}

	if !reflect.DeepEqual(expected, list) {
		t.Errorf("Did not get expected list.\nexpected: %s\nreceived: %s",
			expected, list)
	}
}
