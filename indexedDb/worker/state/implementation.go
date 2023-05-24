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
	"time"

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

	resultChan := make(chan []byte)
	w.wh.SendMessage(SetTag, data,
		func(data []byte) {
			resultChan <- data
		})

	select {
	case result := <-resultChan:
		return errors.New(string(result))
	case <-time.After(worker.ResponseTimeout):
		return errors.Errorf("Timed out after %s waiting for response from the "+
			"worker about Get", worker.ResponseTimeout)
	}
}

func (w *wasmModel) Get(key string) ([]byte, error) {
	resultChan := make(chan []byte)
	w.wh.SendMessage(GetTag, []byte(key),
		func(data []byte) {
			resultChan <- data
		})

	select {
	case result := <-resultChan:
		var msg TransferMessage
		err := json.Unmarshal(result, &msg)
		if err != nil {
			return nil, errors.Errorf(
				"failed to JSON unmarshal %T from main thread: %+v", msg, err)
		}

		if len(msg.Error) > 0 {
			return nil, errors.New(msg.Error)
		}
		return msg.Value, nil
	case <-time.After(worker.ResponseTimeout):
		return nil, errors.Errorf("Timed out after %s waiting for response from the "+
			"worker about Get", worker.ResponseTimeout)
	}
}
