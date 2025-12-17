////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package worker

import (
	"context"
	"strings"
	"sync"
	"syscall/js"
	"time"

	json "github.com/goccy/go-json"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"

	utils "gitlab.com/elixxir/xxdk-wasm/jsutil"
)

// SenderCallback is called when the sender of a message gets a response. The
// message is the response from the receiver.
type SenderCallback func(message []byte)

// ReceiverCallback is called when receiving a message from the sender. Reply
// can optionally be used to send a response to the caller, triggering the
// [SenderCallback].
type ReceiverCallback func(message []byte, reply func(message []byte))

// NewPortCallback is called with a MessagePort Javascript object when received.
type NewPortCallback func(port js.Value, channelName string)

// MessageManager manages the sending and receiving of messages to a remote
// browser context (e.g., Worker and MessagePort)
type MessageManager struct {
	// The underlying Javascript object that sends and receives messages.
	p MessagePort

	// senderCallbacks are a list of SenderCallback that are called when
	// receiving a response. The uint64 is a unique ID that connects each
	// received reply to its original message.
	senderCallbacks map[Tag]map[uint64]SenderCallback

	// receiverCallbacks are a list of ReceiverCallback that are called when
	// receiving a message.
	receiverCallbacks map[Tag]ReceiverCallback

	// responseIDs is a list of the newest ID to assign to each senderCallbacks
	// when registered. The IDs are used to connect a reply to the original
	// message.
	responseIDs map[Tag]uint64

	// messageChannelCB is a list of callbacks that are called when a new
	// message channel is received.
	messageChannelCB map[string]NewPortCallback

	// quit, when triggered, stops the thread that processes received messages.
	quit chan struct{}

	// name names the underlying Javascript object. It is used for debugging and
	// logging purposes.
	name string

	Params

	mux sync.Mutex
}

// NewMessageManager generates a new MessageManager. This functions will only
// return once communication with the remote thread has been established.
// TODO: test
func NewMessageManager(
	v js.Value, name string, p Params) (*MessageManager, error) {
	mm := initMessageManager(name, p)
	mp, err := NewMessagePort(v)
	if err != nil {
		return nil, errors.Wrap(err, "invalid MessagePort value")
	}
	mm.p = mp

	ctx, cancel := context.WithCancel(context.Background())
	events, err := mm.p.Listen(ctx)
	if err != nil {
		cancel()
		return nil, err
	}

	// Start thread to process responses
	go mm.messageReception(events, cancel)

	return mm, nil
}

// initMessageManager initialises a new empty MessageManager.
func initMessageManager(name string, p Params) *MessageManager {
	return &MessageManager{
		senderCallbacks:   make(map[Tag]map[uint64]SenderCallback),
		receiverCallbacks: make(map[Tag]ReceiverCallback),
		responseIDs:       make(map[Tag]uint64),
		messageChannelCB:  make(map[string]NewPortCallback),
		quit:              make(chan struct{}),
		name:              name,
		Params:            p,
	}
}

// Send sends the data to the remote thread with the given tag and waits for a
// response. Returns an error if calling postMessage throws an exception,
// marshalling the message to send fails, or if receiving a response times out.
//
// It is preferable to use [Send] over [SendNoResponse] as it will report a
// timeout when the remote thread crashes and [SendNoResponse] will not.
// TODO: test
func (mm *MessageManager) Send(tag Tag, data []byte) (response []byte, err error) {
	return mm.SendTimeout(tag, data, mm.ResponseTimeout)
}

// SendTimeout sends the data to the remote thread with a custom timeout. Refer
// to [Send] for more information.
// TODO: test
func (mm *MessageManager) SendTimeout(
	tag Tag, data []byte, timeout time.Duration) (response []byte, err error) {
	responseCh := make(chan []byte)
	id := mm.registerSenderCallback(tag, func(msg []byte) { responseCh <- msg })

	err = mm.sendMessage(tag, id, data)
	if err != nil {
		return nil, err
	}

	select {
	case response = <-responseCh:
		return response, nil
	case <-time.After(timeout):
		return nil,
			errors.Errorf("timed out after %s waiting for response", timeout)
	}
}

// SendNoResponse sends the data to the remote thread with the given tag;
// however, unlike [Send], it returns immediately and does not wait for a
// response. Returns an error if calling postMessage throws an exception,
// marshalling the message to send fails, or if receiving a response times out.
//
// It is preferable to use [Send] over [SendNoResponse] as it will report a
// timeout when the remote thread crashes and [SendNoResponse] will not.
// TODO: test
func (mm *MessageManager) SendNoResponse(tag Tag, data []byte) error {
	return mm.sendMessage(tag, initID, data)
}

