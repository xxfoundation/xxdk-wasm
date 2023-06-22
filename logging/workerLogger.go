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
	"io"
	"math"

	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"

	"gitlab.com/elixxir/xxdk-wasm/worker"
)

// workerLogger manages the recording of jwalterweatherman logs to the in-memory
// file buffer in a remote Worker thread.
type workerLogger struct {
	threshold      jww.Threshold
	maxLogFileSize int
	wm             *worker.Manager
}

// newWorkerLogger starts logging to an in-memory log file in a remote Worker
// at the specified threshold. Returns a [workerLogger] that can be used to get
// the log file.
func newWorkerLogger(threshold jww.Threshold, maxLogFileSize int,
	wasmJsPath, workerName string) (*workerLogger, error) {
	// Create new worker manager, which will start the worker and wait until
	// communication has been established
	wm, err := worker.NewManager(wasmJsPath, workerName, false)
	if err != nil {
		return nil, err
	}

	wl := &workerLogger{
		threshold:      threshold,
		maxLogFileSize: maxLogFileSize,
		wm:             wm,
	}

	// Register the callback used by the Javascript to request the log file.
	// This prevents an error print when GetFileExtTag is not registered.
	wl.wm.RegisterCallback(GetFileExtTag, func([]byte, func([]byte)) {
		jww.DEBUG.Print("[LOG] Received file requested from external " +
			"Javascript. Ignoring file.")
	})

	data, err := json.Marshal(wl.maxLogFileSize)
	if err != nil {
		return nil, err
	}

	// Send message to initialize the log file listener
	response, err := wl.wm.SendMessage(NewLogFileTag, data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize the log file listener")
	} else if response != nil {
		return nil, errors.Wrap(errors.New(string(response)),
			"failed to initialize the log file listener")
	}

	jww.FEEDBACK.Printf("[LOG] Outputting log to file of max size %d at level "+
		"%s using web worker %s", wl.maxLogFileSize, wl.threshold, workerName)

	logger = wl
	return wl, nil
}

// Write adheres to the io.Writer interface and sends the log entries to the
// worker to be added to the file buffer. Always returns the length of p.
func (wl *workerLogger) Write(p []byte) (n int, err error) {
	return len(p), wl.wm.SendNoResponse(WriteLogTag, p)
}

// Listen adheres to the [jwalterweatherman.LogListener] type and returns the
// log writer when the threshold is within the set threshold limit.
func (wl *workerLogger) Listen(threshold jww.Threshold) io.Writer {
	if threshold < wl.threshold {
		return nil
	}
	return wl
}

// StopLogging stops log message writes and terminates the worker. Once logging
// is stopped, it cannot be resumed and the log file cannot be recovered.
func (wl *workerLogger) StopLogging() {
	wl.threshold = math.MaxInt

	err := wl.wm.Stop()
	if err != nil {
		jww.ERROR.Printf("[LOG] Failed to terminate log worker: %+v", err)
	} else {
		jww.DEBUG.Printf("[LOG] Terminated log worker.")
	}
}

// GetFile returns the entire log file.
func (wl *workerLogger) GetFile() []byte {
	response, err := wl.wm.SendMessage(GetFileTag, nil)
	if err != nil {
		jww.FATAL.Panicf("[LOG] Failed to get log file from worker: %+v", err)
	}

	return response
}

// Threshold returns the log level threshold used in the file.
func (wl *workerLogger) Threshold() jww.Threshold {
	return wl.threshold
}

// MaxSize returns the max size, in bytes, that the log file is allowed to be.
func (wl *workerLogger) MaxSize() int {
	return wl.maxLogFileSize
}

// Size returns the number of bytes written to the log file.
func (wl *workerLogger) Size() int {
	response, err := wl.wm.SendMessage(SizeTag, nil)
	if err != nil {
		jww.FATAL.Panicf("[LOG] Failed to get log size from worker: %+v", err)
	}

	return int(binary.LittleEndian.Uint64(response))
}

// Worker returns the manager for the Javascript Worker object.
func (wl *workerLogger) Worker() *worker.Manager {
	return wl.wm
}
