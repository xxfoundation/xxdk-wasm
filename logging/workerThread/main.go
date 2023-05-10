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
	"os"
	"syscall/js"

	"github.com/armon/circbuf"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"

	"gitlab.com/elixxir/xxdk-wasm/logging"
	"gitlab.com/elixxir/xxdk-wasm/worker"
)

// SEMVER is the current semantic version of the xxDK Logger web worker.
const SEMVER = "0.1.0"

// workerLogFile manages communication with the main thread and writing incoming
// logging messages to the log file.
type workerLogFile struct {
	wtm *worker.ThreadManager
	b   *circbuf.Buffer
}

func main() {
	// Set to os.Args because the default is os.Args[1:] and in WASM, args start
	// at 0, not 1.
	loggerCmd.SetArgs(os.Args)

	err := loggerCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var loggerCmd = &cobra.Command{
	Use:     "Logger",
	Short:   "Web worker buffer file logger",
	Example: "const go = new Go();\ngo.argv = [\"--logLevel=1\"]",
	Run: func(cmd *cobra.Command, args []string) {
		// Set up basic logging while the worker thread manager is initialised
		ll := logging.NewJsConsoleLogListener(jww.LevelInfo)
		logging.AddLogListener(ll.Listen)
		jww.SetStdoutThreshold(jww.LevelFatal + 1)

		jww.INFO.Printf("xxDK Logger web worker version: v%s", SEMVER)

		jww.INFO.Print("[LOG] Starting xxDK WebAssembly Logger Worker.")

		wlf := workerLogFile{wtm: worker.NewThreadManager("Logger", false)}

		wlf.registerCallbacks()

		wlf.wtm.SignalReady()

		err := logging.EnableWorkerLogging(
			logLevel, fileLogLevel, maxLogFileSizeMB, nil)
		if err != nil {
			fmt.Printf("Failed to intialize logging in logger worker: %+v", err)
			os.Exit(1)
		}

		// Indicate to the Javascript caller that the WASM is ready by resolving
		// a promise created by the caller.
		js.Global().Get("onWasmInitialized").Invoke()

		<-make(chan bool)
		fmt.Println("[WW] Closing xxDK WebAssembly Log Worker.")
		os.Exit(0)
	},
}

var (
	logLevel, fileLogLevel jww.Threshold
	maxLogFileSizeMB       int
)

func init() {
	// Initialize all startup flags
	loggerCmd.Flags().IntVarP((*int)(&logLevel), "logLevel", "l", 2,
		"Sets the log level output when outputting to the Javascript console. "+
			"0 = TRACE, 1 = DEBUG, 2 = INFO, 3 = WARN, 4 = ERROR, "+
			"5 = CRITICAL, 6 = FATAL, -1 = disabled.")
	loggerCmd.Flags().IntVarP((*int)(&fileLogLevel), "fileLogLevel", "m", -1,
		"The log level when outputting to the file buffer. "+
			"0 = TRACE, 1 = DEBUG, 2 = INFO, 3 = WARN, 4 = ERROR, "+
			"5 = CRITICAL, 6 = FATAL, -1 = disabled.")
	loggerCmd.Flags().IntVarP(&maxLogFileSizeMB, "maxLogFileSize", "s", 5,
		"Max file size, in MB, for the file buffer before it rolls over "+
			"over and starts overwriting the oldest entries.")
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

	// Callback for Logging.MaxSize
	wlf.wtm.RegisterCallback(logging.MaxSizeTag, func([]byte) ([]byte, error) {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(wlf.b.Size()))
		return b, nil
	})

	// Callback for Logging.Size
	wlf.wtm.RegisterCallback(logging.SizeTag, func([]byte) ([]byte, error) {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(wlf.b.TotalWritten()))
		return b, nil
	})

	// Callback to let another web worker know they are connected.
	wlf.wtm.RegisterCallback(logging.WorkerReady, func([]byte) ([]byte, error) {
		return []byte{}, nil
	})
}
