////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package worker

import (
	"encoding/json"
	"sync"
	"syscall/js"
	"time"

	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"

	"gitlab.com/elixxir/wasm-utils/utils"
)

// initID is the ID for the first item in the callback list. If the list only
// contains one callback, then this is the ID of that callback. If the list has
// autogenerated unique IDs, this is the initial ID to start at.
const initID = uint64(0)

// Response timeouts.
const (
	// workerInitialConnectionTimeout is the time to wait to receive initial
	// contact from a new worker before timing out.
	workerInitialConnectionTimeout = 90 * time.Second

	// ResponseTimeout is the general time to wait after sending a message to
	// receive a response before timing out.
	ResponseTimeout = 30 * time.Second
)

// receiveQueueChanSize is the size of the channel that received messages are
// put on.
const receiveQueueChanSize = 100

// ReceptionCallback is called with a message received from the worker.
type ReceptionCallback func(data []byte)

// Manager manages the handling of messages received from the worker.
type Manager struct {
	// worker is the Worker Javascript object.
	// Doc: https://developer.mozilla.org/en-US/docs/Web/API/Worker
	worker js.Value

	// callbacks are a list of ReceptionCallback that handle a specific message
	// received from the worker. Each callback is keyed on a tag specifying how
	// the received message should be handled. If the message is a reply to a
	// message sent to the worker, then the callback is also keyed on a unique
	// ID. If the message is not a reply, then it appears on initID.
	callbacks map[Tag]map[uint64]ReceptionCallback

	// responseIDs is a list of the newest ID to assign to each callback when
	// registered. The IDs are used to connect a reply from the worker to the
	// original message sent by the main thread.
	responseIDs map[Tag]uint64

	// receiveQueue is the channel that all received messages are queued on
	// while they wait to be processed.
	receiveQueue chan js.Value

	// quit, when triggered, stops the thread that processes received messages.
	quit chan struct{}

	// name describes the worker. It is used for debugging and logging purposes.
	name string

	// messageLogging determines if debug message logs should be printed every
	// time a message is sent/received to/from the worker.
	messageLogging bool

	mux sync.Mutex
}

// NewManager generates a new Manager. This functions will only return once
// communication with the worker has been established.
func NewManager(aURL, name string, messageLogging bool) (*Manager, error) {
	// Create new worker options with the given name
	opts := newWorkerOptions("", "", name)

	m := &Manager{
		worker:         js.Global().Get("Worker").New(aURL, opts),
		callbacks:      make(map[Tag]map[uint64]ReceptionCallback),
		responseIDs:    make(map[Tag]uint64),
		receiveQueue:   make(chan js.Value, receiveQueueChanSize),
		quit:           make(chan struct{}),
		name:           name,
		messageLogging: messageLogging,
	}

	// Start thread to process responses from worker
	go m.processThread()

	// Register listeners on the Javascript worker object that receive messages
	// and errors from the worker
	m.addEventListeners()

	// Register a callback that will receive initial message from worker
	// indicating that it is ready
	ready := make(chan struct{})
	m.RegisterCallback(readyTag, func([]byte) { ready <- struct{}{} })

	// Wait for the ready signal from the worker
	select {
	case <-ready:
	case <-time.After(workerInitialConnectionTimeout):
		return nil, errors.Errorf("[WW] [%s] timed out after %s waiting for "+
			"initial message from worker",
			m.name, workerInitialConnectionTimeout)
	}

	return m, nil
}

// Stop closes the worker manager and terminates the worker.
func (m *Manager) Stop() {
	// Stop processThread
	select {
	case m.quit <- struct{}{}:
	}

	// Terminate the worker
	go m.terminate()
}

// processThread processes received messages sequentially.
func (m *Manager) processThread() {
	jww.INFO.Printf("[WW] [%s] Starting process thread.", m.name)
	for {
		select {
		case <-m.quit:
			jww.INFO.Printf("[WW] [%s] Quitting process thread.", m.name)
			return
		case msgData := <-m.receiveQueue:

			switch msgData.Type() {
			case js.TypeObject:
				if msgData.Get("constructor").Equal(utils.Uint8Array) {
					err := m.processReceivedMessage(utils.CopyBytesToGo(msgData))
					if err != nil {
						jww.ERROR.Printf("[WW] [%s] Failed to process received "+
							"message from worker: %+v", m.name, err)
					}
					break
				}
				fallthrough

			default:
				jww.ERROR.Printf("[WW] [%s] Cannot handle data of type %s "+
					"from worker: %s", m.name, msgData.Type(),
					utils.JsToJson(msgData))
			}
		}
	}
}

