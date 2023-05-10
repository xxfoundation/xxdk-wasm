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
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/xxdk-wasm/worker"
	"io"
	"math"
	"time"
)

// threadLogger manages the recording of jwalterweatherman logs from a worker to
// the in-memory file buffer in a remote Worker thread.
type threadLogger struct {
	threshold jww.Threshold
	channel   worker.Channel
	wtm       *worker.ThreadManager
}

// newThreadLogger starts logging to an in-memory log file in a remote Worker
// at the specified threshold. Returns a [threadLogger] that can be used to get
// the log file.
func newThreadLogger(threshold jww.Threshold, channel worker.Channel,
	wtm *worker.ThreadManager) (*threadLogger, error) {
	tl := &threadLogger{
		threshold: threshold,
		channel:   channel,
		wtm:       wtm,
	}

	// Wait for ChannelMessage to be created
	channelCreatedChan := make(chan struct{})
	wtm.RegisterChannelCreatedCB(
		channel, func() { channelCreatedChan <- struct{}{} })
	select {
	case <-channelCreatedChan:
	case <-time.After(worker.ResponseTimeout):
		return nil, errors.Errorf("timed out after %s waiting for "+
			"ChannelMessage %s to be created", worker.ResponseTimeout, channel)
	}

	readyChan := make(chan []byte)
	tl.wtm.RegisterCallback(WorkerReady, func(data []byte) ([]byte, error) {
		readyChan <- data
		return nil, nil
	})

	tl.wtm.SendMessage(WorkerReady, channel, []byte{})

	select {
	case <-readyChan:

	case <-time.After(worker.ResponseTimeout):
		return nil, errors.Errorf("timed out after %s waiting for response "+
			"from the logger worker that it is ready", worker.ResponseTimeout)
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
	go tl.wtm.SendMessageQuiet(WriteLogTag, tl.channel, p)
	return len(p), nil
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
	fileChan := make(chan []byte)
	tl.wtm.RegisterCallback(GetFileTag, func(data []byte) ([]byte, error) {
		fileChan <- data
		return nil, nil
	})

	tl.wtm.SendMessage(GetFileTag, tl.channel, []byte{})

	select {
	case file := <-fileChan:
		return file
	case <-time.After(worker.ResponseTimeout):
		jww.FATAL.Panicf("[LOG] Timed out after %s waiting for log "+
			"file from worker", worker.ResponseTimeout)
		return nil
	}
}

// Threshold returns the log level threshold of logs sent to the worker.
func (tl *threadLogger) Threshold() jww.Threshold {
	return tl.threshold
}

// MaxSize returns the max size, in bytes, that the log file is allowed to be.
func (tl *threadLogger) MaxSize() int {
	maxSizeChan := make(chan []byte)
	tl.wtm.RegisterCallback(GetFileTag, func(data []byte) ([]byte, error) {
		maxSizeChan <- data
		return nil, nil
	})

	tl.wtm.SendMessage(MaxSizeTag, tl.channel, []byte{})

	select {
	case data := <-maxSizeChan:
		return int(binary.LittleEndian.Uint64(data))
	case <-time.After(worker.ResponseTimeout):
		jww.FATAL.Panicf("[LOG] Timed out after %s waiting for log "+
			"max file size from worker", worker.ResponseTimeout)
		return 0
	}
}

// Size returns the number of bytes written to the log file.
func (tl *threadLogger) Size() int {
	sizeChan := make(chan []byte)
	tl.wtm.RegisterCallback(GetFileTag, func(data []byte) ([]byte, error) {
		sizeChan <- data
		return nil, nil
	})

	tl.wtm.SendMessage(MaxSizeTag, tl.channel, []byte{})

	select {
	case data := <-sizeChan:
		return int(binary.LittleEndian.Uint64(data))
	case <-time.After(worker.ResponseTimeout):
		jww.FATAL.Panicf("[LOG] Timed out after %s waiting for log "+
			"file size from worker", worker.ResponseTimeout)
		return 0
	}
}

// Worker always returns nil
func (tl *threadLogger) Worker() *worker.Manager {
	return nil
}
