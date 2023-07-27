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
	"gitlab.com/elixxir/wasm-utils/exception"
	"gitlab.com/elixxir/wasm-utils/utils"
)

// ChannelsFileTransfer wraps the [bindings.ChannelsFileTransfer] object so its
// methods can be wrapped to be Javascript compatible.
type ChannelsFileTransfer struct {
	api *bindings.ChannelsFileTransfer
}

// newChannelsFileTransferJS creates a new Javascript compatible object
// (map[string]any) that matches the [ChannelsFileTransfer] structure.
func newChannelsFileTransferJS(api *bindings.ChannelsFileTransfer) map[string]any {
	cft := ChannelsFileTransfer{api}
	channelsFileTransferMap := map[string]any{
		"GetExtensionBuilderID": js.FuncOf(cft.GetExtensionBuilderID),
		"MaxFileNameLen":        js.FuncOf(cft.MaxFileNameLen),
		"MaxFileTypeLen":        js.FuncOf(cft.MaxFileTypeLen),
		"MaxFileSize":           js.FuncOf(cft.MaxFileSize),
		"MaxPreviewSize":        js.FuncOf(cft.MaxPreviewSize),

		// Uploading/Sending
		"Upload":                       js.FuncOf(cft.Upload),
		"Send":                         js.FuncOf(cft.Send),
		"RegisterSentProgressCallback": js.FuncOf(cft.RegisterSentProgressCallback),
		"RetryUpload":                  js.FuncOf(cft.RetryUpload),
		"CloseSend":                    js.FuncOf(cft.CloseSend),

		// Downloading
		"Download":                         js.FuncOf(cft.Download),
		"RegisterReceivedProgressCallback": js.FuncOf(cft.RegisterReceivedProgressCallback),
	}

	return channelsFileTransferMap
}

