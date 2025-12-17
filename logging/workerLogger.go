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
	"fmt"
	"io"
	"math"
	"sync/atomic"
	"time"

	json "github.com/goccy/go-json"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"

	"gitlab.com/elixxir/xxdk-wasm/worker"
)

// workerLogger manages the recording of jwalterweatherman logs to the in-memory
// file buffer in a remote Worker thread.
type workerLogger struct {
	threshold         jww.Threshold
	maxLogFileSize    int
	wm                *worker.Manager
	logChan           chan []byte
	dropCounter       uint64
	lastDropReportSec int64 // Unix timestamp of last drop report (atomic)
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
		logChan:        make(chan []byte, 100), // Buffer up to 100 log messages to limit memory pressure
	}

	// Start background goroutine to drain log messages
	go wl.logWriter()

	// Register the callback used by the Javascript to request the log file.
	// This prevents an error print when GetFileExtTag is not registered.
	// Note: Cannot log here as it would cause infinite recursion through workerLogger
	wl.wm.RegisterCallback(GetFileExtTag, func([]byte, func([]byte)) {
		// Ignoring external file request
	})

	data, err := json.Marshal(wl.maxLogFileSize)
	if err != nil {
		return nil, err
	}

	// Send message to initialize the log file listener
	response, err := wl.wm.SendMessage(NewLogFileTag, data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize the log file listener")
	} else if len(response) > 0 {
		// Note: Use len(response) > 0 instead of response != nil because
		// base64.DecodeString("") returns []byte{} (empty, not nil)
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
// Drops messages if the buffer is full to prevent blocking.
func (wl *workerLogger) Write(p []byte) (n int, err error) {
	// Make a copy since p may be reused by the caller
	msg := make([]byte, len(p))
	copy(msg, p)

	select {
	case wl.logChan <- msg:
		// Successfully buffered
	default:
		// Buffer full, drop message and increment counter
		atomic.AddUint64(&wl.dropCounter, 1)
	}

	return len(p), nil
}

// logWriter runs in a background goroutine and drains the log channel,
// sending messages to the worker thread.
func (wl *workerLogger) logWriter() {
	const dropReportIntervalSec = 5 // Report drops at most once per 5 seconds

	for msg := range wl.logChan {
		// Check if any messages were dropped and report it (throttled)
		dropped := atomic.LoadUint64(&wl.dropCounter)
		if dropped > 0 {
			now := time.Now().Unix()
			lastReport := atomic.LoadInt64(&wl.lastDropReportSec)

			// Only report if enough time has passed since last report
			if now-lastReport >= dropReportIntervalSec {
				if atomic.CompareAndSwapInt64(&wl.lastDropReportSec, lastReport, now) {
					// Reset counter and report
					dropped = atomic.SwapUint64(&wl.dropCounter, 0)
					dropMsg := []byte(fmt.Sprintf("[LOG] Dropped %d log messages due to backpressure\n", dropped))
					_ = wl.wm.SendNoResponse(WriteLogTag, dropMsg)
				}
			}
		}

		// Send the actual log message
		_ = wl.wm.SendNoResponse(WriteLogTag, msg)
	}
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

	// Close the log channel to stop the writer goroutine
	close(wl.logChan)

	// Note: Cannot log here as logChan is closed and could cause recursion
	_ = wl.wm.Stop()
}

// GetFile returns the entire log file.
func (wl *workerLogger) GetFile() []byte {
	response, err := wl.wm.SendMessage(GetFileTag, nil)
	if err != nil {
		// Cannot use jww logging here as it would cause infinite recursion
		panic(fmt.Sprintf("[LOG] Failed to get log file from worker: %+v", err))
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
		// Cannot use jww logging here as it would cause infinite recursion
		panic(fmt.Sprintf("[LOG] Failed to get log size from worker: %+v", err))
	}

	return int(binary.LittleEndian.Uint64(response))
}

// Worker returns the manager for the Javascript Worker object.
func (wl *workerLogger) Worker() *worker.Manager {
	return wl.wm
}
