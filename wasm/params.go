////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"gitlab.com/elixxir/client/v5/bindings"
	"gitlab.com/elixxir/xxdk-wasm/utils"
	"syscall/js"
)

// GetDefaultCMixParams returns a JSON serialized object with all of the cMix
// parameters and their default values. Call this function and modify the JSON
// to change cMix settings.
//
// Returns:
//  - JSON of [xxdk.CMIXParams] (Uint8Array).
func GetDefaultCMixParams(js.Value, []js.Value) interface{} {
	return utils.CopyBytesToJS(bindings.GetDefaultCMixParams())
}

// GetDefaultE2EParams returns a JSON serialized object with all of the E2E
// parameters and their default values. Call this function and modify the JSON
// to change E2E settings.
//
// Returns:
//  - JSON of [xxdk.E2EParams] (Uint8Array).
func GetDefaultE2EParams(js.Value, []js.Value) interface{} {
	return utils.CopyBytesToJS(bindings.GetDefaultE2EParams())
}

// GetDefaultFileTransferParams returns a JSON serialized object with all the
// file transfer parameters and their default values. Call this function and
// modify the JSON to change file transfer settings.
//
// Returns:
//  - JSON of [fileTransfer.Params] (Uint8Array).
func GetDefaultFileTransferParams(js.Value, []js.Value) interface{} {
	return utils.CopyBytesToJS(bindings.GetDefaultFileTransferParams())
}

// GetDefaultSingleUseParams returns a JSON serialized object with all the
// single-use parameters and their default values. Call this function and modify
// the JSON to change single use settings.
//
// Returns:
//  - JSON of [single.RequestParams] (Uint8Array).
func GetDefaultSingleUseParams(js.Value, []js.Value) interface{} {
	return utils.CopyBytesToJS(bindings.GetDefaultSingleUseParams())
}

// GetDefaultE2eFileTransferParams returns a JSON serialized object with all the
// E2E file transfer parameters and their default values. Call this function and
// modify the JSON to change single use settings.
//
// Returns:
//  - JSON of [gitlab.com/elixxir/client/v5/fileTransfer/e2e.Params] (Uint8Array).
func GetDefaultE2eFileTransferParams(js.Value, []js.Value) interface{} {
	return utils.CopyBytesToJS(bindings.GetDefaultE2eFileTransferParams())
}
