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
	"log"
	"syscall/js"
)

// LogLevel sets level of logging. All logs at the set level and below will be
// displayed (e.g., when log level is ERROR, only ERROR, CRITICAL, and FATAL
// messages will be printed).
//
// The default log level without updates is INFO.
func LogLevel(threshold jww.Threshold) error {
	if threshold < jww.LevelTrace || threshold > jww.LevelFatal {
		return errors.Errorf("log level is not valid: log level: %d", threshold)
	}

	jww.SetLogThreshold(threshold)
	jww.SetFlags(log.LstdFlags | log.Lmicroseconds)

	ll := NewJsConsoleLogListener(threshold)
	jww.SetLogListeners(AddLogListener(ll.Listen)...)
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

// LogLevelJS sets level of logging. All logs at the set level and below will be
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
func LogLevelJS(_ js.Value, args []js.Value) any {
	threshold := jww.Threshold(args[0].Int())
	err := LogLevel(threshold)
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return nil
}