// SendMessage sends a message to the worker with the given tag. If a reception
// callback is specified, then the message is given a unique ID to handle the
// reply. Set receptionCB to nil if no reply is expected.
func (m *Manager) SendMessage(
	tag Tag, data []byte, receptionCB ReceptionCallback) {
	var id uint64
	if receptionCB != nil {
		id = m.registerReplyCallback(tag, receptionCB)
	}

	if m.messageLogging {
		jww.DEBUG.Printf("[WW] [%s] Main sending message for %q and ID %d "+
			"with data: %s", m.name, tag, id, data)
	}

	msg := Message{
		Tag:  tag,
		ID:   id,
		Data: data,
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		jww.FATAL.Panicf("[WW] [%s] Main failed to marshal %T for %q and "+
			"ID %d going to worker: %+v", m.name, msg, tag, id, err)
	}

	go m.postMessage(payload)
}

// receiveMessage is registered with the Javascript event listener and is called
// every time a new message from the worker is received.
func (m *Manager) receiveMessage(data js.Value) {
	m.receiveQueue <- data
}

// processReceivedMessage processes the message received from the worker and
// calls the associated callback. This functions blocks until the callback
// returns.
func (m *Manager) processReceivedMessage(data []byte) error {
	var msg Message
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return err
	}

	if m.messageLogging {
		jww.DEBUG.Printf("[WW] [%s] Main received message for %q and ID %d "+
			"with data: %s", m.name, msg.Tag, msg.ID, msg.Data)
	}

	callback, err := m.getCallback(msg.Tag, msg.ID, msg.DeleteCB)
	if err != nil {
		return err
	}

	callback(msg.Data)

	return nil
}

// getCallback returns the callback for the given ID or returns an error if no
// callback is found. The callback is deleted from the map if specified in the
// message. This function is thread safe.
func (m *Manager) getCallback(
	tag Tag, id uint64, deleteCB bool) (ReceptionCallback, error) {
	m.mux.Lock()
	defer m.mux.Unlock()
	callbacks, exists := m.callbacks[tag]
	if !exists {
		return nil, errors.Errorf("no callbacks found for tag %q", tag)
	}

	callback, exists := callbacks[id]
	if !exists {
		return nil, errors.Errorf("no %q callback found for ID %d", tag, id)
	}

	if deleteCB {
		delete(m.callbacks[tag], id)
		if len(m.callbacks[tag]) == 0 {
			delete(m.callbacks, tag)
		}
	}

	return callback, nil
}

// RegisterCallback registers the reception callback for the given tag. If a
// previous callback was registered, it is overwritten. This function is thread
// safe.
func (m *Manager) RegisterCallback(tag Tag, receptionCB ReceptionCallback) {
	m.mux.Lock()
	defer m.mux.Unlock()

	id := initID

	jww.DEBUG.Printf("[WW] [%s] Main registering callback for tag %q and ID %d",
		m.name, tag, id)

	m.callbacks[tag] = map[uint64]ReceptionCallback{id: receptionCB}
}

// RegisterCallback registers the reception callback for the given tag and a new
// unique ID used to associate the reply to the callback. Returns the ID that
// was registered. If a previous callback was registered, it is overwritten.
// This function is thread safe.
func (m *Manager) registerReplyCallback(
	tag Tag, receptionCB ReceptionCallback) uint64 {
	m.mux.Lock()
	defer m.mux.Unlock()
	id := m.getNextID(tag)

	jww.DEBUG.Printf("[WW] [%s] Main registering callback for tag %q and ID %d",
		m.name, tag, id)

	if _, exists := m.callbacks[tag]; !exists {
		m.callbacks[tag] = make(map[uint64]ReceptionCallback)
	}
	m.callbacks[tag][id] = receptionCB

	return id
}

// getNextID returns the next unique ID for the given tag. This function is not
// thread-safe.
func (m *Manager) getNextID(tag Tag) uint64 {
	if _, exists := m.responseIDs[tag]; !exists {
		m.responseIDs[tag] = initID
	}

	id := m.responseIDs[tag]
	m.responseIDs[tag]++
	return id
}

