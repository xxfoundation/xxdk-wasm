////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package utils

import (
	"testing"
)

// Tests that checkAndStoreVersions correct initialises the client and WASM
// versions on first run and upgrades them correctly on subsequent runs.
func Test_checkAndStoreVersions(t *testing.T) {
	ls := GetLocalStorage()
	ls.Clear()
	oldWasmVer := "0.1"
	newWasmVer := "1.0"
	oldClientVer := "2.5"
	newClientVer := "2.6"
	err := checkAndStoreVersions(oldWasmVer, oldClientVer, ls)
	if err != nil {
		t.Errorf("CheckAndStoreVersions error: %+v", err)
	}

	// Check client version
	storedClientVer, err := ls.GetItem(clientVerKey)
	if err != nil {
		t.Errorf("Failed to get client version from storage: %+v", err)
	}
	if string(storedClientVer) != oldClientVer {
		t.Errorf("Loaded client version does not match expected."+
			"\nexpected: %s\nreceived: %s", oldClientVer, storedClientVer)
	}

	// Check WASM version
	storedWasmVer, err := ls.GetItem(semverKey)
	if err != nil {
		t.Errorf("Failed to get WASM version from storage: %+v", err)
	}
	if string(storedWasmVer) != oldWasmVer {
		t.Errorf("Loaded WASM version does not match expected."+
			"\nexpected: %s\nreceived: %s", oldWasmVer, storedWasmVer)
	}

	err = checkAndStoreVersions(newWasmVer, newClientVer, ls)
	if err != nil {
		t.Errorf("CheckAndStoreVersions error: %+v", err)
	}

	// Check client version
	storedClientVer, err = ls.GetItem(clientVerKey)
	if err != nil {
		t.Errorf("Failed to get client version from storage: %+v", err)
	}
	if string(storedClientVer) != newClientVer {
		t.Errorf("Loaded client version does not match expected."+
			"\nexpected: %s\nreceived: %s", newClientVer, storedClientVer)
	}

	// Check WASM version
	storedWasmVer, err = ls.GetItem(semverKey)
	if err != nil {
		t.Errorf("Failed to get WASM version from storage: %+v", err)
	}
	if string(storedWasmVer) != newWasmVer {
		t.Errorf("Loaded WASM version does not match expected."+
			"\nexpected: %s\nreceived: %s", newWasmVer, storedWasmVer)
	}
}

// Tests that initOrLoadStoredSemver initialises the correct version on first run
// and returns the same version on subsequent runs.
func Test_initOrLoadStoredSemver(t *testing.T) {
	ls := GetLocalStorage()
	key := "testKey"
	oldVersion := "0.1"

	loadedVersion, err := initOrLoadStoredSemver(key, oldVersion, ls)
	if err != nil {
		t.Errorf("Failed to intilaise version: %+v", err)
	}

	if loadedVersion != oldVersion {
		t.Errorf("Loaded version does not match expected."+
			"\nexpected: %s\nreceived: %s", oldVersion, loadedVersion)
	}

	loadedVersion, err = initOrLoadStoredSemver(key, "something", ls)
	if err != nil {
		t.Errorf("Failed to load version: %+v", err)
	}

	if loadedVersion != oldVersion {
		t.Errorf("Loaded version does not match expected."+
			"\nexpected: %s\nreceived: %s", oldVersion, loadedVersion)
	}
}
