////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/bindings"
	"gitlab.com/elixxir/xxdk-wasm/utils"
	"io"
	"log"
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
	level := args[0].Int()
	if level < 0 || level > 6 {
		err := errors.Errorf("log level is not valid: log level: %d", level)
		utils.Throw(utils.TypeError, err)
		return nil
	}

	threshold := jww.Threshold(level)
	jww.SetLogThreshold(threshold)
	jww.SetFlags(log.LstdFlags | log.Lmicroseconds)

	ll := &LogListener{threshold, js.Global().Get("console")}
	jww.SetLogListeners(ll.Listen)
	jww.SetStdoutThreshold(jww.LevelFatal + 1)

	switch threshold {
	case jww.LevelTrace:
		fallthrough
	case jww.LevelDebug:
		fallthrough
	case jww.LevelInfo:
		jww.INFO.Printf("Log level set to: %s", threshold)
	case jww.LevelWarn:
		jww.WARN.Printf("Log level set to: %s", threshold)
	case jww.LevelError:
		jww.ERROR.Printf("Log level set to: %s", threshold)
	case jww.LevelCritical:
		jww.CRITICAL.Printf("Log level set to: %s", threshold)
	case jww.LevelFatal:
		jww.FATAL.Printf("Log level set to: %s", threshold)
	}

	return nil
}

// console contains the Javascript console object, which provides access to the
// browser's debugging console. This structure detects logging types and prints
// it using the correct logging method.
type console struct {
	call string
	js.Value
}

func (c *console) Write(p []byte) (n int, err error) {
	c.Call(c.call, string(p))
	return len(p), nil
}

type LogListener struct {
	jww.Threshold
	js.Value
}

func (ll *LogListener) Listen(t jww.Threshold) io.Writer {
	if t < ll.Threshold {
		return nil
	}

	switch t {
	case jww.LevelTrace:
		return &console{"debug", ll.Value}
	case jww.LevelDebug:
		return &console{"log", ll.Value}
	case jww.LevelInfo:
		return &console{"info", ll.Value}
	case jww.LevelWarn:
		return &console{"warn", ll.Value}
	case jww.LevelError:
		return &console{"error", ll.Value}
	case jww.LevelCritical:
		return &console{"error", ll.Value}
	case jww.LevelFatal:
		return &console{"error", ll.Value}
	default:
		return &console{"log", ll.Value}
	}
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
