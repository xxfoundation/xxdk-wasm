////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package storage

import (
	"os"
	"sync"

	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"

	"gitlab.com/elixxir/client/v4/bindings"
)

// SEMVER is the current semantic version of xxDK WASM.
const SEMVER = "0.1.14"

// Storage keys.
const (
	semverKey    = "xxdkWasmSemanticVersion"
	clientVerKey = "xxdkClientSemanticVersion"
)

// CheckAndStoreVersions checks that the stored xxDK WASM version matches the
// current version and if not, upgrades it. It also stored the current xxDK
// client to storage.
//
// On first load, only the xxDK WASM and xxDK client versions are stored.
func CheckAndStoreVersions() error {
	return checkAndStoreVersions(
		SEMVER, bindings.GetVersion(), GetLocalStorage())
}

func checkAndStoreVersions(
	currentWasmVer, currentClientVer string, ls *LocalStorage) error {
	// Get the stored client version, if it exists
	storedClientVer, err :=
		initOrLoadStoredSemver(clientVerKey, currentClientVer, ls)
	if err != nil {
		return err
	}

	// Get the stored WASM versions, if it exists
	storedWasmVer, err := initOrLoadStoredSemver(semverKey, currentWasmVer, ls)
	if err != nil {
		return err
	}

	// Store old versions to memory
	setOldClientSemVersion(storedClientVer)
	setOldWasmSemVersion(storedWasmVer)

	// Check if client needs an update
	if storedClientVer != currentClientVer {
		jww.INFO.Printf("xxDK client out of date; upgrading version: v%s → v%s",
			storedClientVer, currentClientVer)
	} else {
		jww.INFO.Printf("xxDK client version is current: v%s", storedClientVer)
	}

	// Check if WASM needs an update
	if storedWasmVer != currentWasmVer {
		jww.INFO.Printf("xxDK WASM out of date; upgrading version: v%s → v%s",
			storedWasmVer, currentWasmVer)
	} else {
		jww.INFO.Printf("xxDK WASM version is current: v%s", storedWasmVer)
	}

	// Upgrade path code goes here

	// Save current versions
	ls.SetItem(clientVerKey, []byte(currentClientVer))
	ls.SetItem(semverKey, []byte(currentWasmVer))

	return nil
}

// initOrLoadStoredSemver returns the semantic version stored at the key in
// local storage. If no version is stored, then the current version is stored
// and returned.
func initOrLoadStoredSemver(
	key, currentVersion string, ls *LocalStorage) (string, error) {
	storedVersion, err := ls.GetItem(key)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Save the current version if this is the first run
			jww.INFO.Printf("Initialising %s to v%s", key, currentVersion)
			ls.SetItem(key, []byte(currentVersion))
			return currentVersion, nil
		} else {
			// If the item exists, but cannot be loaded, return an error
			return "", errors.Errorf(
				"could not load %s from storage: %+v", key, err)
		}
	}

	// Return the stored version
	return string(storedVersion), nil
}

// oldVersions contains the old versions of xxdk WASM and xxdk client that were
// stored in storage before being overwritten on update.
var oldVersions struct {
	wasm   string
	client string
	sync.Mutex
}

// GetOldWasmSemVersion returns the old version of xxdk WASM before being
// updated.
func GetOldWasmSemVersion() string {
	oldVersions.Lock()
	defer oldVersions.Unlock()
	return oldVersions.wasm
}

// GetOldClientSemVersion returns the old version of xxdk client before being
// updated.
func GetOldClientSemVersion() string {
	oldVersions.Lock()
	defer oldVersions.Unlock()
	return oldVersions.client
}

// setOldWasmSemVersion sets the old version of xxdk WASM. This should be called
// before it is updated.
func setOldWasmSemVersion(v string) {
	oldVersions.Lock()
	defer oldVersions.Unlock()
	oldVersions.wasm = v
}

// setOldClientSemVersion sets the old version of xxdk client. This should be
// called before it is updated.
func setOldClientSemVersion(v string) {
	oldVersions.Lock()
	defer oldVersions.Unlock()
	oldVersions.client = v
}
