////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"syscall/js"

	"gitlab.com/elixxir/wasm-utils/exception"

	"gitlab.com/elixxir/client/v4/bindings"
	"gitlab.com/elixxir/wasm-utils/utils"
)

////////////////////////////////////////////////////////////////////////////////
// RemoteKV Methods                                                           //
////////////////////////////////////////////////////////////////////////////////

// RemoteKV wraps the [bindings.RemoteKV] object so its methods can be wrapped
// to be Javascript compatible.
type RemoteKV struct {
	api *bindings.RemoteKV
}

// newRemoteKvJS creates a new Javascript compatible object (map[string]any)
// that matches the [RemoteKV] structure.
func newRemoteKvJS(api *bindings.RemoteKV) map[string]any {
	rkv := RemoteKV{api}
	rkvMap := map[string]any{
		"Get":               js.FuncOf(rkv.Get),
		"Delete":            js.FuncOf(rkv.Delete),
		"Set":               js.FuncOf(rkv.Set),
		"GetPrefix":         js.FuncOf(rkv.GetPrefix),
		"HasPrefix":         js.FuncOf(rkv.HasPrefix),
		"Prefix":            js.FuncOf(rkv.Prefix),
		"Root":              js.FuncOf(rkv.Root),
		"IsMemStore":        js.FuncOf(rkv.IsMemStore),
		"GetFullKey":        js.FuncOf(rkv.GetFullKey),
		"Transaction":       js.FuncOf(rkv.Transaction),
		"StoreMapElement":   js.FuncOf(rkv.StoreMapElement),
		"StoreMap":          js.FuncOf(rkv.StoreMap),
		"DeleteMapElement":  js.FuncOf(rkv.DeleteMapElement),
		"GetMap":            js.FuncOf(rkv.GetMap),
		"GetMapElement":     js.FuncOf(rkv.GetMapElement),
		"ListenOnRemoteKey": js.FuncOf(rkv.ListenOnRemoteKey),
		"ListenOnRemoteMap": js.FuncOf(rkv.ListenOnRemoteMap),
	}

	return rkvMap
}

// Get returns the object stored at the specified version.
// returns a json of [versioned.Object].
//
// Parameters:
//   - args[0] - key to access, a string
//   - args[1] - version, an integer
//
// Returns a promise:
//   - Resolves to JSON of a [versioned.Object], e.g.:
//     {"Version":1,"Timestamp":"2023-05-13T00:50:03.889192694Z","Data":"bm90IHVwZ3JhZGVk"}
//   - Rejected with an access error. Note: File does not exist errors
//     are returned whent key is not set.
func (r *RemoteKV) Get(_ js.Value, args []js.Value) any {
	key := args[0].String()
	version := int64(args[1].Int())

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		value, err := r.api.Get(key, version)
		if err != nil {
			reject(exception.NewTrace(err))
		} else {
			resolve(utils.CopyBytesToJS(value))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// Delete removes a given key from the data store.
//
// Parameters:
//   - args[0] - key to access, a string
//   - args[1] - version, an integer
//
// Returns a promise:
//   - Rejected with an access error. Note: File does not exist errors
//     are returned whent key is not set.
func (r *RemoteKV) Delete(_ js.Value, args []js.Value) any {
	key := args[0].String()
	version := int64(args[1].Int())

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		err := r.api.Delete(key, version)
		if err != nil {
			reject(exception.NewTrace(err))
		} else {
			resolve()
		}
	}

	return utils.CreatePromise(promiseFn)
}

// Set upserts new data into the storage
// When calling this, you are responsible for prefixing the
// key with the correct type optionally unique id! Call
// GetFullKey() to do so.
// The [Object] should contain the versioning if you are
// maintaining such a functionality.
//
// Parameters:
//   - args[0] - the key string
//   - args[1] - the [versioned.Object] JSON value, e.g.:
//     {"Version":1,"Timestamp":"2023-05-13T00:50:03.889192694Z",
//     "Data":"bm90IHVwZ3JhZGVk"}
//
// Returns a promise:
//   - Rejected with an access error.
func (r *RemoteKV) Set(_ js.Value, args []js.Value) any {
	key := args[0].String()
	value := utils.CopyBytesToGo(args[1])

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		err := r.api.Set(key, value)
		if err != nil {
			reject(exception.NewTrace(err))
		} else {
			resolve()
		}
	}

	return utils.CreatePromise(promiseFn)
}

// GetPrefix returns the full Prefix of the KV
// Returns a string via a Promise
func (r *RemoteKV) GetPrefix(_ js.Value, args []js.Value) any {
	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		prefix := r.api.GetPrefix()
		resolve(prefix)
	}

	return utils.CreatePromise(promiseFn)
}

