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

////////////////////////////////////////////////////////////////////////////////
// File Transfer Structs and Interfaces                                       //
////////////////////////////////////////////////////////////////////////////////

// FileTransfer wraps the [bindings.FileTransfer] object so its methods can be
// wrapped to be Javascript compatible.
type FileTransfer struct {
	api *bindings.FileTransfer
}

// newFileTransferJS creates a new Javascript compatible object (map[string]any)
// that matches the [FileTransfer] structure.
func newFileTransferJS(api *bindings.FileTransfer) map[string]any {
	ft := FileTransfer{api}
	ftMap := map[string]any{
		// Main functions
		"Send":      js.FuncOf(ft.Send),
		"Receive":   js.FuncOf(ft.Receive),
		"CloseSend": js.FuncOf(ft.CloseSend),

		// Callback registration functions
		"RegisterSentProgressCallback": js.FuncOf(
			ft.RegisterSentProgressCallback),
		"RegisterReceivedProgressCallback": js.FuncOf(
			ft.RegisterReceivedProgressCallback),

		// Utility functions
		"MaxFileNameLen": js.FuncOf(ft.MaxFileNameLen),
		"MaxFileTypeLen": js.FuncOf(ft.MaxFileTypeLen),
		"MaxFileSize":    js.FuncOf(ft.MaxFileSize),
		"MaxPreviewSize": js.FuncOf(ft.MaxPreviewSize),
	}

	return ftMap
}

// receiveFileCallback wraps Javascript callbacks to adhere to the
// [bindings.ReceiveFileCallback] interface.
type receiveFileCallback struct {
	callback func(args ...any) js.Value
}

// Callback is called when a new file transfer is received.
//
// Parameters:
//   - payload - Returns the contents of the message. JSON of
//     [bindings.ReceivedFile] (Uint8Array).
func (rfc *receiveFileCallback) Callback(payload []byte) {
	rfc.callback(utils.CopyBytesToJS(payload))
}

// fileTransferSentProgressCallback wraps Javascript callbacks to adhere to the
// [bindings.FileTransferSentProgressCallback] interface.
type fileTransferSentProgressCallback struct {
	callback func(args ...any) js.Value
}

// Callback is called when a new file transfer is received.
//
// Parameters:
//   - payload - Returns the contents of the message. JSON of
//     [bindings.Progress] (Uint8Array).
//   - t - Returns a tracker that allows the lookup of the status of any file
//     part. It is a Javascript object that matches the functions on
//     [FilePartTracker].
//   - err - Returns an error on failure (Error).
func (spc *fileTransferSentProgressCallback) Callback(
	payload []byte, t *bindings.FilePartTracker, err error) {
	spc.callback(utils.CopyBytesToJS(payload), newFilePartTrackerJS(t),
		utils.JsTrace(err))
}

// fileTransferReceiveProgressCallback wraps Javascript callbacks to adhere to
// the [bindings.FileTransferReceiveProgressCallback] interface.
type fileTransferReceiveProgressCallback struct {
	callback func(args ...any) js.Value
}

// Callback is called when a file part is sent or an error occurs.
//
// Parameters:
//   - payload - Returns the contents of the message. JSON of
//     [bindings.Progress] (Uint8Array).
//   - t - Returns a tracker that allows the lookup of the status of any file
//     part. It is a Javascript object that matches the functions on
//     [FilePartTracker].
//   - err - Returns an error on failure (Error).
func (rpc *fileTransferReceiveProgressCallback) Callback(
	payload []byte, t *bindings.FilePartTracker, err error) {
	rpc.callback(utils.CopyBytesToJS(payload), newFilePartTrackerJS(t),
		utils.JsTrace(err))
}

////////////////////////////////////////////////////////////////////////////////
// Main functions                                                             //
////////////////////////////////////////////////////////////////////////////////

