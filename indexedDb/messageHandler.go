////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package indexedDb

import (
	"encoding/json"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	worker "gitlab.com/elixxir/xxdk-wasm/indexedDbWorker"
	"gitlab.com/elixxir/xxdk-wasm/utils"
	"sync"
	"syscall/js"
)

// HandlerFn is the function that handles incoming data from the main thread.
type HandlerFn func(data []byte) ([]byte, error)

// MessageHandler queues incoming messages from the main thread and handles them
// based on their tag.
type MessageHandler struct {
	// messages is a list of queued messages sent from the main thread.
	messages chan js.Value

	// handlers is a list of functions to handle messages that come from the
	// main thread keyed on the handler tag.
	handlers map[worker.Tag]HandlerFn

	// name describes the worker. It is used for debugging and logging purposes.
	name string

	mux sync.Mutex
}

// NewMessageHandler initialises a new MessageHandler.
func NewMessageHandler(name string) *MessageHandler {
	mh := &MessageHandler{
		messages: make(chan js.Value, 100),
		handlers: make(map[worker.Tag]HandlerFn),
		name:     name,
	}

	mh.addEventListeners()

	return mh
}

// SignalReady sends a signal to the main thread indicating that the worker is
// ready. Once the main thread receives this, it will initiate communication.
// Therefore, this should only be run once all listeners are ready.
func (mh *MessageHandler) SignalReady() {
	mh.SendResponse(worker.ReadyTag, worker.InitID, nil)
}

// SendResponse sends a reply to the main thread with the given tag and ID,
func (mh *MessageHandler) SendResponse(
	tag worker.Tag, id uint64, data []byte) {
	msg := worker.WorkerMessage{
		Tag:  tag,
		ID:   id,
		Data: data,
	}
	jww.DEBUG.Printf("[WW] [%s] Worker sending message for %q and ID %d with "+
		"data: %s", mh.name, tag, id, data)

	payload, err := json.Marshal(msg)
	if err != nil {
		jww.FATAL.Panicf("[WW] [%s] Worker failed to marshal %T for %q and ID "+
			"%d going to main: %+v", mh.name, msg, tag, id, err)
	}

	go mh.postMessage(string(payload))
}

// receiveMessage is registered with the Javascript event listener and is called
// everytime a message from the main thread is received. If the registered
// handler returns a response, it is sent to the main thread.
func (mh *MessageHandler) receiveMessage(data []byte) error {
	var msg worker.WorkerMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return err
	}
	jww.DEBUG.Printf("[WW] [%s] Worker received message for %q and ID %d with "+
		"data: %s", mh.name, msg.Tag, msg.ID, msg.Data)

	mh.mux.Lock()
	handler, exists := mh.handlers[msg.Tag]
	mh.mux.Unlock()
	if !exists {
		return errors.Errorf("no handler found for tag %q", msg.Tag)
	}

	// Call handler and register response with its return
	go func() {
		response, err2 := handler(msg.Data)
		if err2 != nil {
			jww.ERROR.Printf("[WW] [%s] Handler for for %q and ID %d returned "+
				"an error: %+v", mh.name, msg.Tag, msg.ID, err)
		}
		if response != nil {
			mh.SendResponse(msg.Tag, msg.ID, response)
		}
	}()

	return nil
}

// RegisterHandler registers the handler with the given tag overwriting any
// previous registered handler with the same tag. This function is thread safe.
//
// If the handler returns anything but nil, it will be returned as a response.
func (mh *MessageHandler) RegisterHandler(tag worker.Tag, handler HandlerFn) {
	jww.DEBUG.Printf(
		"[WW] [%s] Worker registering handler for tag %q", mh.name, tag)
	mh.mux.Lock()
	mh.handlers[tag] = handler
	mh.mux.Unlock()
}

// addEventListeners adds the event listeners needed to receive a message from
// the worker. Two listeners were added; one to receive messages from the worker
// and the other to receive errors.
func (mh *MessageHandler) addEventListeners() {
	// Create a listener for when the message event is fire on the worker. This
	// occurs when a message is received from the main thread.
	// Doc: https://developer.mozilla.org/en-US/docs/Web/API/Worker/message_event
	messageEvent := js.FuncOf(func(_ js.Value, args []js.Value) any {
		err := mh.receiveMessage([]byte(args[0].Get("data").String()))
		if err != nil {
			jww.ERROR.Printf("[WW] [%s] Failed to receive message from "+
				"main thread: %+v", mh.name, err)
		}
		return nil
	})

	// Create listener for when a messageerror event is fired on the worker.
	// This occurs when it receives a message that cannot be deserialized.
	// Doc: https://developer.mozilla.org/en-US/docs/Web/API/Worker/messageerror_event
	messageError := js.FuncOf(func(_ js.Value, args []js.Value) any {
		event := args[0]
		jww.ERROR.Printf("[WW] [%s] Worker received error message from main "+
			"thread: %s", mh.name, utils.JsToJson(event))
		return nil
	})

	// Register each event listener on the worker using addEventListener
	// Doc: https://developer.mozilla.org/en-US/docs/Web/API/EventTarget/addEventListener
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
func (mh *MessageHandler) postMessage(aMessage any) {
	js.Global().Call("postMessage", aMessage)
}
