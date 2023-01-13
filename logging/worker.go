////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package logging

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/xxdk-wasm/utils"
	"gitlab.com/elixxir/xxdk-wasm/worker"
	"io"
	"syscall/js"
	"time"
)

// todo: add ability to import beginning of log

// List of tags that can be used when sending a message or registering a handler
// to receive a message.
const (
	NewLogFileTag worker.Tag = "NewLogFile"
	WriteLogTag   worker.Tag = "WriteLog"
	GetFileTag    worker.Tag = "GetFile"
	GetFileExtTag worker.Tag = "GetFileExt"
	SizeTag       worker.Tag = "Size"
)

// LogFileWorker manages communication with the web worker running the log file
// listener.
type LogFileWorker struct {
	name           string
	threshold      jww.Threshold
	maxLogFileSize int
	wm             *worker.Manager
}

// LogToFileWorker starts a new worker that begins listening for logs and
// writing them to file.
func LogToFileWorker(wasmJsPath, workerName string, threshold jww.Threshold,
	maxLogFileSize int) (*LogFileWorker, error) {
	if threshold < jww.LevelTrace || threshold > jww.LevelFatal {
		return nil,
			errors.Errorf("log level is not valid: log level: %d", threshold)
	}

	wm, err := worker.NewManager(wasmJsPath, workerName, false)
	if err != nil {
		return nil, err
	}

	lfw := &LogFileWorker{
		name:           workerName,
		threshold:      threshold,
		maxLogFileSize: maxLogFileSize,
		wm:             wm,
	}

	lfw.wm.RegisterCallback(GetFileExtTag, func([]byte) {
		jww.DEBUG.Print(
			"Received file requested from external Javascript. Ignoring file.")
	})

	data, err := json.Marshal(maxLogFileSize)
	if err != nil {
		return nil, err
	}

	errChan := make(chan error)
	wm.SendMessage(NewLogFileTag, data, func(data []byte) {
		if len(data) > 0 {
			errChan <- errors.New(string(data))
		} else {
			errChan <- nil
		}
	})

	select {
	case err = <-errChan:
		if err != nil {
			return nil, err
		}
	case <-time.After(worker.ResponseTimeout):
		return nil, errors.Errorf("timed out after %s waiting for new log "+
			"file in worker to initialize", worker.ResponseTimeout)
	}

	// Add the log listener
	jww.SetLogListeners(AddLogListener(lfw.Listen)...)

	msg := fmt.Sprintf("Outputting log to file of max size %d with level %s",
		lfw.MaxSize(), lfw.Threshold())
	switch threshold {
	case jww.LevelTrace:
		fallthrough
	case jww.LevelDebug:
		fallthrough
	case jww.LevelInfo:
		jww.INFO.Print(msg)
	case jww.LevelWarn:
		jww.WARN.Print(msg)
	case jww.LevelError:
		jww.ERROR.Print(msg)
	case jww.LevelCritical:
		jww.CRITICAL.Print(msg)
	case jww.LevelFatal:
		jww.FATAL.Print(msg)
	}

	return lfw, nil
}

// Listen is called for every logging event. This function adheres to the
// [jwalterweatherman.LogListener] type.
func (lfw *LogFileWorker) Listen(t jww.Threshold) io.Writer {
	if t < lfw.threshold {
		return nil
	}

	return lfw
}

// Write sends the bytes to the worker. It does not wait for a response and
// always returns the length of p and a nil error.
func (lfw *LogFileWorker) Write(p []byte) (n int, err error) {
	lfw.wm.SendMessage(WriteLogTag, p, nil)
	return len(p), err
}

// GetFile returns the entire log file.
func (lfw *LogFileWorker) GetFile() []byte {
	fileChan := make(chan []byte)
	lfw.wm.SendMessage(GetFileTag, nil, func(data []byte) { fileChan <- data })

	select {
	case file := <-fileChan:
		return file
	case <-time.After(worker.ResponseTimeout):
		jww.FATAL.Panicf("Timed out after %s waiting for log file from worker",
			worker.ResponseTimeout)
		return nil
	}
}

// Threshold returns the log level threshold used in the file.
func (lfw *LogFileWorker) Threshold() jww.Threshold {
	return lfw.threshold
}

