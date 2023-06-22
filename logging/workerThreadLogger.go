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
	"io"
	"math"
	"syscall/js"

	"github.com/hack-pad/safejs"
	jww "github.com/spf13/jwalterweatherman"

	"gitlab.com/elixxir/xxdk-wasm/worker"
)

// threadLogger manages the recording of jwalterweatherman logs from a worker to
// the in-memory file buffer in a remote Worker thread.
type threadLogger struct {
	threshold jww.Threshold
	mm        *worker.MessageManager
}

// newThreadLogger starts logging to an in-memory log file in a remote Worker
// at the specified threshold. Returns a [threadLogger] that can be used to get
// the log file.
func newThreadLogger(threshold jww.Threshold, channelName string,
	messagePort js.Value) (*threadLogger, error) {
	p := worker.DefaultParams()
	p.MessageLogging = false
	mm, err := worker.NewMessageManager(safejs.Safe(messagePort),
		channelName+"-worker", p)
	if err != nil {
		return nil, err
	}

	tl := &threadLogger{
		threshold: threshold,
		mm:        mm,
	}

	jww.FEEDBACK.Printf("[LOG] Worker outputting log to file at level "+
		"%s using web worker", tl.threshold)

	logger = tl
	return tl, nil
}

// Write adheres to the io.Writer interface and sends the log entries to the
// worker to be added to the file buffer. Always returns the length of p and
// nil. All errors are printed to the log.
func (tl *threadLogger) Write(p []byte) (n int, err error) {
	return len(p), tl.mm.SendNoResponse(WriteLogTag, p)
}

// Listen adheres to the [jwalterweatherman.LogListener] type and returns the
// log writer when the threshold is within the set threshold limit.
func (tl *threadLogger) Listen(threshold jww.Threshold) io.Writer {
	if threshold < tl.threshold {
		return nil
	}
	return tl
}

// StopLogging stops sending log messages to the logging worker. Once logging is
// stopped, it cannot be resumed and the log file cannot be recovered. This does
// not stop the logging worker.
func (tl *threadLogger) StopLogging() {
	tl.threshold = math.MaxInt
}

// GetFile returns the entire log file.
func (tl *threadLogger) GetFile() []byte {
	response, err := tl.mm.Send(GetFileTag, nil)
	if err != nil {
		jww.FATAL.Panicf("[LOG] Failed to get log file from worker: %+v", err)
	}

	return response
}

// Threshold returns the log level threshold of logs sent to the worker.
func (tl *threadLogger) Threshold() jww.Threshold {
	return tl.threshold
}

// MaxSize returns the max size, in bytes, that the log file is allowed to be.
func (tl *threadLogger) MaxSize() int {
	response, err := tl.mm.Send(MaxSizeTag, nil)
	if err != nil {
		jww.FATAL.Panicf("[LOG] Failed to max file size from worker: %+v", err)
	}

	return int(binary.LittleEndian.Uint64(response))
}

// Size returns the number of bytes written to the log file.
func (tl *threadLogger) Size() int {
	response, err := tl.mm.Send(SizeTag, nil)
	if err != nil {
		jww.FATAL.Panicf("[LOG] Failed to file size from worker: %+v", err)
	}

	return int(binary.LittleEndian.Uint64(response))
}

// Worker always returns nil
func (tl *threadLogger) Worker() *worker.Manager {
	return nil
}
