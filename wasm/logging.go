////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"fmt"
	"github.com/armon/circbuf"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/v4/bindings"
	"gitlab.com/elixxir/xxdk-wasm/utils"
	"io"
	"log"
	"syscall/js"
)

// logListeners is a list of all registered log listeners. This is used to add
// additional log listener without overwriting previously registered listeners.
var logListeners []jww.LogListener

// LogLevel sets level of logging. All logs at the set level and below will be
// displayed (e.g., when log level is ERROR, only ERROR, CRITICAL, and FATAL
// messages will be printed).
//
// Log level options:
//	TRACE    - 0
//	DEBUG    - 1
//	INFO     - 2
//	WARN     - 3
//	ERROR    - 4
//	CRITICAL - 5
//	FATAL    - 6
//
// The default log level without updates is INFO.
//
// Parameters:
//  - args[0] - Log level (int).
//
// Returns:
//  - Throws TypeError if the log level is invalid.
func LogLevel(_ js.Value, args []js.Value) interface{} {
	threshold := jww.Threshold(args[0].Int())
	if threshold < jww.LevelTrace || threshold > jww.LevelFatal {
		err := errors.Errorf("log level is not valid: log level: %d", threshold)
		utils.Throw(utils.TypeError, err)
		return nil
	}

	jww.SetLogThreshold(threshold)
	jww.SetFlags(log.LstdFlags | log.Lmicroseconds)

	ll := NewJsConsoleLogListener(threshold)
	logListeners = append(logListeners, ll.Listen)
	jww.SetLogListeners(logListeners...)
	jww.SetStdoutThreshold(jww.LevelFatal + 1)

	msg := fmt.Sprintf("Log level set to: %s", threshold)
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

	return nil
}

// LogToFile enables logging to a file that can be downloaded.
//
// Parameters:
//  - args[0] - Log level (int).
//  - args[1] - Log file name (string).
//  - args[2] - Max log file size, in bytes (int).
//
// Returns:
//  - A Javascript representation of the [LogFile] object, which allows
//    accessing the contents of the log file and other metadata.
func LogToFile(_ js.Value, args []js.Value) interface{} {
	threshold := jww.Threshold(args[0].Int())
	if threshold < jww.LevelTrace || threshold > jww.LevelFatal {
		err := errors.Errorf("log level is not valid: log level: %d", threshold)
		utils.Throw(utils.TypeError, err)
		return nil
	}

	// Create new log file output
	ll, err := NewLogFile(args[1].String(), threshold, args[2].Int())
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	logListeners = append(logListeners, ll.Listen)
	jww.SetLogListeners(logListeners...)

	msg := fmt.Sprintf("Outputting log to file %s of max size %d with level %s",
		ll.name, ll.b.Size(), threshold)
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

	return newLogFileJS(ll)
}

// logWriter wraps Javascript callbacks to adhere to the [bindings.LogWriter]
// interface.
type logWriter struct {
	log func(args ...interface{}) js.Value
}

// Log returns a log message to pass to the log writer.
//
// Parameters:
//  - s - Log message (string).
func (lw *logWriter) Log(s string) { lw.log(s) }

// RegisterLogWriter registers a callback on which logs are written.
//
// Parameters:
//  - args[0] - a function that accepts a string and writes to a log. It must be
//    of the form func(string).
func RegisterLogWriter(_ js.Value, args []js.Value) interface{} {
	bindings.RegisterLogWriter(&logWriter{args[0].Invoke})
	return nil
}

// EnableGrpcLogs sets GRPC trace logging.
//
// Parameters:
//  - args[0] - a function that accepts a string and writes to a log. It must be
//    of the form func(string).
func EnableGrpcLogs(_ js.Value, args []js.Value) interface{} {
	bindings.EnableGrpcLogs(&logWriter{args[0].Invoke})
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Javascript Console Log Listener                                            //
////////////////////////////////////////////////////////////////////////////////

// console contains the Javascript console object, which provides access to the
// browser's debugging console. This structure detects logging types and prints
// it using the correct logging method.
type console struct {
	call string
	js.Value
}

// Write writes the data to the Javascript console at the level specified by the
// call.
func (c *console) Write(p []byte) (n int, err error) {
	c.Call(c.call, string(p))
	return len(p), nil
}

// JsConsoleLogListener redirects log output to the Javascript console.
type JsConsoleLogListener struct {
	jww.Threshold
	js.Value

	trace    *console
	debug    *console
	info     *console
	error    *console
	warn     *console
	critical *console
	fatal    *console
	def      *console
}

// NewJsConsoleLogListener initialises a new log listener that listener for the
// specific threshold and prints the logs to the Javascript console.
func NewJsConsoleLogListener(threshold jww.Threshold) *JsConsoleLogListener {
	consoleObj := js.Global().Get("console")
	return &JsConsoleLogListener{
		Threshold: threshold,
		Value:     consoleObj,
		trace:     &console{"debug", consoleObj},
		debug:     &console{"log", consoleObj},
		info:      &console{"info", consoleObj},
		warn:      &console{"warn", consoleObj},
		error:     &console{"error", consoleObj},
		critical:  &console{"error", consoleObj},
		fatal:     &console{"error", consoleObj},
		def:       &console{"log", consoleObj},
	}
}

// Listen is called for every logging event. This function adheres to the
// [jwalterweatherman.LogListener] type.
func (ll *JsConsoleLogListener) Listen(t jww.Threshold) io.Writer {
	if t < ll.Threshold {
		return nil
	}

	switch t {
	case jww.LevelTrace:
		return ll.trace
	case jww.LevelDebug:
		return ll.debug
	case jww.LevelInfo:
		return ll.info
	case jww.LevelWarn:
		return ll.warn
	case jww.LevelError:
		return ll.error
	case jww.LevelCritical:
		return ll.critical
	case jww.LevelFatal:
		return ll.fatal
	default:
		return ll.def
	}
}

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

// newLogFileJS creates a new Javascript compatible object
// (map[string]interface{}) that matches the [LogFile] structure.
func newLogFileJS(lf *LogFile) map[string]interface{} {
	logFile := map[string]interface{}{
		"Name":      js.FuncOf(lf.Name),
		"Threshold": js.FuncOf(lf.Threshold),
		"GetFile":   js.FuncOf(lf.GetFile),
		"MaxSize":   js.FuncOf(lf.MaxSize),
		"Size":      js.FuncOf(lf.Size),
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
//
// Returns:
//  - File name (string).
func (lf *LogFile) Name(js.Value, []js.Value) interface{} {
	return lf.name
}

// Threshold returns the log level threshold used in the file.
//
// Returns:
//  - Log level (string).
func (lf *LogFile) Threshold(js.Value, []js.Value) interface{} {
	return lf.threshold.String()
}

// GetFile returns the entire log file.
//
// Returns:
//  - Log file contents (string).
func (lf *LogFile) GetFile(js.Value, []js.Value) interface{} {
	return string(lf.b.Bytes())
}

// MaxSize returns the max size, in bytes, that the log file is allowed to be.
//
// Returns:
//  - Max file size (int).
func (lf *LogFile) MaxSize(js.Value, []js.Value) interface{} {
	return lf.b.Size()
}

// Size returns the current size, in bytes, written to the log file.
//
// Returns:
//  - Current file size (int).
func (lf *LogFile) Size(js.Value, []js.Value) interface{} {
	return lf.b.TotalWritten()
}
