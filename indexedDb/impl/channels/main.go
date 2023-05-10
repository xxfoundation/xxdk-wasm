////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package main

import (
	"fmt"
	"os"
	"syscall/js"

	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"

	"gitlab.com/elixxir/xxdk-wasm/logging"
	"gitlab.com/elixxir/xxdk-wasm/worker"
)

// SEMVER is the current semantic version of the xxDK channels web worker.
const SEMVER = "0.1.0"

func main() {
	// Set to os.Args because the default is os.Args[1:] and in WASM, args start
	// at 0, not 1.
	channelsCmd.SetArgs(os.Args)

	err := channelsCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var channelsCmd = &cobra.Command{
	Use:     "channelsIndexedDbWorker",
	Short:   "IndexedDb database for channels.",
	Example: "const go = new Go();\ngo.argv = [\"--logLevel=1\"]",
	Run: func(cmd *cobra.Command, args []string) {
		// Set up basic logging while the worker thread manager is initialised
		ll := logging.NewJsConsoleLogListener(jww.LevelInfo)
		logging.AddLogListener(ll.Listen)
		jww.SetStdoutThreshold(jww.LevelFatal + 1)

		jww.INFO.Printf("xxDK channels web worker version: v%s", SEMVER)

		jww.INFO.Print("[WW] Starting xxDK WebAssembly Channels Database Worker.")
		m := &manager{
			wtm: worker.NewThreadManager("ChannelsIndexedDbWorker", true),
		}

		m.registerCallbacks()

		m.wtm.SignalReady()

		// Start logger first to capture all logging events
		var wtm *worker.ThreadManager
		if workerLogging {
			wtm = m.wtm
		}
		err := logging.EnableWorkerLogging(logLevel, fileLogLevel,
			maxLogFileSizeMB, worker.ChannelsIndexedDbLogging, wtm)
		if err != nil {
			fmt.Printf("Failed to intialize logging in channels indexedDb "+
				"worker: %+v", err)
			os.Exit(1)
		}

		// Indicate to the Javascript caller that the WASM is ready by resolving
		// a promise created by the caller.
		js.Global().Get("onWasmInitialized").Invoke()

		<-make(chan bool)
		fmt.Println("[WW] Closing xxDK WebAssembly Channels Database Worker.")
		os.Exit(0)
	},
}

var (
	logLevel, fileLogLevel jww.Threshold
	maxLogFileSizeMB       int
	workerLogging          bool
)

func init() {
	// Initialize all startup flags
	channelsCmd.Flags().IntVarP((*int)(&logLevel), "logLevel", "l", 2,
		"Sets the log level output when outputting to the Javascript console. "+
			"0 = TRACE, 1 = DEBUG, 2 = INFO, 3 = WARN, 4 = ERROR, "+
			"5 = CRITICAL, 6 = FATAL, -1 = disabled.")
	channelsCmd.Flags().IntVarP((*int)(&fileLogLevel), "fileLogLevel", "m", -1,
		"The log level when outputting to the file buffer. "+
			"0 = TRACE, 1 = DEBUG, 2 = INFO, 3 = WARN, 4 = ERROR, "+
			"5 = CRITICAL, 6 = FATAL, -1 = disabled.")
	channelsCmd.Flags().IntVarP(&maxLogFileSizeMB, "maxLogFileSize", "s", 5,
		"Max file size, in MB, for the file buffer before it rolls over "+
			"over and starts overwriting the oldest entries.")
	channelsCmd.Flags().BoolVarP(&workerLogging, "workerLogging", "w", false,
		"If set, logging is sent to the logging worker instead of a local "+
			"buffer.")
}
