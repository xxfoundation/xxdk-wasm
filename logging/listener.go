////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package logging

import (
	"github.com/armon/circbuf"
	jww "github.com/spf13/jwalterweatherman"
	"io"
	"syscall/js"
)

////////////////////////////////////////////////////////////////////////////////
// Log File Log Listener                                                      //
////////////////////////////////////////////////////////////////////////////////

// LogFile represents a virtual log file in memory. It contains a circular
// buffer that limits the log file, overwriting the oldest logs.
type LogFile struct {
	name      string
	threshold jww.Threshold
	b         *circbuf.Buffer
}

// NewLogFile initialises a new [LogFile] for log writing.
func NewLogFile(
	name string, threshold jww.Threshold, maxSize int) (*LogFile, error) {
	// Create new buffer of the specified size
	b, err := circbuf.NewBuffer(int64(maxSize))
	if err != nil {
		return nil, err
	}

	return &LogFile{
		name:      name,
		threshold: threshold,
		b:         b,
	}, nil
}

// NewLogFileJS creates a new Javascript compatible object (map[string]any) that
// matches the [LogFile] structure.
func NewLogFileJS(lf *LogFile) map[string]any {
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
func (lf *LogFile) Listen(t jww.Threshold) io.Writer {
	if t < lf.threshold {
		return nil
	}

	return lf.b
}

// Name returns the name of the log file.
func (lf *LogFile) Name() string { return lf.name }

// Threshold returns the log level threshold used in the file.
func (lf *LogFile) Threshold() jww.Threshold { return lf.threshold }

// GetFile returns the entire log file.
func (lf *LogFile) GetFile() []byte { return lf.b.Bytes() }

// MaxSize returns the max size, in bytes, that the log file is allowed to be.
func (lf *LogFile) MaxSize() int { return int(lf.b.Size()) }

// Size returns the current size, in bytes, written to the log file.
func (lf *LogFile) Size() int { return int(lf.b.TotalWritten()) }

// NameJS returns the name of the log file.
//
// Returns:
//   - File name (string).
func (lf *LogFile) NameJS(js.Value, []js.Value) any {
	return lf.Name()
}

// ThresholdJS returns the log level threshold used in the file.
//
// Returns:
//   - Log level (string).
func (lf *LogFile) ThresholdJS(js.Value, []js.Value) any {
	return lf.Threshold().String()
}

// GetFileJS returns the entire log file.
//
// Returns:
//   - Log file contents (string).
func (lf *LogFile) GetFileJS(js.Value, []js.Value) any {
	return string(lf.GetFile())
}

// MaxSizeJS returns the max size, in bytes, that the log file is allowed to be.
//
// Returns:
//   - Max file size (int).
func (lf *LogFile) MaxSizeJS(js.Value, []js.Value) any {
	return lf.MaxSize()
}

// SizeJS returns the current size, in bytes, written to the log file.
//
// Returns:
//   - Current file size (int).
func (lf *LogFile) SizeJS(js.Value, []js.Value) any {
	return lf.Size()
}