// HasPrefix returns whether this prefix exists in the KV
//
// Parameters:
//   - args[0] - the prefix string to check for.
//
// Returns a bool via a promise.
func (r *RemoteKV) HasPrefix(_ js.Value, args []js.Value) any {
	prefix := args[0].String()
	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		resolve(r.api.HasPrefix(prefix))
	}

	return utils.CreatePromise(promiseFn)
}

// Prefix returns a new KV with the new prefix appending
//
// Parameters:
//   - args[0] - the prefix to append to the list of prefixes
//
// Returns a promise:
//   - Resolves to a new RemoteKV
//   - Rejected with an error.
func (r *RemoteKV) Prefix(_ js.Value, args []js.Value) any {
	prefix := args[0].String()

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		newAPI, err := r.api.Prefix(prefix)

		if err != nil {
			reject(exception.NewTrace(err))
		} else {
			resolve(newRemoteKvJS(newAPI))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// Root returns the KV with no prefixes
func (r *RemoteKV) Root(_ js.Value, args []js.Value) any {
	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		newAPI, err := r.api.Root()

		if err != nil {
			reject(exception.NewTrace(err))
		} else {
			resolve(newRemoteKvJS(newAPI))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// IsMemStore returns true if the underlying KV is memory based
func (r *RemoteKV) IsMemStore(_ js.Value, args []js.Value) any {
	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		resolve(r.api.IsMemStore())
	}

	return utils.CreatePromise(promiseFn)
}

// GetFullKey returns the key with all prefixes appended
func (r *RemoteKV) GetFullKey(_ js.Value, args []js.Value) any {
	key := args[0].String()
	version := int64(args[1].Int())

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		fullKey := r.api.GetFullKey(key, version)
		resolve(fullKey)
	}

	return utils.CreatePromise(promiseFn)
}

// Transaction locks a key while it is being mutated then stores the result
// and returns the old value and if it existed in a JSON object.
// Transactions cannot be remote operations
// If the op returns an error, the operation will be aborted.
func (r *RemoteKV) Transaction(_ js.Value, args []js.Value) any {
	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		reject("unimplemented")
	}

	return utils.CreatePromise(promiseFn)
}

// StoreMapElement stores a versioned map element into the KV. This relies
// on the underlying remote [KV.StoreMapElement] function to lock and control
// updates, but it uses [versioned.Object] values.
// All Map storage functions update the remote.
// valueJSON is a json of a versioned.Object
//
// Parameters:
//   - args[0] - the mapName string
//   - args[1] - the elementKey string
//   - args[2] - the [versioned.Object] JSON value, e.g.:
//     {"Version":1,"Timestamp":"2023-05-13T00:50:03.889192694Z",
//     "Data":"bm90IHVwZ3JhZGVk"}
//   - args[3] - the version int
//
// Returns a promise with an error if any
func (r *RemoteKV) StoreMapElement(_ js.Value, args []js.Value) any {
	mapName := args[0].String()
	elementKey := args[1].String()
	val := utils.CopyBytesToGo(args[2])
	version := int64(args[3].Int())

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		err := r.api.StoreMapElement(mapName, elementKey, val, version)
		if err != nil {
			reject(exception.NewTrace(err))
		} else {
			resolve()
		}
	}

	return utils.CreatePromise(promiseFn)
}

// StoreMap saves a versioned map element into the KV. This relies
// on the underlying remote [KV.StoreMap] function to lock and control
// updates, but it uses [versioned.Object] values.
// All Map storage functions update the remote.
// valueJSON is a json of map[string]*versioned.Object
//
// Parameters:
//   - args[0] - the mapName string
//   - args[1] - the [map[string]versioned.Object] JSON value, e.g.:
//     {"elementKey": {"Version":1,"Timestamp":"2023-05-13T00:50:03.889192694Z",
//     "Data":"bm90IHVwZ3JhZGVk"}}
//   - args[2] - the version int
//
// Returns a promise with an error if any
func (r *RemoteKV) StoreMap(_ js.Value, args []js.Value) any {
	mapName := args[0].String()
	val := utils.CopyBytesToGo(args[1])
	version := int64(args[2].Int())

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		err := r.api.StoreMap(mapName, val, version)
		if err != nil {
			reject(exception.NewTrace(err))
		} else {
			resolve()
		}
	}

	return utils.CreatePromise(promiseFn)
}

