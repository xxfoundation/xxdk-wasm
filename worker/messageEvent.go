////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package worker

import (
	"syscall/js"

	"github.com/pkg/errors"

	"gitlab.com/elixxir/xxdk-wasm/jsutil"
)

// MessageEventType indicates what kind of data the MessageEvent contains
type MessageEventType int

const (
	MessageEventTypeBytes MessageEventType = iota
	MessageEventTypePort
	MessageEventTypeUnknown
)

// MessageEvent is received from the channel returned by Listen().
// All data is copied to Go types to avoid js.Value lifecycle issues.
type MessageEvent struct {
	eventType MessageEventType
	// For MessageEventTypeBytes - the raw message bytes (copied from Uint8Array)
	dataBytes []byte
	// For MessageEventTypePort - we need to keep js references for port handling
	portData js.Value
	port     js.Value
	// Error if parsing failed
	err error
}

// EventType returns the type of this event
func (e MessageEvent) EventType() MessageEventType {
	return e.eventType
}

// DataBytes returns the message bytes for MessageEventTypeBytes events
func (e MessageEvent) DataBytes() ([]byte, error) {
	if e.err != nil {
		return nil, e.err
	}
	if e.eventType != MessageEventTypeBytes {
		return nil, errors.New("event is not a bytes message")
	}
	return e.dataBytes, nil
}

// PortData returns the port data for MessageEventTypePort events
func (e MessageEvent) PortData() (js.Value, js.Value, error) {
	if e.err != nil {
		return js.Value{}, js.Value{}, e.err
	}
	if e.eventType != MessageEventTypePort {
		return js.Value{}, js.Value{}, errors.New("event is not a port message")
	}
	return e.port, e.portData, nil
}

// Error returns any error that occurred during parsing
func (e MessageEvent) Error() error {
	return e.err
}

// parseMessageEvent extracts data from a JS MessageEvent and copies it to Go types.
// This MUST be called synchronously while the js.Value is still valid.
func parseMessageEvent(v js.Value) MessageEvent {
	data := v.Get("data")
	if data.IsUndefined() || data.IsNull() {
		return MessageEvent{err: errors.New("data is undefined or null"), eventType: MessageEventTypeUnknown}
	}

	// Check if it's a string - JS workers (like KV Worker) may send string messages
	// Go workers send Uint8Array via PostMessageTransferBytes, but we need to
	// handle strings when receiving from pure JS workers
	if data.Type() == js.TypeString {
		str := data.String()
		return MessageEvent{
			eventType: MessageEventTypeBytes,
			dataBytes: []byte(str),
		}
	}

	// Check if it's a Uint8Array - copy bytes immediately
	if data.Type() == js.TypeObject {
		constructor := data.Get("constructor")
		if constructor.Equal(jsutil.Uint8Array) {
			// Copy bytes to Go immediately while js.Value is valid
			length := data.Get("length").Int()
			bytes := make([]byte, length)
			js.CopyBytesToGo(bytes, data)
			return MessageEvent{
				eventType: MessageEventTypeBytes,
				dataBytes: bytes,
			}
		}

		// Check if it's a port message
		port := data.Get("port")
		if port.Truthy() {
			return MessageEvent{
				eventType: MessageEventTypePort,
				port:      port,
				portData:  data,
			}
		}
	}

	// Unknown type - store error
	dataType := data.Type().String()
	return MessageEvent{
		err:       errors.Errorf("unknown message data type: %s", dataType),
		eventType: MessageEventTypeUnknown,
	}
}
