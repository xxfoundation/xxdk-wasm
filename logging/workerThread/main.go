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
	"github.com/hack-pad/safejs"
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
	tm *worker.ThreadManager
	b  *circbuf.Buffer
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
			fmt.Printf(
				"Failed to intialize logging in logging worker: %+v", err)
			os.Exit(1)
		}

		jww.INFO.Printf("xxDK Logger web worker version: v%s", SEMVER)

		jww.INFO.Print("[LOG] Starting xxDK WebAssembly Logger Worker.")

		tm, err := worker.NewThreadManager("Logger", true)
		if err != nil {
			jww.FATAL.Panicf("Failed to get new thread manager: %+v", err)
		}
		wlf := workerLogFile{tm: tm}

		wlf.registerCallbacks()

		wlf.tm.SignalReady()

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
			"0 = TRACE, 1 = DEBUG, 2 = INFO, 3 = WARN, 4 = ERROR, "+
			"5 = CRITICAL, 6 = FATAL, -1 = disabled.")
}

// registerCallbacks registers all the necessary callbacks for the main thread
// to get the file and file metadata.
func (wlf *workerLogFile) registerCallbacks() {
	// Callback for logging.LogToFileWorker
	wlf.tm.RegisterCallback(logging.NewLogFileTag,
		func(message []byte, reply func([]byte)) {

			var maxLogFileSize int64
			err := json.Unmarshal(message, &maxLogFileSize)
			if err != nil {
				reply([]byte(err.Error()))
				return
			}

			wlf.b, err = circbuf.NewBuffer(maxLogFileSize)
			if err != nil {
				reply([]byte(err.Error()))
				return
			}

			jww.DEBUG.Printf("[LOG] Created new worker log file of size %d",
				maxLogFileSize)

			reply(nil)
		})

	// Callback for Logging.GetFile
	wlf.tm.RegisterCallback(logging.WriteLogTag,
		func(message []byte, _ func([]byte)) {
			n, err := wlf.b.Write(message)
			if err != nil {
				jww.ERROR.Printf("[LOG] Failed to write to log: %+v", err)
			} else if n != len(message) {
				jww.ERROR.Printf("[LOG] Failed to write to log: wrote %d "+
					"bytes; expected %d bytes", n, len(message))
			}
		},
	)

	// Callback for Logging.GetFile
	wlf.tm.RegisterCallback(logging.GetFileTag,
		func(_ []byte, reply func([]byte)) { reply(wlf.b.Bytes()) })

	// Callback for Logging.GetFile
	wlf.tm.RegisterCallback(logging.GetFileExtTag,
		func(_ []byte, reply func([]byte)) { reply(wlf.b.Bytes()) })

	// Callback for Logging.Size
	wlf.tm.RegisterCallback(logging.SizeTag, func(_ []byte, reply func([]byte)) {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(wlf.b.TotalWritten()))
		reply(b)
	})

	wlf.tm.RegisterMessageChannelCallback(worker.LoggerTag, wlf.registerLogWorker)
}

func (wlf *workerLogFile) registerLogWorker(port js.Value, channelName string) {
	p := worker.DefaultParams()
	p.MessageLogging = false
	mm, err := worker.NewMessageManager(
		safejs.Safe(port), channelName+"-logger", p)
	if err != nil {
		jww.FATAL.Panic(err)
	}

	mm.RegisterCallback(logging.WriteLogTag,
		func(message []byte, _ func([]byte)) {
			n, err := wlf.b.Write(message)
			if err != nil {
				jww.ERROR.Printf("[LOG] Failed to write to log: %+v", err)
			} else if n != len(message) {
				jww.ERROR.Printf("[LOG] Failed to write to log: wrote %d "+
					"bytes; expected %d bytes", n, len(message))
			}
		},
	)

	// Callback for Logging.GetFile
	mm.RegisterCallback(logging.GetFileTag, func(_ []byte, reply func([]byte)) {
		reply(wlf.b.Bytes())
	})

	// Callback for Logging.MaxSize
	mm.RegisterCallback(logging.GetFileTag, func(_ []byte, reply func([]byte)) {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(wlf.b.Size()))
		reply(b)
	})

	// Callback for Logging.Size
	mm.RegisterCallback(logging.SizeTag, func(_ []byte, reply func([]byte)) {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(wlf.b.TotalWritten()))
		reply(b)
	})
}
