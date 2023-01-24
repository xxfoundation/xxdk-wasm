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
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/xxdk-wasm/utils"
	"sync"
	"syscall/js"
)

// ThreadReceptionCallback is the function that handles incoming data from the
// main thread.
type ThreadReceptionCallback func(data []byte) ([]byte, error)

// ThreadManager queues incoming messages from the main thread and handles them
// based on their tag.
type ThreadManager struct {
	// messages is a list of queued messages sent from the main thread.
	messages chan js.Value

	// callbacks is a list of callbacks to handle messages that come from the
	// main thread keyed on the callback tag.
	callbacks map[Tag]ThreadReceptionCallback

	// name describes the worker. It is used for debugging and logging purposes.
	name string

	// messageLogging determines if debug message logs should be printed every
	// time a message is sent/received to/from the worker.
	messageLogging bool

	mux sync.Mutex
}

// NewThreadManager initialises a new ThreadManager.
func NewThreadManager(name string, messageLogging bool) *ThreadManager {
	mh := &ThreadManager{
		messages:       make(chan js.Value, 100),
		callbacks:      make(map[Tag]ThreadReceptionCallback),
		name:           name,
		messageLogging: messageLogging,
	}

	mh.addEventListeners()

	return mh
}

// SignalReady sends a signal to the main thread indicating that the worker is
// ready. Once the main thread receives this, it will initiate communication.
// Therefore, this should only be run once all listeners are ready.
func (tm *ThreadManager) SignalReady() {
	tm.SendMessage(readyTag, nil)
}

// SendMessage sends a message to the main thread for the given tag.
func (tm *ThreadManager) SendMessage(tag Tag, data []byte) {
	msg := Message{
		Tag:      tag,
		ID:       initID,
		DeleteCB: false,
		Data:     data,
	}

	if tm.messageLogging {
		jww.DEBUG.Printf("[WW] [%s] Worker sending message for %q with data: %s",
			tm.name, tag, data)
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		jww.FATAL.Panicf("[WW] [%s] Worker failed to marshal %T for %q going "+
			"to main: %+v", tm.name, msg, tag, err)
	}

	go tm.postMessage(string(payload))
}

// sendResponse sends a reply to the main thread with the given tag and ID.
func (tm *ThreadManager) sendResponse(
	tag Tag, id uint64, data []byte) {
	msg := Message{
		Tag:      tag,
		ID:       id,
		DeleteCB: true,
		Data:     data,
	}

	if tm.messageLogging {
		jww.DEBUG.Printf("[WW] [%s] Worker sending reply for %q and ID %d "+
			"with data: %s", tm.name, tag, id, data)
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		jww.FATAL.Panicf("[WW] [%s] Worker failed to marshal %T for %q and ID "+
			"%d going to main: %+v", tm.name, msg, tag, id, err)
	}

	go tm.postMessage(string(payload))
}

// receiveMessage is registered with the Javascript event listener and is called
// everytime a message from the main thread is received. If the registered
// callback returns a response, it is sent to the main thread.
func (tm *ThreadManager) receiveMessage(data []byte) error {
	var msg Message
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return err
	}

	if tm.messageLogging {
		jww.DEBUG.Printf("[WW] [%s] Worker received message for %q and ID %d "+
			"with data: %s", tm.name, msg.Tag, msg.ID, msg.Data)
	}

	tm.mux.Lock()
	callback, exists := tm.callbacks[msg.Tag]
	tm.mux.Unlock()
	if !exists {
		return errors.Errorf("no callback found for tag %q", msg.Tag)
	}

	// Call callback and register response with its return
	go func() {
		response, err2 := callback(msg.Data)
		if err2 != nil {
			jww.ERROR.Printf("[WW] [%s] Callback for for %q and ID %d "+
				"returned an error: %+v", tm.name, msg.Tag, msg.ID, err)
		}
		if response != nil {
			tm.sendResponse(msg.Tag, msg.ID, response)
		}
	}()

	return nil
}

// RegisterCallback registers the callback with the given tag overwriting any
// previous registered callbacks with the same tag. This function is thread
// safe.
//
// If the callback returns anything but nil, it will be returned as a response.
func (tm *ThreadManager) RegisterCallback(
	tag Tag, receptionCallback ThreadReceptionCallback) {
	jww.DEBUG.Printf(
		"[WW] [%s] Worker registering callback for tag %q", tm.name, tag)
	tm.mux.Lock()
	tm.callbacks[tag] = receptionCallback
	tm.mux.Unlock()
}

////////////////////////////////////////////////////////////////////////////////
// Javascript Call Wrappers                                                   //
////////////////////////////////////////////////////////////////////////////////

// addEventListeners adds the event listeners needed to receive a message from
// the worker. Two listeners were added; one to receive messages from the worker
// and the other to receive errors.
func (tm *ThreadManager) addEventListeners() {
	// Create a listener for when the message event is fire on the worker. This
	// occurs when a message is received from the main thread.
	// Doc: https://developer.mozilla.org/en-US/docs/Web/API/Worker/message_event
	messageEvent := js.FuncOf(func(_ js.Value, args []js.Value) any {
		err := tm.receiveMessage([]byte(args[0].Get("data").String()))
		if err != nil {
			jww.ERROR.Printf("[WW] [%s] Failed to receive message from "+
				"main thread: %+v", tm.name, err)
		}
		return nil
	})

	// Create listener for when an error event is fired on the worker. This
	// occurs when an error occurs in the worker.
	// Doc: https://developer.mozilla.org/en-US/docs/Web/API/Worker/error_event
	errorEvent := js.FuncOf(func(_ js.Value, args []js.Value) any {
		event := args[0]
		jww.ERROR.Printf("[WW] [%s] Worker received error event: %s",
			tm.name, utils.JsErrorToJson(event))
		return nil
	})

	// Create listener for when a messageerror event is fired on the worker.
	// This occurs when it receives a message that cannot be deserialized.
	// Doc: https://developer.mozilla.org/en-US/docs/Web/API/Worker/messageerror_event
	messageerrorEvent := js.FuncOf(func(_ js.Value, args []js.Value) any {
		event := args[0]
		jww.ERROR.Printf("[WW] [%s] Worker received message error event: %s",
			tm.name, utils.JsErrorToJson(event))
		return nil
	})

	// Register each event listener on the worker using addEventListener
	// Doc: https://developer.mozilla.org/en-US/docs/Web/API/EventTarget/addEventListener
	js.Global().Call("addEventListener", "message", messageEvent)
	js.Global().Call("addEventListener", "error", errorEvent)
	js.Global().Call("addEventListener", "messageerror", messageerrorEvent)
}

// postMessage sends a message from this worker to the main WASM thread.
//
// aMessage must be a js.Value or a primitive type that can be converted via
// js.ValueOf. The Javascript object must be "any value or JavaScript object
// handled by the structured clone algorithm". See the doc for more information.
//
// Doc: https://developer.mozilla.org/docs/Web/API/DedicatedWorkerGlobalScope/postMessage
func (tm *ThreadManager) postMessage(aMessage any) {
	js.Global().Call("postMessage", aMessage)
}
