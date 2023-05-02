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
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/xxdk-wasm/logging"
	"gitlab.com/elixxir/xxdk-wasm/worker"
	"os"
	"syscall/js"
)

// SEMVER is the current semantic version of the xxDK Logger web worker.
const SEMVER = "0.1.0"

func init() {
	// Set up Javascript console listener set at level INFO
	ll := logging.NewJsConsoleLogListener(jww.LevelDebug)
	logging.AddLogListener(ll.Listen)
	jww.SetStdoutThreshold(jww.LevelFatal + 1)
}

// workerLogFile manages communication with the main thread and writing incoming
// logging messages to the log file.
type workerLogFile struct {
	wtm *worker.ThreadManager
	b   *circbuf.Buffer
}

func main() {
	// Set to os.Args because the default is os.Args[1:] and in WASM, args start
	// at 0, not 1.
	LoggerCmd.SetArgs(os.Args)

	err := LoggerCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var LoggerCmd = &cobra.Command{
	Use:     "Logger",
	Short:   "Web worker buffer file logger",
	Example: "const go = new Go();\ngo.argv = [\"--logLevel=1\"]",
	Run: func(cmd *cobra.Command, args []string) {
		// Start logger first to capture all logging events
		err := logging.EnableLogging(logLevel, -1, 0, "", "")
		if err != nil {
			fmt.Printf("Failed to intialize logging: %+v", err)
			os.Exit(1)
		}

		jww.INFO.Printf("xxDK Logger web worker version: v%s", SEMVER)

		jww.INFO.Print("[LOG] Starting xxDK WebAssembly Logger Worker.")

		wlf := workerLogFile{wtm: worker.NewThreadManager("Logger", false)}

		wlf.registerCallbacks()

		wlf.wtm.SignalReady()

		// Indicate to the Javascript caller that the WASM is ready by resolving
		// a promise created by the caller.
		js.Global().Get("onWasmInitialized").Invoke()

		<-make(chan bool)
		fmt.Println("[WW] Closing xxDK WebAssembly Log Worker.")
		os.Exit(0)
	},
}

var (
	logLevel jww.Threshold
)

func init() {
	// Initialize all startup flags
	LoggerCmd.Flags().IntVarP((*int)(&logLevel), "logLevel", "l", 2,
		"Sets the log level output when outputting to the Javascript console. "+
			"Set to -1 to disable.")
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
