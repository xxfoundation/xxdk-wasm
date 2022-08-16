////////////////////////////////////////////////////////////////////////////////
// Copyright © 2020 xx network SEZC                                           //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file                                                               //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"gitlab.com/elixxir/client/bindings"
	"syscall/js"
)

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
//  - args[0] - log level (int).
//
// Returns:
//  - Throws TypeError if the log level is invalid.
func LogLevel(_ js.Value, args []js.Value) interface{} {
	err := bindings.LogLevel(args[0].Int())
	if err != nil {
		Throw(TypeError, err.Error())
		return nil
	}

	return nil
}

// logWriter wraps Javascript callbacks to adhere to the [bindings.LogWriter]
// interface.
type logWriter struct {
	log func(args ...interface{}) js.Value
}

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