// sendMessage packages the data into a Message with the tag and ID and sends it
// to the remote thread.
// TODO: test
func (mm *MessageManager) sendMessage(tag Tag, id uint64, data []byte) error {
	// Note: Cannot use jww logging here as it may recursively call this function
	// through workerLogger, causing deadlock with marshalLock

	msg := Message{Tag: tag, ID: id, Response: false, Data: data}
	payload, err := json.Marshal(msg)
	if err != nil {
		return errors.Wrap(err, "failed to marshal message")
	}
	return mm.p.PostMessageTransferBytes(payload)
}

// sendResponse sends a reply to the remote thread with the given tag and ID.
// TODO: test
func (mm *MessageManager) sendResponse(tag Tag, id uint64, data []byte) error {
	// Note: Cannot use jww logging here as it may recursively call this function
	// through workerLogger, causing deadlock with marshalLock

	msg := Message{Tag: tag, ID: id, Response: true, Data: data}
	payload, err := json.Marshal(msg)
	if err != nil {
		return errors.Wrap(err, "failed to marshal response")
	}
	return mm.p.PostMessageTransferBytes(payload)
}

// messageReception processes received messages sequentially.
// TODO: test
func (mm *MessageManager) messageReception(
	events <-chan MessageEvent, cancel context.CancelFunc) {
	jww.INFO.Printf("[WW] [%s] Starting message reception thread.", mm.name)
	for {
		select {
		case <-mm.quit:
			cancel()
			jww.INFO.Printf(
				"[WW] [%s] Quitting message reception thread.", mm.name)
			return
		case event := <-events:
			// Check for parse errors first
			if err := event.Error(); err != nil {
				// Suppress "Go program has already exited" spam - these flood the
				// console after a crash and hide the original crash trace
				errStr := err.Error()
				if !strings.Contains(errStr, "Go program has already exited") {
					jww.ERROR.Printf("[WW] [%s] Failed to parse message event: %+v", mm.name, err)
				}
				continue
			}

			switch event.EventType() {
			case MessageEventTypeBytes:
				// Data was already copied to Go bytes in parseMessageEvent
				data, err := event.DataBytes()
				if err != nil {
					jww.ERROR.Printf("[WW] [%s] Failed to get message bytes: %+v", mm.name, err)
					continue
				}
				err = mm.processReceivedMessage(data)
				if err != nil {
					jww.ERROR.Printf("[WW] [%s] Failed to process "+
						"received message: %+v", mm.name, err)
				}

			case MessageEventTypePort:
				port, portData, err := event.PortData()
				if err != nil {
					jww.ERROR.Printf("[WW] [%s] Failed to get port data: %+v", mm.name, err)
					continue
				}
				err = mm.processReceivedPort(port, portData)
				if err != nil {
					jww.ERROR.Printf("[WW] [%s] Failed to process "+
						"received MessagePort: %+v", mm.name, err)
				}

			case MessageEventTypeUnknown:
				jww.ERROR.Printf("[WW] [%s] Received unknown message event type", mm.name)
			}
		}
	}
}

// processReceivedMessage processes the received message and calls the
// associated callback. This functions blocks until the callback returns.
func (mm *MessageManager) processReceivedMessage(data []byte) error {
	// CRITICAL: Copy data FIRST, before ANY use including logging.
	// json.Unmarshal creates string headers that point into the input buffer.
	// Even string(data) for logging can hold references that outlive the buffer.
	// If the original buffer gets GC'd while strings are still in use
	// (especially in async logging), we get "marked free object in span" errors.
	// By copying first, all subsequent operations use memory we own.
	dataCopy := append([]byte(nil), data...)

	// DEBUG LOGGING DISABLED - investigating if %q formatting causes zombie pointers
	// jww.DEBUG.Printf("[WW] [%s] Raw received data: %q", mm.name, string(dataCopy))

	var msg Message
	err := json.Unmarshal(dataCopy, &msg)
	if err != nil {
		jww.ERROR.Printf("[WW] [%s] Failed to unmarshal message: %+v", mm.name, err)
		return err
	}

	// Copy msg.Tag to ensure it doesn't reference dataCopy after this function returns
	// (closures below capture msg.Tag which could outlive dataCopy)
	tagCopy := Tag(string([]byte(msg.Tag)))
	idCopy := msg.ID

	// DEBUG LOGGING DISABLED - investigating zombie pointer issue
	_ = mm.MessageLogging // suppress unused warning

	if msg.Response {
		callback, err := mm.getSenderCallback(tagCopy, idCopy)
		if err != nil {
			return err
		}

		callback(msg.Data)
	} else {
		callback, err := mm.getReceiverCallback(tagCopy)
		if err != nil {
			return err
		}

		callback(msg.Data, func(message []byte) {
			if err = mm.sendResponse(tagCopy, idCopy, message); err != nil {
				jww.FATAL.Panicf("[WW] [%s] Failed to send response for %q "+
					"and ID %d: %+v", mm.name, tagCopy, idCopy, err)
			}
		})
	}

	return nil
}

