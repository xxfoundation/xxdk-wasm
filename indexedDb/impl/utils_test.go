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
	jww "github.com/spf13/jwalterweatherman"
	"strings"
	"syscall/js"
	"testing"
	"time"
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

// TestBenchmark ensures IndexedDb can take at least n operations per second.
func TestBenchmark(t *testing.T) {
	jww.SetStdoutThreshold(jww.LevelInfo)
	benchmarkDb(50, t)
}

// benchmarkDb sends n operations to IndexedDb and prints errors.
func benchmarkDb(n int, t *testing.T) {
	jww.INFO.Printf("Benchmarking IndexedDb: %d total.", n)

	objectStoreName := "test"
	testValue := js.ValueOf(make(map[string]interface{}))
	db := newTestDB(objectStoreName, "index", t)

	type metric struct {
		didSucceed bool
		duration   time.Duration
	}
	done := make(chan metric)

	// Spawn n operations at the same time
	startTime := time.Now()
	for i := 0; i < n; i++ {
		go func() {
			opStart := time.Now()
			_, err := Put(db, objectStoreName, testValue)
			done <- metric{
				didSucceed: err == nil,
				duration:   time.Since(opStart),
			}
		}()
	}

	// Wait for all to complete
	didSucceed := true
	for i := 0; i < n; i++ {
		result := <-done
		if !result.didSucceed {
			didSucceed = false
		}
		jww.DEBUG.Printf("Operation time: %s", result.duration)
	}

	timeElapsed := time.Since(startTime)
	jww.INFO.Printf("Benchmarking complete. Succeeded: %t\n"+
		"Took %s, Average of %s.",
		didSucceed, timeElapsed, timeElapsed/time.Duration(n))
}
