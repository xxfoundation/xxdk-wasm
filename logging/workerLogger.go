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
	"time"

	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"

	"gitlab.com/elixxir/xxdk-wasm/worker"
)

// TODO: add ability to import worker so that multiple threads can send logs: https://stackoverflow.com/questions/8343781/how-to-do-worker-to-worker-communication
// TODO: test

// WorkerLogger manages the recording of jwalterweatherman logs to the in-memory
// file buffer in a remote Worker thread.
type WorkerLogger struct {
	threshold      jww.Threshold
	maxLogFileSize int
	wm             *worker.Manager
}

// NewWorkerLogger starts logging to an in-memory log file in a remote Worker
// at the specified threshold. Returns a [WorkerLogger] that can be used to get
// the log file.
func NewWorkerLogger(threshold jww.Threshold, maxLogFileSize int,
	wasmJsPath, workerName string) (*WorkerLogger, error) {
	wl := &WorkerLogger{
		threshold:      threshold,
		maxLogFileSize: maxLogFileSize,
	}

	// Create new worker manager, which will start the worker and wait until
	// communication has been established
	wm, err := worker.NewManager(wasmJsPath, workerName, false)
	if err != nil {
		return nil, err
	}
	wl.wm = wm

	// Register the callback used by the Javascript to request the log file.
	// This prevents an error print when GetFileExtTag is not registered.
	wl.wm.RegisterCallback(GetFileExtTag, func([]byte) {
		jww.DEBUG.Print("[LOG] Received file requested from external " +
			"Javascript. Ignoring file.")
	})

	data, err := json.Marshal(wl.maxLogFileSize)
	if err != nil {
		return nil, err
	}

	// Send message to initialize the log file listener
	errChan := make(chan error)
	wl.wm.SendMessage(NewLogFileTag, data, func(data []byte) {
		if len(data) > 0 {
			errChan <- errors.New(string(data))
		} else {
			errChan <- nil
		}
	})

	// Wait for worker to respond
	select {
	case err = <-errChan:
		if err != nil {
			return nil, err
		}
	case <-time.After(worker.ResponseTimeout):
		return nil, errors.Errorf("timed out after %s waiting for new log "+
			"file in worker to initialize", worker.ResponseTimeout)
	}

	jww.FEEDBACK.Printf("[LOG] Outputting log to file of max size %d at level "+
		"%s using web worker %s", wl.maxLogFileSize, wl.threshold, workerName)

	logger = wl
	return wl, nil
}

// Write adheres to the io.Writer interface and writes log entries to the
// buffer.
func (wl *WorkerLogger) Write(p []byte) (n int, err error) {
	wl.wm.SendMessage(WriteLogTag, p, nil)
	return len(p), nil
}

// Listen adheres to the [jwalterweatherman.LogListener] type and returns the
// log writer when the threshold is within the set threshold limit.
func (wl *WorkerLogger) Listen(t jww.Threshold) io.Writer {
	if t < wl.threshold {
		return nil
	}
	return wl
}

// StopLogging stops log message writes and terminates the worker. Once logging
// is stopped, it cannot be resumed and the log file cannot be recovered.
func (wl *WorkerLogger) StopLogging() {
	wl.threshold = 20

	wl.wm.Stop()
	jww.DEBUG.Printf("[LOG] Terminated log worker.")
}

// GetFile returns the entire log file.
func (wl *WorkerLogger) GetFile() []byte {
	fileChan := make(chan []byte)
	wl.wm.SendMessage(GetFileTag, nil, func(data []byte) { fileChan <- data })

	select {
	case file := <-fileChan:
		return file
	case <-time.After(worker.ResponseTimeout):
		jww.FATAL.Panicf("[LOG] Timed out after %s waiting for log "+
			"file from worker", worker.ResponseTimeout)
		return nil
	}
}

// Threshold returns the log level threshold used in the file.
func (wl *WorkerLogger) Threshold() jww.Threshold {
	return wl.threshold
}

// MaxSize returns the max size, in bytes, that the log file is allowed to be.
func (wl *WorkerLogger) MaxSize() int {
	return wl.maxLogFileSize
}

// Size returns the number of bytes written to the log file.
func (wl *WorkerLogger) Size() int {
	sizeChan := make(chan []byte)
	wl.wm.SendMessage(SizeTag, nil, func(data []byte) { sizeChan <- data })

	select {
	case data := <-sizeChan:
		return int(jww.Threshold(binary.LittleEndian.Uint64(data)))
	case <-time.After(worker.ResponseTimeout):
		jww.FATAL.Panicf("[LOG] Timed out after %s waiting for log "+
			"file size from worker", worker.ResponseTimeout)
		return 0
	}
}

// Worker returns the manager for the Javascript Worker object.
func (wl *WorkerLogger) Worker() *worker.Manager {
	return wl.wm
}
