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

// SEMVER is the current semantic version of the xxDK DM web worker.
const SEMVER = "0.1.0"

func main() {
	// Set to os.Args because the default is os.Args[1:] and in WASM, args start
	// at 0, not 1.
	dmCmd.SetArgs(os.Args)

	err := dmCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var dmCmd = &cobra.Command{
	Use:     "dmIndexedDbWorker",
	Short:   "IndexedDb database for DMs.",
	Example: "const go = new Go();\ngo.argv = [\"--logLevel=1\"]",
	Run: func(cmd *cobra.Command, args []string) {
		// Start logger first to capture all logging events
		err := logging.EnableLogging(logLevel, -1, 0, "", "")
		if err != nil {
			fmt.Printf(
				"Failed to intialize logging in DM indexedDb worker: %+v", err)
			os.Exit(1)
		}

		jww.INFO.Printf("xxDK DM web worker version: v%s", SEMVER)

		jww.INFO.Print("[WW] Starting xxDK WebAssembly DM Database Worker.")
		m := &manager{
			wtm: worker.NewThreadManager("DmIndexedDbWorker", true),
		}
		m.registerCallbacks()
		m.wtm.SignalReady()

		// Indicate to the Javascript caller that the WASM is ready by resolving
		// a promise created by the caller.
		js.Global().Get("onWasmInitialized").Invoke()

		<-make(chan bool)
		fmt.Println("[WW] Closing xxDK WebAssembly Channels Database Worker.")
		os.Exit(0)
	},
}

var (
	logLevel jww.Threshold
)

func init() {
	// Initialize all startup flags
	dmCmd.Flags().IntVarP((*int)(&logLevel), "logLevel", "l", 2,
		"Sets the log level output when outputting to the Javascript console. "+
			"0 = TRACE, 1 = DEBUG, 2 = INFO, 3 = WARN, 4 = ERROR, "+
			"5 = CRITICAL, 6 = FATAL, -1 = disabled.")
}
