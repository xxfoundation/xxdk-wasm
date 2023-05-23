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

	"gitlab.com/elixxir/client/v4/storage/utility"
	stateWorker "gitlab.com/elixxir/xxdk-wasm/indexedDb/worker/state"
	"gitlab.com/elixxir/xxdk-wasm/worker"
)

// manager handles the message callbacks, which is used to
// send information between the model and the main thread.
type manager struct {
	wtm   *worker.ThreadManager
	model utility.WebState
}

// registerCallbacks registers all the reception callbacks to manage messages
// from the main thread.
func (m *manager) registerCallbacks() {
	m.wtm.RegisterCallback(stateWorker.NewStateTag, m.newStateCB)
	m.wtm.RegisterCallback(stateWorker.SetTag, m.setCB)
	m.wtm.RegisterCallback(stateWorker.GetTag, m.getCB)
}

// newStateCB is the callback for NewState. Returns an empty
// slice on success or an error message on failure.
func (m *manager) newStateCB(data []byte) ([]byte, error) {
	var msg stateWorker.NewStateMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return []byte{}, errors.Errorf(
			"failed to JSON unmarshal %T from main thread: %+v", msg, err)
	}

	m.model, err = NewState(msg.DatabaseName)
	if err != nil {
		return []byte(err.Error()), nil
	}

	return []byte{}, nil
}

// setCB is the callback for stateModel.Set.
// Returns nil on error or the resulting byte data on success.
func (m *manager) setCB(data []byte) ([]byte, error) {
	var msg stateWorker.TransferMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, errors.Errorf(
			"failed to JSON unmarshal %T from main thread: %+v", msg, err)
	}

	return nil, m.model.Set(msg.Key, msg.Value)
}

// getCB is the callback for stateModel.Get.
// Returns nil on error or the resulting byte data on success.
func (m *manager) getCB(data []byte) ([]byte, error) {
	key := string(data)
	result, err := m.model.Get(key)
	msg := stateWorker.TransferMessage{
		Key:   key,
		Value: result,
		Error: err.Error(),
	}

	return json.Marshal(msg)
}
