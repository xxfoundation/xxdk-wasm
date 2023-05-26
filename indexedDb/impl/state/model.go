////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package main

const (
	// Text representation of primary key value (keyPath).
	pkeyName = "id"

	// Text representation of the names of the various [idb.ObjectStore].
	stateStoreName = "states"
)

// State defines the IndexedDb representation of a single KV data store.
type State struct {
	// Id is a unique identifier for a given State.
	Id string `json:"id"` // Matches pkeyName

	// Value stores the data contents of the State.
	Value []byte `json:"value"`
}
