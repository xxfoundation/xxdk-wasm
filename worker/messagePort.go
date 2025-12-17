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
	"syscall/js"

	"github.com/pkg/errors"

	"gitlab.com/elixxir/xxdk-wasm/jsutil"
)

// MessagePort wraps a Javascript MessagePort object.
//
// Doc: https://developer.mozilla.org/en-US/docs/Web/API/MessagePort
type MessagePort struct {
	js.Value
}

// NewMessagePort wraps the given MessagePort.
func NewMessagePort(v js.Value) (MessagePort, error) {
	method, err := jsutil.Get(v, "postMessage")
	if err != nil {
		return MessagePort{}, err
	}
	if method.Type() != js.TypeFunction {
		return MessagePort{}, errors.New("postMessage is not a function")
	}
	return MessagePort{v}, nil
}

// PostMessage sends a message from the port.
func (mp MessagePort) PostMessage(message any) error {
	_, err := jsutil.Call(mp.Value, "postMessage", message)
	return err
}

// PostMessageTransfer sends a message from the port and transfers ownership of
// objects to other browsing contexts.
func (mp MessagePort) PostMessageTransfer(message any, transfer ...any) error {
	_, err := jsutil.Call(mp.Value, "postMessage", message, transfer)
	return err
}

// PostMessageTransferBytes sends the message bytes from the port via transfer.
func (mp MessagePort) PostMessageTransferBytes(message []byte) error {
	// Create Uint8Array and copy bytes
	buffer := jsutil.Uint8Array.New(len(message))
	js.CopyBytesToJS(buffer, message)

	// Transfer the underlying ArrayBuffer
	mp.Value.Call("postMessage", buffer, []any{buffer.Get("buffer")})
	return nil
}

// PostMessageString sends a string message from the port.
// This is simpler and more efficient than PostMessageTransferBytes for JSON messages.
func (mp MessagePort) PostMessageString(message string) error {
	mp.Value.Call("postMessage", message)
	return nil
}

// Listen registers listeners on the MessagePort and returns all events on the
// returned channel.
func (mp MessagePort) Listen(
	ctx context.Context) (_ <-chan MessageEvent, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		if err != nil {
			cancel()
		}
	}()

	events := make(chan MessageEvent)

	// IMPORTANT: We must extract data from args SYNCHRONOUSLY (while js.Value is valid),
	// but send to channel in a GOROUTINE (to not block the JS event loop).
	// This prevents both "marked free object in span" errors AND deadlocks.

	messageHandler, err := jsutil.FuncOf(func(_ js.Value, args []js.Value) any {
		// Extract synchronously while args[0] is valid
		event := parseMessageEvent(args[0])
		// Send in goroutine to not block JS
		go func() { events <- event }()
		return nil
	})
	if err != nil {
		return nil, err
	}
	errorHandler, err := jsutil.FuncOf(func(_ js.Value, args []js.Value) any {
		// Extract error message synchronously while args[0] is valid
		jsErr := js.Error{Value: args[0]}
		event := MessageEvent{
			err:       errors.New(jsErr.Error()),
			eventType: MessageEventTypeUnknown,
		}
		// Send in goroutine to not block JS
		go func() { events <- event }()
		return nil
	})
	if err != nil {
		return nil, err
	}
	messageErrorHandler, err := jsutil.FuncOf(func(_ js.Value, args []js.Value) any {
		// Extract synchronously while args[0] is valid
		event := parseMessageEvent(args[0])
		// Send in goroutine to not block JS
		go func() { events <- event }()
		return nil
	})
	if err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		jsutil.Call(mp.Value, "removeEventListener", "message", messageHandler)
		jsutil.ReleaseFunc(messageHandler)
		jsutil.Call(mp.Value, "removeEventListener", "error", errorHandler)
		jsutil.ReleaseFunc(errorHandler)
		jsutil.Call(mp.Value, "removeEventListener", "messageerror", messageErrorHandler)
		jsutil.ReleaseFunc(messageErrorHandler)
		close(events)
	}()

	_, err = jsutil.Call(mp.Value, "addEventListener", "message", messageHandler)
	if err != nil {
		return nil, err
	}
	_, err = jsutil.Call(mp.Value, "addEventListener", "error", errorHandler)
	if err != nil {
		return nil, err
	}
	_, err = jsutil.Call(mp.Value, "addEventListener", "messageerror", messageErrorHandler)
	if err != nil {
		return nil, err
	}

	// Check if port needs to be started (MessagePort from MessageChannel)
	start, err := jsutil.Get(mp.Value, "start")
	if err == nil && start.Truthy() {
		if _, err := jsutil.Call(mp.Value, "start"); err != nil {
			return nil, err
		}
	}

	return events, nil
}
