////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package worker

import (
	"context"
	"syscall/js"

	"github.com/hack-pad/safejs"
	"github.com/pkg/errors"

	"gitlab.com/elixxir/wasm-utils/utils"
)

// MessagePort wraps a Javascript MessagePort object.
//
// Doc: https://developer.mozilla.org/en-US/docs/Web/API/MessagePort
type MessagePort struct {
	safejs.Value
}

// NewMessagePort wraps the given MessagePort.
func NewMessagePort(v safejs.Value) (MessagePort, error) {
	method, err := v.Get("postMessage")
	if err != nil {
		return MessagePort{}, err
	}
	if method.Type() != safejs.TypeFunction {
		return MessagePort{}, errors.New("postMessage is not a function")
	}
	return MessagePort{v}, nil
}

// PostMessage sends a message from the port.
func (mp MessagePort) PostMessage(message any) error {
	_, err := mp.Call("postMessage", message)
	return err
}

// PostMessageTransfer sends a message from the port and transfers ownership of
// objects to other browsing contexts.
func (mp MessagePort) PostMessageTransfer(message any, transfer ...any) error {
	_, err := mp.Call("postMessage", message, transfer)
	return err
}

// PostMessageTransferBytes sends the message bytes from the port via transfer.
func (mp MessagePort) PostMessageTransferBytes(message []byte) error {
	buffer := utils.CopyBytesToJS(message)
	return mp.PostMessageTransfer(buffer, buffer.Get("buffer"))
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
	messageHandler, err := nonBlocking(func(args []safejs.Value) {
		events <- parseMessageEvent(args[0])
	})
	if err != nil {
		return nil, err
	}
	errorHandler, err := nonBlocking(func(args []safejs.Value) {
		events <- MessageEvent{err: js.Error{Value: safejs.Unsafe(args[0])}}
	})
	if err != nil {
		return nil, err
	}
	messageErrorHandler, err := nonBlocking(func(args []safejs.Value) {
		events <- parseMessageEvent(args[0])
	})
	if err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		_, err := mp.Call("removeEventListener", "message", messageHandler)
		if err == nil {
			messageHandler.Release()
		}
		_, err = mp.Call("removeEventListener", "error", errorHandler)
		if err == nil {
			errorHandler.Release()
		}
		_, err = mp.Call("removeEventListener", "messageerror", messageErrorHandler)
		if err == nil {
			messageErrorHandler.Release()
		}
		close(events)
	}()
	_, err = mp.Call("addEventListener", "message", messageHandler)
	if err != nil {
		return nil, err
	}
	_, err = mp.Call("addEventListener", "error", errorHandler)
	if err != nil {
		return nil, err
	}
	_, err = mp.Call("addEventListener", "messageerror", messageErrorHandler)
	if err != nil {
		return nil, err
	}
	if start, err := mp.Get("start"); err == nil {
		if truthy, err := start.Truthy(); err == nil && truthy {
			if _, err := mp.Call("start"); err != nil {
				return nil, err
			}
		}
	}
	return events, nil
}

func nonBlocking(fn func(args []safejs.Value)) (safejs.Func, error) {
	return safejs.FuncOf(func(_ safejs.Value, args []safejs.Value) any {
		go fn(args)
		return nil
	})
}