// GetWorker returns the web worker object. This returned so the worker object
// can be returned to the Javascript layer for it to communicate with the worker
// thread.
func (m *Manager) GetWorker() js.Value { return m.worker }

// Name returns the name of the web worker object.
func (m *Manager) Name() string { return m.name }

////////////////////////////////////////////////////////////////////////////////
// Javascript Call Wrappers                                                   //
////////////////////////////////////////////////////////////////////////////////

// addEventListeners adds the event listeners needed to receive a message from
// the worker. Two listeners were added; one to receive messages from the worker
// and the other to receive errors.
func (m *Manager) addEventListeners() {
	// Create a listener for when the message event is fired on the worker. This
	// occurs when a message is received from the worker.
	// Doc: https://developer.mozilla.org/en-US/docs/Web/API/Worker/message_event
	messageEvent := js.FuncOf(func(_ js.Value, args []js.Value) any {
		m.receiveMessage(args[0].Get("data"))
		return nil
	})

	// Create listener for when an error event is fired on the worker. This
	// occurs when an error occurs in the worker.
	// Doc: https://developer.mozilla.org/en-US/docs/Web/API/Worker/error_event
	errorEvent := js.FuncOf(func(_ js.Value, args []js.Value) any {
		event := args[0]
		jww.FATAL.Panicf("[WW] [%s] Main received error event: %+v",
			m.name, js.Error{Value: event})
		return nil
	})

	// Create listener for when a messageerror event is fired on the worker.
	// This occurs when it receives a message that cannot be deserialized.
	// Doc: https://developer.mozilla.org/en-US/docs/Web/API/Worker/messageerror_event
	messageerrorEvent := js.FuncOf(func(_ js.Value, args []js.Value) any {
		event := args[0]
		jww.ERROR.Printf("[WW] [%s] Main received message error event: %+v",
			m.name, js.Error{Value: event})
		return nil
	})

	// Register each event listener on the worker using addEventListener
	// Doc: https://developer.mozilla.org/en-US/docs/Web/API/EventTarget/addEventListener
	m.worker.Call("addEventListener", "message", messageEvent)
	m.worker.Call("addEventListener", "error", errorEvent)
	m.worker.Call("addEventListener", "messageerror", messageerrorEvent)
}

// postMessage sends a message to the worker.
//
// msg is the object to deliver to the worker; this will be in the data
// field in the event delivered to the worker. It must be a transferable object
// because this function transfers ownership of the message instead of copying
// it for better performance. See the doc for more information.
//
// If the message parameter is not provided, a SyntaxError will be thrown by the
// parser. If the data to be passed to the worker is unimportant, js.Null or
// js.Undefined can be passed explicitly.
//
// Doc: https://developer.mozilla.org/en-US/docs/Web/API/Worker/postMessage
func (m *Manager) postMessage(msg []byte) {
	buffer := utils.CopyBytesToJS(msg)
	m.worker.Call("postMessage", buffer, []any{buffer.Get("buffer")})
}

// terminate immediately terminates the Worker. This does not offer the worker
// an opportunity to finish its operations; it is stopped at once.
//
// Doc: https://developer.mozilla.org/en-US/docs/Web/API/Worker/terminate
func (m *Manager) terminate() {
	m.worker.Call("terminate")
}

// newWorkerOptions creates a new Javascript object containing optional
// properties that can be set when creating a new worker.
//
// Each property is optional; leave a property empty to use the defaults (as
// documented). The available properties are:
//   - type - The type of worker to create. The value can be either "classic" or
//     "module". If not specified, the default used is "classic".
//   - credentials - The type of credentials to use for the worker. The value
//     can be "omit", "same-origin", or "include". If it is not specified, or if
//     the type is "classic", then the default used is "omit" (no credentials
//     are required).
//   - name - An identifying name for the worker, used mainly for debugging
//     purposes.
//
// Doc: https://developer.mozilla.org/en-US/docs/Web/API/Worker/Worker#options
func newWorkerOptions(workerType, credentials, name string) js.Value {
	options := make(map[string]any, 3)
	if workerType != "" {
		options["type"] = workerType
	}
	if credentials != "" {
		options["credentials"] = credentials
	}
	if name != "" {
		options["name"] = name
	}
	return js.ValueOf(options)
}
