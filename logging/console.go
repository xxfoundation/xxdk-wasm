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
	"syscall/js"

	jww "github.com/spf13/jwalterweatherman"
)

var consoleObj = js.Global().Get("console")

// Console contains the Javascript console object, which provides access to the
// browser's debugging console. This structure is defined for only a single
// method on the console object. For example, if the method is set to debug,
// then all calls to console.Write will print a debug message to the Javascript
// console.
//
// Doc: https://developer.mozilla.org/en-US/docs/Web/API/console
type Console struct {
	method string
	js.Value
}

// Write writes the data to the Javascript console with preset method. Returns
// the number of bytes written.
func (c *Console) Write(p []byte) (n int, err error) {
	c.Call(c.method, string(p))
	return len(p), nil
}

// JsConsoleLogListener redirects log output to the Javascript console using the
// correct console method.
type JsConsoleLogListener struct {
	jww.Threshold
	js.Value

	trace    *Console
	debug    *Console
	info     *Console
	error    *Console
	warn     *Console
	critical *Console
	fatal    *Console
	def      *Console
}

// NewJsConsoleLogListener initialises a new log listener that listener for the
// specific threshold and prints the logs to the Javascript console.
func NewJsConsoleLogListener(threshold jww.Threshold) *JsConsoleLogListener {
	return &JsConsoleLogListener{
		Threshold: threshold,
		Value:     consoleObj,
		trace:     &Console{"debug", consoleObj},
		debug:     &Console{"log", consoleObj},
		info:      &Console{"info", consoleObj},
		warn:      &Console{"warn", consoleObj},
		error:     &Console{"error", consoleObj},
		critical:  &Console{"error", consoleObj},
		fatal:     &Console{"error", consoleObj},
		def:       &Console{"log", consoleObj},
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
