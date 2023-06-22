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

	"gitlab.com/elixxir/xxdk-wasm/indexedDb/impl"
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

// WebState defines an interface for setting persistent state in a KV format
// specifically for web-based implementations.
type WebState interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte) error
}

// NewState returns a [utility.WebState] backed by indexeddb.
// The name should be a base64 encoding of the users public key.
func NewState(path, wasmJsPath string) (impl.WebState, error) {
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

	response, err := wh.SendMessage(NewStateTag, payload)
	if err != nil {
		jww.FATAL.Panicf("Failed to send message to %q: %+v", NewStateTag, err)
	} else if len(response) > 0 {
		return nil, errors.New(string(response))
	}

	return &wasmModel{wh}, nil
}
