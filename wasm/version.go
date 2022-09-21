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
	"syscall/js"
)

// GetVersion returns the [xxdk.SEMVER].
//
// Returns:
//  - Version (string).
func GetVersion(js.Value, []js.Value) interface{} {
	return bindings.GetVersion()
}

// GetGitVersion returns the [xxdk.GITVERSION].
//
// Returns:
//  - Git version (string).
func GetGitVersion(js.Value, []js.Value) interface{} {
	return bindings.GetGitVersion()
}

// GetDependencies returns the [xxdk.DEPENDENCIES].
//
// Returns:
//  - Git dependencies (string).
func GetDependencies(js.Value, []js.Value) interface{} {
	return bindings.GetDependencies()
}
