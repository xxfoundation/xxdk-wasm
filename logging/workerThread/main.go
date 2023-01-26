////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/armon/circbuf"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/xxdk-wasm/logging"
	"gitlab.com/elixxir/xxdk-wasm/worker"
	"syscall/js"
)

// SEMVER is the current semantic version of the xxDK Logger web worker.
const SEMVER = "0.1.0"

func init() {
	// Set up Javascript console listener set at level INFO
	ll := logging.NewJsConsoleLogListener(jww.LevelDebug)
	logging.AddLogListener(ll.Listen)
	jww.SetStdoutThreshold(jww.LevelFatal + 1)
	jww.INFO.Printf("xxDK Logger web worker version: v%s", SEMVER)
}

// workerLogFile manages communication with the main thread and writing incoming
// logging messages to the log file.
type workerLogFile struct {
	wtm *worker.ThreadManager
	b   *circbuf.Buffer
}

func main() {
	jww.INFO.Print("[LOG] Starting xxDK WebAssembly Logger Worker.")

	js.Global().Set("LogLevel", js.FuncOf(logging.LogLevelJS))

	wlf := workerLogFile{wtm: worker.NewThreadManager("Logger", false)}

	wlf.registerCallbacks()

	wlf.wtm.SignalReady()
	<-make(chan bool)
	fmt.Println("[WW] Closing xxDK WebAssembly Log Worker.")
}

// registerCallbacks registers all the necessary callbacks for the main thread
// to get the file and file metadata.
func (wlf *workerLogFile) registerCallbacks() {
	// Callback for logging.LogToFileWorker
	wlf.wtm.RegisterCallback(logging.NewLogFileTag,
		func(data []byte) ([]byte, error) {
			var maxLogFileSize int64
			err := json.Unmarshal(data, &maxLogFileSize)
			if err != nil {
				return []byte(err.Error()), err
			}

			wlf.b, err = circbuf.NewBuffer(maxLogFileSize)
			if err != nil {
				return []byte(err.Error()), err
			}

			jww.DEBUG.Printf("[LOG] Created new worker log file of size %d",
				maxLogFileSize)

			return []byte{}, nil
		})

	// Callback for Logging.GetFile
	wlf.wtm.RegisterCallback(logging.WriteLogTag,
		func(data []byte) ([]byte, error) {
			n, err := wlf.b.Write(data)
			if err != nil {
				return nil, err
			} else if n != len(data) {
				return nil, errors.Errorf(
					"wrote %d bytes; expected %d bytes", n, len(data))
			}

			return nil, nil
		},
	)

	// Callback for Logging.GetFile
	wlf.wtm.RegisterCallback(logging.GetFileTag, func([]byte) ([]byte, error) {
		return wlf.b.Bytes(), nil
	})

	// Callback for Logging.GetFile
	wlf.wtm.RegisterCallback(logging.GetFileExtTag, func([]byte) ([]byte, error) {
		return wlf.b.Bytes(), nil
	})

	// Callback for Logging.Size
	wlf.wtm.RegisterCallback(logging.SizeTag, func([]byte) ([]byte, error) {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(wlf.b.TotalWritten()))
		return b, nil
	})
}