// DeleteMapElement removes a versioned map element from the KV.
//
// Parameters:
//   - args[0] - the mapName string
//   - args[1] - the elementKey string
//   - args[2] - the version int
//
// Returns a promise with an error if any or the json of the deleted
// [versioned.Object], e.g.:
//
//	{"Version":1,"Timestamp":"2023-05-13T00:50:03.889192694Z",
//	"Data":"bm90IHVwZ3JhZGVk"}
func (r *RemoteKV) DeleteMapElement(_ js.Value, args []js.Value) any {
	mapName := args[0].String()
	elementKey := args[1].String()
	version := int64(args[2].Int())

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		deleted, err := r.api.DeleteMapElement(mapName, elementKey,
			version)
		if err != nil {
			reject(exception.NewTrace(err))
		} else {
			resolve(utils.CopyBytesToJS(deleted))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// GetMap loads a versioned map from the KV. This relies
// on the underlying remote [KV.GetMap] function to lock and control
// updates, but it uses [versioned.Object] values.
//
// Parameters:
//   - args[0] - the mapName string
//   - args[1] - the version int
//
// Returns a promise with an error if any or the
// the [map[string]versioned.Object] JSON value, e.g.:
//
//	{"elementKey": {"Version":1,"Timestamp":"2023-05-13T00:50:03.889192694Z",
//	"Data":"bm90IHVwZ3JhZGVk"}}
func (r *RemoteKV) GetMap(_ js.Value, args []js.Value) any {
	mapName := args[0].String()
	version := int64(args[1].Int())

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		mapJSON, err := r.api.GetMap(mapName, version)
		if err != nil {
			reject(exception.NewTrace(err))
		} else {
			resolve(utils.CopyBytesToJS(mapJSON))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// GetMapElement loads a versioned map element from the KV. This relies
// on the underlying remote [KV.GetMapElement] function to lock and control
// updates, but it uses [versioned.Object] values.
// Parameters:
//   - args[0] - the mapName string
//   - args[1] - the elementKey string
//   - args[2] - the version int
//
// Returns a promise with an error if any or the json of the
// [versioned.Object], e.g.:
//
//	{"Version":1,"Timestamp":"2023-05-13T00:50:03.889192694Z",
//	"Data":"bm90IHVwZ3JhZGVk"}
func (r *RemoteKV) GetMapElement(_ js.Value, args []js.Value) any {
	mapName := args[0].String()
	elementKey := args[1].String()
	version := int64(args[2].Int())

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		deleted, err := r.api.GetMapElement(mapName, elementKey,
			version)
		if err != nil {
			reject(exception.NewTrace(err))
		} else {
			resolve(utils.CopyBytesToJS(deleted))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// ListenOnRemoteKey sets up a callback listener for the object specified
// by the key and version. It returns the current [versioned.Object] JSON
// of the value.
// Parameters:
//   - args[0] - the key string
//   - args[1] - the version int
//   - args[2] - the [KeyChangedByRemoteCallback] javascript callback
//
// Returns a promise with an error if any or the json of the existing
// [versioned.Object], e.g.:
//
//	{"Version":1,"Timestamp":"2023-05-13T00:50:03.889192694Z",
//	"Data":"bm90IHVwZ3JhZGVk"}
func (r *RemoteKV) ListenOnRemoteKey(_ js.Value, args []js.Value) any {
	key := args[0].String()
	version := int64(args[1].Int())
	cb := newKeyChangedByRemoteCallback(args[2])

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		deleted, err := r.api.ListenOnRemoteKey(key, version, cb)
		if err != nil {
			reject(exception.NewTrace(err))
		} else {
			resolve(utils.CopyBytesToJS(deleted))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// ListenOnRemoteMap allows the caller to receive updates when
// the map or map elements are updated. Returns a JSON of
// map[string]versioned.Object of the current map value.
// Parameters:
//   - args[0] - the mapName string
//   - args[1] - the version int
//   - args[2] - the [MapChangedByRemoteCallback] javascript callback
//
// Returns a promise with an error if any or the json of the existing
// the [map[string]versioned.Object] JSON value, e.g.:
//
//	{"elementKey": {"Version":1,"Timestamp":"2023-05-13T00:50:03.889192694Z",
//	"Data":"bm90IHVwZ3JhZGVk"}}
func (r *RemoteKV) ListenOnRemoteMap(_ js.Value, args []js.Value) any {
	mapName := args[0].String()
	version := int64(args[1].Int())
	cb := newMapChangedByRemoteCallback(args[2])

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		deleted, err := r.api.ListenOnRemoteMap(mapName, version, cb)
		if err != nil {
			reject(exception.NewTrace(err))
		} else {
			resolve(utils.CopyBytesToJS(deleted))
		}
	}

	return utils.CreatePromise(promiseFn)
}

////////////////////////////////////////////////////////////////////////////////
// RemoteStore                                                                //
////////////////////////////////////////////////////////////////////////////////

// RemoteStore wraps Javascript callbacks to adhere to the
// [bindings.RemoteStore] interface.
type RemoteStore struct {
	read            func(args ...any) js.Value
	write           func(args ...any) js.Value
	getLastModified func(args ...any) js.Value
	getLastWrite    func(args ...any) js.Value
	readDir         func(args ...any) js.Value
}

// newRemoteStoreCallbacks maps the functions of the Javascript object matching
// [bindings.RemoteStore] to a RemoteStoreCallbacks.
func newRemoteStore(arg js.Value) *RemoteStore {
	return &RemoteStore{
		read:            utils.WrapCB(arg, "Read"),
		write:           utils.WrapCB(arg, "Write"),
		getLastModified: utils.WrapCB(arg, "GetLastModified"),
		getLastWrite:    utils.WrapCB(arg, "GetLastWrite"),
		readDir:         utils.WrapCB(arg, "ReadDir"),
	}
}

// Read impelements [bindings.RemoteStore.Read]
//
// Parameters:
//   - path - The file path to read from (string).
//
// Returns:
//   - The file data (Uint8Array).
//   - Catches any thrown errors (of type Error) and returns it as an error.
func (rsCB *RemoteStore) Read(path string) ([]byte, error) {

	fn := func() js.Value { return rsCB.read(path) }
	v, err := exception.RunAndCatch(fn)
	if err != nil {
		return nil, err
	}
	return utils.CopyBytesToGo(v), err
}

// Write implements [bindings.RemoteStore.Write]
//
// Parameters:
//   - path - The file path to write to (string).
//   - data - The file data to write (Uint8Array).
//
// Returns:
//   - Catches any thrown errors (of type Error) and returns it as an error.
func (rsCB *RemoteStore) Write(path string, data []byte) error {
	fn := func() js.Value { return rsCB.write(path, utils.CopyBytesToJS(data)) }
	_, err := exception.RunAndCatch(fn)
	return err
}

// GetLastModified implements [bindings.RemoteStore.GetLastModified]
//
// Parameters:
//   - path - The file path (string).
//
// Returns:
//   - JSON of [bindings.RemoteStoreReport] (Uint8Array).
//   - Catches any thrown errors (of type Error) and returns it as an error.
func (rsCB *RemoteStore) GetLastModified(path string) ([]byte, error) {
	fn := func() js.Value { return rsCB.getLastModified(path) }
	v, err := exception.RunAndCatch(fn)
	if err != nil {
		return nil, err
	}
	return utils.CopyBytesToGo(v), err
}

// GetLastWrite implements [bindings.RemoteStore.GetLastWrite()
//
// Returns:
//   - JSON of [bindings.RemoteStoreReport] (Uint8Array).
//   - Catches any thrown errors (of type Error) and returns it as an error.
func (rsCB *RemoteStore) GetLastWrite() ([]byte, error) {
	fn := func() js.Value { return rsCB.getLastWrite() }
	v, err := exception.RunAndCatch(fn)
	if err != nil {
		return nil, err
	}
	return utils.CopyBytesToGo(v), err
}

// ReadDir implements [bindings.RemoteStore.ReadDir]
//
// Parameters:
//   - path - The file path (string).
//
// Returns:
//   - JSON of []string (Uint8Array).
//   - Catches any thrown errors (of type Error) and returns it as an error.
func (rsCB *RemoteStore) ReadDir(path string) ([]byte, error) {
	fn := func() js.Value { return rsCB.readDir(path) }
	v, err := exception.RunAndCatch(fn)
	if err != nil {
		return nil, err
	}
	return utils.CopyBytesToGo(v), err
}

////////////////////////////////////////////////////////////////////////////////
// Callbacks                                                                  //
////////////////////////////////////////////////////////////////////////////////

// KeyChangedByRemoteCallback wraps the passed javascript function and
// implements [bindings.KeyChangedByRemoteCallback]
type KeyChangedByRemoteCallback struct {
	callback func(args ...any) js.Value
}

func (k *KeyChangedByRemoteCallback) Callback(key string, old, new []byte,
	opType int8) {
	k.callback(key, utils.CopyBytesToJS(old), utils.CopyBytesToJS(new),
		opType)
}

func newKeyChangedByRemoteCallback(
	jsFunc js.Value) *KeyChangedByRemoteCallback {
	return &KeyChangedByRemoteCallback{
		callback: utils.WrapCB(jsFunc, "Callback"),
	}
}

// MapChangedByRemoteCallback wraps the passed javascript function and
// implements [bindings.KeyChangedByRemoteCallback]
type MapChangedByRemoteCallback struct {
	callback func(args ...any) js.Value
}

func (m *MapChangedByRemoteCallback) Callback(mapName string,
	editsJSON []byte) {
	m.callback(mapName, utils.CopyBytesToJS(editsJSON))
}

func newMapChangedByRemoteCallback(
	jsFunc js.Value) *MapChangedByRemoteCallback {
	return &MapChangedByRemoteCallback{
		callback: utils.WrapCB(jsFunc, "Callback"),
	}
}
