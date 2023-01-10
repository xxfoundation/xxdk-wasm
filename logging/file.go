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
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/xxdk-wasm/utils"
	"syscall/js"
)

// logListeners is a list of all registered log listeners. This is used to add
// additional log listener without overwriting previously registered listeners.
var logListeners []jww.LogListener

// AddLogListener appends to the log listener list. Call this and pass the
// return into jwalterweatherman.SetLogListeners.
func AddLogListener(ll jww.LogListener) []jww.LogListener {
	logListeners = append(logListeners, ll)
	return logListeners
}

// LogToFile enables logging to a file that can be downloaded.
func LogToFile(threshold jww.Threshold, logFileName string,
	maxLogFileSize int) (*LogFile, error) {
	if threshold < jww.LevelTrace || threshold > jww.LevelFatal {
		return nil,
			errors.Errorf("log level is not valid: log level: %d", threshold)
	}

	// Create new log file output
	lf, err := NewLogFile(logFileName, threshold, maxLogFileSize)
	if err != nil {
		return nil, err
	}

	jww.SetLogListeners(AddLogListener(lf.Listen)...)

	msg := fmt.Sprintf("Outputting log to file %s of max size %d with level %s",
		lf.Name(), lf.Size(), threshold)
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
//   - A Javascript representation of the [LogFile] object, which allows
//     accessing the contents of the log file and other metadata.
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
