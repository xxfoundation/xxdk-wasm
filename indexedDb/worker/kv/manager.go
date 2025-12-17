////////////////////////////////////////////////////////////////////////////////
// Copyright © 2024 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package kv

import (
	"encoding/json"
	"sync"
	"syscall/js"

	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
)

// Global KV store for the main thread.
var (
	globalClient   *KVWorkerClient
	globalStore    Store
	globalWorkerJS js.Value // Raw Worker JS object for sending MessageChannel ports
	globalMu       sync.Mutex
)

// SetKVWorkerManager is called from TypeScript after the KV Worker is created.
// It creates a MessageChannel to communicate with the KV Worker and sets up
// the global Store that Go code can use for KV operations.
//
// Parameters:
//   - args[0] - JavaScript Worker object for the KV Worker.
//
// Returns:
//   - Error string on failure, empty string on success.
// kvWorkerGlobalKey is the key used to store the KV Worker on globalThis.
// This must match the key used in TypeScript (kv.ts).
const kvWorkerGlobalKey = "__xxdkKVWorker"

// SetKVWorkerManager is called from TypeScript after the KV Worker is created.
// It can also be called with no arguments to check globalThis for the worker.
func SetKVWorkerManager(_ js.Value, args []js.Value) any {
	jww.DEBUG.Print("[KV] SetKVWorkerManager called")
	globalMu.Lock()
	defer globalMu.Unlock()

	// Already initialized?
	if globalStore != nil {
		jww.DEBUG.Print("[KV] SetKVWorkerManager: already initialized")
		return ""
	}

	var workerJS js.Value

	// Check if worker was passed as argument
	if len(args) >= 1 && !args[0].IsUndefined() && !args[0].IsNull() {
		workerJS = args[0]
		jww.DEBUG.Printf("[KV] SetKVWorkerManager: using worker from args, type=%v", workerJS.Type())
	} else {
		// Try to get worker from globalThis
		workerJS = js.Global().Get(kvWorkerGlobalKey)
		if workerJS.IsUndefined() || workerJS.IsNull() {
			jww.DEBUG.Printf("[KV] SetKVWorkerManager: no worker found on globalThis.%s", kvWorkerGlobalKey)
			return "SetKVWorkerManager: worker not available yet"
		}
		jww.DEBUG.Printf("[KV] SetKVWorkerManager: found worker on globalThis.%s, type=%v", kvWorkerGlobalKey, workerJS.Type())
	}

	// Create the KVWorkerClient which handles MessageChannel setup
	jww.DEBUG.Print("[KV] Creating KVWorkerClient...")
	client, err := NewKVWorkerClient(workerJS, "main-kv")
	if err != nil {
		errMsg := errors.Wrap(err, "failed to create KV worker client").Error()
		jww.ERROR.Print(errMsg)
		return errMsg
	}
	jww.DEBUG.Print("[KV] KVWorkerClient created successfully")

	globalClient = client

	// Wrap with base32768 encoding for bytes support
	globalStore = NewEncodedStore(client)
	jww.DEBUG.Print("[KV] EncodedStore wrapper created")

	// Store raw Worker JS object for SendPortToKVWorker
	globalWorkerJS = workerJS

	jww.INFO.Print("[KV] Global KV store initialized via SetKVWorkerManager")
	return ""
}

// InitKVFromGlobal checks globalThis for the KV Worker and initializes if found.
// This is called during WASM startup to pick up a worker created by TypeScript.
func InitKVFromGlobal() {
	jww.DEBUG.Printf("[KV] InitKVFromGlobal: checking globalThis.%s", kvWorkerGlobalKey)
	result := SetKVWorkerManager(js.Undefined(), nil)
	if result != "" {
		jww.DEBUG.Printf("[KV] InitKVFromGlobal: %s", result)
	}
}

// GetStore returns the global Store for KV operations.
// Returns nil if SetKVWorkerManager hasn't been called yet.
func GetStore() Store {
	globalMu.Lock()
	defer globalMu.Unlock()
	return globalStore
}

// GetKVWorker returns the raw Worker JS object for the KV Worker.
// This can be used to send MessageChannel ports directly.
// Returns js.Undefined() if not initialized.
func GetKVWorker() js.Value {
	globalMu.Lock()
	defer globalMu.Unlock()
	return globalWorkerJS
}

