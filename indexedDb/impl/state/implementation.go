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
	"github.com/hack-pad/go-indexeddb/idb"
	"github.com/pkg/errors"
	"gitlab.com/elixxir/xxdk-wasm/indexedDb/impl"
	"gitlab.com/elixxir/xxdk-wasm/utils"
	"syscall/js"
)

// ClientState
type ClientState interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte) error
}

// stateModel implements [ClientState] interface backed by IndexedDb.
// NOTE: This model is NOT thread safe - it is the responsibility of the
// caller to ensure that its methods are called sequentially.
type stateModel struct {
	db *idb.Database
}

func (s *stateModel) Get(key string) ([]byte, error) {
	result, err := impl.Get(s.db, stateStoreName, js.ValueOf(key))
	if err != nil {
		return nil, err
	}

	stateObj := &State{}
	err = json.Unmarshal([]byte(utils.JsToJson(result)), stateObj)
	if err != nil {
		return nil, err
	}

	return stateObj.Value, err
}

func (s *stateModel) Set(key string, value []byte) error {
	state := &State{
		Id:    key,
		Value: value,
	}

	// Convert to jsObject
	newStateJSON, err := json.Marshal(state)
	if err != nil {
		return errors.Errorf("Unable to marshal State: %+v", err)
	}
	stateObj, err := utils.JsonToJS(newStateJSON)
	if err != nil {
		return errors.Errorf("Unable to marshal State: %+v", err)
	}

	// Store State to database
	_, err = impl.Put(s.db, stateStoreName, stateObj)
	if err != nil {
		return errors.Errorf("Unable to put State: %+v\n%s",
			err, newStateJSON)
	}
	return nil
}