// processReceivedPort processes the received Javascript MessagePort and calls
// the associated NewPortCallback callback. This functions blocks until the
// callback returns.
func (mm *MessageManager) processReceivedPort(port js.Value, data js.Value) error {
	channel := string(utils.CopyBytesToGo(data.Get("channel")))
	key := string(utils.CopyBytesToGo(data.Get("key")))

	jww.INFO.Printf("[WW] [%s] Received new MessageChannel %q for key %q.",
		mm.name, channel, key)

	mm.mux.Lock()
	cb, exists := mm.messageChannelCB[key]
	mm.mux.Unlock()

	if !exists {
		return errors.Errorf(
			"Failed to find callback for channel %q and key %q.", channel, key)
	} else {
		cb(port, channel)
	}

	return nil
}

// RegisterCallback registers the callback for the given tag. Previous tags are
// overwritten. This function is thread safe.
func (mm *MessageManager) RegisterCallback(tag Tag, receiverCB ReceiverCallback) {
	mm.mux.Lock()
	defer mm.mux.Unlock()

	jww.DEBUG.Printf("[WW] [%s] Registering receiver callback for tag %q",
		mm.name, tag)

	mm.receiverCallbacks[tag] = receiverCB
}

// getReceiverCallback returns the ReceiverCallback for the given Tag or returns
// an error if no callback is found. This function is thread safe.
func (mm *MessageManager) getReceiverCallback(tag Tag) (ReceiverCallback, error) {
	mm.mux.Lock()
	defer mm.mux.Unlock()

	callback, exists := mm.receiverCallbacks[tag]
	if !exists {
		return nil, errors.Errorf("no receiver callbacks found for tag %q", tag)
	}

	return callback, nil
}

// registerSenderCallback registers the callback for the given tag and a new
// unique ID used to associate the reply to the callback. Returns the ID that
// was registered. If a previous callback was registered, it is overwritten.
// This function is thread safe.
func (mm *MessageManager) registerSenderCallback(
	tag Tag, senderCB SenderCallback) uint64 {
	mm.mux.Lock()
	defer mm.mux.Unlock()
	id := mm.getNextID(tag)

	jww.DEBUG.Printf("[WW] [%s] Registering callback for tag %q and ID %d",
		mm.name, tag, id)

	if _, exists := mm.senderCallbacks[tag]; !exists {
		mm.senderCallbacks[tag] = make(map[uint64]SenderCallback)
	}
	mm.senderCallbacks[tag][id] = senderCB

	return id
}

// getSenderCallback returns the SenderCallback for the given Tag and ID or
// returns an error if no callback is found. The callback is deleted from the
// map once found. This function is thread safe.
func (mm *MessageManager) getSenderCallback(
	tag Tag, id uint64) (SenderCallback, error) {
	mm.mux.Lock()
	defer mm.mux.Unlock()
	callbacks, exists := mm.senderCallbacks[tag]
	if !exists {
		return nil, errors.Errorf("no sender callbacks found for tag %q", tag)
	}

	callback, exists := callbacks[id]
	if !exists {
		return nil,
			errors.Errorf("no %q sender callback found for ID %d", tag, id)
	}

	delete(mm.senderCallbacks[tag], id)
	if len(mm.senderCallbacks[tag]) == 0 {
		delete(mm.senderCallbacks, tag)
	}

	return callback, nil
}

// RegisterMessageChannelCallback registers a callback that will be called when
// a MessagePort with the given Channel is received.
func (mm *MessageManager) RegisterMessageChannelCallback(
	key string, fn NewPortCallback) {
	mm.mux.Lock()
	defer mm.mux.Unlock()
	mm.messageChannelCB[key] = fn
}

// Stop closes the message reception thread and closes the port.
// TODO: test
func (mm *MessageManager) Stop() {
	// Stop messageReception
	select {
	case mm.quit <- struct{}{}:
	}
}

// getNextID returns the next unique ID for the given tag. This function is not
// thread-safe.
func (mm *MessageManager) getNextID(tag Tag) uint64 {
	if _, exists := mm.responseIDs[tag]; !exists {
		mm.responseIDs[tag] = initID
	}

	id := mm.responseIDs[tag]
	mm.responseIDs[tag]++
	return id
}
