////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"gitlab.com/elixxir/client/bindings"
	"gitlab.com/elixxir/xxdk-wasm/storage"
	"syscall/js"
)

// GetVersion returns the current xxDK WASM semantic version.
//
// Returns:
//  - Current version (string).
func GetVersion(js.Value, []js.Value) interface{} {
	return storage.SEMVER
}

// GetClientVersion returns the current client xxDK semantic version
// ([xxdk.SEMVER]).
//
// Returns:
//  - Current version (string).
func GetClientVersion(js.Value, []js.Value) interface{} {
	return bindings.GetVersion()
}

// GetClientGitVersion returns the current client xxDK git version
// ([xxdk.GITVERSION]).
//
// Returns:
//  - Git version (string).
func GetClientGitVersion(js.Value, []js.Value) interface{} {
	return bindings.GetGitVersion()
}

// GetClientDependencies returns the client's dependencies
// ([xxdk.DEPENDENCIES]).
//
// Returns:
//  - Dependency list (string).
func GetClientDependencies(js.Value, []js.Value) interface{} {
	return bindings.GetDependencies()
}
