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
	"time"

	"github.com/pkg/errors"

	"gitlab.com/elixxir/client/v4/storage/utility"
	"gitlab.com/elixxir/xxdk-wasm/storage"
	"gitlab.com/elixxir/xxdk-wasm/worker"
)

// databaseSuffix is the suffix to be appended to the name of the database.
const databaseSuffix = "_speakeasy_state"

// NewStateMessage is JSON marshalled and sent to the worker for
// [NewState].
type NewStateMessage struct {
	DatabaseName string `json:"databaseName"`
}

// NewState returns a [utility.WebState] backed by indexeddb.
// The name should be a base64 encoding of the users public key.
func NewState(path, wasmJsPath string) (utility.WebState, error) {
	databaseName := path + databaseSuffix

	wh, err := worker.NewManager(wasmJsPath, "stateIndexedDb", true)
	if err != nil {
		return nil, err
	}

	// Store the database name
	err = storage.StoreIndexedDb(databaseName)
	if err != nil {
		return nil, err
	}

	msg := NewStateMessage{
		DatabaseName: databaseName,
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	dataChan := make(chan []byte)
	wh.SendMessage(NewStateTag, payload,
		func(data []byte) { dataChan <- data })

	select {
	case data := <-dataChan:
		if len(data) > 0 {
			return nil, errors.New(string(data))
		}
	case <-time.After(worker.ResponseTimeout):
		return nil, errors.Errorf("timed out after %s waiting for indexedDB "+
			"database in worker to initialize", worker.ResponseTimeout)
	}

	return &wasmModel{wh}, nil
}
