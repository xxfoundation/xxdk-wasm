////////////////////////////////////////////////////////////////////////////////
// Copyright © 2024 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package kv

import (
	"io/fs"
	"strings"
	"sync"
	"sync/atomic"
	"syscall/js"
	"time"

	json "github.com/goccy/go-json"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
)

// KVWorkerClient communicates with the JavaScript KV Worker via MessageChannel.
// This is a self-contained implementation that doesn't use the worker package.
type KVWorkerClient struct {
	port      js.Value           // Our end of the MessageChannel
	name      string             // For logging
	timeout   time.Duration      // Timeout for operations
	nextID    uint64             // Atomic counter for message IDs
	callbacks sync.Map           // map[uint64]chan *KVMessage
	onMessage js.Func            // Message handler (prevent GC)
	quit      chan struct{}      // Signal to stop
	mu        sync.Mutex         // Protect initialization
}

// DefaultKVTimeout is the default timeout for KV operations.
const DefaultKVTimeout = 30 * time.Second

// NewKVWorkerClient creates a new client that communicates with the KV Worker.
// It creates a MessageChannel, sends one port to the worker, and keeps the other.
//
// Parameters:
//   - worker: The JavaScript Worker object for the KV Worker
//   - channelName: Name for this channel (for logging and registration)
func NewKVWorkerClient(worker js.Value, channelName string) (*KVWorkerClient, error) {
	if worker.IsUndefined() || worker.IsNull() {
		return nil, errors.New("worker is undefined or null")
	}

	// Create a MessageChannel
	messageChannel := js.Global().Get("MessageChannel").New()
	port1 := messageChannel.Get("port1") // Our port
	port2 := messageChannel.Get("port2") // Worker's port

	client := &KVWorkerClient{
		port:    port1,
		name:    channelName,
		timeout: DefaultKVTimeout,
		quit:    make(chan struct{}),
	}

	// Set up message handler on our port
	client.onMessage = js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) == 0 {
			return nil
		}
		event := args[0]
		data := event.Get("data")
		client.handleResponse(data)
		return nil
	})
	port1.Set("onmessage", client.onMessage)
	port1.Call("start")

	// Send port2 to the worker with registration message
	// The worker expects: { channel: "channelName" } with the port in the transfer list
	regMsg := map[string]any{
		"channel": channelName,
	}
	regJSON, err := json.Marshal(regMsg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal registration message")
	}

	// PostMessage with transfer list to send port2 to worker
	transferList := js.Global().Get("Array").New(port2)
	worker.Call("postMessage", string(regJSON), transferList)

	jww.INFO.Printf("[KV] KVWorkerClient created for channel: %s", channelName)
	return client, nil
}

// handleResponse processes incoming messages from the KV Worker.
func (c *KVWorkerClient) handleResponse(data js.Value) {
	// Convert data to string
	var text string
	if data.Type() == js.TypeString {
		text = data.String()
	} else {
		// Might be Uint8Array or other
		jww.WARN.Printf("[KV] [%s] Unexpected data type: %v", c.name, data.Type())
		return
	}

	// Parse the KVMessage
	var msg KVMessage
	if err := json.Unmarshal([]byte(text), &msg); err != nil {
		jww.ERROR.Printf("[KV] [%s] Failed to unmarshal response: %+v", c.name, err)
		return
	}

	// Only process responses
	if !msg.Response {
		return
	}

	// Find and notify the waiting callback
	if ch, ok := c.callbacks.LoadAndDelete(msg.ID); ok {
		ch.(chan *KVMessage) <- &msg
	} else {
		jww.WARN.Printf("[KV] [%s] No callback for response ID %d", c.name, msg.ID)
	}
}

// getNextID returns the next unique message ID.
func (c *KVWorkerClient) getNextID() uint64 {
	return atomic.AddUint64(&c.nextID, 1)
}

