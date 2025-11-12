////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package dm

import (
	"encoding/json"

	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"

	"gitlab.com/elixxir/xxdk-wasm/worker"
)

type wasmModel struct {
	wh *worker.Manager
}

// TransferMessage is JSON marshalled and sent to the worker.
type TransferMessage struct {
	Key   string `json:"key"`
	Value []byte `json:"value"`
	Error string `json:"error"`
}

func (w *wasmModel) Set(key string, value []byte) error {
	msg := TransferMessage{
		Key:   key,
		Value: value,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return errors.Errorf(
			"Could not JSON marshal payload for TransferMessage: %+v", err)
	}

	response, err := w.wh.SendMessage(SetTag, data)
	if err != nil {
		jww.FATAL.Panicf("Failed to send message to %q: %+v", SetTag, err)
	} else if len(response) > 0 {
		return errors.New(string(response))
	}

	return nil
}

func (w *wasmModel) Get(key string) ([]byte, error) {

	response, err := w.wh.SendMessage(GetTag, []byte(key))
	if err != nil {
		jww.FATAL.Panicf("Failed to send message to %q: %+v", GetTag, err)
	}

	var msg TransferMessage
	if err = json.Unmarshal(response, &msg); err != nil {
		return nil, errors.Errorf(
			"failed to JSON unmarshal %T from worker: %+v", msg, err)
	}

	if len(msg.Error) > 0 {
		return nil, errors.New(msg.Error)
	}

	return msg.Value, nil
}

func (w *wasmModel) Delete(key string) error {
	response, err := w.wh.SendMessage(DeleteTag, []byte(key))
	if err != nil {
		jww.FATAL.Panicf("Failed to send message to %q: %+v", DeleteTag, err)
	} else if len(response) > 0 {
		return errors.New(string(response))
	}

	return nil
}

func (w *wasmModel) Keys() ([]byte, error) {
	response, err := w.wh.SendMessage(KeysTag, nil)
	if err != nil {
		jww.FATAL.Panicf("Failed to send message to %q: %+v", KeysTag, err)
	}

	// Response is JSON array of keys or error message
	// Try to parse as JSON array first
	var keys []string
	if err = json.Unmarshal(response, &keys); err != nil {
		// If not valid JSON, treat as error message
		return nil, errors.New(string(response))
	}

	return response, nil
}