// InitChannelsFileTransfer creates a file transfer manager for channels.
//
// Parameters:
//   - args[0] - ID of [E2e] object in tracker (int).
//   - args[1] - JSON of [channelsFileTransfer.Params] (Uint8Array).
//
// Returns:
//   - New [ChannelsFileTransfer] object.
//
// Returns a promise:
//   - Resolves to a Javascript representation of the [ChannelsFileTransfer]
//     object.
//   - Rejected with an error if creating the file transfer object fails.
func InitChannelsFileTransfer(_ js.Value, args []js.Value) any {
	e2eID := args[0].Int()
	paramsJson := utils.CopyBytesToGo(args[1])

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		cft, err := bindings.InitChannelsFileTransfer(e2eID, paramsJson)
		if err != nil {
			reject(exception.NewTrace(err))
		} else {
			resolve(newChannelsFileTransferJS(cft))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// GetExtensionBuilderID returns the ID of the extension builder in the tracker.
// Pass this ID into the channel manager creator to use file transfer manager in
// conjunction with channels.
//
// Returns:
//   - Extension builder ID (int).
func (cft *ChannelsFileTransfer) GetExtensionBuilderID(js.Value, []js.Value) any {
	return cft.api.GetExtensionBuilderID()
}

// MaxFileNameLen returns the max number of bytes allowed for a file name.
//
// Returns:
//   - Max number of bytes (int).
func (cft *ChannelsFileTransfer) MaxFileNameLen(js.Value, []js.Value) any {
	return cft.api.MaxFileNameLen()
}

// MaxFileTypeLen returns the max number of bytes allowed for a file type.
//
// Returns:
//   - Max number of bytes (int).
func (cft *ChannelsFileTransfer) MaxFileTypeLen(js.Value, []js.Value) any {
	return cft.api.MaxFileNameLen()
}

// MaxFileSize returns the max number of bytes allowed for a file.
//
// Returns:
//   - Max number of bytes (int).
func (cft *ChannelsFileTransfer) MaxFileSize(js.Value, []js.Value) any {
	return cft.api.MaxFileSize()
}

// MaxPreviewSize returns the max number of bytes allowed for a file preview.
//
// Returns:
//   - Max number of bytes (int).
func (cft *ChannelsFileTransfer) MaxPreviewSize(js.Value, []js.Value) any {
	return cft.api.MaxFileSize()
}

////////////////////////////////////////////////////////////////////////////////
// Uploading/Sending                                                          //
////////////////////////////////////////////////////////////////////////////////

// Upload starts uploading the file to a new ID that can be sent to the
// specified channel when complete. To get progress information about the
// upload, a [bindings.FtSentProgressCallback] must be registered. All errors
// returned on the callback are fatal and the user must take action to either
// [ChannelsFileTransfer.RetryUpload] or [ChannelsFileTransfer.CloseSend].
//
// The file is added to the event model at the returned file ID with the status
// [channelsFileTransfer.Uploading]. Once the upload is complete, the file link
// is added to the event model with the status [channelsFileTransfer.Complete].
//
// The [bindings.FtSentProgressCallback] only indicates the progress of the file
// upload, not the status of the file in the event model. You must rely on
// updates from the event model to know when it can be retrieved.
//
// Parameters:
//   - args[0] - File contents. Max size defined by
//     [ChannelsFileTransfer.MaxFileSize] (Uint8Array).
//   - args[1] - The number of sending retries allowed on send failure (e.g. a
//     retry of 2.0 with 6 parts means 12 total possible sends) (float).
//   - args[2] - The progress callback, which is a callback that reports the
//     progress of the file upload. The callback is called once on
//     initialization, on every progress update (or less if restricted by the
//     period), or on fatal error. It must be a Javascript object that
//     implements the [bindings.FtSentProgressCallback] interface.
//   - args[3] - Progress callback period. A progress callback will be limited
//     from triggering only once per period, in milliseconds (int).
//
// Returns a promise:
//   - Resolves to the marshalled bytes of [fileTransfer.ID] that uniquely
//     identifies the file (Uint8Array).
//   - Rejected with an error if initiating the upload fails.
func (cft *ChannelsFileTransfer) Upload(_ js.Value, args []js.Value) any {
	var (
		fileData   = utils.CopyBytesToGo(args[0])
		retry      = float32(args[1].Float())
		progressCB = &ftSentCallback{utils.WrapCB(args[2], "Callback")}
		period     = args[3].Int()
	)

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		fileID, err := cft.api.Upload(fileData, retry, progressCB, period)
		if err != nil {
			reject(exception.NewTrace(err))
		} else {
			resolve(utils.CopyBytesToJS(fileID))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// Send sends the specified file info to the channel. Once a file is uploaded
// via [ChannelsFileTransfer.Upload], its file info (found in the event model)
// can be sent to any channel.
//
// Parameters:
//   - args[0] - Marshalled bytes of the channel's [id.ID] to send the file to
//     (Uint8Array).
//   - args[1] - JSON of [channelsFileTransfer.FileLink] stored in the event
//     model (Uint8Array).
//   - args[2] - Human-readable file name. Max length defined by
//     [ChannelsFileTransfer.MaxFileNameLen] (string).
//   - args[3] - Shorthand that identifies the type of file. Max length defined
//     by [ChannelsFileTransfer.MaxFileTypeLen] (string).
//   - args[4] - A preview of the file data (e.g. a thumbnail). Max size defined
//     by [ChannelsFileTransfer.MaxPreviewSize] (Uint8Array).
//   - args[5] - The duration, in milliseconds, that the file is available in
//     the channel (int). For the maximum amount of time, use [ValidForever].
//   - args[6] - JSON of [xxdk.CMIXParams] (Uint8Array). If left empty,
//     [GetDefaultCMixParams] will be used internally.
//   - args[7] - JSON of a slice of public keys of users that should receive
//     mobile notifications for the message.
//
// Returns a promise:
//   - Resolves to the JSON of [bindings.ChannelSendReport] (Uint8Array).
//   - Rejected with an error if sending fails.
func (cft *ChannelsFileTransfer) Send(_ js.Value, args []js.Value) any {
	var (
		channelIdBytes = utils.CopyBytesToGo(args[0])
		fileLinkJSON   = utils.CopyBytesToGo(args[1])
		fileName       = args[2].String()
		fileType       = args[3].String()
		preview        = utils.CopyBytesToGo(args[4])
		validUntilMS   = args[5].Int()
		cmixParamsJSON = utils.CopyBytesToGo(args[6])
		pingsJSON      = utils.CopyBytesToGo(args[7])
	)

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		fileID, err := cft.api.Send(channelIdBytes, fileLinkJSON,
			fileName, fileType, preview, validUntilMS,
			cmixParamsJSON, pingsJSON)
		if err != nil {
			reject(exception.NewTrace(err))
		} else {
			resolve(utils.CopyBytesToJS(fileID))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// RegisterSentProgressCallback allows for the registration of a callback to
// track the progress of an individual file upload. A
// [bindings.FtSentProgressCallback] is auto-registered on
// [ChannelsFileTransfer.Send]; this function should be called when resuming
// clients or registering extra callbacks.
//
// The callback will be called immediately when added to report the current
// progress of the transfer. It will then call every time a file part arrives,
// the transfer completes, or a fatal error occurs. It is called at most once
// every period regardless of the number of progress updates.
//
// In the event that the client is closed and resumed, this function must be
// used to re-register any callbacks previously registered with this function or
// [ChannelsFileTransfer.Send].
//
// The [bindings.FtSentProgressCallback] only indicates the progress of the file
// upload, not the status of the file in the event model. You must rely on
// updates from the event model to know when it can be retrieved.
//
// Parameters:
//   - args[0] - Marshalled bytes of the file's [fileTransfer.ID] (Uint8Array).
//   - args[1] - The progress callback, which is a callback that reports the
//     progress of the file upload. The callback is called once on
//     initialization, on every progress update (or less if restricted by the
//     period), or on fatal error. It must be a Javascript object that
//     implements the [bindings.FtSentProgressCallback] interface.
//   - args[2] - Progress callback period. A progress callback will be limited
//     from triggering only once per period, in milliseconds (int).
//
// Returns a promise:
//   - Resolves on success (void).
//   - Rejected with an error if registering the callback fails.
func (cft *ChannelsFileTransfer) RegisterSentProgressCallback(
	_ js.Value, args []js.Value) any {
	var (
		fileIDBytes = utils.CopyBytesToGo(args[0])
		progressCB  = &ftSentCallback{utils.WrapCB(args[1], "Callback")}
		periodMS    = args[2].Int()
	)

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		err := cft.api.RegisterSentProgressCallback(
			fileIDBytes, progressCB, periodMS)
		if err != nil {
			reject(exception.NewTrace(err))
		} else {
			resolve()
		}
	}

	return utils.CreatePromise(promiseFn)
}

// RetryUpload retries uploading a failed file upload. Returns an error if the
// transfer has not failed.
//
// This function should be called once a transfer errors out (as reported by the
// progress callback).
//
// A new progress callback must be registered on retry. Any previously
// registered callbacks are defunct when the upload fails.
//
// Parameters:
//   - args[0] - Marshalled bytes of the file's [fileTransfer.ID] (Uint8Array).
//   - args[1] - The progress callback, which is a callback that reports the
//     progress of the file upload. The callback is called once on
//     initialization, on every progress update (or less if restricted by the
//     period), or on fatal error. It must be a Javascript object that
//     implements the [bindings.FtSentProgressCallback] interface.
//   - args[2] - Progress callback period. A progress callback will be limited
//     from triggering only once per period, in milliseconds (int).
//
// Returns a promise:
//   - Resolves on success (void).
//   - Rejected with an error if registering retrying the upload fails.
func (cft *ChannelsFileTransfer) RetryUpload(_ js.Value, args []js.Value) any {
	var (
		fileIDBytes = utils.CopyBytesToGo(args[0])
		progressCB  = &ftSentCallback{utils.WrapCB(args[1], "Callback")}
		periodMS    = args[2].Int()
	)

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		err := cft.api.RetryUpload(fileIDBytes, progressCB, periodMS)
		if err != nil {
			reject(exception.NewTrace(err))
		} else {
			resolve()
		}
	}

	return utils.CreatePromise(promiseFn)
}

// CloseSend deletes a file from the internal storage once a transfer has
// completed or reached the retry limit. If neither of those condition are met,
// an error is returned.
//
// This function should be called once a transfer completes or errors out (as
// reported by the progress callback).
//
// Parameters:
//   - args[0] - Marshalled bytes of the file's [fileTransfer.ID] (Uint8Array).
//
// Returns a promise:
//   - Resolves on success (void).
//   - Rejected with an error if the file has not failed or completed or if
//     closing failed.
func (cft *ChannelsFileTransfer) CloseSend(_ js.Value, args []js.Value) any {
	fileIDBytes := utils.CopyBytesToGo(args[0])

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		err := cft.api.CloseSend(fileIDBytes)
		if err != nil {
			reject(exception.NewTrace(err))
		} else {
			resolve()
		}
	}

	return utils.CreatePromise(promiseFn)
}

////////////////////////////////////////////////////////////////////////////////
// Download                                                                   //
////////////////////////////////////////////////////////////////////////////////

// Download begins the download of the file described in the marshalled
// [channelsFileTransfer.FileInfo]. The progress of the download is reported on
// the [bindings.FtReceivedProgressCallback].
//
// Once the download completes, the file will be stored in the event model with
// the given file ID and with the status [channels.ReceptionProcessingComplete].
//
// The [bindings.FtReceivedProgressCallback] only indicates the progress of the
// file download, not the status of the file in the event model. You must rely
// on updates from the event model to know when it can be retrieved.
//
// Parameters:
//   - args[0] - The JSON of [channelsFileTransfer.FileInfo] received on a
//     channel (Uint8Array).
//   - args[1] - The progress callback, which is a callback that reports the
//     progress of the file download. The callback is called once on
//     initialization, on every progress update (or less if restricted by the
//     period), or on fatal error. It must be a Javascript object that
//     implements the [bindings.FtReceivedProgressCallback] interface.
//   - args[2] - Progress callback period. A progress callback will be limited
//     from triggering only once per period, in milliseconds (int).
//
// Returns:
//   - Marshalled bytes of [fileTransfer.ID] that uniquely identifies the file.
//
// Returns a promise:
//   - Resolves to the marshalled bytes of [fileTransfer.ID] that uniquely
//     identifies the file. (Uint8Array).
//   - Rejected with an error if downloading fails.
func (cft *ChannelsFileTransfer) Download(_ js.Value, args []js.Value) any {
	var (
		fileInfoJSON = utils.CopyBytesToGo(args[0])
		progressCB   = &ftReceivedCallback{utils.WrapCB(args[1], "Callback")}
		periodMS     = args[2].Int()
	)

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		fileID, err := cft.api.Download(fileInfoJSON, progressCB, periodMS)
		if err != nil {
			reject(exception.NewTrace(err))
		} else {
			resolve(utils.CopyBytesToJS(fileID))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// RegisterReceivedProgressCallback allows for the registration of a callback to
// track the progress of an individual file download.
//
// The callback will be called immediately when added to report the current
// progress of the transfer. It will then call every time a file part is
// received, the transfer completes, or a fatal error occurs. It is called at
// most once every period regardless of the number of progress updates.
//
// In the event that the client is closed and resumed, this function must be
// used to re-register any callbacks previously registered.
//
// Once the download completes, the file will be stored in the event model with
// the given file ID and with the status [channelsFileTransfer.Complete].
//
// The [bindings.FtReceivedProgressCallback] only indicates the progress of the
// file download, not the status of the file in the event model. You must rely
// on updates from the event model to know when it can be retrieved.
//
// Parameters:
//   - args[0] - Marshalled bytes of the file's [fileTransfer.ID] (Uint8Array).
//   - args[1] - The progress callback, which is a callback that reports the
//     progress of the file download. The callback is called once on
//     initialization, on every progress update (or less if restricted by the
//     period), or on fatal error. It must be a Javascript object that
//     implements the [bindings.FtReceivedProgressCallback] interface.
//   - args[2] - Progress callback period. A progress callback will be limited
//     from triggering only once per period, in milliseconds (int).
//
// Returns a promise:
//   - Resolves on success (void).
//   - Rejected with an error if registering the callback fails.
func (cft *ChannelsFileTransfer) RegisterReceivedProgressCallback(
	_ js.Value, args []js.Value) any {
	var (
		fileIDBytes = utils.CopyBytesToGo(args[0])
		progressCB  = &ftReceivedCallback{utils.WrapCB(args[1], "Callback")}
		periodMS    = args[2].Int()
	)

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		err := cft.api.RegisterReceivedProgressCallback(
			fileIDBytes, progressCB, periodMS)
		if err != nil {
			reject(exception.NewTrace(err))
		} else {
			resolve()
		}
	}

	return utils.CreatePromise(promiseFn)
}

////////////////////////////////////////////////////////////////////////////////
// Callbacks                                                                  //
////////////////////////////////////////////////////////////////////////////////

// ftSentCallback wraps Javascript callbacks to adhere to the
// [bindings.FtSentProgressCallback] interface.
type ftSentCallback struct {
	callback func(args ...any) js.Value
}

// Callback is called when the status of the sent file changes.
//
// Parameters:
//   - payload - Returns the contents of the message. JSON of
//     [bindings.Progress] (Uint8Array).
//   - t - Returns a tracker that allows the lookup of the status of any file
//     part. It is a Javascript object that matches the functions on
//     [FilePartTracker].
//   - err - Returns an error on failure (Error).

// Callback is called when the progress on a sent file changes or an error
// occurs in the transfer.
//
// The [ChFilePartTracker] can be used to look up the status of individual file
// parts. Note, when completed == true, the [ChFilePartTracker] may be nil.
//
// Any error returned is fatal and the file must either be retried with
// [ChannelsFileTransfer.RetryUpload] or canceled with
// [ChannelsFileTransfer.CloseSend].
//
// This callback only indicates the status of the file transfer, not the status
// of the file in the event model. Do NOT use this callback as an indicator of
// when the file is available in the event model.
//
// Parameters:
//   - payload - JSON of [bindings.FtSentProgress], which describes the progress
//     of the current sent transfer.
//   - fpt - File part tracker that allows the lookup of the status of
//     individual file parts.
//   - err - Fatal errors during sending.
func (fsc *ftSentCallback) Callback(
	payload []byte, t *bindings.ChFilePartTracker, err error) {
	fsc.callback(utils.CopyBytesToJS(payload), newChFilePartTrackerJS(t),
		exception.NewTrace(err))
}

// ftReceivedCallback wraps Javascript callbacks to adhere to the
// [bindings.FtReceivedProgressCallback] interface.
type ftReceivedCallback struct {
	callback func(args ...any) js.Value
}

// Callback is called when
// the progress on a received file changes or an error occurs in the transfer.
//
// The [ChFilePartTracker] can be used to look up the status of individual file
// parts. Note, when completed == true, the [ChFilePartTracker] may be nil.
//
// This callback only indicates the status of the file transfer, not the status
// of the file in the event model. Do NOT use this callback as an indicator of
// when the file is available in the event model.
//
// Parameters:
//   - payload - JSON of [bindings.FtReceivedProgress], which describes the
//     progress of the current received transfer.
//   - fpt - File part tracker that allows the lookup of the status of
//     individual file parts.
//   - err - Fatal errors during receiving.
func (frc *ftReceivedCallback) Callback(
	payload []byte, t *bindings.ChFilePartTracker, err error) {
	frc.callback(utils.CopyBytesToJS(payload), newChFilePartTrackerJS(t),
		exception.NewTrace(err))
}

////////////////////////////////////////////////////////////////////////////////
// File Part Tracker                                                          //
////////////////////////////////////////////////////////////////////////////////

// ChFilePartTracker wraps the [bindings.ChFilePartTracker] object so its
// methods can be wrapped to be Javascript compatible.
type ChFilePartTracker struct {
	api *bindings.ChFilePartTracker
}

// newChFilePartTrackerJS creates a new Javascript compatible object
// (map[string]any) that matches the [FilePartTracker] structure.
func newChFilePartTrackerJS(api *bindings.ChFilePartTracker) map[string]any {
	fpt := ChFilePartTracker{api}
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
func (fpt *ChFilePartTracker) GetPartStatus(_ js.Value, args []js.Value) any {
	return fpt.api.GetPartStatus(args[0].Int())
}

// GetNumParts returns the total number of file parts in the transfer.
//
// Returns:
//   - Number of parts (int).
func (fpt *ChFilePartTracker) GetNumParts(js.Value, []js.Value) any {
	return fpt.api.GetNumParts()
}
