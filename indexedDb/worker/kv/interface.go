////////////////////////////////////////////////////////////////////////////////
// Copyright © 2024 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

// Package kv provides a unified key-value storage interface that can be used
// by both the main thread and worker threads. All components that need KV
// storage should use this interface, which is backed by the KV Worker for
// persistent IndexedDB storage.
//
// All KV operations go through the KV Worker via MessageChannel, ensuring
// consistent access to IndexedDB from any context.
//
// Setup (from TypeScript, before Go code runs):
//
//	// Create KV Worker and register with Go
//	const kvWorker = new Worker('kvWorker.js');
//	SetKVWorkerManager(kvWorker);
//
// Usage (same API for main thread and worker threads):
//
//	// Get the global store (after initialization)
//	store := kv.GetStore()
//	value, err := store.Get("myKey")
//
// Initialization differs by context:
//
// Main thread: SetKVWorkerManager automatically sets up the store.
//
// Worker threads: Register a callback for the KV port, then call SetStore:
//
//	tm.RegisterMessageChannelCallback(kv.PortTag, func(port js.Value, channelName string) {
//	    kvStore, err := kv.NewWorkerThread(port, channelName)
//	    kv.SetStore(kvStore)  // Register globally
//	})
package kv

import "io/fs"

// StringStore is the raw interface for string-only key-value storage.
// This matches the TypeScript/KV Worker interface directly.
// Values are stored and retrieved as strings - no encoding is performed.
type StringStore interface {
	// Get retrieves a string value by key. Returns fs.ErrNotExist if key doesn't exist.
	Get(key string) (string, error)

	// Set stores a string value for a key.
	Set(key string, value string) error

	// Delete removes a key. Does nothing if key doesn't exist.
	Delete(key string) error

	// Keys returns all key names.
	Keys() ([]string, error)
}

// Store is the interface for byte-based key-value storage operations.
// This is the interface expected by Go callers (e.g., bindings.GenericKeyValue).
// Values are automatically encoded/decoded using base32768 for storage.
type Store interface {
	// Get retrieves a value by key. Returns fs.ErrNotExist if key doesn't exist.
	Get(key string) ([]byte, error)

	// Set stores a value for a key.
	Set(key string, value []byte) error

	// Delete removes a key. Does nothing if key doesn't exist.
	Delete(key string) error

	// Keys returns JSON-encoded array of all key names.
	Keys() ([]byte, error)
}

// ErrNotExist is returned when a key does not exist.
// This is an alias for fs.ErrNotExist for convenience.
var ErrNotExist = fs.ErrNotExist
