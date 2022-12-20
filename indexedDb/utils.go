////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

// This file contains several generic IndexedDB helper functions that
// may be useful for any IndexedDB implementations.

package indexedDb

import (
	"context"
	"github.com/hack-pad/go-indexeddb/idb"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/xxdk-wasm/utils"
	"syscall/js"
	"time"
)

const (
	// dbTimeout is the global timeout for operations with the storage
	// [context.Context].
	dbTimeout = time.Second
	// ErrDoesNotExist is an error string for got undefined on Get operations.
	ErrDoesNotExist = "result is undefined"
)

// NewContext builds a context for indexedDb operations.
func NewContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), dbTimeout)
}

// Get is a generic helper for getting values from the given [idb.ObjectStore].
func Get(db *idb.Database, objectStoreName string, key js.Value) (js.Value, error) {
	parentErr := errors.Errorf("failed to Get %s/%s", objectStoreName, key)

	// Prepare the Transaction
	txn, err := db.Transaction(idb.TransactionReadOnly, objectStoreName)
	if err != nil {
		return js.Undefined(), errors.WithMessagef(parentErr,
			"Unable to create Transaction: %+v", err)
	}
	store, err := txn.ObjectStore(objectStoreName)
	if err != nil {
		return js.Undefined(), errors.WithMessagef(parentErr,
			"Unable to get ObjectStore: %+v", err)
	}

	// Perform the operation
	getRequest, err := store.Get(key)
	if err != nil {
		return js.Undefined(), errors.WithMessagef(parentErr,
			"Unable to Get from ObjectStore: %+v", err)
	}

	// Wait for the operation to return
	ctx, cancel := NewContext()
	resultObj, err := getRequest.Await(ctx)
	cancel()
	if err != nil {
		return js.Undefined(), errors.WithMessagef(parentErr,
			"Unable to get from ObjectStore: %+v", err)
	} else if resultObj.IsUndefined() {
		return js.Undefined(), errors.WithMessagef(parentErr,
			"Unable to get from ObjectStore: %s", ErrDoesNotExist)
	}

	// Process result into string
	jww.DEBUG.Printf("Got from %s/%s: %s",
		objectStoreName, key, utils.JsToJson(resultObj))
	return resultObj, nil
}

// GetIndex is a generic helper for getting values from the given
// [idb.ObjectStore] using the given [idb.Index].
func GetIndex(db *idb.Database, objectStoreName string,
	indexName string, key js.Value) (js.Value, error) {
	parentErr := errors.Errorf("failed to GetIndex %s/%s/%s",
		objectStoreName, indexName, key)

	// Prepare the Transaction
	txn, err := db.Transaction(idb.TransactionReadOnly, objectStoreName)
	if err != nil {
		return js.Undefined(), errors.WithMessagef(parentErr,
			"Unable to create Transaction: %+v", err)
	}
	store, err := txn.ObjectStore(objectStoreName)
	if err != nil {
		return js.Undefined(), errors.WithMessagef(parentErr,
			"Unable to get ObjectStore: %+v", err)
	}
	idx, err := store.Index(indexName)
	if err != nil {
		return js.Undefined(), errors.WithMessagef(parentErr,
			"Unable to get Index: %+v", err)
	}

	// Perform the operation
	getRequest, err := idx.Get(key)
	if err != nil {
		return js.Undefined(), errors.WithMessagef(parentErr,
			"Unable to Get from ObjectStore: %+v", err)
	}

	// Wait for the operation to return
	ctx, cancel := NewContext()
	resultObj, err := getRequest.Await(ctx)
	cancel()
	if err != nil {
		return js.Undefined(), errors.WithMessagef(parentErr,
			"Unable to get from ObjectStore: %+v", err)
	} else if resultObj.IsUndefined() {
		return js.Undefined(), errors.WithMessagef(parentErr,
			"Unable to get from ObjectStore: %s", ErrDoesNotExist)
	}

	// Process result into string
	jww.DEBUG.Printf("Got from %s/%s/%s: %s",
		objectStoreName, indexName, key, utils.JsToJson(resultObj))
	return resultObj, nil
}

