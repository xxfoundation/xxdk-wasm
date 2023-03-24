////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"gitlab.com/elixxir/client/v4/bindings"
	"gitlab.com/elixxir/xxdk-wasm/src/api/logging"
	"syscall/js"
)

// LogLevel sets level of logging. All logs at the set level and below will be
// displayed (e.g., when log level is ERROR, only ERROR, CRITICAL, and FATAL
// messages will be printed).
//
// Log level options:
//
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
//   - args[0] - Log level (int).
//
// Returns:
//   - Throws TypeError if the log level is invalid.
func LogLevel(this js.Value, args []js.Value) any {
	return logging.LogLevelJS(this, args)
}

// logWriter wraps Javascript callbacks to adhere to the [bindings.LogWriter]
// interface.
type logWriter struct {
	log func(args ...any) js.Value
}

// Log returns a log message to pass to the log writer.
//
// Parameters:
//   - s - Log message (string).
func (lw *logWriter) Log(s string) { lw.log(s) }

// RegisterLogWriter registers a callback on which logs are written.
//
// Parameters:
//   - args[0] - a function that accepts a string and writes to a log. It must
//     be of the form func(string).
func RegisterLogWriter(_ js.Value, args []js.Value) any {
	bindings.RegisterLogWriter(&logWriter{args[0].Invoke})
	return nil
}

// EnableGrpcLogs sets GRPC trace logging.
//
// Parameters:
//   - args[0] - a function that accepts a string and writes to a log. It must
//     be of the form func(string).
func EnableGrpcLogs(_ js.Value, args []js.Value) any {
	bindings.EnableGrpcLogs(&logWriter{args[0].Invoke})
	return nil
}
