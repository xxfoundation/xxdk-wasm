////////////////////////////////////////////////////////////////////////////////
// Copyright © 2024 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package kv

// KVMessage is a simplified message format for KV operations.
// Uses direct string fields instead of base64-encoded data to avoid
// transport-level encoding overhead.
type KVMessage struct {
	Op       string `json:"op"`
	ID       uint64 `json:"id"`
	Response bool   `json:"response"`
	// Request fields
	Key    string `json:"key,omitempty"`
	Value  string `json:"value"`            // base32768-encoded bytes for storage (no omitempty - empty values are valid)
	DbName string `json:"dbName,omitempty"` // For init operation
	// Response fields
	Error string `json:"error,omitempty"`
}
