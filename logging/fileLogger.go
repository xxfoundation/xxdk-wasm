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

// FileLogger manages the recording of jwalterweatherman logs to the local
// in-memory file buffer.
type FileLogger struct {
	threshold      jww.Threshold
	maxLogFileSize int
	cb             *circbuf.Buffer
}

// NewFileLogger starts logging to a local, in-memory log file at the specified
// threshold. Returns a [FileLogger] that can be used to get the log file.
func NewFileLogger(threshold jww.Threshold, maxLogFileSize int) (*FileLogger, error) {
	fl := &FileLogger{
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
func (fl *FileLogger) Write(p []byte) (n int, err error) {
	return fl.cb.Write(p)
}

// Listen adheres to the [jwalterweatherman.LogListener] type and returns the
// log writer when the threshold is within the set threshold limit.
func (fl *FileLogger) Listen(t jww.Threshold) io.Writer {
	if t < fl.threshold {
		return nil
	}
	return fl
}

// StopLogging stops log message writes. Once logging is stopped, it cannot be
// resumed and the log file cannot be recovered.
func (fl *FileLogger) StopLogging() {
	fl.threshold = 20
}

// GetFile returns the entire log file.
func (fl *FileLogger) GetFile() []byte {
	return fl.cb.Bytes()
}

// Threshold returns the log level threshold used in the file.
func (fl *FileLogger) Threshold() jww.Threshold {
	return fl.threshold
}

// MaxSize returns the max size, in bytes, that the log file is allowed to be.
func (fl *FileLogger) MaxSize() int {
	return fl.maxLogFileSize
}

// Size returns the current size, in bytes, written to the log file.
func (fl *FileLogger) Size() int {
	return int(fl.cb.Size())
}

// Worker returns nil.
func (fl *FileLogger) Worker() *worker.Manager {
	return nil
}
