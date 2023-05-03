////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package logging

import (
	"io"

	"github.com/armon/circbuf"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"

	"gitlab.com/elixxir/xxdk-wasm/worker"
)

// TODO: test

// fileLogger manages the recording of jwalterweatherman logs to the local
// in-memory file buffer.
type fileLogger struct {
	threshold      jww.Threshold
	maxLogFileSize int
	cb             *circbuf.Buffer
}

// newFileLogger starts logging to a local, in-memory log file at the specified
// threshold. Returns a [fileLogger] that can be used to get the log file.
func newFileLogger(threshold jww.Threshold, maxLogFileSize int) (*fileLogger, error) {
	fl := &fileLogger{
		threshold:      threshold,
		maxLogFileSize: maxLogFileSize,
	}

	b, err := circbuf.NewBuffer(int64(maxLogFileSize))
	if err != nil {
		return nil, errors.Wrap(err, "could not create new circular buffer")
	}
	fl.cb = b

	jww.FEEDBACK.Printf("[LOG] Outputting log to file of max size %d at level %s",
		fl.maxLogFileSize, fl.threshold)

	logger = fl
	return fl, nil
}

// Write adheres to the io.Writer interface and writes log entries to the
// buffer.
func (fl *fileLogger) Write(p []byte) (n int, err error) {
	return fl.cb.Write(p)
}

// Listen adheres to the [jwalterweatherman.LogListener] type and returns the
// log writer when the threshold is within the set threshold limit.
func (fl *fileLogger) Listen(t jww.Threshold) io.Writer {
	if t < fl.threshold {
		return nil
	}
	return fl
}

// StopLogging stops log message writes. Once logging is stopped, it cannot be
// resumed and the log file cannot be recovered.
func (fl *fileLogger) StopLogging() {
	fl.threshold = 20
}

// GetFile returns the entire log file.
func (fl *fileLogger) GetFile() []byte {
	return fl.cb.Bytes()
}

// Threshold returns the log level threshold used in the file.
func (fl *fileLogger) Threshold() jww.Threshold {
	return fl.threshold
}

// MaxSize returns the max size, in bytes, that the log file is allowed to be.
func (fl *fileLogger) MaxSize() int {
	return fl.maxLogFileSize
}

// Size returns the current size, in bytes, written to the log file.
func (fl *fileLogger) Size() int {
	return int(fl.cb.Size())
}

// Worker returns nil.
func (fl *fileLogger) Worker() *worker.Manager {
	return nil
}
