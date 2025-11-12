////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package state

import (
	"fmt"
	"os"
	"syscall/js"

	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"

	"gitlab.com/elixxir/xxdk-wasm/logging"
	"gitlab.com/elixxir/xxdk-wasm/worker"
)

// SEMVER is the current semantic version of the xxDK web worker.
const SEMVER = "0.1.0"

// RunStateWorker returns the state worker command for main.go to execute
func RunStateWorker() *cobra.Command {
	return stateCmd
}

var stateCmd = &cobra.Command{
	Use:     "stateIndexedDbWorker",
	Short:   "IndexedDb database for state.",
	Example: "const go = new Go();\ngo.argv = [\"--logLevel=1\"]",
	Run: func(cmd *cobra.Command, args []string) {
		// Start logger first to capture all logging events
		err := logging.EnableLogging(logLevel, -1, 0, "", "")
		if err != nil {
			fmt.Printf("Failed to intialize logging: %+v", err)
			os.Exit(1)
		}

		jww.INFO.Printf("xxDK state web worker version: v%s", SEMVER)

		jww.INFO.Print("[WW] Starting xxDK WebAssembly State Database Worker.")
		tm, err := worker.NewThreadManager("DmIndexedDbWorker", true)
		if err != nil {
			jww.FATAL.Panicf("Failed to create thread manager: %+v", err)
		}
		m := &manager{wtm: tm}
		m.registerCallbacks()

		m.wtm.RegisterMessageChannelCallback(worker.LoggerTag,
			func(port js.Value, channelName string) {
				p := worker.DefaultParams()
				p.MessageLogging = false
				err = logging.EnableThreadLogging(
					logLevel, threadLogLevel, 0, channelName, port)
				if err != nil {
					fmt.Printf("Failed to intialize logging: %+v", err)
					os.Exit(1)
				}

				jww.INFO.Print("TEST channel")
			})

		m.wtm.SignalReady()

		// Indicate to the Javascript caller that the WASM is ready by resolving
		// a promise created by the caller.
		js.Global().Get("onWasmInitialized").Invoke()

		<-make(chan bool)
		fmt.Println("[WW] Closing xxDK WebAssembly State Database Worker.")
		os.Exit(0)
	},
}

var (
	logLevel       jww.Threshold
	threadLogLevel jww.Threshold
)

func init() {
	// Initialize all startup flags
	stateCmd.Flags().IntVarP((*int)(&logLevel),
		"logLevel", "l", int(jww.LevelDebug),
		"Sets the log level output when outputting to the Javascript console. "+
			"0 = TRACE, 1 = DEBUG, 2 = INFO, 3 = WARN, 4 = ERROR, "+
			"5 = CRITICAL, 6 = FATAL, -1 = disabled.")
	stateCmd.Flags().IntVarP((*int)(&threadLogLevel),
		"threadLogLevel", "m", int(jww.LevelDebug),
		"The log level when outputting to the worker file buffer. "+
			"0 = TRACE, 1 = DEBUG, 2 = INFO, 3 = WARN, 4 = ERROR, "+
			"5 = CRITICAL, 6 = FATAL, -1 = disabled.")
}