// InitFileTransfer creates a bindings-level file transfer manager.
//
// Parameters:
//   - args[0] - ID of [E2e] object in tracker (int).
//   - args[1] - Javascript object that has functions that implement the
//     [bindings.ReceiveFileCallback] interface.
//   - args[2] - JSON of [gitlab.com/elixxir/client/v4/fileTransfer/e2e.Params]
//     (Uint8Array).
//   - args[3] - JSON of [fileTransfer.Params] (Uint8Array).
//
// Returns:
//   - Javascript representation of the [FileTransfer] object.
//   - Throws a TypeError initialising the file transfer manager fails.
func InitFileTransfer(_ js.Value, args []js.Value) any {
	rfc := &receiveFileCallback{utils.WrapCB(args[1], "Callback")}
	e2eFileTransferParamsJson := utils.CopyBytesToGo(args[2])
	fileTransferParamsJson := utils.CopyBytesToGo(args[3])

	api, err := bindings.InitFileTransfer(
		args[0].Int(), rfc, e2eFileTransferParamsJson, fileTransferParamsJson)
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return newFileTransferJS(api)
}

// Send is the bindings-level function for sending a file.
//
// Parameters:
//   - args[0] - JSON of [bindings.FileSend] (Uint8Array).
//   - args[1] - Marshalled bytes of the recipient [id.ID] (Uint8Array).
//   - args[2] - Number of retries allowed (float).
//   - args[3] - Javascript object that has functions that implement the
//     [bindings.FileTransferSentProgressCallback] interface.
//   - args[4] - Duration, in milliseconds, to wait between progress callbacks
//     triggering (int).
//
// Returns a promise:
//   - Resolves to a unique ID for this file transfer (Uint8Array).
//   - Rejected with an error if sending fails.
func (f *FileTransfer) Send(_ js.Value, args []js.Value) any {
	payload := utils.CopyBytesToGo(args[0])
	recipientID := utils.CopyBytesToGo(args[1])
	retry := float32(args[2].Float())
	spc := &fileTransferSentProgressCallback{utils.WrapCB(args[3], "Callback")}

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		ftID, err := f.api.Send(payload, recipientID, retry, spc, args[4].Int())
		if err != nil {
			reject(utils.JsTrace(err))
		} else {
			resolve(utils.CopyBytesToJS(ftID))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// Receive returns the full file on the completion of the transfer. It deletes
// internal references to the data and unregisters any attached progress
// callbacks. Returns an error if the transfer is not complete, the full file
// cannot be verified, or if the transfer cannot be found.
//
// Receive can only be called once the progress callback returns that the
// file transfer is complete.
//
// Parameters:
//   - args[0] - File transfer [fileTransfer.TransferID] (Uint8Array).
//
// Returns:
//   - File contents (Uint8Array).
//   - Throws a TypeError the file transfer is incomplete or Receive has already
//     been called.
func (f *FileTransfer) Receive(_ js.Value, args []js.Value) any {
	file, err := f.api.Receive(utils.CopyBytesToGo(args[0]))
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return utils.CopyBytesToJS(file)
}

// CloseSend deletes a file from the internal storage once a transfer has
// completed or reached the retry limit. Returns an error if the transfer has
// not run out of retries.
//
// This function should be called once a transfer completes or errors out (as
// reported by the progress callback).
//
// Parameters:
//   - args[0] - File transfer [fileTransfer.TransferID] (Uint8Array).
//
// Returns:
//   - Throws a TypeError if the file transfer is incomplete.
func (f *FileTransfer) CloseSend(_ js.Value, args []js.Value) any {
	err := f.api.CloseSend(utils.CopyBytesToGo(args[0]))
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Callback Registration Functions                                            //
////////////////////////////////////////////////////////////////////////////////

// RegisterSentProgressCallback allows for the registration of a callback to
// track the progress of an individual sent file transfer.
//
// SentProgressCallback is auto registered on Send; this function should be
// called when resuming clients or registering extra callbacks.
//
// Parameters:
//   - args[0] - File transfer [fileTransfer.TransferID] (Uint8Array).
//   - args[1] - Javascript object that has functions that implement the
//     [bindings.FileTransferSentProgressCallback] interface.
//   - args[2] - Duration, in milliseconds, to wait between progress callbacks
//     triggering (int).
//
// Returns:
//   - Throws a TypeError if registering the callback fails.
func (f *FileTransfer) RegisterSentProgressCallback(
	_ js.Value, args []js.Value) any {
	tidBytes := utils.CopyBytesToGo(args[0])
	spc := &fileTransferSentProgressCallback{utils.WrapCB(args[1], "Callback")}

	err := f.api.RegisterSentProgressCallback(tidBytes, spc, args[2].Int())
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return nil
}

// RegisterReceivedProgressCallback allows for the registration of a callback to
// track the progress of an individual received file transfer.
//
// This should be done when a new transfer is received on the ReceiveCallback.
//
// Parameters:
//   - args[0] - File transfer [fileTransfer.TransferID] (Uint8Array).
//   - args[1] - Javascript object that has functions that implement the
//     [bindings.FileTransferReceiveProgressCallback] interface.
//   - args[2] - Duration, in milliseconds, to wait between progress callbacks
//     triggering (int).
//
// Returns:
//   - Throws a TypeError if registering the callback fails.
func (f *FileTransfer) RegisterReceivedProgressCallback(
	_ js.Value, args []js.Value) any {
	tidBytes := utils.CopyBytesToGo(args[0])
	rpc := &fileTransferReceiveProgressCallback{utils.WrapCB(args[1], "Callback")}

	err := f.api.RegisterReceivedProgressCallback(
		tidBytes, rpc, args[2].Int())
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Utility Functions                                                          //
////////////////////////////////////////////////////////////////////////////////

// MaxFileNameLen returns the max number of bytes allowed for a file name.
//
// Returns:
//   - Max file name length (int).
func (f *FileTransfer) MaxFileNameLen(js.Value, []js.Value) any {
	return f.api.MaxFileNameLen()
}

// MaxFileTypeLen returns the max number of bytes allowed for a file type.
//
// Returns:
//   - Max file type length (int).
func (f *FileTransfer) MaxFileTypeLen(js.Value, []js.Value) any {
	return f.api.MaxFileTypeLen()
}

// MaxFileSize returns the max number of bytes allowed for a file.
//
// Returns:
//   - Max file size (int).
func (f *FileTransfer) MaxFileSize(js.Value, []js.Value) any {
	return f.api.MaxFileSize()
}

// MaxPreviewSize returns the max number of bytes allowed for a file preview.
//
// Returns:
//   - Max preview size (int).
func (f *FileTransfer) MaxPreviewSize(js.Value, []js.Value) any {
	return f.api.MaxPreviewSize()
}

////////////////////////////////////////////////////////////////////////////////
// File Part Tracker                                                          //
////////////////////////////////////////////////////////////////////////////////

// FilePartTracker wraps the [bindings.FilePartTracker] object so its methods
// can be wrapped to be Javascript compatible.
type FilePartTracker struct {
	api *bindings.FilePartTracker
}

// newFilePartTrackerJS creates a new Javascript compatible object
// (map[string]any) that matches the [FilePartTracker] structure.
func newFilePartTrackerJS(api *bindings.FilePartTracker) map[string]any {
	fpt := FilePartTracker{api}
	ftMap := map[string]any{
		"GetPartStatus": js.FuncOf(fpt.GetPartStatus),
		"GetNumParts":   js.FuncOf(fpt.GetNumParts),
	}

	return ftMap
}

// GetPartStatus returns the status of the file part with the given part number.
//
// The possible values for the status are:
//   - 0 < Part does not exist
//   - 0 = unsent
//   - 1 = arrived (sender has sent a part, and it has arrived)
//   - 2 = received (receiver has received a part)
//
// Parameters:
//   - args[0] - Index of part (int).
//
// Returns:
//   - Part status (int).
func (fpt *FilePartTracker) GetPartStatus(_ js.Value, args []js.Value) any {
	return fpt.api.GetPartStatus(args[0].Int())
}

// GetNumParts returns the total number of file parts in the transfer.
//
// Returns:
//   - Number of parts (int).
func (fpt *FilePartTracker) GetNumParts(js.Value, []js.Value) any {
	return fpt.api.GetNumParts()
}
