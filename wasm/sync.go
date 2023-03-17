////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"gitlab.com/elixxir/client/v4/bindings"
	"gitlab.com/elixxir/xxdk-wasm/utils"
	"syscall/js"
)

// TODO: add tests

////////////////////////////////////////////////////////////////////////////////
// Local Storage Interface & Implementation(s)                                //
////////////////////////////////////////////////////////////////////////////////

// LocalStoreEKV wraps the [bindings.LocalStoreEKV] object so its methods can be
// wrapped to be Javascript compatible.
type LocalStoreEKV struct {
	api *bindings.LocalStoreEKV
}

// newLocalStoreEkvJS creates a new Javascript compatible object
// (map[string]any) that matches the [LocalStoreEKV] structure.
func newLocalStoreEkvJS(api *bindings.LocalStoreEKV) map[string]any {
	ls := LocalStoreEKV{api}
	lsMap := map[string]any{
		"Read":  js.FuncOf(ls.Read),
		"Write": js.FuncOf(ls.Write),
	}

	return lsMap
}

// NewEkvLocalStore is a constructor for [LocalStoreEKV].
//
// Parameters:
//   - args[0] - The base directory that all file operations will be performed.
//     It must contain a file delimiter (i.e., `/`) (string).
//   - args[1] - Password (string).
//
// Returns a promise:
//   - Resolves to a Javascript representation of the [LocalStoreEKV] object.
//   - Rejected with an error if initialising the EKV local store fails.
func NewEkvLocalStore(_ js.Value, args []js.Value) any {
	baseDir := args[0].String()
	password := args[1].String()

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		api, err := bindings.NewEkvLocalStore(baseDir, password)
		if err != nil {
			reject(utils.JsTrace(err))
		} else {
			resolve(newLocalStoreEkvJS(api))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// Read reads data from path. This returns an error if it fails to read from the
// file path.
//
// This utilizes [ekv.KeyValue] under the hood.
//
// Parameters:
//   - args[0] - The file path to read from (string).
//
// Returns a promise:
//   - Resolves to the file data (Uint8Array)
//   - Rejected with an error if reading from the file fails.
func (ls *LocalStoreEKV) Read(_ js.Value, args []js.Value) any {
	path := args[0].String()

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		data, err := ls.api.Read(path)
		if err != nil {
			reject(utils.JsTrace(err))
		} else {
			resolve(utils.CopyBytesToJS(data))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// Write writes data to the path. This returns an error if it fails to write.
//
// This utilizes [ekv.KeyValue] under the hood.
//
// Parameters:
//   - args[0] - The file path to write to (string).
//   - args[1] - The file data to write (Uint8Array).
//
// Returns a promise:
//   - Resolves on success (void).
//   - Rejected with an error if writing to the file fails.
func (ls *LocalStoreEKV) Write(_ js.Value, args []js.Value) any {
	path := args[0].String()
	data := utils.CopyBytesToGo(args[1])

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		err := ls.api.Write(path, data)
		if err != nil {
			reject(utils.JsTrace(err))
		} else {
			resolve()
		}
	}

	return utils.CreatePromise(promiseFn)
}

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
//   - args[1] - The path that the state data for this device will be written to
//     locally (e.g. sync/txLog.txt) (string).
//   - args[2] - The key update callback that will be called when the value of
//     a key is modified by another device. A Javascript object that implements
//     the functions in the [bindings.KeyUpdateCallback] interface.
//   - args[3] - A [bindings.RemoteStoreCallback] that will be called to
//     report the results of a write to the remote storage option AFTER a key
//     has been saved locally. This will be used to report any previously called
//     sets that had unsuccessful reports. A Javascript object that implements
//     the functions in the [bindings.RemoteStoreCallback] interface. newTx
//     should be a Uint8Array.
//   - remote - A [RemoteStore]. This should be what the remote storage operation
//     wrapper wrapped should adhere.
//   - local - A [LocalStore]. This should be what a local storage option adheres
//     to.
//
// Returns a promise:
//   - Resolves to a Javascript representation of the [RemoteKV] object.
//   - Rejected with an error if initialising the remote KV fails.
//
// TODO: fix remote and local
func NewOrLoadSyncRemoteKV(_ js.Value, args []js.Value) any {
	e2eID := args[0].Int()
	txLogPath := args[1].String()
	keyUpdateCb := &keyUpdateCallback{utils.WrapCB(args[2], "Callback")}
	remoteStoreCb := &remoteStoreCallback{utils.WrapCB(args[3], "Callback")}

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		api, err := bindings.NewOrLoadSyncRemoteKV(e2eID, txLogPath,
			keyUpdateCb, remoteStoreCb, nil, nil, nil)
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
//     name) (string).
//   - args[1] - The data that will be stored (i.e., state data) (Uint8Array).
//   - args[2] - A Javascript object that implements the functions in the
//     [bindings.RemoteStoreCallback] interface. newTx should be a Uint8Array.
//     This may be nil if you do not care about the network report.
//
// Returns a promise:
//   - Resolves on success (void).
//   - Rejected with an error if writing to the file fails.
func (rkv *RemoteKV) Write(_ js.Value, args []js.Value) any {
	path := args[0].String()
	data := utils.CopyBytesToGo(args[1])
	cb := &remoteStoreCallback{utils.WrapCB(args[2], "Callback")}

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

// remoteStoreCallback wraps Javascript callbacks to adhere to the
// [bindings.RemoteStoreCallback] interface.
type remoteStoreCallback struct {
	callback func(args ...any) js.Value
}

// Callback reports the status of writing the new transaction to remote storage.
//
// Parameters:
//   - newTx - Returns the transaction (Uint8Array).
//   - err - Returns an error on failure (Error).
func (rsCB *remoteStoreCallback) Callback(newTx []byte, err string) {
	rsCB.callback(utils.CopyBytesToJS(newTx), err)
}

// keyUpdateCallback wraps Javascript callbacks to adhere to the
// [bindings.KeyUpdateCallback] interface.
type keyUpdateCallback struct {
	callback func(args ...any) js.Value
}

// Callback reports the event.
//
// Parameters:
//   - newTx - Returns the transaction (Uint8Array).
//   - err - Returns an error on failure (Error).
func (kuCB *keyUpdateCallback) Callback(key, val string) {
	kuCB.callback(key, val)
}
