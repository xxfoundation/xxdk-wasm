////////////////////////////////////////////////////////////////////////////////
// Copyright © 2024 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package kv

// Op is a string identifier for KV operations.
type Op string

// KV Worker operations.
// These match the tags in src/workers/kvWorker.ts.
const (
	// InitOp initializes the KV Worker with a database name.
	InitOp Op = "KVInit"

	// GetOp retrieves a value from the KV store.
	GetOp Op = "KVGet"

	// SetOp stores a value in the KV store.
	SetOp Op = "KVSet"

	// DeleteOp removes a key from the KV store.
	DeleteOp Op = "KVDelete"

	// KeysOp returns all keys from the KV store.
	KeysOp Op = "KVKeys"

	// PortOp is the key used when registering a MessageChannel callback
	// for workers to connect to the KV Worker.
	PortOp Op = "KVPort"
)
