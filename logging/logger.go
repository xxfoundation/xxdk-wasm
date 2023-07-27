////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package logging

import (
	"github.com/pkg/errors"
	"syscall/js"

	jww "github.com/spf13/jwalterweatherman"

	"gitlab.com/elixxir/xxdk-wasm/utils"
	"gitlab.com/elixxir/xxdk-wasm/worker"
)

// List of tags that can be used when sending a message or registering a handler
// to receive a message.
const (
	NewLogFileTag worker.Tag = "NewLogFile"
	WriteLogTag   worker.Tag = "WriteLog"
	GetFileTag    worker.Tag = "GetFile"
	GetFileExtTag worker.Tag = "GetFileExt"
	SizeTag       worker.Tag = "Size"
)

// logger is the global that all jwalterweatherman logging is sent to.
var logger Logger

// GetLogger returns the Logger object, used to manager where logging is
// recorded.
func GetLogger() Logger {
	return logger
}

type Logger interface {
	// StopLogging stops log message writes. Once logging is stopped, it cannot
	// be resumed and the log file cannot be recovered.
	StopLogging()

	// GetFile returns the entire log file.
	GetFile() []byte

	// Threshold returns the log level threshold used in the file.
	Threshold() jww.Threshold

	// MaxSize returns the maximum size, in bytes, of the log file before it
	// rolls over and starts overwriting the oldest entries
	MaxSize() int

	// Size returns the number of bytes written to the log file.
	Size() int

	// Worker returns the manager for the Javascript Worker object. If the
	// worker has not been initialized, it returns nil.
	Worker() *worker.Manager
}

// EnableLogging enables logging to the Javascript console and to a local or
// worker file buffer. This must be called only once at initialisation.
func EnableLogging(logLevel, fileLogLevel jww.Threshold, maxLogFileSizeMB int,
	workerScriptURL, workerName string) error {

	var listeners []jww.LogListener
	if logLevel > -1 {
		// Overwrites setting the log level to INFO done in bindings so that the
		// Javascript console can be used
		ll := NewJsConsoleLogListener(logLevel)
		listeners = append(listeners, ll.Listen)
		jww.SetStdoutThreshold(jww.LevelFatal + 1)
		jww.FEEDBACK.Printf("[LOG] Log level for console set to %s", logLevel)
	} else {
		jww.FEEDBACK.Print("[LOG] Disabling logging to console.")
	}

	if fileLogLevel > -1 {
		maxLogFileSize := maxLogFileSizeMB * 1_000_000
		if workerScriptURL == "" {
			fl, err := newFileLogger(fileLogLevel, maxLogFileSize)
			if err != nil {
				return errors.Wrap(err, "could not initialize logging to file")
			}
			listeners = append(listeners, fl.Listen)
		} else {
			wl, err := newWorkerLogger(
				fileLogLevel, maxLogFileSize, workerScriptURL, workerName)
			if err != nil {
				return errors.Wrap(err, "could not initialize logging to worker file")
			}

			listeners = append(listeners, wl.Listen)
		}

		js.Global().Set("GetLogger", js.FuncOf(GetLoggerJS))
	}
	jww.SetLogListeners(listeners...)

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Javascript Bindings                                                        //
////////////////////////////////////////////////////////////////////////////////

// GetLoggerJS returns the Logger object, used to manager where logging is
// recorded and accessing the log file.
//
// Returns:
//   - A Javascript representation of the [Logger] object.
func GetLoggerJS(js.Value, []js.Value) any {
	// l := GetLogger()
	// if l != nil {
	// 	return newLoggerJS(LoggerJS{GetLogger()})
	// }
	// return js.Null()
	return newLoggerJS(LoggerJS{GetLogger()})
}

type LoggerJS struct {
	api Logger
}

// newLoggerJS creates a new Javascript compatible object (map[string]any) that
// matches the [Logger] structure.
func newLoggerJS(l LoggerJS) map[string]any {
	logFileWorker := map[string]any{
		"StopLogging": js.FuncOf(l.StopLogging),
		"GetFile":     js.FuncOf(l.GetFile),
		"Threshold":   js.FuncOf(l.Threshold),
		"MaxSize":     js.FuncOf(l.MaxSize),
		"Size":        js.FuncOf(l.Size),
		"Worker":      js.FuncOf(l.Worker),
	}

	return logFileWorker
}

// StopLogging stops the logging of log messages and disables the log
// listener. If the log worker is running, it is terminated. Once logging is
// stopped, it cannot be resumed the log file cannot be recovered.
func (l *LoggerJS) StopLogging(js.Value, []js.Value) any {
	l.api.StopLogging()

	return nil
}

// GetFile returns the entire log file.
//
// If the log file is listening locally, it returns it from the local buffer. If
// it is listening from the worker, it blocks until the file is returned.
//
// Returns a promise:
//   - Resolves to the log file contents (string).
func (l *LoggerJS) GetFile(js.Value, []js.Value) any {
	promiseFn := func(resolve, _ func(args ...any) js.Value) {
		resolve(string(l.api.GetFile()))
	}

	return utils.CreatePromise(promiseFn)
}

// Threshold returns the log level threshold used in the file.
//
// Returns:
//   - Log level (int).
func (l *LoggerJS) Threshold(js.Value, []js.Value) any {
	return int(l.api.Threshold())
}

// MaxSize returns the max size, in bytes, that the log file is allowed to be.
//
// Returns:
//   - Max file size (int).
func (l *LoggerJS) MaxSize(js.Value, []js.Value) any {
	return l.api.MaxSize()
}

// Size returns the current size, in bytes, written to the log file.
//
// If the log file is listening locally, it returns it from the local buffer. If
// it is listening from the worker, it blocks until the size is returned.
//
// Returns a promise:
//   - Resolves to the current file size (int).
func (l *LoggerJS) Size(js.Value, []js.Value) any {
	promiseFn := func(resolve, _ func(args ...any) js.Value) {
		resolve(l.api.Size())
	}

	return utils.CreatePromise(promiseFn)
}

// Worker returns the web worker object.
//
// Returns:
//   - Javascript worker object. If the worker has not been initialized, it
//     returns null.
func (l *LoggerJS) Worker(js.Value, []js.Value) any {
	wm := l.api.Worker()
	if wm == nil {
		return js.Null()
	}

	return wm.GetWorker()
}