// MaxSize returns the max size, in bytes, that the log file is allowed to be.
func (lfw *LogFileWorker) MaxSize() int {
	return lfw.maxLogFileSize
}

// Size returns the current size, in bytes, written to the log file.
func (lfw *LogFileWorker) Size() int {
	sizeChan := make(chan []byte)
	lfw.wm.SendMessage(SizeTag, nil, func(data []byte) { sizeChan <- data })

	select {
	case data := <-sizeChan:
		return int(jww.Threshold(binary.LittleEndian.Uint64(data)))
	case <-time.After(worker.ResponseTimeout):
		jww.FATAL.Panicf("Timed out after %s waiting for log file size "+
			"from worker", worker.ResponseTimeout)
		return 0
	}
}

////////////////////////////////////////////////////////////////////////////////
// Javascript Bindings                                                        //
////////////////////////////////////////////////////////////////////////////////

// LogToFileWorkerJS starts a new worker that begins listening for logs and
// writing them to file.
//
// Parameters:
//   - args[0] - Path to Javascript start file for the worker WASM (string).
//   - args[1] - Name of the worker (used in logs) (string).
//   - args[2] - Log level (int).
//   - args[3] - Max log file size, in bytes (int).
//
// Returns a promise:
//   - Resolves to a Javascript representation of the [LogFileWorker] object,
//     which allows accessing the contents of the log file and other metadata.
//   - Rejected with an error if starting the worker fails.
func LogToFileWorkerJS(_ js.Value, args []js.Value) any {
	wasmJsPath := args[0].String()
	workerName := args[1].String()
	threshold := jww.Threshold(args[2].Int())
	maxLogFileSize := args[3].Int()

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		lfw, err := LogToFileWorker(
			wasmJsPath, workerName, threshold, maxLogFileSize)
		if err != nil {
			reject(utils.JsTrace(err))
		} else {
			resolve(NewLogFileWorkerJS(lfw))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// NewLogFileWorkerJS creates a new Javascript compatible object
// (map[string]any) that matches the [LogFileWorker] structure.
func NewLogFileWorkerJS(lfw *LogFileWorker) map[string]any {
	logFileWorker := map[string]any{
		"GetFile":   js.FuncOf(lfw.GetFileJS),
		"Threshold": js.FuncOf(lfw.ThresholdJS),
		"MaxSize":   js.FuncOf(lfw.MaxSizeJS),
		"Size":      js.FuncOf(lfw.SizeJS),
		"Worker":    js.FuncOf(lfw.WorkerJS),
	}

	return logFileWorker
}

// GetFileJS returns the entire log file.
//
// Returns promise:
//   - Log file contents (string).
//
// Returns a promise:
//   - Resolves to the log file contents (string).
func (lfw *LogFileWorker) GetFileJS(js.Value, []js.Value) any {
	promiseFn := func(resolve, _ func(args ...any) js.Value) {
		resolve(string(lfw.GetFile()))
	}

	return utils.CreatePromise(promiseFn)
}

// ThresholdJS returns the log level threshold used in the file.
//
// Returns:
//   - Log level (string).
func (lfw *LogFileWorker) ThresholdJS(js.Value, []js.Value) any {
	return lfw.Threshold().String()
}

// MaxSizeJS returns the max size, in bytes, that the log file is allowed to be.
//
// Returns:
//   - Max file size (int).
func (lfw *LogFileWorker) MaxSizeJS(js.Value, []js.Value) any {
	return lfw.MaxSize()
}

// SizeJS returns the current size, in bytes, written to the log file.
//
// Returns a promise:
//   - Resolves to the current file size (int).
func (lfw *LogFileWorker) SizeJS(js.Value, []js.Value) any {
	promiseFn := func(resolve, _ func(args ...any) js.Value) {
		resolve(lfw.Size())
	}

	return utils.CreatePromise(promiseFn)
}

// WorkerJS returns the web worker object.
//
// Returns:
//   - Javascript worker object.
func (lfw *LogFileWorker) WorkerJS(js.Value, []js.Value) any {
	return lfw.wm.GetWorker()
}