// Put is a generic helper for putting values into the given [idb.ObjectStore].
// Equivalent to insert if not exists else update.
func Put(db *idb.Database, objectStoreName string, value js.Value) (*idb.Request, error) {
	// Prepare the Transaction
	txn, err := db.Transaction(idb.TransactionReadWrite, objectStoreName)
	if err != nil {
		return nil, errors.Errorf("Unable to create Transaction: %+v", err)
	}
	store, err := txn.ObjectStore(objectStoreName)
	if err != nil {
		return nil, errors.Errorf("Unable to get ObjectStore: %+v", err)
	}

	// Perform the operation
	request, err := store.Put(value)
	if err != nil {
		return nil, errors.Errorf("Unable to Put: %+v", err)
	}

	// Wait for the operation to return
	ctx, cancel := NewContext()
	err = txn.Await(ctx)
	cancel()
	if err != nil {
		return nil, errors.Errorf("Putting value failed: %+v", err)
	}
	jww.DEBUG.Printf("Successfully put value in %s: %v",
		objectStoreName, utils.JsToJson(value))
	return request, nil
}

// Delete is a generic helper for removing values from the given [idb.ObjectStore].
func Delete(db *idb.Database, objectStoreName string, key js.Value) error {
	parentErr := errors.Errorf("failed to Delete %s/%s", objectStoreName, key)

	// Prepare the Transaction
	txn, err := db.Transaction(idb.TransactionReadOnly, objectStoreName)
	if err != nil {
		return errors.WithMessagef(parentErr,
			"Unable to create Transaction: %+v", err)
	}
	store, err := txn.ObjectStore(objectStoreName)
	if err != nil {
		return errors.WithMessagef(parentErr,
			"Unable to get ObjectStore: %+v", err)
	}

	// Perform the operation
	deleteRequest, err := store.Delete(key)
	if err != nil {
		return errors.WithMessagef(parentErr,
			"Unable to Get from ObjectStore: %+v", err)
	}

	// Wait for the operation to return
	ctx, cancel := NewContext()
	err = deleteRequest.Await(ctx)
	cancel()
	if err != nil {
		return errors.WithMessagef(parentErr,
			"Unable to delete from ObjectStore: %+v", err)
	}
	return nil
}

// Dump returns the given [idb.ObjectStore] contents to string slice for
// testing and debugging purposes.
func Dump(db *idb.Database, objectStoreName string) ([]string, error) {
	parentErr := errors.Errorf("failed to Dump %s", objectStoreName)

	txn, err := db.Transaction(idb.TransactionReadOnly, objectStoreName)
	if err != nil {
		return nil, errors.WithMessagef(parentErr,
			"Unable to create Transaction: %+v", err)
	}
	store, err := txn.ObjectStore(objectStoreName)
	if err != nil {
		return nil, errors.WithMessagef(parentErr,
			"Unable to get ObjectStore: %+v", err)
	}
	cursorRequest, err := store.OpenCursor(idb.CursorNext)
	if err != nil {
		return nil, errors.WithMessagef(parentErr,
			"Unable to open Cursor: %+v", err)
	}

	// Run the query
	jww.DEBUG.Printf("%s values:", objectStoreName)
	results := make([]string, 0)
	ctx, cancel := NewContext()
	err = cursorRequest.Iter(ctx,
		func(cursor *idb.CursorWithValue) error {
			value, err := cursor.Value()
			if err != nil {
				return err
			}
			valueStr := utils.JsToJson(value)
			results = append(results, valueStr)
			jww.DEBUG.Printf("- %v", valueStr)
			return nil
		})
	cancel()
	if err != nil {
		return nil, errors.WithMessagef(parentErr,
			"Unable to dump ObjectStore: %+v", err)
	}
	return results, nil
}
