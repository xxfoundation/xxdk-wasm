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

	"gitlab.com/elixxir/xxdk-wasm/indexedDb/impl"
	stateWorker "gitlab.com/elixxir/xxdk-wasm/indexedDb/worker/state"
	"gitlab.com/elixxir/xxdk-wasm/worker"
)

// manager handles the message callbacks, which is used to
// send information between the model and the main thread.
type manager struct {
	wtm   *worker.ThreadManager
	model impl.WebState
}

// registerCallbacks registers all the reception callbacks to manage messages
// from the main thread.
func (m *manager) registerCallbacks() {
	m.wtm.RegisterCallback(stateWorker.NewStateTag, m.newStateCB)
	m.wtm.RegisterCallback(stateWorker.SetTag, m.setCB)
	m.wtm.RegisterCallback(stateWorker.GetTag, m.getCB)
	m.wtm.RegisterCallback(stateWorker.DeleteTag, m.deleteCB)
	m.wtm.RegisterCallback(stateWorker.KeysTag, m.keysCB)
}

// newStateCB is the callback for NewState. Returns an empty
// slice on success or an error message on failure.
func (m *manager) newStateCB(message []byte, reply func(message []byte)) {
	var msg stateWorker.NewStateMessage
	err := json.Unmarshal(message, &msg)
	if err != nil {
		reply([]byte(errors.Wrapf(err,
			"failed to JSON unmarshal %T from main thread", msg).Error()))
		return
	}

	m.model, err = NewState(msg.DatabaseName)
	if err != nil {
		reply([]byte(err.Error()))
		return
	}

	reply(nil)
}

// setCB is the callback for stateModel.Set.
// Returns nil on error or the resulting byte data on success.
func (m *manager) setCB(message []byte, reply func(message []byte)) {
	var msg stateWorker.TransferMessage
	err := json.Unmarshal(message, &msg)
	if err != nil {
		reply([]byte(errors.Wrapf(err,
			"failed to JSON unmarshal %T from main thread", msg).Error()))
		return
	}

	err = m.model.Set(msg.Key, msg.Value)
	if err != nil {
		reply([]byte(err.Error()))
		return
	}

	reply(nil)
}

// getCB is the callback for stateModel.Get.
// Returns nil on error or the resulting byte data on success.
func (m *manager) getCB(message []byte, reply func(message []byte)) {
	key := string(message)
	result, err := m.model.Get(key)
	msg := stateWorker.TransferMessage{
		Key:   key,
		Value: result,
		Error: err.Error(),
	}

	replyMessage, err := json.Marshal(msg)
	if err != nil {
		jww.FATAL.Panicf("Could not JSON marshal %T for Get: %+v", msg, err)
	}

	reply(replyMessage)
}

// deleteCB is the callback for stateModel.Delete.
// Returns nil on success or an error message on failure.
func (m *manager) deleteCB(message []byte, reply func(message []byte)) {
	key := string(message)
	err := m.model.Delete(key)
	if err != nil {
		reply([]byte(err.Error()))
		return
	}

	reply(nil)
}

// keysCB is the callback for stateModel.Keys.
// Returns JSON array of keys on success or an error message on failure.
func (m *manager) keysCB(message []byte, reply func(message []byte)) {
	result, err := m.model.Keys()
	if err != nil {
		reply([]byte(err.Error()))
		return
	}

	reply(result)
}