// sendMessage sends a KVMessage and waits for response.
func (c *KVWorkerClient) sendMessage(msg *KVMessage) (*KVMessage, error) {
	// Create response channel
	responseCh := make(chan *KVMessage, 1)
	c.callbacks.Store(msg.ID, responseCh)

	// Clean up on exit
	defer c.callbacks.Delete(msg.ID)

	// Marshal and send message
	payload, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal KV message")
	}

	c.port.Call("postMessage", string(payload))

	// Wait for response with timeout
	select {
	case resp := <-responseCh:
		return resp, nil
	case <-time.After(c.timeout):
		return nil, errors.Errorf("timeout after %s waiting for KV response (op=%s, key=%s)",
			c.timeout, msg.Op, msg.Key)
	case <-c.quit:
		return nil, errors.New("KV client stopped")
	}
}

// Get retrieves a string value by key.
func (c *KVWorkerClient) Get(key string) (string, error) {
	msg := &KVMessage{
		Op:  string(GetOp),
		ID:  c.getNextID(),
		Key: key,
	}

	resp, err := c.sendMessage(msg)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get key %q", key)
	}

	if resp.Error != "" {
		if strings.Contains(resp.Error, "does not exist") {
			return "", fs.ErrNotExist
		}
		return "", errors.New(resp.Error)
	}

	return resp.Value, nil
}

// Set stores a string value for a key.
func (c *KVWorkerClient) Set(key string, value string) error {
	msg := &KVMessage{
		Op:    string(SetOp),
		ID:    c.getNextID(),
		Key:   key,
		Value: value,
	}

	resp, err := c.sendMessage(msg)
	if err != nil {
		return errors.Wrapf(err, "failed to set key %q", key)
	}

	if resp.Error != "" {
		return errors.New(resp.Error)
	}

	return nil
}

// Delete removes a key.
func (c *KVWorkerClient) Delete(key string) error {
	msg := &KVMessage{
		Op:  string(DeleteOp),
		ID:  c.getNextID(),
		Key: key,
	}

	resp, err := c.sendMessage(msg)
	if err != nil {
		return errors.Wrapf(err, "failed to delete key %q", key)
	}

	if resp.Error != "" {
		return errors.New(resp.Error)
	}

	return nil
}

// Keys returns all key names.
func (c *KVWorkerClient) Keys() ([]string, error) {
	msg := &KVMessage{
		Op: string(KeysOp),
		ID: c.getNextID(),
	}

	resp, err := c.sendMessage(msg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get keys")
	}

	if resp.Error != "" {
		return nil, errors.New(resp.Error)
	}

	// Parse keys from value (JSON array)
	var keys []string
	if resp.Value != "" {
		if err := json.Unmarshal([]byte(resp.Value), &keys); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal keys")
		}
	}

	return keys, nil
}

// Stop closes the client and releases resources.
func (c *KVWorkerClient) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-c.quit:
		// Already stopped
		return
	default:
		close(c.quit)
	}

	// Release the JS function
	c.onMessage.Release()

	// Close the port
	c.port.Call("close")

	jww.INFO.Printf("[KV] [%s] KVWorkerClient stopped", c.name)
}

// Ensure KVWorkerClient implements StringStore
var _ StringStore = (*KVWorkerClient)(nil)

// NewWorkerThread creates a Store from a MessagePort received by a worker thread.
// This is used by channels/dm workers that receive a KV port via MessageChannel.
//
// Parameters:
//   - port: The MessagePort js.Value received from the main thread
//   - channelName: Name for this channel (for logging)
func NewWorkerThread(port js.Value, channelName string) (Store, error) {
	if port.IsUndefined() || port.IsNull() {
		return nil, errors.New("port is undefined or null")
	}

	client := &KVWorkerClient{
		port:    port,
		name:    channelName,
		timeout: DefaultKVTimeout,
		quit:    make(chan struct{}),
	}

	// Set up message handler on the port
	client.onMessage = js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) == 0 {
			return nil
		}
		event := args[0]
		data := event.Get("data")
		client.handleResponse(data)
		return nil
	})
	port.Set("onmessage", client.onMessage)
	port.Call("start")

	jww.INFO.Printf("[KV] WorkerThread client created for channel: %s", channelName)

	// Wrap with base32768 encoding for bytes support
	return NewEncodedStore(client), nil
}
