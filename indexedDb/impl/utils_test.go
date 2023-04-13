////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package impl

import (
	"github.com/hack-pad/go-indexeddb/idb"
	"strings"
	"syscall/js"
	"testing"
)

// Error path: Tests that Get returns an error when trying to get a message that
// does not exist.
func TestGet_NoMessageError(t *testing.T) {
	db := newTestDB("messages", "index", t)

	_, err := Get(db, "messages", js.ValueOf(5))
	if err == nil || !strings.Contains(err.Error(), "undefined") {
		t.Errorf("Did not get expected error when getting a message that "+
			"does not exist: %+v", err)
	}
}

// Error path: Tests that GetIndex returns an error when trying to get a message
// that does not exist.
func TestGetIndex_NoMessageError(t *testing.T) {
	db := newTestDB("messages", "index", t)

	_, err := GetIndex(db, "messages", "index", js.ValueOf(5))
	if err == nil || !strings.Contains(err.Error(), "undefined") {
		t.Errorf("Did not get expected error when getting a message that "+
			"does not exist: %+v", err)
	}
}

// Test simple put on empty DB is successful
func TestPut(t *testing.T) {
	objectStoreName := "messages"
	db := newTestDB(objectStoreName, "index", t)
	testValue := js.ValueOf(make(map[string]interface{}))
	result, err := Put(db, objectStoreName, testValue)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if !result.Equal(js.ValueOf(1)) {
		t.Fatalf("Failed to generate autoincremented key")
	}
}

// newTestDB creates a new idb.Database for testing.
func newTestDB(name, index string, t *testing.T) *idb.Database {
	// Attempt to open database object
	ctx, cancel := NewContext()
	defer cancel()
	openRequest, err := idb.Global().Open(ctx, "databaseName", 0,
		func(db *idb.Database, _ uint, _ uint) error {
			storeOpts := idb.ObjectStoreOptions{
				KeyPath:       js.ValueOf("id"),
				AutoIncrement: true,
			}

			// Build Message ObjectStore and Indexes
			messageStore, err := db.CreateObjectStore(name, storeOpts)
			if err != nil {
				return err
			}

			_, err = messageStore.CreateIndex(
				index, js.ValueOf("id"), idb.IndexOptions{})
			if err != nil {
				return err
			}

			return nil
		})
	if err != nil {
		t.Fatal(err)
	}

	// Wait for database open to finish
	db, err := openRequest.Await(ctx)
	if err != nil {
		t.Fatal(err)
	}

	return db
}
