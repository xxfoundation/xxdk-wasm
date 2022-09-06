////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"gitlab.com/elixxir/client/bindings"
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

// newFileTransferJS creates a new Javascript compatible object
// (map[string]interface{}) that matches the FileTransfer structure.
func newFileTransferJS(api *bindings.FileTransfer) map[string]interface{} {
	ft := FileTransfer{api}
	ftMap := map[string]interface{}{
		// Main functions
		"Send":      js.FuncOf(ft.Send),
		"Receive":   js.FuncOf(ft.Receive),
		"CloseSend": js.FuncOf(ft.CloseSend),

		// Callback registration functions
		"RegisterSentProgressCallback":     js.FuncOf(ft.RegisterSentProgressCallback),
		"RegisterReceivedProgressCallback": js.FuncOf(ft.RegisterReceivedProgressCallback),

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
	callback func(args ...interface{}) js.Value
}

func (rfc *receiveFileCallback) Callback(payload []byte, err error) {
	rfc.callback(CopyBytesToJS(payload), err.Error())
}

// fileTransferSentProgressCallback wraps Javascript callbacks to adhere to the
// [bindings.FileTransferSentProgressCallback] interface.
type fileTransferSentProgressCallback struct {
	callback func(args ...interface{}) js.Value
}

func (spc *fileTransferSentProgressCallback) Callback(
	payload []byte, t *bindings.FilePartTracker, err error) {
	spc.callback(CopyBytesToJS(payload), newFilePartTrackerJS(t), err.Error())
}

// fileTransferReceiveProgressCallback wraps Javascript callbacks to adhere to
// the [bindings.FileTransferReceiveProgressCallback] interface.
type fileTransferReceiveProgressCallback struct {
	callback func(args ...interface{}) js.Value
}

func (rpc *fileTransferReceiveProgressCallback) Callback(
	payload []byte, t *bindings.FilePartTracker, err error) {
	rpc.callback(CopyBytesToJS(payload), newFilePartTrackerJS(t), err.Error())
}

////////////////////////////////////////////////////////////////////////////////
// Main functions                                                             //
////////////////////////////////////////////////////////////////////////////////

// InitFileTransfer creates a bindings-level file transfer manager.
//
// Parameters:
//  - args[0] - ID of E2e object in tracker (int).
//  - args[1] - Javascript object that has functions that implement the
//    [bindings.ReceiveFileCallback] interface.
//  - args[2] - JSON of [fileTransfer.e2e.Params] (Uint8Array).
//  - args[3] - JSON of [fileTransfer.Params] (Uint8Array).
//
// Returns:
//  - Javascript representation of the FileTransfer object.
//  - Throws a TypeError initialising the file transfer manager fails.
func InitFileTransfer(_ js.Value, args []js.Value) interface{} {
	rfc := &receiveFileCallback{WrapCB(args[1], "Callback")}
	e2eFileTransferParamsJson := CopyBytesToGo(args[2])
	fileTransferParamsJson := CopyBytesToGo(args[3])

	api, err := bindings.InitFileTransfer(
		args[0].Int(), rfc, e2eFileTransferParamsJson, fileTransferParamsJson)
	if err != nil {
		Throw(TypeError, err)
		return nil
	}

	return newFileTransferJS(api)
}

// Send is the bindings-level function for sending a file.
//
// Parameters:
//  - args[0] - JSON of [bindings.FileSend] (Uint8Array).
//  - args[1] - marshalled recipient [id.ID] (Uint8Array).
//  - args[2] - number of retries allowed (float)
//  - args[3] - Javascript object that has functions that implement the
//    [bindings.FileTransferSentProgressCallback] interface.
//  - args[4] - duration to wait between progress callbacks triggering (string).
//    Reference [time.ParseDuration] for info on valid duration strings.
//
// Returns:
//  - A unique ID for this file transfer (Uint8Array).
//  - Throws a TypeError if sending fails.
func (f *FileTransfer) Send(_ js.Value, args []js.Value) interface{} {
	payload := CopyBytesToGo(args[0])
	recipientID := CopyBytesToGo(args[1])
	retry := float32(args[2].Float())
	spc := &fileTransferSentProgressCallback{WrapCB(args[3], "Callback")}

	ftID, err := f.api.Send(payload, recipientID, retry, spc, args[4].String())
	if err != nil {
		Throw(TypeError, err)
		return nil
	}

	return CopyBytesToJS(ftID)
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
//  - args[0] - file transfer ID (Uint8Array).
//
// Returns:
//  - File contents (Uint8Array).
//  - Throws a TypeError the file transfer is incomplete or Receive has already
//    been called.
func (f *FileTransfer) Receive(_ js.Value, args []js.Value) interface{} {
	file, err := f.api.Receive(CopyBytesToGo(args[0]))
	if err != nil {
		Throw(TypeError, err)
		return nil
	}

	return CopyBytesToJS(file)
}

// CloseSend deletes a file from the internal storage once a transfer has
// completed or reached the retry limit. Returns an error if the transfer has
// not run out of retries.
//
// This function should be called once a transfer completes or errors out (as
// reported by the progress callback).
//
// Parameters:
//  - args[0] - file transfer ID (Uint8Array).
//
// Returns:
//  - Throws a TypeError if the file transfer is incomplete.
func (f *FileTransfer) CloseSend(_ js.Value, args []js.Value) interface{} {
	err := f.api.CloseSend(CopyBytesToGo(args[0]))
	if err != nil {
		Throw(TypeError, err)
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
//  - args[0] - file transfer ID (Uint8Array).
//  - args[1] - Javascript object that has functions that implement the
//    [bindings.FileTransferSentProgressCallback] interface.
//  - args[2] - duration to wait between progress callbacks triggering (string).
//    Reference [time.ParseDuration] for info on valid duration strings.
//
// Returns:
//  - Throws a TypeError if registering the callback fails.
func (f *FileTransfer) RegisterSentProgressCallback(
	_ js.Value, args []js.Value) interface{} {
	tidBytes := CopyBytesToGo(args[0])
	spc := &fileTransferSentProgressCallback{WrapCB(args[1], "Callback")}

	err := f.api.RegisterSentProgressCallback(tidBytes, spc, args[2].String())
	if err != nil {
		Throw(TypeError, err)
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
//  - args[0] - file transfer ID (Uint8Array).
//  - args[1] - Javascript object that has functions that implement the
//    [bindings.FileTransferReceiveProgressCallback] interface.
//  - args[2] - duration to wait between progress callbacks triggering (string).
//    Reference [time.ParseDuration] for info on valid duration strings.
//
// Returns:
//  - Throws a TypeError if registering the callback fails.
func (f *FileTransfer) RegisterReceivedProgressCallback(
	_ js.Value, args []js.Value) interface{} {
	tidBytes := CopyBytesToGo(args[0])
	rpc := &fileTransferReceiveProgressCallback{WrapCB(args[1], "Callback")}

	err := f.api.RegisterReceivedProgressCallback(
		tidBytes, rpc, args[2].String())
	if err != nil {
		Throw(TypeError, err)
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
//  - int
func (f *FileTransfer) MaxFileNameLen(js.Value, []js.Value) interface{} {
	return f.api.MaxFileNameLen()
}

// MaxFileTypeLen returns the max number of bytes allowed for a file type.
//
// Returns:
//  - int
func (f *FileTransfer) MaxFileTypeLen(js.Value, []js.Value) interface{} {
	return f.api.MaxFileTypeLen()
}

// MaxFileSize returns the max number of bytes allowed for a file.
//
// Returns:
//  - int
func (f *FileTransfer) MaxFileSize(js.Value, []js.Value) interface{} {
	return f.api.MaxFileSize()
}

// MaxPreviewSize returns the max number of bytes allowed for a file preview.
//
// Returns:
//  - int
func (f *FileTransfer) MaxPreviewSize(js.Value, []js.Value) interface{} {
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
// (map[string]interface{}) that matches the filePartTracker structure.
func newFilePartTrackerJS(api *bindings.FilePartTracker) map[string]interface{} {
	fpt := FilePartTracker{api}
	ftMap := map[string]interface{}{
		"GetPartStatus": js.FuncOf(fpt.GetPartStatus),
		"GetNumParts":   js.FuncOf(fpt.GetNumParts),
	}

	return ftMap
}

// GetPartStatus returns the status of the file part with the given part number.
//
// The possible values for the status are:
//  - 0 < Part does not exist
//  - 0 = unsent
//  - 1 = arrived (sender has sent a part, and it has arrived)
//  - 2 = received (receiver has received a part)
//
// Parameters:
//  - args[0] - index of part (int).
//
// Returns:
//  - Part status (int).
func (fpt *FilePartTracker) GetPartStatus(_ js.Value, args []js.Value) interface{} {
	return fpt.api.GetPartStatus(args[0].Int())
}

// GetNumParts returns the total number of file parts in the transfer.
//
// Returns:
//  - int
func (fpt *FilePartTracker) GetNumParts(js.Value, []js.Value) interface{} {
	return fpt.api.GetNumParts()
}