// HasKVWorker returns true if the KV Worker is available.
func HasKVWorker() bool {
	globalMu.Lock()
	defer globalMu.Unlock()
	return !globalWorkerJS.IsUndefined() && !globalWorkerJS.IsNull()
}

// SendPortToKVWorker sends a MessagePort to the KV Worker for a given channel.
// This allows other workers to establish direct communication with the KV Worker.
// The port message includes channel name for the KV Worker to register.
func SendPortToKVWorker(port js.Value, channelName string) error {
	globalMu.Lock()
	workerJS := globalWorkerJS
	globalMu.Unlock()

	if workerJS.IsUndefined() || workerJS.IsNull() {
		return errors.New("KV Worker not initialized")
	}

	// Create registration message with the port
	// The KV Worker expects: { channel: "channelName" } with port in transfer list
	regMsg := map[string]any{
		"channel": channelName,
	}
	regJSON, err := json.Marshal(regMsg)
	if err != nil {
		return errors.Wrap(err, "failed to marshal registration message")
	}

	// Send port to KV Worker with transfer
	transferList := js.Global().Get("Array").New(port)
	workerJS.Call("postMessage", string(regJSON), transferList)

	jww.DEBUG.Printf("[KV] Sent port to KV Worker for channel: %s", channelName)
	return nil
}

// CreateKVChannelForWorker creates a MessageChannel between a Go worker and
// the KV Worker. It sends one port to the KV Worker and sends the other
// port to the Go worker.
//
// Parameters:
//   - goWorkerJS: The raw js.Value for the Go worker (from Manager.GetWorker())
//   - channelName: Name for this channel (for logging and registration)
//
// Returns:
//   - error: Error if KV Worker is not initialized or channel creation fails
func CreateKVChannelForWorker(goWorkerJS js.Value, channelName string) error {
	globalMu.Lock()
	workerJS := globalWorkerJS
	globalMu.Unlock()

	if workerJS.IsUndefined() || workerJS.IsNull() {
		return errors.New("KV Worker not initialized")
	}

	if goWorkerJS.IsUndefined() || goWorkerJS.IsNull() {
		return errors.New("Go worker is undefined or null")
	}

	// Create a MessageChannel
	messageChannel := js.Global().Get("MessageChannel").New()
	port1 := messageChannel.Get("port1") // For KV Worker
	port2 := messageChannel.Get("port2") // For Go worker

	// Send port1 to KV Worker with channel name
	// KV Worker expects: { channel: "channelName" } with port in transfer list
	kvRegMsg := map[string]any{
		"channel": channelName,
	}
	kvRegJSON, err := json.Marshal(kvRegMsg)
	if err != nil {
		return errors.Wrap(err, "failed to marshal KV registration message")
	}
	kvTransferList := js.Global().Get("Array").New(port1)
	workerJS.Call("postMessage", string(kvRegJSON), kvTransferList)

	// Send port2 to Go worker with the format it expects
	// Go workers expect: { port: MessagePort, channel: Uint8Array, key: Uint8Array }
	channelBytes := js.Global().Get("Uint8Array").New(len(channelName))
	js.CopyBytesToJS(channelBytes, []byte(channelName))

	keyStr := string(PortOp)
	keyBytes := js.Global().Get("Uint8Array").New(len(keyStr))
	js.CopyBytesToJS(keyBytes, []byte(keyStr))

	goWorkerMsg := map[string]any{
		"port":    port2,
		"channel": channelBytes,
		"key":     keyBytes,
	}
	goTransferList := js.Global().Get("Array").New(port2)
	goWorkerJS.Call("postMessage", goWorkerMsg, goTransferList)

	jww.DEBUG.Printf("[KV] Created KV channel between KV Worker and Go worker: %s", channelName)
	return nil
}

// HasStore returns true if the global store has been initialized.
func HasStore() bool {
	globalMu.Lock()
	defer globalMu.Unlock()
	return globalStore != nil
}

// SetStore sets the global store directly. This is used by worker threads
// that receive their KV port via a different mechanism.
func SetStore(s Store) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalStore = s
	jww.INFO.Print("[KV] Global store set directly")
}

// StopKVClient stops the global KV client and releases resources.
func StopKVClient() {
	globalMu.Lock()
	defer globalMu.Unlock()

	if globalClient != nil {
		globalClient.Stop()
		globalClient = nil
	}
	globalStore = nil

	jww.INFO.Print("[KV] Global KV client stopped")
}
