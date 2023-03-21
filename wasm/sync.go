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

	"gitlab.com/elixxir/client/v4/bindings"
	"gitlab.com/elixxir/xxdk-wasm/utils"
)

// TODO: add tests

////////////////////////////////////////////////////////////////////////////////
// Remote Storage Interface and Implementation(s)                             //
////////////////////////////////////////////////////////////////////////////////

// RemoteStoreFileSystem wraps the [bindings.RemoteStoreFileSystem] object so
// its methods can be wrapped to be Javascript compatible.
type RemoteStoreFileSystem struct {
	api *bindings.RemoteStoreFileSystem
}

// newRemoteStoreFileSystemJS creates a new Javascript compatible object
// (map[string]any) that matches the [RemoteStoreFileSystem] structure.
func newRemoteStoreFileSystemJS(api *bindings.RemoteStoreFileSystem) map[string]any {
	rsf := RemoteStoreFileSystem{api}
	rsfMap := map[string]any{
		"Read":            js.FuncOf(rsf.Read),
		"Write":           js.FuncOf(rsf.Write),
		"GetLastModified": js.FuncOf(rsf.GetLastModified),
		"GetLastWrite":    js.FuncOf(rsf.GetLastWrite),
	}

	return rsfMap
}

// NewFileSystemRemoteStorage is a constructor for [RemoteStoreFileSystem].
//
// Parameters:
//   - args[0] - The base directory that all file operations will be performed.
//     It must contain a file delimiter (i.e., `/`) (string).
//
// Returns:
//   - A Javascript representation of the [RemoteStoreFileSystem] object.
func NewFileSystemRemoteStorage(_ js.Value, args []js.Value) any {
	baseDir := args[0].String()
	api := bindings.NewFileSystemRemoteStorage(baseDir)

	return newRemoteStoreFileSystemJS(api)
}

// Read reads from the provided file path and returns the data at that path.
// An error is returned if it failed to read the file.
//
// Parameters:
//   - args[0] - The file path to read from (string).
//
// Returns a promise:
//   - Resolves to the file data (Uint8Array)
//   - Rejected with an error if reading from the file fails.
func (rsf *RemoteStoreFileSystem) Read(_ js.Value, args []js.Value) any {
	path := args[0].String()

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		data, err := rsf.api.Read(path)
		if err != nil {
			reject(utils.JsTrace(err))
		} else {
			resolve(utils.CopyBytesToJS(data))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// Write writes to the file path the provided data. An error is returned if it
// fails to write to file.
//
// Parameters:
//   - args[0] - The file path to write to (string).
//   - args[1] - The file data to write (Uint8Array).
//
// Returns a promise:
//   - Resolves on success (void).
//   - Rejected with an error if writing to the file fails.
func (rsf *RemoteStoreFileSystem) Write(_ js.Value, args []js.Value) any {
	path := args[0].String()
	data := utils.CopyBytesToGo(args[1])

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		err := rsf.api.Write(path, data)
		if err != nil {
			reject(utils.JsTrace(err))
		} else {
			resolve()
		}
	}

	return utils.CreatePromise(promiseFn)
}

// GetLastModified returns when the file at the given file path was last
// modified. If the implementation that adheres to this interface does not
// support this, [Write] or [Read] should be implemented to either write a
// separate timestamp file or add a prefix.
//
// Parameters:
//   - args[0] - The file path (string).
//
// Returns a promise:
//   - Resolves to the JSON of [bindings.RemoteStoreReport] (Uint8Array).
//   - Rejected with an error on failure.
func (rsf *RemoteStoreFileSystem) GetLastModified(_ js.Value, args []js.Value) any {
	path := args[0].String()

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		report, err := rsf.api.GetLastModified(path)
		if err != nil {
			reject(utils.JsTrace(err))
		} else {
			resolve(utils.CopyBytesToJS(report))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// GetLastWrite retrieves the most recent successful write operation that was
// received by [RemoteStoreFileSystem].
//
// Returns a promise:
//   - Resolves to the JSON of [bindings.RemoteStoreReport] (Uint8Array).
//   - Rejected with an error on failure.
func (rsf *RemoteStoreFileSystem) GetLastWrite(js.Value, []js.Value) any {
	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		report, err := rsf.api.GetLastWrite()
		if err != nil {
			reject(utils.JsTrace(err))
		} else {
			resolve(utils.CopyBytesToJS(report))
		}
	}

	return utils.CreatePromise(promiseFn)
}

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
		"Write": js.FuncOf(rkv.Write),
		"Read":  js.FuncOf(rkv.Read),
	}

	return rkvMap
}

