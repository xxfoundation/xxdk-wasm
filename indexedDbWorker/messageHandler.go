////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package main

import (
	"encoding/json"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/v4/channels"
	"gitlab.com/elixxir/xxdk-wasm/indexedDb"
	"gitlab.com/elixxir/xxdk-wasm/utils"
	"sync"
	"syscall/js"
)

// handlerFn is the function that handles incoming data from the main thread.
type handlerFn func(data []byte) []byte

// messageHandler queues incoming messages from the main thread and handles them
// based on their tag.
type messageHandler struct {
	// messages is a list of queued messages sent from the main thread.
	messages chan js.Value

	// handlers is a list of functions to handle messages that come from the
	// main thread keyed on the handler tag.
	handlers map[indexedDb.Tag]handlerFn

	model channels.EventModel

	mux sync.Mutex
}

// newMessageHandler initialises a new messageHandler.
func newMessageHandler() *messageHandler {
	mh := &messageHandler{
		messages: make(chan js.Value, 100),
		handlers: make(map[indexedDb.Tag]handlerFn),
	}

	mh.addEventListeners()

	return mh
}

// signalReady sends a signal to the main thread indicating that the worker is
// ready. Once the main thread receives this, it will initiate communication.
// Therefore, this should only be run once all listeners are ready.
func (mh *messageHandler) signalReady() {
	mh.sendResponse(indexedDb.ReadyTag, indexedDb.InitialID, nil)
}

// sendResponse sends a reply to the main thread with the given tag and ID,
func (mh *messageHandler) sendResponse(tag indexedDb.Tag, id uint64, data []byte) {
	message := indexedDb.WorkerMessage{
		Tag:  tag,
		ID:   id,
		Data: data,
	}

	payload, err := json.Marshal(message)
	if err != nil {
		jww.FATAL.Panicf("Failed to marshal payload with tag %q and ID %d "+
			"going to main thread: %+v", tag, id, err)
	}

	postMessage(string(payload))
}

// receiveMessage is registered with the Javascript event listener and is called
// everytime a message from the main thread is received. If the registered
// handler returns a response, it is sent to the main thread.
func (mh *messageHandler) receiveMessage(data []byte) error {
	var message indexedDb.WorkerMessage
	err := json.Unmarshal(data, &message)
	if err != nil {
		return errors.Errorf(
			"could not unmarshal payload from worker: %+v", err)
	}

	mh.mux.Lock()
	handler, exists := mh.handlers[message.Tag]
	mh.mux.Unlock()
	if !exists {
		return errors.Errorf("no handler found for tag %q", message.Tag)
	}

	// Call handler and register response with its return
	go func() {
		response := handler(message.Data)
		if response != nil {
			mh.sendResponse(message.Tag, message.ID, response)
		}
	}()

	return nil
}

// registerHandler registers the handler with the given tag overwriting any
// previous registered handler with the same tag. This function is thread safe.
func (mh *messageHandler) registerHandler(tag indexedDb.Tag, handler handlerFn) {
	mh.mux.Lock()
	mh.handlers[tag] = handler
	mh.mux.Unlock()
}

// addEventListeners adds the event listeners needed to receive a message from
// the worker. Two listeners were added; one to receive messages from the worker
// and the other to receive errors.
func (mh *messageHandler) addEventListeners() {
	// Create a listener for when the message event is fire on the worker. This
	// occurs when a message is received from the main thread.
	// Doc: https://developer.mozilla.org/en-US/docs/Web/API/Worker/message_event
	messageEvent := js.FuncOf(func(this js.Value, args []js.Value) any {
		err := mh.receiveMessage([]byte(args[0].Get("data").String()))
		if err != nil {
			jww.ERROR.Printf("Failed to receive message from main thread: %+v", err)
		}
		return nil
	})

	// Create listener for when a messageerror event is fired on the worker.
	// This occurs when it receives a message that can't be deserialized.
	// Doc: https://developer.mozilla.org/en-US/docs/Web/API/Worker/messageerror_event
	messageError := js.FuncOf(func(this js.Value, args []js.Value) any {
		event := args[0]
		jww.ERROR.Printf(
			"Error receiving message from main thread: %s", utils.JsToJson(event))
		return nil
	})

	js.Global().Call("addEventListener", "message", messageEvent)
	js.Global().Call("addEventListener", "messageerror", messageError)
}

// postMessage sends a message from this worker to the main WASM thread.
//
// aMessage must be a js.Value or a primitive type that can be converted via
// js.ValueOf. The Javascript object must be "any value or JavaScript object
// handled by the structured clone algorithm". See the doc for more information.
//
// Doc: https://developer.mozilla.org/docs/Web/API/DedicatedWorkerGlobalScope/postMessage
func postMessage(aMessage any) {
	js.Global().Call("postMessage", aMessage)
}

// postMessageTransferList sends an array of Transferable objects to transfer to
// the main thread. This is meant to be used to transfer large amounts of binary
// data using a high-performance, zero-copy operation. Refer to the doc for more
// information.
//
// Note: The binary data cannot simply be passed as the transferList. The
// traversable objects must be specified in the aMessage and included in the
// transferList
//
// Doc: https://developer.mozilla.org/en-US/docs/Web/API/DedicatedWorkerGlobalScope/postMessage
func postMessageTransferList(aMessage, transferList any) {
	js.Global().Call("postMessage", aMessage, transferList)
}
