////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package logging

import (
	"fmt"
	"github.com/armon/circbuf"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/xxdk-wasm/utils"
	"io"
	"syscall/js"
)

// LogToFile enables logging to a file that can be downloaded.
func LogToFile(threshold jww.Threshold, logFileName string,
	maxLogFileSize int) (*LogFile2, error) {
	if threshold < jww.LevelTrace || threshold > jww.LevelFatal {
		return nil,
			errors.Errorf("log level is not valid: log level: %d", threshold)
	}

	// Create new log file output
	lf, err := NewLogFile(logFileName, threshold, maxLogFileSize)
	if err != nil {
		return nil, err
	}

	AddLogListener(lf.Listen)

	msg := fmt.Sprintf("Outputting log to file %s of max size %d with level %s",
		lf.Name(), lf.MaxSize(), threshold)
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

	return lf, nil
}

// LogToFileJS enables logging to a file that can be downloaded.
//
// Parameters:
//   - args[0] - Log level (int).
//   - args[1] - Log file name (string).
//   - args[2] - Max log file size, in bytes (int).
//
// Returns:
//   - A Javascript representation of the [LogFile2] object, which allows
//     accessing the contents of the log file and other metadata.
//   - Throws a TypeError if starting the log file writer fails.
func LogToFileJS(_ js.Value, args []js.Value) any {
	threshold := jww.Threshold(args[0].Int())
	logFileName := args[1].String()
	maxLogFileSize := args[2].Int()

	lf, err := LogToFile(threshold, logFileName, maxLogFileSize)
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return NewLogFileJS(lf)
}

////////////////////////////////////////////////////////////////////////////////
// Log File Log Listener                                                      //
////////////////////////////////////////////////////////////////////////////////

// LogFile2 represents a virtual log file in memory. It contains a circular
// buffer that limits the log file, overwriting the oldest logs.
type LogFile2 struct {
	name      string
	threshold jww.Threshold
	b         *circbuf.Buffer
}

// NewLogFile initialises a new [LogFile2] for log writing.
func NewLogFile(
	name string, threshold jww.Threshold, maxSize int) (*LogFile2, error) {
	// Create new buffer of the specified size
	b, err := circbuf.NewBuffer(int64(maxSize))
	if err != nil {
		return nil, err
	}

	return &LogFile2{
		name:      name,
		threshold: threshold,
		b:         b,
	}, nil
}

// NewLogFileJS creates a new Javascript compatible object (map[string]any) that
// matches the [LogFile2] structure.
func NewLogFileJS(lf *LogFile2) map[string]any {
	logFile := map[string]any{
		"Name":      js.FuncOf(lf.NameJS),
		"Threshold": js.FuncOf(lf.ThresholdJS),
		"GetFile":   js.FuncOf(lf.GetFileJS),
		"MaxSize":   js.FuncOf(lf.MaxSizeJS),
		"Size":      js.FuncOf(lf.SizeJS),
	}

	return logFile
}

// Listen is called for every logging event. This function adheres to the
// [jwalterweatherman.LogListener] type.
func (lf *LogFile2) Listen(t jww.Threshold) io.Writer {
	if t < lf.threshold {
		return nil
	}

	return lf.b
}

// Name returns the name of the log file.
func (lf *LogFile2) Name() string { return lf.name }

// Threshold returns the log level threshold used in the file.
func (lf *LogFile2) Threshold() jww.Threshold { return lf.threshold }

// GetFile returns the entire log file.
func (lf *LogFile2) GetFile() []byte { return lf.b.Bytes() }

// MaxSize returns the max size, in bytes, that the log file is allowed to be.
func (lf *LogFile2) MaxSize() int { return int(lf.b.Size()) }

// Size returns the current size, in bytes, written to the log file.
func (lf *LogFile2) Size() int { return int(lf.b.TotalWritten()) }

// NameJS returns the name of the log file.
//
// Returns:
//   - File name (string).
func (lf *LogFile2) NameJS(js.Value, []js.Value) any {
	return lf.Name()
}

// ThresholdJS returns the log level threshold used in the file.
//
// Returns:
//   - Log level (string).
func (lf *LogFile2) ThresholdJS(js.Value, []js.Value) any {
	return lf.Threshold().String()
}

// GetFileJS returns the entire log file.
//
// Returns:
//   - Log file contents (string).
func (lf *LogFile2) GetFileJS(js.Value, []js.Value) any {
	return string(lf.GetFile())
}

// MaxSizeJS returns the max size, in bytes, that the log file is allowed to be.
//
// Returns:
//   - Max file size (int).
func (lf *LogFile2) MaxSizeJS(js.Value, []js.Value) any {
	return lf.MaxSize()
}

// SizeJS returns the current size, in bytes, written to the log file.
//
// Returns:
//   - Current file size (int).
func (lf *LogFile2) SizeJS(js.Value, []js.Value) any {
	return lf.Size()
}