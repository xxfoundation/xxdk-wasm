////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package state

import (
	"encoding/json"
	"github.com/hack-pad/go-indexeddb/idb"
	"github.com/pkg/errors"
	"gitlab.com/elixxir/wasm-utils/utils"
	"gitlab.com/elixxir/xxdk-wasm/indexedDb/impl"
	"syscall/js"
)

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

func (s *stateModel) Delete(key string) error {
	err := impl.Delete(s.db, stateStoreName, js.ValueOf(key))
	if err != nil {
		return errors.Errorf("Unable to delete key %s: %+v", key, err)
	}
	return nil
}

func (s *stateModel) Keys() ([]byte, error) {
	keys, err := impl.GetAllKeys(s.db, stateStoreName)
	if err != nil {
		return nil, errors.Errorf("Unable to get all keys: %+v", err)
	}

	// Convert JS array of keys to Go string slice
	keysLength := keys.Length()
	keysList := make([]string, keysLength)
	for i := 0; i < keysLength; i++ {
		keysList[i] = keys.Index(i).String()
	}

	// Return JSON-encoded list of keys
	keysJSON, err := json.Marshal(keysList)
	if err != nil {
		return nil, errors.Errorf("Unable to marshal keys: %+v", err)
	}

	return keysJSON, nil
}
