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
	"testing"
)

// mockKVStore implements kv.Store interface for testing
type mockKVStore struct {
	data map[string][]byte
}

func newMockKVStore() *mockKVStore {
	return &mockKVStore{data: make(map[string][]byte)}
}

func (m *mockKVStore) Get(key string) ([]byte, error) {
	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return nil, fs.ErrNotExist
}

func (m *mockKVStore) Set(key string, value []byte) error {
	m.data[key] = value
	return nil
}

func (m *mockKVStore) Delete(key string) error {
	delete(m.data, key)
	return nil
}

func (m *mockKVStore) Keys() ([]byte, error) {
	return nil, nil
}

func (m *mockKVStore) Clear() {
	m.data = make(map[string][]byte)
}

// Tests that checkAndStoreVersionsKV correct initialises the client and WASM
// versions on first run and upgrades them correctly on subsequent runs.
func Test_checkAndStoreVersionsKV(t *testing.T) {
	store := newMockKVStore()
	oldWasmVer := "0.1"
	newWasmVer := "1.0"
	oldClientVer := "2.5"
	newClientVer := "2.6"
	err := checkAndStoreVersionsKV(oldWasmVer, oldClientVer, store)
	if err != nil {
		t.Errorf("checkAndStoreVersionsKV error: %+v", err)
	}

	// Check client version
	storedClientVer, err := store.Get(clientVerKey)
	if err != nil {
		t.Errorf("Failed to get client version from storage: %+v", err)
	}
	if string(storedClientVer) != oldClientVer {
		t.Errorf("Loaded client version does not match expected."+
			"\nexpected: %s\nreceived: %s", oldClientVer, storedClientVer)
	}

	// Check WASM version
	storedWasmVer, err := store.Get(semverKey)
	if err != nil {
		t.Errorf("Failed to get WASM version from storage: %+v", err)
	}
	if string(storedWasmVer) != oldWasmVer {
		t.Errorf("Loaded WASM version does not match expected."+
			"\nexpected: %s\nreceived: %s", oldWasmVer, storedWasmVer)
	}

	err = checkAndStoreVersionsKV(newWasmVer, newClientVer, store)
	if err != nil {
		t.Errorf("checkAndStoreVersionsKV error: %+v", err)
	}

	// Check client version
	storedClientVer, err = store.Get(clientVerKey)
	if err != nil {
		t.Errorf("Failed to get client version from storage: %+v", err)
	}
	if string(storedClientVer) != newClientVer {
		t.Errorf("Loaded client version does not match expected."+
			"\nexpected: %s\nreceived: %s", newClientVer, storedClientVer)
	}

	// Check WASM version
	storedWasmVer, err = store.Get(semverKey)
	if err != nil {
		t.Errorf("Failed to get WASM version from storage: %+v", err)
	}
	if string(storedWasmVer) != newWasmVer {
		t.Errorf("Loaded WASM version does not match expected."+
			"\nexpected: %s\nreceived: %s", newWasmVer, storedWasmVer)
	}
}

// Tests that initOrLoadStoredSemverKV initialises the correct version on first
// run and returns the same version on subsequent runs.
func Test_initOrLoadStoredSemverKV(t *testing.T) {
	store := newMockKVStore()
	key := "testKey"
	oldVersion := "0.1"

	loadedVersion, err := initOrLoadStoredSemverKV(key, oldVersion, store)
	if err != nil {
		t.Errorf("Failed to intilaise version: %+v", err)
	}

	if loadedVersion != oldVersion {
		t.Errorf("Loaded version does not match expected."+
			"\nexpected: %s\nreceived: %s", oldVersion, loadedVersion)
	}

	loadedVersion, err = initOrLoadStoredSemverKV(key, "something", store)
	if err != nil {
		t.Errorf("Failed to load version: %+v", err)
	}

	if loadedVersion != oldVersion {
		t.Errorf("Loaded version does not match expected."+
			"\nexpected: %s\nreceived: %s", oldVersion, loadedVersion)
	}
}