// NewOrLoadSyncRemoteKV constructs a [RemoteKV].
//
// Parameters:
//   - args[0] - ID of [E2e] object in tracker (int).
//   - args[1] - A Javascript object that implements the functions on
//     [RemoteKVCallbacks]. These will be the callbacks that are called for
//     [bindings.RemoteStore] operations.
//   - args[2] - A [RemoteStoreCallbacks]. This will be a structure the consumer
//     implements. This acts as a wrapper around the remote storage API
//     (e.g., Google Drive's API, DropBox's API, etc.).
//
// Returns a promise:
//   - Resolves to a Javascript representation of the [RemoteKV] object.
//   - Rejected with an error if initialising the remote KV fails.
func NewOrLoadSyncRemoteKV(_ js.Value, args []js.Value) any {
	e2eID := args[0].Int()
	remoteKvCallbacks := newRemoteKVCallbacks(args[1])
	remote := newRemoteStoreCallbacks(args[2])

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		api, err :=
			bindings.NewOrLoadSyncRemoteKV(e2eID, remoteKvCallbacks, remote)
		if err != nil {
			reject(utils.JsTrace(err))
		} else {
			resolve(newRemoteKvJS(api))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// Write writes a transaction to the remote and local store.
//
// Parameters:
//   - args[0] - The key that this data will be written to (i.e., the device
//     name, the channel name, etc.). Certain keys should follow a pattern and
//     contain special characters (see [RemoteKV.GetList] for details) (string).
//   - args[1] - The data that will be stored (i.e., state data) (Uint8Array).
//   - args[2] - A Javascript object that implements the functions on
//     [RemoteKVCallbacks]. This may be nil if you do not care about the network
//     report.
//
// Returns a promise:
//   - Resolves on success (void).
//   - Rejected with an error if writing to the file fails.
func (rkv *RemoteKV) Write(_ js.Value, args []js.Value) any {
	path := args[0].String()
	data := utils.CopyBytesToGo(args[1])
	cb := newRemoteKVCallbacks(args[1])

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		err := rkv.api.Write(path, data, cb)
		if err != nil {
			reject(utils.JsTrace(err))
		} else {
			resolve()
		}
	}

	return utils.CreatePromise(promiseFn)
}

// Read retrieves the data stored in the underlying KV. Returns an error if the
// data at this key cannot be retrieved.
//
// Parameters:
//   - args[0] - The path that this data will be written to (i.e., the device
//     name) (string).
//
// Returns a promise:
//   - Resolves to the file data (Uint8Array)
//   - Rejected with an error if reading from the file fails.
func (rkv *RemoteKV) Read(_ js.Value, args []js.Value) any {
	path := args[0].String()

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		data, err := rkv.api.Read(path)
		if err != nil {
			reject(utils.JsTrace(err))
		} else {
			resolve(utils.CopyBytesToJS(data))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// GetList returns all entries for a path (or key) that contain the name
// parameter from the local store.
//
// For example, assuming the usage of the [sync.LocalStoreKeyDelimiter], if both
// "channels-123" and "channels-abc" are written to [RemoteKV], then
// GetList("channels") will retrieve the data for both channels. All data that
// contains no [sync.LocalStoreKeyDelimiter] can be retrieved using GetList("").
//
// Parameters:
//   - args[0] - Some prefix to a Write operation. If no prefix applies, simply
//     use the empty string. (string).
//
// Returns:
//   - The file data (Uint8Array)
//   - Throws a TypeError if getting the list fails.
func (rkv *RemoteKV) GetList(_ js.Value, args []js.Value) any {
	name := args[0].String()

	data, err := rkv.api.GetList(name)
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return utils.CopyBytesToJS(data)
}

// RemoteStoreCallbacks wraps Javascript callbacks to adhere to the
// [bindings.RemoteStore] interface.
type RemoteStoreCallbacks struct {
	read            func(args ...any) js.Value
	write           func(args ...any) js.Value
	getLastModified func(args ...any) js.Value
	getLastWrite    func(args ...any) js.Value
}

// newRemoteStoreCallbacks maps the functions of the Javascript object matching
// [bindings.RemoteStore] to a RemoteStoreCallbacks.
func newRemoteStoreCallbacks(arg js.Value) *RemoteStoreCallbacks {
	return &RemoteStoreCallbacks{
		read:            utils.WrapCB(arg, "Read"),
		write:           utils.WrapCB(arg, "Write"),
		getLastModified: utils.WrapCB(arg, "GetLastModified"),
		getLastWrite:    utils.WrapCB(arg, "GetLastWrite"),
	}
}

// Read reads from the provided file path and returns the data at that path.
// An error is returned if it failed to read the file.
//
// Parameters:
//   - path - The file path to read from (string).
//
// Returns:
//   - The file data (Uint8Array).
//   - Catches any thrown errors (of type Error) and returns it as an error.
func (rsCB *RemoteStoreCallbacks) Read(path string) ([]byte, error) {

	fn := func() js.Value { return rsCB.read(path) }
	v, err := utils.RunAndCatch(fn)
	if err != nil {
		return nil, err
	}
	return utils.CopyBytesToGo(v), err
}

// Write writes to the file path the provided data. An error is returned if it
// fails to write to file.
//
// Parameters:
//   - path - The file path to write to (string).
//   - data - The file data to write (Uint8Array).
//
// Returns:
//   - Catches any thrown errors (of type Error) and returns it as an error.
func (rsCB *RemoteStoreCallbacks) Write(path string, data []byte) error {
	fn := func() js.Value { return rsCB.write(path, utils.CopyBytesToJS(data)) }
	_, err := utils.RunAndCatch(fn)
	return err
}

// GetLastModified returns when the file at the given file path was last
// modified. If the implementation that adheres to this interface does not
// support this, [Write] or [Read] should be implemented to either write a
// separate timestamp file or add a prefix.
//
// Parameters:
//   - path - The file path (string).
//
// Returns:
//   - JSON of [bindings.RemoteStoreReport] (Uint8Array).
//   - Catches any thrown errors (of type Error) and returns it as an error.
func (rsCB *RemoteStoreCallbacks) GetLastModified(path string) ([]byte, error) {
	fn := func() js.Value { return rsCB.getLastModified(path) }
	v, err := utils.RunAndCatch(fn)
	if err != nil {
		return nil, err
	}
	return utils.CopyBytesToGo(v), err
}

// GetLastWrite retrieves the most recent successful write operation that was
// received by [RemoteStoreFileSystem].
//
// Returns:
//   - JSON of [bindings.RemoteStoreReport] (Uint8Array).
//   - Catches any thrown errors (of type Error) and returns it as an error.
func (rsCB *RemoteStoreCallbacks) GetLastWrite() ([]byte, error) {
	fn := func() js.Value { return rsCB.getLastWrite() }
	v, err := utils.RunAndCatch(fn)
	if err != nil {
		return nil, err
	}
	return utils.CopyBytesToGo(v), err
}

// RemoteKVCallbacks wraps Javascript callbacks to adhere to the
// [bindings.RemoteKVCallbacks] interface.
type RemoteKVCallbacks struct {
	keyUpdated        func(args ...any) js.Value
	remoteStoreResult func(args ...any) js.Value
}

// newRemoteKVCallbacks maps the functions of the Javascript object matching
// [bindings.RemoteKVCallbacks] to a RemoteKVCallbacks.
func newRemoteKVCallbacks(arg js.Value) *RemoteKVCallbacks {
	return &RemoteKVCallbacks{
		keyUpdated:        utils.WrapCB(arg, "KeyUpdated"),
		remoteStoreResult: utils.WrapCB(arg, "RemoteStoreResult"),
	}
}

// KeyUpdated is the callback to be called any time a key is updated by another
// device tracked by the [RemoteKV] store.
//
// Parameters:
//   - key - (string).
//   - oldVal - (Uint8Array).
//   - newVal - (Uint8Array).
//   - updated - (Boolean)
func (rkvCB *RemoteKVCallbacks) KeyUpdated(
	key string, oldVal, newVal []byte, updated bool) {
	rkvCB.keyUpdated(
		key, utils.CopyBytesToJS(oldVal), utils.CopyBytesToJS(newVal), updated)
}

// RemoteStoreResult is called to report network save results after the key has
// been updated locally.
//
// NOTE: Errors originate from the authentication and writing code in regard to
// remote which is handled by the user of this API. As a result, this callback
// provides no information in simple implementations.
//
// Parameters:
//   - remoteStoreReport - JSON of [bindings.RemoteStoreReport] (Uint8Array).
func (rkvCB *RemoteKVCallbacks) RemoteStoreResult(remoteStoreReport []byte) {
	rkvCB.remoteStoreResult(utils.CopyBytesToJS(remoteStoreReport))
}
