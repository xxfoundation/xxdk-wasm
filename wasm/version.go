////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"encoding/json"
	"gitlab.com/elixxir/client/v4/bindings"
	"gitlab.com/elixxir/xxdk-wasm/storage"
	"gitlab.com/elixxir/xxdk-wasm/utils"
	"syscall/js"
)

// GetVersion returns the current xxDK WASM semantic version.
//
// Returns:
//   - Current version (string).
func GetVersion(js.Value, []js.Value) any {
	return storage.SEMVER
}

// GetClientVersion returns the current client xxDK semantic version
// ([xxdk.SEMVER]).
//
// Returns:
//   - Current version (string).
func GetClientVersion(js.Value, []js.Value) any {
	return bindings.GetVersion()
}

// GetClientGitVersion returns the current client xxDK git version
// ([xxdk.GITVERSION]).
//
// Returns:
//   - Git version (string).
func GetClientGitVersion(js.Value, []js.Value) any {
	return bindings.GetGitVersion()
}

// GetClientDependencies returns the client's dependencies
// ([xxdk.DEPENDENCIES]).
//
// Returns:
//   - Dependency list (string).
func GetClientDependencies(js.Value, []js.Value) any {
	return bindings.GetDependencies()
}

// VersionInfo contains information about the current and old version of the
// API.
type VersionInfo struct {
	Current string `json:"current"`
	Updated bool   `json:"updated"`
	Old     string `json:"old"`
}

// GetWasmSemanticVersion returns the current version of the WASM client, it's
// old version before being updated, and if it has been updated.
//
// Returns:
//   - JSON of [VersionInfo] (Uint8Array).
//   - Throws a TypeError if getting the version failed.
func GetWasmSemanticVersion(js.Value, []js.Value) any {
	vi := VersionInfo{
		Current: storage.SEMVER,
		Updated: true,
		Old:     storage.GetOldWasmSemVersion(),
	}

	if vi.Current == vi.Old {
		vi.Updated = false
	}

	data, err := json.Marshal(vi)
	if err != nil {
		utils.Throw(utils.TypeError, err)
	}

	return utils.CopyBytesToJS(data)
}

// GetXXDKSemanticVersion returns the current version of the xxdk client, it's
// old version before being updated, and if it has been updated.
//
// Returns:
//   - JSON of [VersionInfo] (Uint8Array).
//   - Throws a TypeError if getting the version failed.
func GetXXDKSemanticVersion(js.Value, []js.Value) any {
	vi := VersionInfo{
		Current: bindings.GetVersion(),
		Updated: true,
		Old:     storage.GetOldClientSemVersion(),
	}
	if vi.Current == vi.Old {
		vi.Updated = false
	}

	data, err := json.Marshal(vi)
	if err != nil {
		utils.Throw(utils.TypeError, err)
	}

	return utils.CopyBytesToJS(data)
}
